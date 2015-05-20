package pagination

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emicklei/go-restful"
)

const (
	DefaultLimitName    = "limit"
	DefaultSkipName     = "skip"
	DefaultLimitDefault = 20
	DefaultLimitMax     = 100
)

type Paginator struct {
	LimitName    string
	SkipName     string
	LimitDefault int
	LimitMax     int
	Host         string // if there is no host, then url for previous and next will be relative
}

func New() *Paginator {
	return &Paginator{
		LimitName:    DefaultLimitName,
		SkipName:     DefaultSkipName,
		LimitDefault: DefaultLimitDefault,
		LimitMax:     DefaultLimitMax,
	}
}

func (p *Paginator) LimitParam() *restful.Parameter {
	return restful.QueryParameter(p.LimitName, "limit objects")
}

func (p *Paginator) SkipParam() *restful.Parameter {
	return restful.QueryParameter(p.SkipName, "skip n objects")
}

func (p *Paginator) Parse(req *restful.Request) (skip, limit int) {
	skip = p.ParseSkip(req)
	limit = p.ParseLimit(req)
	return skip, limit
}

func (p *Paginator) Urls(req *restful.Request, skip, limit, count int) (previous, next string) {
	u := req.Request.URL
	prev, err := url.Parse(p.Host)
	if err != nil {
		prev = &url.URL{}
	}
	prev.Path = u.Path
	val := u.Query()
	pSkip := skip - limit
	pLimit := limit
	if pSkip < 0 {
		pLimit += pSkip
		pSkip = 0
	}
	if skip != 0 {
		val.Set(p.SkipName, fmt.Sprintf("%d", pSkip))
		val.Set(p.LimitName, fmt.Sprintf("%d", pLimit))
		prev.RawQuery = val.Encode()
		previous = prev.String()
	}
	if (skip + limit) < count {
		val.Set(p.SkipName, fmt.Sprintf("%d", skip+limit))
		val.Set(p.LimitName, fmt.Sprintf("%d", limit))
		prev.RawQuery = val.Encode()
		next = prev.String()
	}
	return previous, next
}

// parse request and return skip
func (p *Paginator) ParseSkip(req *restful.Request) int {
	skip := 0
	if p := req.QueryParameter(p.SkipName); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			skip = val
		}
	}
	if skip < 0 {
		skip = 0
	}
	return skip

}

// parse request and return limit
func (p *Paginator) ParseLimit(req *restful.Request) int {
	limit := p.LimitDefault
	if p := req.QueryParameter(p.LimitName); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			limit = val
		}
	}
	if limit < 0 {
		limit = p.LimitDefault
	}
	if limit > p.LimitMax {
		limit = p.LimitMax
	}
	return limit
}

func (p *Paginator) PreviousUrl(req *restful.Request) string {
	skip := p.ParseSkip(req)
	limit := p.ParseLimit(req)
	u := req.Request.URL
	prev, err := url.Parse(p.Host)
	if err != nil {
		prev = &url.URL{}
	}
	prev.Path = u.Path
	val := u.Query()
	skip = skip - limit
	if skip < 0 {
		skip = 0
	}
	val.Set(p.SkipName, fmt.Sprintf("%d", skip))
	val.Set(p.LimitName, fmt.Sprintf("%d", limit))
	prev.RawQuery = val.Encode()
	return prev.String()
}

func (p *Paginator) NextUrl(req *restful.Request) string {
	skip := p.ParseSkip(req)
	limit := p.ParseLimit(req)
	u := req.Request.URL
	prev, err := url.Parse(p.Host)
	if err != nil {
		prev = &url.URL{}
	}
	prev.Path = u.Path
	val := u.Query()
	val.Set(p.SkipName, fmt.Sprintf("%d", skip+limit))
	val.Set(p.LimitName, fmt.Sprintf("%d", limit))
	prev.RawQuery = val.Encode()
	return prev.String()
}
