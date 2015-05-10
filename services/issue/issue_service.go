package issue

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/validator.v2"

	"github.com/bearded-web/bearded/models/comment"
	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/pkg/filters"
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
	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusInternalServerError,
	))
}

func (s *IssueService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/issues")
	ws.Doc("Manage Issues")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthTokenFilter(s.BaseManager()))
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	addDefaults(r)
	r.Doc("list")
	r.Operation("list")
	s.SetParams(r, fltr.GetParams(ws, manager.IssueFltr{}))
	r.Writes(issue.TargetIssueList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	r = ws.POST("").To(s.create)
	addDefaults(r)
	r.Doc("create")
	r.Operation("create")
	r.Writes(issue.TargetIssue{})
	r.Reads(TargetIssueEntity{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
		http.StatusConflict,
	))
	ws.Route(r)

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

	r = ws.GET(fmt.Sprintf("{%s}/comments", ParamId)).To(s.TakeIssue(s.comments))
	r.Doc("comments")
	r.Operation("comments")
	r.Param(ws.PathParameter(ParamId, ""))
	//	s.SetParams(r, fltr.GetParams(ws, manager.CommentFltr{}))
	r.Writes(comment.CommentList{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	ws.Route(r)

	r = ws.POST(fmt.Sprintf("{%s}/comments", ParamId)).To(s.TakeIssue(s.commentsAdd))
	r.Doc("commentsAdd")
	r.Operation("commentsAdd")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Reads(CommentEntity{})
	r.Writes(comment.Comment{})
	r.Do(services.Returns(
		http.StatusCreated,
		http.StatusNotFound))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *IssueService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	raw := &TargetIssueEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(
			http.StatusBadRequest,
			services.WrongEntityErr,
		)
		return
	}
	// check target field, it must be present
	if !s.IsId(raw.Target) {
		resp.WriteServiceError(
			http.StatusBadRequest,
			services.NewBadReq("Target is wrong"),
		)
		return
	}
	// validate other fields
	if err := validator.WithTag("creating").Validate(raw); err != nil {
		resp.WriteServiceError(
			http.StatusBadRequest,
			services.NewBadReq("Validation error: %s", err.Error()),
		)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	// load target and project
	t, err := mgr.Targets.GetById(mgr.ToId(raw.Target))
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Target not found"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	p, err := mgr.Projects.GetById(t.Project)
	if err != nil {
		if mgr.IsNotFound(err) {
			// This situation is really strange
			logrus.Errorf("Target %s is existed, but his project %s isn't",
				raw.Target, mgr.FromId(t.Project))
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Project not found"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	//	current user should have a permission to create issue there
	u := filters.GetUser(req)
	if !mgr.Permission.HasProjectAccess(p, u) {
		resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
		return
	}

	newObj := &issue.TargetIssue{
		Project: p.Id,
		Target:  t.Id,
	}
	updateTargetIssue(raw, newObj)
	newObj.AddUserReportActivity(u.Id)

	obj, err := mgr.Issues.Create(newObj)
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

	// update issue object from entity
	updateTargetIssue(raw, issueObj)

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

func (s *IssueService) comments(_ *restful.Request, resp *restful.Response, obj *issue.TargetIssue) {
	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Comments.FilterBy(&manager.CommentFltr{Type: comment.Issue, Link: obj.Id})

	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &comment.CommentList{
		Meta:    pagination.Meta{Count: count},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *IssueService) commentsAdd(req *restful.Request, resp *restful.Response, t *issue.TargetIssue) {
	ent := &CommentEntity{}
	if err := req.ReadEntity(ent); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	if len(ent.Text) == 0 {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Text is required"))
		return
	}

	u := filters.GetUser(req)
	raw := &comment.Comment{
		Owner: u.Id,
		Type:  comment.Issue,
		Link:  t.Id,
		Text:  ent.Text,
	}

	mgr := s.Manager()
	defer mgr.Close()

	obj, err := mgr.Comments.Create(raw)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

// Helpers

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
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteErrorString(http.StatusNotFound, "Not found")
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}

		if !s.hasProjectPermission(req, resp, obj.Project) {
			resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
			return
		}

		mgr.Close()
		fn(req, resp, obj)
	}
}

func (s *IssueService) hasProjectPermission(req *restful.Request, resp *restful.Response,
	projectId bson.ObjectId) bool {

	mgr := s.Manager()
	defer mgr.Close()

	p, err := mgr.Projects.GetById(projectId)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Project not found"))
			return false
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return false
	}

	//	current user should have a permission to create issue there
	u := filters.GetUser(req)
	if !mgr.Permission.HasProjectAccess(p, u) {
		logrus.Warnf("User %s try to access to project %s", u, p)
		return false
	}
	return true
}
