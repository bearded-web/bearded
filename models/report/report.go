package report

type Type string

const (
	TypeRaw Type = "raw"
)

type Report struct {
	Type Type   `json:"type"`
	Raw  string `json:"raw"`
}
