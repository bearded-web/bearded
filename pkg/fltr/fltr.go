package fltr

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	restful "github.com/emicklei/go-restful"
	"github.com/fatih/structs"
	"gopkg.in/mgo.v2/bson"
)

var (
	eq  string = "eq"
	gt         = "gt"
	gte        = "gte"
	lt         = "lt"
	lte        = "lte"
	in         = "in"
	neq        = "neq"
	nin        = "nin"

	all = []string{eq, gt, gte, lt, lte, in, neq, nin}
)

var (
	FilterTag       = "fltr"
	ModifierDivider = "_"
)

type Enumer interface {
	Enum() []interface{}
}

type Converter interface {
	Convert(string) (interface{}, error)
}

type EnumConverter interface {
	Enumer
	Converter
}

func GetQuery(raw interface{}) bson.M {
	result := bson.M{}
	if raw == nil {
		return result
	}
	s := structs.New(raw)
	for _, field := range s.Fields() {
		if field.IsZero() {
			continue
		}
		name := field.Name()
		if tagValue := field.Tag(FilterTag); tagValue != "" {
			tags := strings.Split(tagValue, ",")
			if len(tags) > 0 {
				if tags[0] != "" {
					name = tags[0] // take field name from tag
				}
			}
		}
		result[name] = field.Value()
	}
	return result
}

//
func GetParams(ws *restful.WebService, raw interface{}) []*restful.Parameter {
	params := []*restful.Parameter{}
	s := structs.New(raw)
	for _, field := range s.Fields() {
		name := field.Name()
		tags := strings.Split(field.Tag(FilterTag), ",")
		if len(tags) > 0 {
			if tags[0] != "" {
				name = tags[0] // take field name from tag
			}
			tags = tags[1:]
		}
		modifiers := getModifiers(tags)

		desc := field.Tag("description")
		// generate description automatically
		if desc == "" {
			desc = fmt.Sprintf("filter by %s", name)
			if enum, casted := field.Value().(Enumer); casted {
				enumValues := enum.Enum()
				desc = fmt.Sprintf("%s, one of %v", desc, enumValues)
			}
		}
		if desc == "-" {
			desc = ""
		}

		param := ws.QueryParameter(name, desc)
		if hasTag(tags, "required") {
			param.Required(true)
		}

		params = append(params, param)

		for _, m := range modifiers {
			mName := fmt.Sprintf("%s%s%s", name, ModifierDivider, m)
			param := ws.QueryParameter(mName, desc)
			if m == in || m == nin {
				param.AllowMultiple(true)
			}
			params = append(params, param)
		}
	}

	return params
}

func getModifiers(tags []string) []string {
	modifiers := []string{}
	if hasTag(tags, "all") {
		return append(modifiers, all...)
	} else {
		for _, m := range all {
			if hasTag(tags, m) {
				modifiers = append(modifiers, m)
			}
		}
	}
	return modifiers
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func FromRequest(req *restful.Request, raw interface{}) (bson.M, error) {
	result := bson.M{}
	s := structs.New(raw)
fields:
	for _, field := range s.Fields() {
		name := field.Name()
		tags := strings.Split(field.Tag(FilterTag), ",")
		if len(tags) > 0 {
			name = tags[0] // take field name from tag
			tags = tags[1:]
		}
		val := req.QueryParameter(name)
		if val != "" {
			if v, err := parseValue(field, val); err != nil {
				return nil, fmt.Errorf("param %s: %v", name, err)
			} else {
				result[name] = v
				// if there is an eq value then we just skip modifiers
				continue fields
			}
		}
		modifiers := getModifiers(tags)
	modifiers:
		for _, m := range modifiers {
			mName := fmt.Sprintf("%s%s%s", name, ModifierDivider, m)
			val := req.QueryParameter(mName)
			if val == "" {
				continue modifiers
			}
			if m == in || m == nin {
				ins := []interface{}{}
				for _, val := range strings.Split(val, ",") {
					if v, err := parseValue(field, val); err != nil {
						return nil, fmt.Errorf("param %s: %v", mName, err)
					} else {
						ins = append(ins, v)
					}
				}
				result[name] = bson.M{fmt.Sprintf("$%s", m): ins}
				continue modifiers
			}
			// gt, gte, lt, lte, ne
			if v, err := parseValue(field, val); err != nil {
				return nil, fmt.Errorf("param %s: %v", name, err)
			} else {
				result[name] = bson.M{fmt.Sprintf("$%s", m): v}
				// if there is an eq value then we just skip modifiers
				continue fields
			}
		}
	}
	return result, nil
}

func parseValue(field *structs.Field, val string) (interface{}, error) {
	switch field.Kind() {
	case reflect.Int:
		v, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		return v, nil
	case reflect.Float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return v, nil
	default:
		fieldVal := field.Value()

		if _, casted := fieldVal.(bson.ObjectId); casted {
			if bson.IsObjectIdHex(val) {
				return bson.ObjectIdHex(val), nil
			} else {
				return nil, fmt.Errorf("should be bson.ObjectId hex")
			}
		}
		if _, casted := (fieldVal).(time.Time); casted {
			v := &time.Time{}
			return v, v.UnmarshalText([]byte(val))
		}
		if convertable, casted := fieldVal.(Converter); casted {
			converted, err := convertable.Convert(val)
			if err != nil {
				return nil, err
			}
			if enum, casted := fieldVal.(Enumer); casted {
				println("enum!")
				enumValues := enum.Enum()
				for _, eV := range enumValues {
					if eV == converted {
						return converted, nil
					}
				}
				return nil, fmt.Errorf("should be one of %v", enumValues)
			}
			return converted, nil
		}
		return val, nil
	}

}
