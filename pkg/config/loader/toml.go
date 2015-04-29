package loader

import (
	"io"

	"github.com/BurntSushi/toml"
)

func init() {
	loaders[".toml"] = LoadToml
	loaders[".tml"] = LoadToml
}

func LoadToml(r io.Reader, dst interface{}) error {
	_, err := toml.DecodeReader(r, dst)
	return err
}
