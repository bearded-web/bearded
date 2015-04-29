package loader

import (
	"github.com/BurntSushi/toml"
	"io"
)

func init() {
	loaders[".toml"] = LoadToml
	loaders[".tml"] = LoadToml
}

func LoadToml(r io.Reader, dst interface{}) error {
	_, err := toml.DecodeReader(r, dst)
	return err
}
