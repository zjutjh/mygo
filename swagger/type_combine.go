package swagger

import (
	"mime/multipart"
	"reflect"
	"time"
)

func StructEmpty(t reflect.Type, desc string, nameKey string) *Property {
	property := &Property{
		Properties:  make(map[string]*Property),
		Description: desc,
		Type:        "object",
		Required:    []string{},
	}
	return property
}

func Struct(t reflect.Type, desc string, nameKey string, requiredFunc func(reflect.StructField) bool) *Property {
	property := &Property{
		Properties:  make(map[string]*Property),
		Description: desc,
		Type:        "object",
		Required:    []string{},
	}
	for i := 0; i < t.NumField(); i++ {
		required := false
		tx := t.Field(i)
		txt := tx.Type
	structstart:
		switch txt.Kind() {
		case reflect.Bool:
			property.Properties[Name(tx, nameKey)], required = Boolean(txt, Desc(tx)), requiredFunc(tx)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			property.Properties[Name(tx, nameKey)], required = IntegerTag(tx, Integer(txt, Desc(tx))), requiredFunc(tx)
		case reflect.Float32, reflect.Float64:
			property.Properties[Name(tx, nameKey)], required = NumberTag(tx, Number(txt, Desc(tx))), requiredFunc(tx)
		case reflect.String:
			property.Properties[Name(tx, nameKey)], required = StringTag(tx, String(txt, Desc(tx))), requiredFunc(tx)
		case reflect.Array, reflect.Slice:
			property.Properties[Name(tx, nameKey)], required = ArrayTag(tx, Array(txt, Desc(tx), nameKey, requiredFunc)), requiredFunc(tx)
		case reflect.Struct:
			if tx.Type == reflect.TypeOf(time.Time{}) {
				property.Properties[Name(tx, nameKey)], required = Time(txt, Desc(tx)), requiredFunc(tx)
			} else if tx.Type == reflect.TypeOf(multipart.FileHeader{}) {
				Output("不支持的响应类型[multipart.FileHeader]\n")
			} else if tx.Anonymous {
				op := Struct(txt, Desc(tx), nameKey, requiredFunc)
				for k, v := range op.Properties {
					property.Properties[k] = v
				}
				property.Required = append(property.Required, property.Required...)
			} else {
				property.Properties[Name(tx, nameKey)], required = Struct(txt, Desc(tx), nameKey, requiredFunc), requiredFunc(tx)
			}
		case reflect.Interface:
			property.Properties[Name(tx, nameKey)], required = StructEmpty(txt, Desc(tx), nameKey), requiredFunc(tx)
		case reflect.Map:
			property.Properties[Name(tx, nameKey)], required = StructEmpty(txt, Desc(tx), nameKey), requiredFunc(tx)
		case reflect.Uintptr:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		case reflect.Pointer:
			vx := reflect.New(txt)
			txt = vx.Type().Elem().Elem()
			goto structstart
		case reflect.UnsafePointer:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		default:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		}
		if required {
			property.Required = append(property.Required, Name(tx, nameKey))
		}
	}
	return property
}

func Array(t reflect.Type, desc string, nameKey string, requiredFunc func(reflect.StructField) bool) *Property {
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
	itemKind := t.Elem().Kind()
arraystart:
	switch itemKind {
	case reflect.Bool:
		property.Items = Boolean(t.Elem(), "["+itemKind.String()+"]子项信息")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		property.Items = Integer(t.Elem(), "["+itemKind.String()+"]子项信息")
	case reflect.Float32, reflect.Float64:
		property.Items = Number(t.Elem(), "["+itemKind.String()+"]子项信息")
	case reflect.String:
		property.Items = String(t.Elem(), "["+itemKind.String()+"]子项信息")
	case reflect.Array, reflect.Slice:
		property.Items = Array(t.Elem(), "["+itemKind.String()+"]子项信息", nameKey, requiredFunc)
	case reflect.Struct:
		property.Items = Struct(t.Elem(), "["+itemKind.String()+"]子项信息", nameKey, requiredFunc)
	case reflect.Interface:
		property.Items = StructEmpty(t.Elem(), "["+itemKind.String()+"]子项信息", nameKey)
	case reflect.Map:
		property.Items = StructEmpty(t.Elem(), "["+itemKind.String()+"]子项信息", nameKey)
	case reflect.Uintptr:
		Output("不支持的响应类型[%s]\n", itemKind)
	case reflect.Pointer:
		vx := reflect.New(t.Elem())
		t = vx.Elem().Type()
		itemKind = vx.Type().Elem().Elem().Kind()
		goto arraystart
	case reflect.UnsafePointer:
		Output("不支持的响应类型[%s]\n", itemKind)
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
		Output("不支持的响应类型[%s]\n", itemKind)
	default:
		Output("不支持的响应类型[%s]\n", itemKind)
	}
	return property
}

func ArrayTag(sf reflect.StructField, property *Property) *Property {
	bindingList := Binding(sf)
	for _, binding := range bindingList {
		if binding == "required" {
			continue
		}
		if !(LEN(binding, property, sf.Type.Kind()) ||
			MAX(binding, property, sf.Type.Kind()) ||
			MIN(binding, property, sf.Type.Kind()) ||
			EQ(binding, property, sf.Type.Kind()) ||
			GT(binding, property, sf.Type.Kind()) ||
			GTE(binding, property, sf.Type.Kind()) ||
			LT(binding, property, sf.Type.Kind()) ||
			LTE(binding, property, sf.Type.Kind())) {
			Output("无需处理的tag[%s]\n", binding)
		}
	}
	return property
}

func String(t reflect.Type, desc string) *Property {
	property := &Property{
		Description: desc,
		Type:        "string",
		Enum:        nil,
		Format:      "",
		MaxLength:   0,
		MinLength:   0,
		Pattern:     "",
	}
	return property
}

func StringTag(sf reflect.StructField, property *Property) *Property {
	bindingList := Binding(sf)
	for _, binding := range bindingList {
		if binding == "required" {
			continue
		}
		if !(LEN(binding, property, sf.Type.Kind()) ||
			MAX(binding, property, sf.Type.Kind()) ||
			MIN(binding, property, sf.Type.Kind()) ||
			EQ(binding, property, sf.Type.Kind()) ||
			ONEOF(binding, property, sf.Type.Kind()) ||
			GT(binding, property, sf.Type.Kind()) ||
			GTE(binding, property, sf.Type.Kind()) ||
			LT(binding, property, sf.Type.Kind()) ||
			LTE(binding, property, sf.Type.Kind()) ||
			DATETIME(binding, property) ||
			EMAIL(binding, property) ||
			IP(binding, property) ||
			HOSTNAME(binding, property)) {
			Output("无需处理的tag[%s]\n", binding)
		}
	}
	return property
}

func Number(t reflect.Type, desc string) *Property {
	property := &Property{
		Description:      desc,
		Type:             "number",
		Enum:             nil,
		MultipleOf:       0,
		Maximum:          0,
		Minimum:          0,
		ExclusiveMaximum: 0,
		ExclusiveMinimum: 0,
	}
	if t.Kind() == reflect.Float32 {
		property.Format = "float"
	} else if t.Kind() == reflect.Float64 {
		property.Format = "double"
	}
	return property
}

func NumberTag(sf reflect.StructField, property *Property) *Property {
	bindingList := Binding(sf)
	for _, binding := range bindingList {
		if binding == "required" {
			continue
		}
		if !(MAX(binding, property, sf.Type.Kind()) ||
			MIN(binding, property, sf.Type.Kind()) ||
			EQ(binding, property, sf.Type.Kind()) ||
			ONEOF(binding, property, sf.Type.Kind()) ||
			GT(binding, property, sf.Type.Kind()) ||
			GTE(binding, property, sf.Type.Kind()) ||
			LT(binding, property, sf.Type.Kind()) ||
			LTE(binding, property, sf.Type.Kind())) {
			Output("无需处理的tag[%s]\n", binding)
		}
	}
	return property
}

func Integer(t reflect.Type, desc string) *Property {
	property := &Property{
		Description:      desc,
		Type:             "integer",
		Enum:             []any{},
		MultipleOf:       0,
		Maximum:          0,
		Minimum:          0,
		ExclusiveMaximum: 0,
		ExclusiveMinimum: 0,
	}
	if t.Kind() == reflect.Int32 {
		property.Format = "int32"
	} else if t.Kind() == reflect.Int64 {
		property.Format = "int64"
	}
	return property
}

func IntegerTag(sf reflect.StructField, property *Property) *Property {
	bindingList := Binding(sf)
	for _, binding := range bindingList {
		if binding == "required" {
			continue
		}
		if !(MAX(binding, property, sf.Type.Kind()) ||
			MIN(binding, property, sf.Type.Kind()) ||
			EQ(binding, property, sf.Type.Kind()) ||
			ONEOF(binding, property, sf.Type.Kind()) ||
			GT(binding, property, sf.Type.Kind()) ||
			GTE(binding, property, sf.Type.Kind()) ||
			LT(binding, property, sf.Type.Kind()) ||
			LTE(binding, property, sf.Type.Kind())) {
			Output("无需处理的tag[%s]\n", binding)
		}
	}
	return property
}

func Boolean(t reflect.Type, desc string) *Property {
	return &Property{
		Description: desc,
		Type:        "boolean",
	}
}

func Time(t reflect.Type, desc string) *Property {
	return &Property{
		Description: desc,
		Type:        "string",
		Format:      "date-time",
	}
}
