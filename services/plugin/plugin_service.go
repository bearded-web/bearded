package plugin

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

type PluginService struct {
	*services.BaseService
}

func New(base *services.BaseService) *PluginService {
	return &PluginService{
		BaseService: base,
	}
}

func (s *PluginService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/plugins")
	ws.Doc("Manage Plugins")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("").To(s.list)
	// docs
	r.Doc("list")
	r.Operation("list")
	r.Writes(plugin.PluginList{})
	r.Do(services.Returns(http.StatusOK))
	r.Do(services.ReturnsE(http.StatusInternalServerError))
	ws.Route(r)

	r = ws.POST("").To(s.create)
	// docs
	r.Doc("create")
	r.Operation("create")
	r.Writes(plugin.Plugin{})
	r.Reads(plugin.Plugin{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(
		http.StatusConflict,
		http.StatusInternalServerError))
	ws.Route(r)

	r = ws.GET("{plugin-id}").To(s.get)
	// docs
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter("plugin-id", ""))
	r.Writes(plugin.Plugin{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
		http.StatusInternalServerError))
	ws.Route(r)

	r = ws.PUT("{plugin-id}").To(s.update)
	// docs
	r.Doc("update")
	r.Operation("update")
	r.Param(ws.PathParameter("plugin-id", ""))
	r.Writes(plugin.Plugin{})
	r.Reads(plugin.Plugin{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
		http.StatusInternalServerError))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *PluginService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	raw := &plugin.Plugin{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	obj, err := mgr.Plugins.Create(raw)
	if err != nil {
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "plugin with this name and version is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *PluginService) list(_ *restful.Request, resp *restful.Response) {
	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Plugins.All()
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &plugin.PluginList{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *PluginService) get(req *restful.Request, resp *restful.Response) {
	pluginId := req.PathParameter("plugin-id")
	if !s.IsId(pluginId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	u, err := mgr.Plugins.GetById(pluginId)
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

func (s *PluginService) update(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	pluginId := req.PathParameter("plugin-id")
	if !s.IsId(pluginId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	raw := &plugin.Plugin{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	mgr := s.Manager()
	defer mgr.Close()

	raw.Id = mgr.ToId(pluginId)

	if err := mgr.Plugins.Update(raw); err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "plugin with this name and version is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(raw)
}
