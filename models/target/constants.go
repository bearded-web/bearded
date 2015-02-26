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

type IssueStatus string

const (
	IssueConfirmed = IssueStatus("confirmed") // when a pentester confirm that issue is real
	IssueFalse     = IssueStatus("false")     // it's a false issue
	IssueFixed     = IssueStatus("fixed")
)

var issueStatuses = []interface{}{IssueConfirmed, IssueFalse, IssueFixed}

// It's a hack to show custom type as string in swagger
func (t IssueStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t IssueStatus) Enum() []interface{} {
	return issueStatuses
}

func (t IssueStatus) Convert(text string) (interface{}, error) {
	return IssueStatus(text), nil
}
