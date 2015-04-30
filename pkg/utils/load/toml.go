package load

import (
	"io"

	"github.com/BurntSushi/toml"
)

const TomlFormat = Format("toml")

func LoadToml(r io.Reader, dst interface{}) error {
	_, err := toml.DecodeReader(r, dst)
	return err
}
