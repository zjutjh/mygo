package swagger

import (
	"fmt"
	"mime/multipart"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

type schemaRegistry struct {
	schemas    map[string]Property
	typeToName map[reflect.Type]string
	nameToType map[string]reflect.Type
}

func newSchemaRegistry() *schemaRegistry {
	return &schemaRegistry{
		schemas:    make(map[string]Property),
		typeToName: make(map[reflect.Type]string),
		nameToType: make(map[string]reflect.Type),
	}
}

func (r *schemaRegistry) buildSchemaForType(t reflect.Type, desc string, nameKey string, requiredFunc func(reflect.StructField) bool) *Property {
	t = derefSchemaType(t)
	switch t.Kind() {
	case reflect.Bool:
		return Boolean(t, desc)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Integer(t, desc)
	case reflect.Float32, reflect.Float64:
		return Number(t, desc)
	case reflect.String:
		return String(t, desc)
	case reflect.Array, reflect.Slice:
		return r.buildArraySchema(t, desc, nameKey, requiredFunc)
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return Time(t, desc)
		}
		if t == reflect.TypeOf(multipart.FileHeader{}) {
			Output("不支持的响应类型[multipart.FileHeader]\n")
			return nil
		}
		if t.Name() != "" {
			return r.ensureStructSchema(t, desc, nameKey, requiredFunc)
		}
		return r.buildStructSchema(t, desc, nameKey, requiredFunc)
	case reflect.Interface, reflect.Map:
		return StructEmpty(t, desc, nameKey)
	case reflect.Uintptr, reflect.UnsafePointer, reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
		Output("不支持的响应类型[%s]\n", t.Kind())
		return nil
	default:
		Output("不支持的响应类型[%s]\n", t.Kind())
		return nil
	}
}

func (r *schemaRegistry) buildSchemaForField(sf reflect.StructField, nameKey string, requiredFunc func(reflect.StructField) bool) *Property {
	desc := Desc(sf)
	t := derefSchemaType(sf.Type)
	sfCopy := sf
	sfCopy.Type = t

	switch t.Kind() {
	case reflect.Bool:
		return Boolean(t, desc)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return IntegerTag(sfCopy, Integer(t, desc))
	case reflect.Float32, reflect.Float64:
		return NumberTag(sfCopy, Number(t, desc))
	case reflect.String:
		return StringTag(sfCopy, String(t, desc))
	case reflect.Array, reflect.Slice:
		property := r.buildArraySchema(t, desc, nameKey, requiredFunc)
		if property == nil {
			return nil
		}
		return ArrayTag(sfCopy, property)
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return Time(t, desc)
		}
		if t == reflect.TypeOf(multipart.FileHeader{}) {
			Output("不支持的响应类型[multipart.FileHeader]\n")
			return nil
		}
		if t.Name() != "" && !sf.Anonymous {
			return r.ensureStructSchema(t, desc, nameKey, requiredFunc)
		}
		return r.buildStructSchema(t, desc, nameKey, requiredFunc)
	case reflect.Interface, reflect.Map:
		return StructEmpty(t, desc, nameKey)
	case reflect.Uintptr, reflect.UnsafePointer, reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
		Output("不支持的响应类型[%s]\n", t.Kind())
		return nil
	default:
		Output("不支持的响应类型[%s]\n", t.Kind())
		return nil
	}
}

func (r *schemaRegistry) buildStructSchema(t reflect.Type, desc string, nameKey string, requiredFunc func(reflect.StructField) bool) *Property {
	property := &Property{
		Properties:  make(map[string]*Property),
		Description: desc,
		Type:        "object",
		Required:    []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		tx := t.Field(i)
		txt := derefSchemaType(tx.Type)

		if tx.Anonymous && txt.Kind() == reflect.Struct && txt != reflect.TypeOf(time.Time{}) && txt != reflect.TypeOf(multipart.FileHeader{}) {
			op := r.buildStructSchema(txt, Desc(tx), nameKey, requiredFunc)
			if op != nil {
				for k, v := range op.Properties {
					property.Properties[k] = v
				}
				for _, field := range op.Required {
					property.Required = appendUniqueRequired(property.Required, field)
				}
			}
			continue
		}

		fieldName := Name(tx, nameKey)
		if fieldName == "" || fieldName == "-" {
			continue
		}

		schema := r.buildSchemaForField(tx, nameKey, requiredFunc)
		if schema == nil {
			continue
		}
		property.Properties[fieldName] = schema
		if requiredFunc(tx) {
			property.Required = appendUniqueRequired(property.Required, fieldName)
		}
	}

	return property
}

func (r *schemaRegistry) buildArraySchema(t reflect.Type, desc string, nameKey string, requiredFunc func(reflect.StructField) bool) *Property {
	property := &Property{
		Items:       nil,
		Description: desc,
		Type:        "array",
		Enum:        nil,
		MaxItems:    0,
		MinItems:    0,
		UniqueItems: false,
		MaxContains: 0,
		MinContains: 0,
	}

	itemType := derefSchemaType(t.Elem())
	itemDesc := "[" + itemType.Kind().String() + "]子项信息"
	property.Items = r.buildSchemaForType(itemType, itemDesc, nameKey, requiredFunc)
	return property
}

func (r *schemaRegistry) ensureStructSchema(t reflect.Type, desc string, nameKey string, requiredFunc func(reflect.StructField) bool) *Property {
	t = derefSchemaType(t)
	if schemaName, ok := r.typeToName[t]; ok {
		return schemaRefProperty(schemaName)
	}

	schemaName := r.uniqueSchemaName(defaultSchemaName(t), t)
	r.typeToName[t] = schemaName
	r.nameToType[schemaName] = t
	r.schemas[schemaName] = Property{
		Type:       "object",
		Properties: map[string]*Property{},
		Required:   []string{},
	}

	schema := r.buildStructSchema(t, desc, nameKey, requiredFunc)
	if schema == nil {
		delete(r.schemas, schemaName)
		delete(r.typeToName, t)
		delete(r.nameToType, schemaName)
		return nil
	}
	r.schemas[schemaName] = *schema
	return schemaRefProperty(schemaName)
}

func appendUniqueRequired(required []string, fieldName string) []string {
	if fieldName == "" || fieldName == "-" || slices.Contains(required, fieldName) {
		return required
	}
	return append(required, fieldName)
}

func derefSchemaType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func schemaRefProperty(schemaName string) *Property {
	return &Property{
		Ref: fmt.Sprintf("#/components/schemas/%s", escapeJSONPointerToken(schemaName)),
	}
}

func defaultSchemaName(t reflect.Type) string {
	pkg := path.Base(t.PkgPath())
	if pkg == "" {
		return t.Name()
	}
	return pkg + "." + t.Name()
}

func (r *schemaRegistry) uniqueSchemaName(base string, t reflect.Type) string {
	base = sanitizeSchemaName(base)
	if base == "" {
		base = "Schema"
	}

	if current, ok := r.nameToType[base]; !ok || current == t {
		return base
	}

	for i := 2; ; i++ {
		candidate := base + "_" + strconv.Itoa(i)
		if current, ok := r.nameToType[candidate]; !ok || current == t {
			return candidate
		}
	}
}

func sanitizeSchemaName(name string) string {
	var sb strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '.' || ch == '_' || ch == '-' {
			sb.WriteRune(ch)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

func escapeJSONPointerToken(v string) string {
	v = strings.ReplaceAll(v, "~", "~0")
	v = strings.ReplaceAll(v, "/", "~1")
	return v
}
