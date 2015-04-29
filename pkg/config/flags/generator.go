package flags

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/m0sth8/cli"
	"reflect"
	"strings"
)

const (
	DefaultDescTag = "desc"
	DefaultFlagTag = "flag"
	DefaultEnvTag  = "env"
)

type Opts struct {
	DescTag   string
	FlagTag   string
	Prefix    string
	EnvPrefix string
}

func GenerateFlags(cfg interface{}, opts ...Opts) []cli.Flag {
	prefix := ""
	envPrefix := ""
	descTag := DefaultDescTag
	flagTag := DefaultFlagTag
	for _, opt := range opts {
		if opt.DescTag != "" {
			descTag = opt.DescTag
		}
		if opt.Prefix != "" {
			prefix = opt.Prefix
		}
		if opt.EnvPrefix != "" {
			envPrefix = opt.EnvPrefix
		}
	}
	flags := []cli.Flag{}
	s := structs.New(cfg)
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
		envVar := FlagToEnv(flagName)
		ignoreEnvPrefix := false
		if envTags := strings.Split(field.Tag(DefaultEnvTag), ","); len(envTags) > 0 {
			switch envName := envTags[0]; envName {
			case "-":
				// if tag is `env:"-"` then remove env var
				envVar = ""
			case "":
				// if tag is `env:""` then env var is taken from flag name
			default:
				// if tag is `env:"NAME"` then env var is envPrefix_flagPrefix_NAME
				// if tag is `env:"~NAME"` then env var is NAME
				if strings.HasPrefix(envName, "~") {
					envVar = envName[1:]
					ignoreEnvPrefix = true
				} else {
					envVar = envName
					if prefix != "" {
						envVar = fmt.Sprintf("%s%s%s", FlagToEnv(prefix), envDivider, envVar)
					}
				}
			}
		}
		if envVar != "" && !ignoreEnvPrefix && envPrefix != "" {
			envVar = fmt.Sprintf("%s%s%s", envPrefix, envDivider, envVar)
		}
		usage := field.Tag(descTag)
		var f cli.Flag
		switch field.Kind() {
		case reflect.String:
			f = cli.StringFlag{
				Name:   flagName,
				Value:  field.Value().(string),
				EnvVar: envVar,
				Usage:  usage,
			}
		case reflect.Int:
			f = cli.IntFlag{
				Name:   flagName,
				Value:  field.Value().(int),
				EnvVar: envVar,
				Usage:  usage,
			}
		case reflect.Bool:
			f = cli.BoolFlag{
				Name:   flagName,
				EnvVar: envVar,
				Usage:  usage,
			}
		case reflect.Struct:
			opt := Opts{
				DescTag:   descTag,
				Prefix:    flagName,
				FlagTag:   flagTag,
				EnvPrefix: envPrefix,
			}
			flags = append(flags, GenerateFlags(field.Value(), opt)...)
		}

		if f != nil {
			flags = append(flags, f)
		}
	}
	return flags
}
