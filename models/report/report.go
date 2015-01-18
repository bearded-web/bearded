package report

import "encoding/json"

type ReportType string

const (
	TypeRaw ReportType = "raw"
)

// It's a hack to show custom type as string in swagger
func (t ReportType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

type Report struct {
	Type ReportType   `json:"type"`
	Raw  string `json:"raw"`
}
