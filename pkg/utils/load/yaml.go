package load

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const YamlFormat = Format("yaml")

func LoadYaml(r io.Reader, dst interface{}) error {
	var data []byte
	data, err := ioutil.ReadAll(r)
	if err == nil {
		err = yaml.Unmarshal(data, dst)
	}
	return err
}
