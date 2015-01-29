package target

import "encoding/json"

type TargetType string

const (
	TypeWeb    TargetType = "web"
	TypeMobile TargetType = "mobile"
)

var targetTypes = []interface{}{TypeWeb, TypeMobile}

// It's a hack to show custom type as string in swagger
func (t TargetType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t TargetType) Enum() []interface{} {
	return targetTypes
}

func (t TargetType) Convert(text string) (interface{}, error) {
	return TargetType(text), nil
}
