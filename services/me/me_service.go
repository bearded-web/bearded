package me

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/models/me"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/services"
)

type MeService struct {
	*services.BaseService
}

func New(base *services.BaseService) *MeService {
	return &MeService{
		BaseService: base,
	}
}

func (s *MeService) Init() error {
	return nil
}

func (s *MeService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/me")
	ws.Doc("Current user management")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("").To(s.info)
	// docs
	r.Doc("Information")
	r.Operation("info")
	r.Writes(me.Info{}) // on the response
	r.Do(services.Returns(http.StatusOK))
	r.Do(services.ReturnsE(
		http.StatusInternalServerError,
		http.StatusUnauthorized))
	ws.Route(r)

	container.Add(ws)
}

func (s *MeService) info(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)

	userId, existed := session.Get("userId")
	if !existed {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
		return
	}
	if !s.IsId(userId) {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
		// TODO (m0sth8): logout here
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	u, err := mgr.Users.GetById(userId)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	info := me.Info{
		User: u,
	}

	resp.WriteEntity(info)
}
