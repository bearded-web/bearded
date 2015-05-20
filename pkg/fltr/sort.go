package fltr

import (
	"fmt"
	"strings"

	"github.com/emicklei/go-restful"
)

const (
	DefaultSortName       = "sort"
	DefaultSortSeparator  = ","
	DefaultSortFieldLimit = 3
)

type Sorter struct {
	SortName       string
	SortSeparator  string
	SortFieldLimit int
	Fields         []string
}

func NewSorter(fields ...string) *Sorter {
	return &Sorter{
		SortName:       DefaultSortName,
		SortSeparator:  DefaultSortSeparator,
		SortFieldLimit: DefaultSortFieldLimit,
		Fields:         fields,
	}
}

func (s *Sorter) Param() *restful.Parameter {
	return restful.QueryParameter(s.SortName, fmt.Sprintf("sort by %s [-%s]", strings.Join(s.Fields, "|"), s.SortSeparator))
}

func (s *Sorter) Parse(req *restful.Request) []string {
	result := []string{}
	p := req.QueryParameter(s.SortName)
	if p == "" {
		return result
	}
	fields := strings.Split(p, s.SortSeparator)
	if len(fields) > s.SortFieldLimit {
		fields = fields[:s.SortFieldLimit]
	}
	for _, field := range fields {
		for _, exField := range s.Fields {
			if exField == field || (strings.HasPrefix(field, "-") && len(field) == (len(exField)+1) && field[1:] == exField) {
				result = append(result, field)
			}
		}
	}
	return result
}
