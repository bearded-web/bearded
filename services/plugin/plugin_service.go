package plugin

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "plugin-id"

type PluginService struct {
	*services.BaseService
}

func New(base *services.BaseService) *PluginService {
	return &PluginService{
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

// Fix for IntelijIdea inpsections. Cause it can't investigate anonymous method results =(
func (s *PluginService) Manager() *manager.Manager {
	return s.BaseService.Manager()
}

func (s *PluginService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/plugins")
	ws.Doc("Manage Plugins")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("").To(s.list)
	addDefaults(r)
	r.Doc("list")
	r.Operation("list")
	s.SetParams(r, fltr.GetParams(ws, manager.PluginFltr{}))

	r.Writes(plugin.PluginList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	r = ws.POST("").To(s.create)
	addDefaults(r)
	r.Doc("create")
	r.Operation("create")
	r.Writes(plugin.Plugin{})
	r.Reads(plugin.Plugin{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
		http.StatusConflict,
	))
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakePlugin(s.get))
	addDefaults(r)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(plugin.Plugin{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.TakePlugin(s.update))
	// docs
	r.Doc("update")
	r.Operation("update")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(plugin.Plugin{})
	r.Reads(plugin.Plugin{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
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

func (s *PluginService) list(req *restful.Request, resp *restful.Response) {
	query, err := fltr.FromRequest(req, manager.PluginFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Plugins.FilterByQuery(query)
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

func (s *PluginService) get(_ *restful.Request, resp *restful.Response, pl *plugin.Plugin) {
	resp.WriteEntity(pl)
}

func (s *PluginService) update(req *restful.Request, resp *restful.Response, pl *plugin.Plugin) {
	// TODO (m0sth8): Check permissions

	raw := &plugin.Plugin{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	mgr := s.Manager()
	defer mgr.Close()

	raw.Id = pl.Id

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

func (s *PluginService) TakePlugin(fn func(*restful.Request,
	*restful.Response, *plugin.Plugin)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		pluginId := req.PathParameter(ParamId)
		if !s.IsId(pluginId) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		pl, err := mgr.Plugins.GetById(pluginId)
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
		fn(req, resp, pl)
	}
}
