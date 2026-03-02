package swagger

type OpenAPI struct {
	Openapi    string              `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	Info       Info                `json:"info,omitempty" yaml:"info,omitempty"`
	Servers    []Server            `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths" yaml:"paths"`
	Components Components          `json:"components,omitempty" yaml:"components,omitempty"`
}

type Info struct {
	Title       string `json:"title,omitempty" yaml:"title,omitempty" mapstructure:"title"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty" mapstructure:"summary"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	Version     string `json:"version,omitempty" yaml:"version,omitempty" mapstructure:"version"`
}

type Server struct {
	Url         string `json:"url" yaml:"url" mapstructure:"url"`
	Description string `json:"description" yaml:"description" mapstructure:"description"`
}

type PathItem struct {
	Summary     string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation  `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation  `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation  `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation  `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation  `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation  `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation  `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation  `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []Server    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type ExampleObject struct {
	Summary       string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string `json:"description,omitempty" yaml:"description,omitempty"`
	Value         any    `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
	// Ref 与上面的字段不能共存，请在使用时确保只使用一种结构
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type SecurityType string

const (
	SecurityTypeApiKey        SecurityType = "apiKey"
	SecurityTypeHttp          SecurityType = "http"
	SecurityTypeMutualTLS     SecurityType = "mutualTLS"
	SecurityTypeOAuth2        SecurityType = "oauth2"
	SecurityTypeOpenIdConnect SecurityType = "openIdConnect"
)

type SecurityIn string

const (
	SecurityInHeader SecurityIn = "header"
	SecurityInQuery  SecurityIn = "query"
	SecurityInCookie SecurityIn = "cookie"
)

type SecurityScheme struct {
	Type        SecurityType `json:"type,omitempty" yaml:"type,omitempty"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	In          SecurityIn   `json:"in,omitempty" yaml:"in,omitempty"`
	Name        string       `json:"name,omitempty" yaml:"name,omitempty"`
	Scheme      string       `json:"scheme,omitempty" yaml:"scheme,omitempty"`
}

type Components struct {
	Schemas         map[string]Property       `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Examples        map[string]ExampleObject  `json:"examples,omitempty" yaml:"examples,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

type SecurityItem map[string][]string

type Operation struct {
	Tags        []string            `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses,omitempty" yaml:"responses,omitempty"`
	Servers     []Server            `json:"servers,omitempty" yaml:"servers,omitempty"`
	Security    []SecurityItem      `json:"security,omitempty" yaml:"security,omitempty"`
}

type Parameter struct {
	Name        string   `json:"name,omitempty" yaml:"name,omitempty"`
	In          string   `json:"in,omitempty" yaml:"in,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool     `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      Property `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

type Response struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

type MediaType struct {
	Schema   Property                 `json:"schema,omitempty" yaml:"schema,omitempty"`
	Examples map[string]ExampleObject `json:"examples,omitempty" yaml:"examples,omitempty"`
}
