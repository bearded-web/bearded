//go:generate stringer -type=Type
package comment

type Type string

const (
	Scan Type = "scan"
)
