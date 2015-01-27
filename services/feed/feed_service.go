package feed

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/feed"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "feed-id"

type FeedService struct {
	*services.BaseService
}

func New(base *services.BaseService) *FeedService {
	return &FeedService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	r.Notes("Authorization required. Feed by default is returned in descending order by updated field")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

// Fix for IntelijIdea inpsections. Cause it can't investigate anonymous method results =(
func (s *FeedService) Manager() *manager.Manager {
	return s.BaseService.Manager()
}

func (s *FeedService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/feed")
	ws.Doc("Manage Feed")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	addDefaults(r)
	r.Doc("list")
	r.Operation("list")
	r.Param(ws.QueryParameter("project", "filter by project id"))
	r.Param(ws.QueryParameter("target", "filter by target id"))
	r.Param(ws.QueryParameter("type", "filter by type [scan]"))
	r.Writes(feed.Feed{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeFeed(s.get))
	addDefaults(r)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(feed.FeedItem{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.TakeFeed(s.delete))
	r.Doc("delete")
	r.Operation("delete")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusNoContent,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	//	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *FeedService) list(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): check if this user has access to feed items
	mgr := s.Manager()
	defer mgr.Close()
	fltr := mgr.Feed.Fltr()

	if p := req.QueryParameter("project"); p != "" {
		if !s.IsId(p) {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("project must be hex id"))
			return
		}
		fltr.Project = mgr.ToId(p)
	}
	if p := req.QueryParameter("target"); p != "" {
		if !s.IsId(p) {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("target must be hex id"))
			return
		}
		fltr.Target = mgr.ToId(p)
	}
	if p := req.QueryParameter("type"); p != "" {
		fltr.Type = feed.ItemType(p)
	}

	results, count, err := mgr.Feed.FilterBy(fltr)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &feed.Feed{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *FeedService) get(_ *restful.Request, resp *restful.Response, pl *feed.FeedItem) {
	resp.WriteEntity(pl)
}

func (s *FeedService) delete(_ *restful.Request, resp *restful.Response, obj *feed.FeedItem) {
	mgr := s.Manager()
	defer mgr.Close()

	mgr.Feed.Remove(obj)
	resp.WriteHeader(http.StatusNoContent)
}

func (s *FeedService) TakeFeed(fn func(*restful.Request,
	*restful.Response, *feed.FeedItem)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Feed.GetById(mgr.ToId(id))
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
