package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/m0sth8/cli"
)

func ParseFlags(cfg interface{}, ctx *cli.Context, opts ...Opts) error {
	prefix := ""
	flagTag := DefaultFlagTag
	for _, opt := range opts {
		if opt.Prefix != "" {
			prefix = opt.Prefix
		}
	}
	s := structs.New(cfg)
	srcStruct := reflect.ValueOf(cfg)
	if srcStruct.Kind() == reflect.Ptr {
		srcStruct = reflect.Indirect(srcStruct)
	}
fields:
	for _, field := range s.Fields() {
		flagName := CamelToFlag(field.Name())
		if flagTags := strings.Split(field.Tag(flagTag), ","); len(flagTags) > 0 {
			switch fName := flagTags[0]; fName {
			case "-":
				continue fields
			case "":
			default:
				flagName = fName
			}
		}
		if prefix != "" {
			flagName = fmt.Sprintf("%s%s%s", prefix, flagDivider, flagName)
		}
		var err error
		switch field.Kind() {
		case reflect.String:
			val := ctx.String(flagName)
			// environment doesn't count in IsSet operation, don't know why.
			if val == "" {
				if !ctx.IsSet(flagName) {
					continue fields
				}
			}
			err = field.Set(val)
		case reflect.Bool:
			val := ctx.Bool(flagName)
			// environment doesn't count in IsSet operation, don't know why.
			if val == false {
				if !ctx.IsSet(flagName) {
					continue fields
				}
			}
			err = field.Set(val)
		case reflect.Int:
			val := ctx.Int(flagName)
			// environment doesn't count in IsSet operation, don't know why.
			if val == 0 {
				if !ctx.IsSet(flagName) {
					continue fields
				}
			}
			err = field.Set(val)
		case reflect.Struct:
			opt := Opts{
				Prefix:  flagName,
				FlagTag: flagTag,
			}
			realField := srcStruct.FieldByName(field.Name())
			err = ParseFlags(realField.Addr().Interface(), ctx, opt)
		case reflect.Slice:
			realField := srcStruct.FieldByName(field.Name())
			switch realField.Type().Elem().Kind() {
			case reflect.String:
				val := ctx.StringSlice(flagName)
				if len(val) == 0 {
					if !ctx.IsSet(flagName) {
						continue fields
					}
				}
				err = field.Set(val)
			case reflect.Int:
				val := ctx.IntSlice(flagName)
				if len(val) == 0 {
					if !ctx.IsSet(flagName) {
						continue fields
					}
				}
				err = field.Set(val)
			}
		}
		if err != nil {
			return fmt.Errorf("Flag %s parsing error: %s", flagName, err.Error())
		}
	}
	return nil
}
