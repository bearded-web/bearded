// Loader package helps to load configuration from files with different formats
package loader

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
)

type loader func(io.Reader, interface{}) error

var loaders map[string]loader = map[string]loader{
	".json": LoadJson,
}

// Encode file base on extension
// .toml by toml, .yml by yaml, .json by json
func LoadFromFile(filepath string, dst interface{}) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	ext := path.Ext(filepath)
	if loader, ok := loaders[ext]; !ok {
		return fmt.Errorf("Unknown file type %s", ext)
	} else {
		return loader(f, dst)
	}
}

func LoadJson(r io.Reader, dst interface{}) error {
	return json.NewDecoder(r).Decode(dst)
}
