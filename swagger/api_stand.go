package swagger

import (
	"fmt"
	"mime/multipart"
	"reflect"
	"slices"
	"time"

	"github.com/zjutjh/mygo/kit"
)

func ParseApiStandRequestParameters(t reflect.Type, requestPos, validatePos, in string) []Parameter {
	var parameters []Parameter
	t1, exist := t.FieldByName("Request")
	if !exist {
		return parameters
	}
	t2, exist := t1.Type.FieldByName(requestPos)
	if !exist {
		return parameters
	}
	parameters = append(parameters, getRequestFields(t2.Type, validatePos, in)...)
	return parameters
}

func getRequestFields(t reflect.Type, validatePos, in string) []Parameter {
	var parameters []Parameter
	for i := 0; i < t.NumField(); i++ {
		tx := t.Field(i)
		parameter := Parameter{
			Name:        tx.Tag.Get(validatePos),
			In:          in,
			Description: Desc(tx),
			Required:    RequestRequired(tx),
		}
		var property *Property
		txt := tx.Type
	fieldstart:
		switch txt.Kind() {
		case reflect.Bool:
			property = Boolean(txt, Desc(tx))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			property = IntegerTag(tx, Integer(txt, Desc(tx)))
		case reflect.Float32, reflect.Float64:
			property = NumberTag(tx, Number(txt, Desc(tx)))
		case reflect.String:
			property = StringTag(tx, String(txt, Desc(tx)))
		case reflect.Array, reflect.Slice:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		case reflect.Struct:
			if txt == reflect.TypeOf(time.Time{}) {
				property = Time(txt, Desc(tx))
			} else if txt == reflect.TypeOf(multipart.FileHeader{}) {
				Output("不支持的响应类型[multipart.FileHeader]\n")
			} else if tx.Anonymous {
				// 如果是匿名的结构体字段,则采用平铺的方式
				parameters = append(parameters, getRequestFields(txt, validatePos, in)...)
			} else {
				Output("不支持的响应类型[%s]\n", txt.Kind())
			}
		case reflect.Uintptr:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		case reflect.Pointer:
			vx := reflect.New(txt)
			txt = vx.Type().Elem().Elem()
			goto fieldstart
		case reflect.UnsafePointer:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		case reflect.Interface:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		case reflect.Map:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		default:
			Output("不支持的响应类型[%s]\n", txt.Kind())
		}
		if nil != property {
			parameter.Schema = *property
			parameters = append(parameters, parameter)
		}
	}
	return parameters
}

func ParseApiStandRequestBody(t reflect.Type) *RequestBody {
	requestBody := &RequestBody{}
	t1, exist := t.FieldByName("Request")
	if !exist {
		return nil
	}
	t2, exist := t1.Type.FieldByName("Body")
	if !exist {
		return nil
	}
	var property Property
	switch t2.Type.Kind() {
	case reflect.Struct:
		property = *Struct(t2.Type, "业务请求", "json", RequestRequired)
	case reflect.Slice, reflect.Array:
		property = *Array(t2.Type, "业务请求", "json", RequestRequired)
	case reflect.Interface:
		property = *StructEmpty(t2.Type, "业务请求", "json")
	case reflect.Map:
		property = *StructEmpty(t2.Type, "业务请求", "json")
	default:
		Output("不支持的Request Body类型[%s]\n", t2.Type.Kind())
		return nil
	}

	requestBody.Description = "请求body"
	requestBody.Required = true
	requestBody.Content = map[string]MediaType{
		"application/json": {Schema: property},
	}
	return requestBody
}

func getComponentExampleRef(code kit.Code) string {
	return fmt.Sprintf("#/components/examples/response_business_status_code_%d", code.Code)
}

type commResponse struct {
	Code    int64  `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

func ParseApiStandResponse(t reflect.Type, businessStatusCodes []kit.Code) Response {
	standResponse := Response{}
	t1, exist := t.FieldByName("Response")
	if !exist {
		return standResponse
	}
	codeProperty := &Property{
		Description: "业务响应Code",
		Type:        "integer",
	}
	messageProperty := &Property{
		Description: "业务响应Message",
		Type:        "string",
	}
	var dataProperty *Property
	switch t1.Type.Kind() {
	case reflect.Struct:
		dataProperty = Struct(t1.Type, "业务响应Data", "json", ResponseRequired)
	case reflect.Slice, reflect.Array:
		dataProperty = Array(t1.Type, "业务响应Data", "json", ResponseRequired)
	case reflect.Interface:
		dataProperty = StructEmpty(t1.Type, "业务响应Data", "json")
	case reflect.Map:
		dataProperty = StructEmpty(t1.Type, "业务响应Data", "json")
	default:
		Output("不支持的Response类型[%s]\n", t1.Type.Kind())
		return standResponse
	}
	standProperty := Property{
		Properties: map[string]*Property{
			"code":    codeProperty,
			"message": messageProperty,
			"data":    dataProperty,
		},
		Description: "业务响应",
		Type:        "object",
		Required:    []string{"code", "message", "data"},
	}
	// 嵌入业务状态码信息
	if slices.Contains(businessStatusCodes, kit.CodeOK) {
		messageProperty.Default = kit.CodeOK.Message
	}
	standResponse.Description = "HTTP 状态码为 200 时的业务响应"
	standResponse.Content = map[string]MediaType{"application/json": {Schema: standProperty}}
	return standResponse
}

func GenerateApiFailureResponse(businessStatusCodes []kit.Code) (Response, bool) {
	failureResponse := Response{}
	codeProperty := &Property{
		Description: "业务响应Code",
		Type:        "integer",
	}
	messageProperty := &Property{
		Description: "业务响应Message",
		Type:        "string",
	}
	dataProperty := &Property{
		Description: "业务响应Data",
		Type:        "null",
	}
	failureProperty := Property{
		Properties: map[string]*Property{
			"code":    codeProperty,
			"message": messageProperty,
			"data":    dataProperty,
		},
		Description: "业务响应",
		Type:        "object",
		Required:    []string{"code", "message", "data"},
	}
	// 嵌入业务状态码清单
	examples := map[string]ExampleObject{}
	for _, code := range businessStatusCodes {
		name := fmt.Sprintf("code-%d", code.Code)
		if code == kit.CodeOK {
			continue
		}
		examples[name] = ExampleObject{
			Ref: getComponentExampleRef(code),
		}
	}
	if len(examples) == 0 {
		return failureResponse, false
	}
	failureResponse.Description = "处理失败时的业务响应 (HTTP 状态码仍为 200)"
	failureResponse.Content = map[string]MediaType{"application/json": {Schema: failureProperty, Examples: examples}}
	return failureResponse, true
}
