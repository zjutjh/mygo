package swagger

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Name(sf reflect.StructField, key string) string {
	name := sf.Tag.Get(key)
	if name == "" {
		return sf.Name
	}
	return strings.Split(name, ",")[0]
}

func Desc(sf reflect.StructField) string {
	desc := sf.Tag.Get("desc")
	if desc == "" {
		desc = sf.Name
	}
	return "[" + sf.Type.Kind().String() + "]" + desc
}

func Binding(sf reflect.StructField) []string {
	binding := sf.Tag.Get("binding")
	if binding == "" {
		return []string{}
	}
	return strings.Split(binding, ",")
}

func RequestRequired(sf reflect.StructField) bool {
	bindingList := strings.SplitSeq(sf.Tag.Get("binding"), ",")
	for binding := range bindingList {
		if binding == "required" {
			return true
		}
	}
	return false
}

func ResponseRequired(sf reflect.StructField) bool {
	jsonTag := sf.Tag.Get("json")
	if jsonTag == "-" {
		return false
	}

	jsonList := strings.SplitSeq(jsonTag, ",")
	for part := range jsonList {
		if part == "omitempty" {
			return false
		}
	}
	return true
}

func LEN(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 4 || binding[:4] != "len=" {
		return false
	}
	vi, err := strconv.ParseInt(binding[4:], 10, 64)
	if nil != err {
		return false
	}
	if reflect.String == kind {
		property.MinLength = vi
		property.MaxLength = vi
	} else {
		property.MinItems = vi
		property.MaxItems = vi
	}
	return true
}

func MAX(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 4 || binding[:4] != "max=" {
		return false
	}
	if reflect.String == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MaxLength = vi
			return true
		}
	} else if reflect.Slice == kind || reflect.Array == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MaxItems = vi
			return true
		}
	} else {
		vf, err := strconv.ParseFloat(binding[4:], 64)
		if nil == err {
			property.Maximum = vf
			property.ExclusiveMaximum = 0
			return true
		}
	}
	return false
}

func MIN(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 4 || binding[:4] != "min=" {
		return false
	}
	if reflect.String == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MinLength = vi
			return true
		}
	} else if reflect.Slice == kind || reflect.Array == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MinItems = vi
			return true
		}
	} else {
		vf, err := strconv.ParseFloat(binding[4:], 64)
		if nil == err {
			property.Minimum = vf
			property.ExclusiveMinimum = 0
			return true
		}
	}
	return false
}

func EQ(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 3 || binding[:3] != "eq=" {
		return false
	}
	if reflect.String == kind {
		property.Enum = append(property.Enum, binding[3:])
		return true
	} else if reflect.Slice == kind || reflect.Array == kind {
		vi, err := strconv.ParseInt(binding[3:], 10, 64)
		if nil == err {
			property.MinItems = vi
			property.MaxItems = vi
			return true
		}
	} else if reflect.Float32 == kind || reflect.Float64 == kind {
		vf, err := strconv.ParseFloat(binding[3:], 64)
		if nil == err {
			property.Enum = append(property.Enum, vf)
			return true
		}
	} else {
		vi, err := strconv.ParseInt(binding[3:], 10, 64)
		if nil == err {
			property.Enum = append(property.Enum, vi)
			return true
		}
	}
	return false
}

func ONEOF(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 6 || binding[:6] != "oneof=" {
		return false
	}
	list := strings.Split(binding[6:], " ")
	for _, s := range list {
		if reflect.String == kind {
			property.Enum = append(property.Enum, s)
		} else if reflect.Float32 == kind || reflect.Float64 == kind {
			vf, err := strconv.ParseFloat(s, 64)
			if nil == err {
				property.Enum = append(property.Enum, vf)
			}
		} else {
			vi, err := strconv.ParseInt(s, 10, 64)
			if nil == err {
				property.Enum = append(property.Enum, vi)
			}
		}
	}
	return true
}

func GT(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 3 || binding[:3] != "gt=" {
		return false
	}
	if reflect.String == kind {
		vi, err := strconv.ParseInt(binding[3:], 10, 64)
		if nil == err {
			property.MinLength = vi + 1
			return true
		}
	} else if reflect.Slice == kind || reflect.Array == kind {
		vi, err := strconv.ParseInt(binding[3:], 10, 64)
		if nil == err {
			property.MinItems = vi + 1
			return true
		}
	} else {
		v, err := strconv.ParseFloat(binding[3:], 64)
		if nil == err {
			property.Minimum = v
			property.ExclusiveMinimum = v
			return true
		}
	}
	return false
}

func GTE(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 4 || binding[:4] != "gte=" {
		return false
	}
	if reflect.String == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MinLength = vi
			return true
		}
	} else if reflect.Slice == kind || reflect.Array == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MinItems = vi
			return true
		}
	} else {
		vf, err := strconv.ParseFloat(binding[4:], 64)
		if nil == err {
			property.Minimum = vf
			property.ExclusiveMinimum = 0
			return true
		}
	}
	return false
}

func LT(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 3 || binding[:3] != "lt=" {
		return false
	}
	if reflect.String == kind {
		vi, err := strconv.ParseInt(binding[3:], 10, 64)
		if nil == err {
			property.MaxLength = vi - 1
			return true
		}
	} else if reflect.Slice == kind || reflect.Array == kind {
		vi, err := strconv.ParseInt(binding[3:], 10, 64)
		if nil == err {
			property.MaxItems = vi - 1
			return true
		}
	} else {
		v, err := strconv.ParseFloat(binding[3:], 64)
		if nil == err {
			property.Maximum = v
			property.ExclusiveMaximum = v
			return true
		}
	}
	return false
}

func LTE(binding string, property *Property, kind reflect.Kind) bool {
	if len(binding) < 4 || binding[:4] != "lte=" {
		return false
	}
	if reflect.String == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MaxLength = vi
			return true
		}
	} else if reflect.Slice == kind || reflect.Array == kind {
		vi, err := strconv.ParseInt(binding[4:], 10, 64)
		if nil == err {
			property.MaxItems = vi
			return true
		}
	} else {
		vf, err := strconv.ParseFloat(binding[4:], 64)
		if nil == err {
			property.Maximum = vf
			property.ExclusiveMaximum = 0
			return true
		}
	}
	return false
}

func DATETIME(binding string, property *Property) bool {
	if len(binding) < 9 || binding[:9] != "datetime=" {
		return false
	}
	if time.DateTime == binding[9:] {
		property.Format = "date-time"
	} else if time.DateOnly == binding[9:] {
		property.Format = "date"
	} else if time.TimeOnly == binding[9:] {
		property.Format = "time"
	} else {
		return false
	}
	return true
}

func EMAIL(binding string, property *Property) bool {
	if binding != "email" {
		return false
	}
	property.Format = "email"
	return true
}

func IP(binding string, property *Property) bool {
	switch binding {
	case "ip":
		property.Format = "ip"
	case "ipv4":
		property.Format = "ipv4"
	case "ipv6":
		property.Format = "ipv6"
	default:
		return false
	}
	return true
}

func HOSTNAME(binding string, property *Property) bool {
	if binding == "hostname" || binding == "hostname_rfc1123" {
		property.Format = "hostname"
		return true
	}
	return false
}
