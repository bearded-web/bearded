package vulndb

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

type VulndbService struct {
	*services.BaseService
}

func New(base *services.BaseService) *VulndbService {
	return &VulndbService{
		BaseService: base,
	}
}

func (s *VulndbService) Init() error {
	return nil
}

func addDefaults(r *restful.RouteBuilder) {
	r.Do(services.ReturnsE(
		http.StatusInternalServerError,
	))
}

func (s *VulndbService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/vulndb")
	ws.Doc("Vulnerability database")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("compact").To(s.compact)
	r.Doc("compact list")
	r.Operation("compact")
	r.Writes(CompactVulnList{}) // on the response
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	container.Add(ws)
}

func (s *VulndbService) compact(req *restful.Request, resp *restful.Response) {
	results := []*CompactVuln{
		&CompactVuln{Id: 1, Title: "Allowed HTTP methods", Severity: issue.SeverityInfo},
		&CompactVuln{Id: 45, Title: "SQL Injection", Severity: issue.SeverityHigh},
		&CompactVuln{Id: 55, Title: "Reflected Cross-Site Scripting (XSS)", Severity: issue.SeverityHigh},
	}

	data := CompactVulnList{
		Meta:    pagination.Meta{Count: len(results)},
		Results: results,
	}

	resp.WriteEntity(data)
}
