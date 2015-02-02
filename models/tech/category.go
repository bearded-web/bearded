//go:generate stringer -type=Category
package tech

type Category int

const (
	CMS Category = iota
	JavascriptFrameworks
)
