package plugin

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

type PluginService struct {
	pluginCol *mgo.Collection
}

func New(pluginCol *mgo.Collection) *PluginService {
	return &PluginService{
		pluginCol: pluginCol,
	}
}

func (s *PluginService) Init() error {
	logrus.Infof("Initialize plugin indexes")
	s.pluginCol.EnsureIndex(mgo.Index{
		Key:        []string{"name", "version"},
		Unique:     true,
		Background: false,
	})
	return nil
}

func (s *PluginService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.
		Path("/api/v1/plugins").
		Doc("Manage Plugins").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").To(s.list).
		// docs
		Doc("List plugins").
		Operation("list").
		Writes(plugin.PluginList{}). // on the response
		Do(
		services.Returns(http.StatusOK),
		services.ReturnsE(http.StatusInternalServerError),
	))

	ws.Route(ws.POST("").To(s.create).
		// docs
		Doc("Create plugin").
		Operation("create").
		Writes(plugin.Plugin{}). // on the response
		Reads(plugin.Plugin{}).
		Do(
		services.Returns(http.StatusCreated),
		services.ReturnsE(http.StatusConflict, http.StatusInternalServerError),
	))

	ws.Route(ws.GET("{plugin-id}").To(s.get).
		// docs
		Doc("Get plugin").
		Operation("get").
		Param(ws.PathParameter("plugin-id", "")).
		Writes(plugin.Plugin{}). // on the response
		Do(
		services.Returns(http.StatusOK, http.StatusNotFound),
		services.ReturnsE(http.StatusBadRequest, http.StatusInternalServerError),
	))

	ws.Route(ws.PUT("{plugin-id}").To(s.update).
		// docs
		Doc("Update plugin").
		Operation("update").
		Param(ws.PathParameter("plugin-id", "")).
		Writes(plugin.Plugin{}). // on the response
		Reads(plugin.Plugin{}).
		Do(
		services.Returns(http.StatusOK, http.StatusNotFound),
		services.ReturnsE(http.StatusBadRequest, http.StatusInternalServerError),
	))

	container.Add(ws)
}

// Get plugins collection with mongo session
func (s *PluginService) plugins(session *mgo.Session) *mgo.Collection {
	return s.pluginCol.With(session)
}

// ====== service operations

func (s *PluginService) create(req *restful.Request, resp *restful.Response) {
	plugins := s.plugins(filters.GetMongo(req))

	pl := &plugin.Plugin{}

	if err := req.ReadEntity(pl); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// TODO: add validation
	pl.Id = bson.NewObjectId()
	if err := plugins.Insert(pl); err != nil {
		if mgo.IsDup(err) {
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
	resp.WriteEntity(pl)
}

func (s *PluginService) list(req *restful.Request, resp *restful.Response) {
	plugins := s.plugins(filters.GetMongo(req))

	results := []*plugin.Plugin{}

	query := &bson.M{}
	q := plugins.Find(query)
	if err := q.All(&results); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	count, err := q.Count()
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
	plugins := s.plugins(filters.GetMongo(req))

	pl := &plugin.Plugin{}

	pluginId := req.PathParameter("plugin-id")
	if !bson.IsObjectIdHex(pluginId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	if err := plugins.FindId(bson.ObjectIdHex(pluginId)).One(pl); err != nil {
		if err == mgo.ErrNotFound {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteEntity(pl)
}

func (s *PluginService) update(req *restful.Request, resp *restful.Response) {
	plugins := s.plugins(filters.GetMongo(req))

	pluginId := req.PathParameter("plugin-id")
	if !bson.IsObjectIdHex(pluginId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	pl := &plugin.Plugin{}

	if err := req.ReadEntity(pl); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// TODO: add validation
	pl.Id = bson.ObjectIdHex(pluginId)

	if err := plugins.UpdateId(pl.Id, pl); err != nil {
		if err == mgo.ErrNotFound {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(pl)
}
