package services

import (
	"bytes"
	"html/template"
	"path"

	//"github.com/bearded-web/bearded/pkg/config"
)

func GetHTML(file string, data interface{}) (out []byte, error error) {
	var buf bytes.Buffer
	//TODO use config path
	t, err := template.ParseFiles(path.Join("./extra/templates/", file))
	if err != nil {
		return nil, err
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
