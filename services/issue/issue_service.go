package issue

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "issueId"

type IssueService struct {
	*services.BaseService
}

func New(base *services.BaseService) *IssueService {
	return &IssueService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	//	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		//		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

func (s *IssueService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/issues")
	ws.Doc("Manage Issues")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("").To(s.list)
	addDefaults(r)
	r.Doc("list")
	r.Operation("list")
	s.SetParams(r, fltr.GetParams(ws, manager.IssueFltr{}))
	r.Writes(issue.TargetIssueList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	//	r = ws.POST("").To(s.create)
	//	addDefaults(r)
	//	r.Doc("create")
	//	r.Operation("create")
	//	r.Writes(issue.TargetIssue{})
	//	r.Reads(issue.TargetIssue{})
	//	r.Do(services.Returns(http.StatusCreated))
	//	r.Do(services.ReturnsE(
	//		http.StatusBadRequest,
	//		http.StatusConflict,
	//	))
	//	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeIssue(s.get))
	addDefaults(r)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(issue.TargetIssue{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.TakeIssue(s.update))
	// docs
	r.Doc("update")
	r.Operation("update")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(issue.TargetIssue{})
	r.Reads(TargetIssueEntity{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.TakeIssue(s.delete))
	// docs
	r.Doc("delete")
	r.Operation("delete")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusNoContent,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *IssueService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	raw := &issue.TargetIssue{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	obj, err := mgr.Issues.Create(raw)
	if err != nil {
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.DuplicateErr)
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *IssueService) list(req *restful.Request, resp *restful.Response) {
	query, err := fltr.FromRequest(req, manager.IssueFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Issues.FilterByQuery(query)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &issue.TargetIssueList{
		Meta:    pagination.Meta{Count: count},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *IssueService) get(_ *restful.Request, resp *restful.Response, issueObj *issue.TargetIssue) {
	resp.WriteEntity(issueObj)
}

func (s *IssueService) update(req *restful.Request, resp *restful.Response, issueObj *issue.TargetIssue) {
	// TODO (m0sth8): Check permissions

	raw := &TargetIssueEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	mgr := s.Manager()
	defer mgr.Close()

	if raw.Confirmed != nil {
		issueObj.Confirmed = *raw.Confirmed
	}
	if raw.False != nil {
		issueObj.False = *raw.False
	}
	if raw.Resolved != nil {
		issueObj.Resolved = *raw.Resolved
	}
	if raw.Muted != nil {
		issueObj.Muted = *raw.Muted
	}

	if err := mgr.Issues.Update(issueObj); err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.DuplicateErr)
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(issueObj)
}

func (s *IssueService) delete(_ *restful.Request, resp *restful.Response, obj *issue.TargetIssue) {
	mgr := s.Manager()
	defer mgr.Close()

	mgr.Issues.Remove(obj)
	resp.WriteHeader(http.StatusNoContent)
}

func (s *IssueService) TakeIssue(fn func(*restful.Request,
	*restful.Response, *issue.TargetIssue)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Issues.GetById(mgr.ToId(id))
		mgr.Close()
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteErrorString(http.StatusNotFound, "Not found")
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}
		fn(req, resp, obj)
	}
}
