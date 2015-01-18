package target

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/pkg/filters"
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

// Fix for IntelijIdea inpsections. Cause it can't investigate anonymous method results =(
func (s *TargetService) Manager() *manager.Manager {
	return s.BaseService.Manager()
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
	r.Writes(target.TargetList{})
	r.Param(ws.QueryParameter("project", "filter by project id"))
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

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.get)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(target.Target{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.delete)
	r.Doc("delete")
	r.Operation("delete")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(http.StatusNoContent))
	addDefaults(r)
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
	if raw.Type == target.Web {
		if raw.Web == nil || raw.Web.Domain == "" { // TODO (m0sth8): check domain format
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("web.domain is required for target.type=web"))
			return
		}
	}

	user := filters.GetUser(req)

	mgr := s.Manager()
	defer mgr.Close()

	// TODO (m0sth8): check if the user has permission to add a target to the project
	proj, err := mgr.Projects.GetById(mgr.FromId(raw.Project))
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
	var (
		results []*target.Target
		count   int
		err     error
	)
	mgr := s.Manager()
	defer mgr.Close()

	projectId := req.QueryParameter("project")
	// TODO (m0sth8): check project existence and permissions

	if projectId != "" {
		if !s.IsId(projectId) {
			resp.WriteServiceError(
				http.StatusBadRequest,
				services.NewError(services.CodeIdHex, "project should be bson uuid in hex form"))
			return
		}
		results, count, err = mgr.Targets.GetByProject(projectId)
	} else {
		results, count, err = mgr.Targets.All()
	}
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

func (s *TargetService) get(req *restful.Request, resp *restful.Response) {
	targetId := req.PathParameter(ParamId)
	if !s.IsId(targetId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	// TODO (m0sth8): check permissions for the user for the project of this target
	u, err := mgr.Targets.GetById(targetId)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteEntity(u)
}

func (s *TargetService) delete(req *restful.Request, resp *restful.Response) {
	targetId := req.PathParameter(ParamId)
	if !s.IsId(targetId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	// TODO (m0sth8): check permissions for the user for the project of this target
	obj, err := mgr.Targets.GetById(targetId)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	mgr.Targets.Remove(obj)

	resp.WriteHeader(http.StatusNoContent)
}
