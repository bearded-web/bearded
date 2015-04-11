package report

import "encoding/json"

type ReportType string

const (
	TypeEmpty  = ReportType("empty")
	TypeRaw    = ReportType("raw")
	TypeMulti  = ReportType("multi")
	TypeIssues = ReportType("issues")
	TypeTechs  = ReportType("techs")
)

var reportTypes = []interface{}{
	TypeEmpty,
	TypeRaw,
	TypeMulti,
	TypeIssues,
	TypeTechs,
}

// It's a hack to show custom type as string in swagger
func (t ReportType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t ReportType) Enum() []interface{} {
	return reportTypes
}

func (t ReportType) Convert(text string) (interface{}, error) {
	return ReportType(text), nil
}
