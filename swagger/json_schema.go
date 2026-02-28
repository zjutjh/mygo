package swagger

type Property struct {
	Ref         string               `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Items       *Property            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties  map[string]*Property `json:"properties,omitempty" yaml:"properties,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`

	// null/boolean/number/string/array/object
	Type    string `json:"type,omitempty" yaml:"type,omitempty"`
	Enum    []any  `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default any    `json:"default,omitempty" yaml:"default,omitempty"`

	// string type: date-time/date/time/duration
	// string type: email/idn-email
	// string type: hostname/idn-hostname
	// string type: ipv4/ipv6
	// string type: uri/uri-reference/iri/iri-reference/uuid
	// string type: uri-template
	// string type: json-pointer/relative-json-pointer
	// string type: regex
	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	// for numbers instances
	MultipleOf       float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Minimum          float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMaximum float64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum float64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`

	// for string instances
	MaxLength int64  `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength int64  `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern   string `json:"pattern,omitempty" yaml:"pattern,omitempty"`

	// for array instances
	MaxItems    int64 `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems    int64 `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems bool  `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxContains int64 `json:"maxContains,omitempty" yaml:"maxContains,omitempty"`
	MinContains int64 `json:"minContains,omitempty" yaml:"minContains,omitempty"`

	// for object instances
	MaxProperties int64    `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties int64    `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required      []string `json:"required,omitempty" yaml:"required,omitempty"`

	AdditionalProperties bool `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
}
