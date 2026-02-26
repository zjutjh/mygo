package swagger

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/kit"
)

type middlewareNames []string

// 获取中间件映射
type middlewareMap struct {
	mapping map[string]map[string]middlewareNames
}

func (m *middlewareMap) add(method, fullpath string, names middlewareNames) {
	if method == "" || fullpath == "" || len(names) == 0 {
		return
	}
	if _, ok := m.mapping[method]; !ok {
		m.mapping[method] = make(map[string]middlewareNames)
	}
	m.mapping[method][fullpath] = names
}

func (m *middlewareMap) get(method, fullpath string) middlewareNames {
	if m == nil {
		return nil
	}
	v := m.mapping[method]
	if v == nil {
		return nil
	}
	return v[fullpath]
}

func dereference(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func gatherMiddlewares(engine *gin.Engine) (*middlewareMap, error) {
	middlewareMap := &middlewareMap{
		mapping: make(map[string]map[string]middlewareNames),
	}

	// 反射获取 Engine 的 trees 字段
	engineValue := reflect.ValueOf(engine).Elem()
	treesField := engineValue.FieldByName("trees")
	if !treesField.IsValid() || treesField.Kind() != reflect.Slice {
		return nil, errors.New("无法获取 gin.Engine 中的“trees”字段")
	}

	// 遍历 trees (HTTP 方法 -> 路由树)
	for i := 0; i < treesField.Len(); i++ {
		tree := treesField.Index(i)
		tree = dereference(tree)
		methodField := tree.FieldByName("method")
		if !methodField.IsValid() || methodField.Kind() != reflect.String {
			return nil, errors.New("无法获取 HTTP 方法")
		}
		method := methodField.String()

		// 获取树的根节点
		rootField := tree.FieldByName("root")
		if !rootField.IsValid() {
			return nil, errors.New("无法获取树的根节点")
		}

		// 递归遍历树节点
		if err := parseNode(method, rootField, "", middlewareMap); err != nil {
			return nil, err
		}
	}

	return middlewareMap, nil
}

// 递归解析树节点
func parseNode(method string, node reflect.Value, parentPath string, middlewareMap *middlewareMap) error {
	node = dereference(node)
	// 获取当前节点的路径
	pathField := node.FieldByName("path")
	if !pathField.IsValid() {
		return errors.New("无法获取当前节点的路径")
	}
	path := pathField.String()

	// 拼接完整路径
	fullPath := parentPath + path

	// 获取当前节点的处理函数
	handlersField := node.FieldByName("handlers")
	if !handlersField.IsValid() || handlersField.Kind() != reflect.Slice {
		return errors.New("无法获取当前节点的处理函数")
	}
	handlerNames := make([]string, 0, handlersField.Len())
	for i := 0; i < handlersField.Len(); i++ {
		handler := handlersField.Index(i)
		if handler.Kind() != reflect.Func {
			return errors.New("当前节点的处理函数不是函数")
		}
		h := runtime.FuncForPC(handler.Pointer()).Name()
		handlerNames = append(handlerNames, h)
	}
	// 如果处理函数数量少于1，则没有中间件
	if len(handlerNames) > 1 {
		middlewareMap.add(method, fullPath, handlerNames[:len(handlerNames)-1])
	}

	// 递归处理子节点
	childrenField := node.FieldByName("children")
	if !childrenField.IsValid() {
		return nil
	}
	if childrenField.Kind() != reflect.Slice {
		return errors.New("无法获取子节点")
	}
	for i := 0; i < childrenField.Len(); i++ {
		child := childrenField.Index(i)
		if err := parseNode(method, child, fullPath, middlewareMap); err != nil {
			return err
		}
	}
	return nil
}

func DocumentHandler(engine *gin.Engine) gin.HandlerFunc {
	var selfFuncName string

	handler := func(ctx *gin.Context) {
		// openapi版本
		// 项目基本信息
		// 项目服务实例信息
		// projectKey := "API Document"
		servers := []Server{}
		_ = config.Pick().UnmarshalKey("openapi.servers", &servers)
		openapi := OpenAPI{
			Openapi: "3.1.0",
			Info: Info{
				Title:       fmt.Sprintf("接口[%s]文档", config.AppName()),
				Summary:     fmt.Sprintf("接口[%s]文档", config.AppName()),
				Description: fmt.Sprintf("接口[%s]文档", config.AppName()),
				Version:     "",
			},
			Servers: servers,
			Paths:   map[string]PathItem{},
			Components: Components{
				Examples:        map[string]ExampleObject{},
				SecuritySchemes: map[string]SecurityScheme{},
			},
		}

		// 确定group规则
		groupKeyStart := 2
		groupKeyEnd := 3
		if config.Pick().IsSet("openapi.group_key_start") {
			groupKeyStart = config.Pick().GetInt("openapi.group_key_start")
		}
		if config.Pick().IsSet("openapi.group_key_end") {
			groupKeyEnd = config.Pick().GetInt("openapi.group_key_end")
		}

		middlewareMap, err := gatherMiddlewares(engine)
		if err != nil {
			Output("无法获取路由中间件信息: %s\n", err.Error())
		}

		// 取出所有接口
		routes := engine.Routes()
		for _, route := range routes {
			api, exist := CM[route.Handler]
			if !exist {
				if route.Handler != selfFuncName {
					Output("发现未注册CM的接口[%s]\n", route.Path)
				}
				continue
			}

			// 获取接口group、path、method
			group := strings.Join(strings.Split(route.Path, "/")[groupKeyStart:groupKeyEnd], "/")
			pathSlice := strings.Split(route.Path, "/")
			for i, s := range pathSlice {
				if s != "" && (s[0] == ':' || s[0] == '*') {
					pathSlice[i] = "{" + s[1:] + "}"
				}
			}
			path := strings.Join(pathSlice, "/")
			method := strings.ToUpper(route.Method)

			// 获取api的type、value反射
			t := reflect.TypeOf(api)
			// v := reflect.ValueOf(api)

			// 获取接口name、description、summary
			d, ok := t.FieldByName("Info")
			name := "API名称"
			desc := "API描述"
			if ok {
				name = d.Tag.Get("name")
				desc = d.Tag.Get("desc")
			}

			// 获取接口request query、header、uri、cookie
			parameters := ParseApiStandRequestParameters(t, "Query", "form", "query")
			parameters = append(parameters, ParseApiStandRequestParameters(t, "Header", "header", "header")...)
			parameters = append(parameters, ParseApiStandRequestParameters(t, "Uri", "uri", "path")...)

			// 获取接口request body
			request := ParseApiStandRequestBody(t)

			// 获取所有中间件（不包含末端的处理器）
			middlewares := middlewareMap.get(method, route.Path)
			// 获取所有状态码
			fullChain := append(middlewares, route.Handler)
			businessStatusCodes := getAllBusinessStatusCodes(fullChain...)
			registerCommonResponseExamples(openapi.Components.Examples, businessStatusCodes)

			// 按标准模式获取接口response
			responses := map[string]Response{
				"200": ParseApiStandResponse(t, businessStatusCodes),
			}
			if failureResponse, exist := GenerateApiFailureResponse(businessStatusCodes); exist {
				responses["default"] = failureResponse
			}

			// 组装operation
			operation := &Operation{
				Tags:        []string{group},
				Summary:     name,
				Description: desc,
				Parameters:  parameters,
				RequestBody: request,
				Responses:   responses,
				Servers:     nil,
			}
			// 鉴权中间件解析
			securitySchemes := parseAuthenticationMiddleware(middlewares)
			operation.Security = make([]SecurityItem, 0, len(securitySchemes))
			for _, securityScheme := range securitySchemes {
				if _, ok := openapi.Components.SecuritySchemes[securityScheme.key]; !ok {
					openapi.Components.SecuritySchemes[securityScheme.key] = securityScheme.scheme
				}
				operation.Security = append(operation.Security, SecurityItem{
					securityScheme.key: []string{},
				})
			}

			// 放入pathItem
			pathItem, exist := openapi.Paths[path]
			if !exist {
				pathItem = PathItem{}
			}
			switch method {
			case "GET":
				pathItem.Get = operation
			case "PUT":
				pathItem.Put = operation
			case "POST":
				pathItem.Post = operation
			case "DELETE":
				pathItem.Delete = operation
			case "OPTIONS":
				pathItem.Options = operation
			case "HEAD":
				pathItem.Head = operation
			case "PATCH":
				pathItem.Patch = operation
			case "TRACE":
				pathItem.Trace = operation
			}
			openapi.Paths[path] = pathItem
		}

		format := ctx.Query("format")
		if format == "yaml" {
			ctx.Header("Content-Disposition", "attachment; filename=swagger.yaml")
			ctx.YAML(http.StatusOK, openapi)
		} else {
			ctx.JSON(http.StatusOK, openapi)
		}
	}

	selfFuncName = runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	return handler
}

func parseAuthenticationMiddleware(middlewareNames []string) []*securitySchemeInfo {
	for _, funcName := range middlewareNames {
		if authSchema := getMidAuthScheme(funcName); len(authSchema) != 0 {
			return authSchema
		}
	}
	return nil
}

// limitString 限制输出的字符数
func limitString(s string, n int) string {
	// n 为负数，或字节数都小于等于 n
	if n < 0 || len(s) <= n {
		return s
	}
	// 转换为 []rune，以正确判断字符数
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n]) + "…"
}

func registerCommonResponseExamples(commExamples map[string]ExampleObject, codes []kit.Code) {
	for _, code := range codes {
		commName := fmt.Sprintf("response_business_status_code_%d", code.Code)
		// 注册通用业务状态码响应示例
		if _, ok := commExamples[commName]; ok {
			continue
		}
		msg := code.Message
		commExamples[commName] = ExampleObject{
			Summary: fmt.Sprintf("状态码 %d: %s", code.Code, limitString(msg, 5)),
			Value: commResponse{
				Code:    code.Code,
				Message: msg,
			},
		}
	}
}
