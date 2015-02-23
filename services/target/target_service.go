package target

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/comment"
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "target-id"

type TargetService struct {
	*services.BaseService
}

func New(base *services.BaseService) *TargetService {
	return &TargetService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
		http.StatusForbidden,
		http.StatusBadRequest,
	))
}

func (s *TargetService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/targets")
	ws.Doc("Manage Targets")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	r.Doc("list")
	r.Operation("list")
	s.SetParams(r, fltr.GetParams(ws, manager.TargetFltr{}))
	r.Writes(target.TargetList{})
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	r = ws.POST("").To(s.create)
	r.Doc("create")
	r.Operation("create")
	r.Writes(target.Target{})
	r.Reads(target.Target{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(http.StatusConflict))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeTarget(s.get))
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(target.Target{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.TakeTarget(s.delete))
	r.Doc("delete")
	r.Operation("delete")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(http.StatusNoContent))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}/comments", ParamId)).To(s.TakeTarget(s.comments))
	r.Doc("comments")
	r.Operation("comments")
	r.Param(ws.PathParameter(ParamId, ""))
	//	s.SetParams(r, fltr.GetParams(ws, manager.CommentFltr{}))
	r.Writes(comment.CommentList{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	ws.Route(r)

	r = ws.POST(fmt.Sprintf("{%s}/comments", ParamId)).To(s.TakeTarget(s.commentsAdd))
	r.Doc("commentsAdd")
	r.Operation("commentsAdd")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Reads(CommentEntity{})
	r.Writes(comment.Comment{})
	r.Do(services.Returns(
		http.StatusCreated,
		http.StatusNotFound))
	ws.Route(r)

	//	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.update)
	//	// docs
	//	r.Doc("update")
	//	r.Operation("update")
	//	r.Param(ws.PathParameter(ParamId, ""))
	//	r.Writes(target.Target{})
	//	r.Reads(target.Target{})
	//	r.Do(services.Returns(
	//		http.StatusOK,
	//		http.StatusNotFound))
	//	r.Do(services.ReturnsE(
	//		http.StatusBadRequest,
	//		http.StatusInternalServerError))
	//	ws.Route(r)

	container.Add(ws)
}

func (s *TargetService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions for the user, he is might be blocked or removed

	raw := &target.Target{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// TODO (m0sth8): add validation and extract it to manager
	if raw.Type == target.TypeWeb {
		if raw.Web == nil || raw.Web.Domain == "" { // TODO (m0sth8): check domain format
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("web.domain is required for target.type=web"))
			return
		}
		addr, err := url.Parse(raw.Web.Domain)
		if err != nil {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
			return
		}
		if addr.Scheme == "" || !(addr.Scheme == "http" || addr.Scheme == "https") {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("scheme must be http or https"))
			return
		}
		raw.Web.Domain = addr.String()
	}

	user := filters.GetUser(req)

	mgr := s.Manager()
	defer mgr.Close()

	// TODO (m0sth8): check if the user has permission to add a target to the project
	proj, err := mgr.Projects.GetById(raw.Project)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteServiceError(http.StatusForbidden, services.NewError(services.CodeAuthForbid, "project doesn't exist"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	if proj.Owner != user.Id {
		resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
		return
	}

	obj, err := mgr.Targets.Create(raw)
	if err != nil {
		//		if mgr.IsDup(err) {
		//			resp.WriteServiceError(
		//				http.StatusConflict,
		//				services.NewError(services.CodeDuplicate, "target with this name and owner is existed"))
		//			return
		//		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *TargetService) list(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): check project existence and permissions
	query, err := fltr.FromRequest(req, manager.TargetFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Targets.FilterByQuery(query)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &target.TargetList{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *TargetService) get(_ *restful.Request, resp *restful.Response, obj *target.Target, _ *project.Project) {
	resp.WriteEntity(obj)
}

func (s *TargetService) delete(_ *restful.Request, resp *restful.Response, obj *target.Target, _ *project.Project) {
	// TODO (m0sth8): do not remove target, just mark as deleted
	mgr := s.Manager()
	defer mgr.Close()

	mgr.Targets.Remove(obj)

	resp.WriteHeader(http.StatusNoContent)
}

func (s *TargetService) comments(_ *restful.Request, resp *restful.Response, obj *target.Target, _ *project.Project) {
	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Comments.FilterBy(&manager.CommentFltr{Type: comment.Scan, Link: obj.Id})

	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &comment.CommentList{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *TargetService) commentsAdd(req *restful.Request, resp *restful.Response, t *target.Target, _ *project.Project) {
	ent := &CommentEntity{}
	if err := req.ReadEntity(ent); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	u := filters.GetUser(req)
	raw := &comment.Comment{
		Owner: u.Id,
		Type:  comment.Scan,
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

func (s *TargetService) TakeTarget(fn func(*restful.Request,
	*restful.Response, *target.Target, *project.Project)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		// TODO (m0sth8): check permissions for the user for the project of this target
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		t, err := mgr.Targets.GetById(mgr.ToId(id))
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteErrorString(http.StatusNotFound, "Target not found")
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}
		p, err := mgr.Projects.GetById(t.Project)
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteErrorString(http.StatusNotFound, "Project not found")
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}
		u := filters.GetUser(req)
		admin := false
		if !admin && p.Owner != u.Id && p.GetMember(u.Id) == nil {
			resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
			return
		}
		mgr.Close()
		fn(req, resp, t, p)
	}
}
