package project

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "project-id"

type ProjectService struct {
	*services.BaseService
}

func New(base *services.BaseService) *ProjectService {
	return &ProjectService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

func (s *ProjectService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/projects")
	ws.Doc("Manage Projects")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	r.Doc("list")
	r.Operation("list")
	r.Writes(project.ProjectList{})
	s.SetParams(r, fltr.GetParams(ws, manager.ProjectFltr{}))
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	r = ws.POST("").To(s.create)
	r.Doc("create")
	r.Operation("create")
	r.Reads(ProjectEntity{})
	r.Writes(project.Project{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(http.StatusConflict))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeProject(s.get))
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(project.Project{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.TakeProject(s.update))
	// docs
	r.Doc("update")
	r.Operation("update")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Reads(ProjectEntity{})
	r.Writes(project.Project{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
		http.StatusInternalServerError))
	ws.Route(r)

	s.RegisterMembers(ws)

	container.Add(ws)
}

func (s *ProjectService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions for the user, he is might be blocked or removed

	raw := &ProjectEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	user := filters.GetUser(req)

	mgr := s.Manager()
	defer mgr.Close()

	obj := &project.Project{
		Name:  raw.Name,
		Owner: user.Id,
	}
	obj, err := mgr.Projects.Create(obj)
	if err != nil {
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "project with this name and owner is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *ProjectService) list(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): check  permissions

	query, err := fltr.FromRequest(req, manager.ProjectFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	u := filters.GetUser(req)
	admin := false
	// if user is not admin then show him only his projects or where he has membership
	if !admin {
		query = manager.Or(fltr.GetQuery(&manager.ProjectFltr{Owner: u.Id, Member: u.Id}))
	}
	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Projects.FilterByQuery(query)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &project.ProjectList{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *ProjectService) get(_ *restful.Request, resp *restful.Response, p *project.Project) {
	resp.WriteEntity(p)
}

func (s *ProjectService) update(req *restful.Request, resp *restful.Response, p *project.Project) {
	raw := &ProjectEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	user := filters.GetUser(req)

	if p.Owner != user.Id {
		resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	if raw.Name != "" {
		p.Name = raw.Name
	}
	if err := mgr.Projects.Update(p); err != nil {
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "project with this name and owner is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	resp.WriteEntity(p)
}

// Helpers

type ProjectFunction func(*restful.Request, *restful.Response, *project.Project)

func (s *ProjectService) TakeProject(fn ProjectFunction) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		// TODO (m0sth8): check permissions for the user
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}
		mgr := s.Manager()
		defer mgr.Close()

		p, err := mgr.Projects.GetById(mgr.ToId(id))
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

		u := filters.GetUser(req)
		admin := false
		if !admin && p.Owner != u.Id && p.GetMember(u.Id) == nil {
			resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
			return
		}

		fn(req, resp, p)
	}
}
