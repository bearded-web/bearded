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

type Severity string

const (
	SeverityInfo   = Severity("info")
	SeverityLow    = Severity("low")
	SeverityMedium = Severity("medium")
	SeverityHigh   = Severity("high")
	SeverityError  = Severity("error")
)

var severities = []interface{}{
	SeverityInfo,
	SeverityLow,
	SeverityMedium,
	SeverityHigh,
	SeverityError,
}

// It's a hack to show custom type as string in swagger
func (t Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t Severity) Enum() []interface{} {
	return severities
}

func (t Severity) Convert(text string) (interface{}, error) {
	return Severity(text), nil
}
