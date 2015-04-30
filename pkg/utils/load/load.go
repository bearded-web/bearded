package load

import (
	"fmt"
	"io"
	"os"
	"path"
)

type Format string

const (
	UnknownFormat = Format("")
)

type Opts struct {
	Format Format
}

type LoaderFunc func(io.Reader, interface{}) error

type Loader struct {
	ExtToFormat map[string]Format
	Loaders     map[Format]LoaderFunc
}

func New() *Loader {
	return &Loader{
		ExtToFormat: map[string]Format{},
		Loaders:     map[Format]LoaderFunc{},
	}
}

// load data to dst struct from file according to format
func (l *Loader) FromFile(filename string, dst interface{}, opts ...Opts) error {
	// take format from filename extension
	format := UnknownFormat
	for _, opt := range opts {
		format = opt.Format
	}
	if format == UnknownFormat {
		ext := path.Ext(filename)
		if f, ok := l.ExtToFormat[ext]; !ok {
			return fmt.Errorf("Cant't recognize format from extension %s", ext)
		} else {
			format = f
		}
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return l.FromReader(f, dst, format)
}

// load data to dst struct from f according to format
func (l *Loader) FromReader(r io.Reader, dst interface{}, format Format) error {
	var err error
	if loaderFunc, ok := l.Loaders[format]; !ok {
		err = fmt.Errorf("Unknown format %s", format)
	} else {
		err = loaderFunc(r, dst)
	}
	return err
}

var DefaultLoader = Loader{
	Loaders: map[Format]LoaderFunc{
		TomlFormat: LoadToml,
		JsonFormat: LoadJson,
		YamlFormat: LoadYaml,
	},
	ExtToFormat: map[string]Format{
		".toml": TomlFormat,
		".tml":  TomlFormat,
		".json": JsonFormat,
		".yaml": YamlFormat,
		".yml":  YamlFormat,
	},
}

// load data to dst struct from file according to format with default loader
func FromFile(filename string, dst interface{}, opts ...Opts) error {
	return DefaultLoader.FromFile(filename, dst, opts...)
}

// load data to dst struct from f according to format with default loader
func FromReader(r io.Reader, dst interface{}, format Format) error {
	return DefaultLoader.FromReader(r, dst, format)

}
