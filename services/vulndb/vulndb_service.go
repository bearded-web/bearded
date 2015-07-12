package vulndb

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"

	"github.com/bearded-web/bearded/models/vuln"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "vulnId"

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

	r := ws.GET("").To(s.list)
	r.Doc("list")
	r.Operation("list")
	r.Param(s.Paginator.SkipParam())
	r.Param(s.Paginator.LimitParam())
	r.Writes(vuln.VulnList{}) // on the response
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeVuln(s.get))
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(vuln.Vuln{}) // on the response
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET("compact").To(s.compact)
	r.Doc("compact list")
	r.Operation("compact")
	r.Writes(vuln.CompactVulnList{}) // on the response
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	container.Add(ws)
}

func (s *VulndbService) list(req *restful.Request, resp *restful.Response) {
	mgr := s.Manager()
	defer mgr.Close()

	results := mgr.Vulndb.GetVulns()
	count := len(results)
	skip, limit := s.Paginator.Parse(req)
	if skip < 0 || skip >= count {
		skip = count
	}
	rLimit := skip + limit
	if rLimit >= count {
		rLimit = count
	}
	results = results[skip:rLimit]

	previous, next := s.Paginator.Urls(req, skip, limit, count)
	data := vuln.VulnList{
		Meta: pagination.Meta{
			Count:    count,
			Previous: previous,
			Next:     next,
		},
		Results: results,
	}
	resp.WriteEntity(data)
}

func (s *VulndbService) get(req *restful.Request, resp *restful.Response, v *vuln.Vuln) {
	resp.WriteEntity(v)
}

func (s *VulndbService) compact(req *restful.Request, resp *restful.Response) {
	results := []*vuln.CompactVuln{}

	mgr := s.Manager()
	defer mgr.Close()

	for _, vuln := range mgr.Vulndb.GetVulns() {
		results = append(results, vuln.Compact())
	}

	data := vuln.CompactVulnList{
		Meta:    pagination.Meta{Count: len(results)},
		Results: results,
	}
	resp.WriteEntity(data)
}

// Helpers

func (s *VulndbService) TakeVuln(fn func(*restful.Request,
	*restful.Response, *vuln.Vuln)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		idStr := req.PathParameter(ParamId)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("id must be int"))
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj := mgr.Vulndb.GetById(id)
		if obj == nil {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		mgr.Close()

		fn(req, resp, obj)
	}
}
