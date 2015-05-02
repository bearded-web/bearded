package load

import (
	"encoding/json"
	"io"
)

const JsonFormat = Format("json")

func LoadJson(r io.Reader, dst interface{}) error {
	return json.NewDecoder(r).Decode(dst)
}
