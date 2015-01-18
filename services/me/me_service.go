package me

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/models/me"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
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

// Fix for IntelijIdea inpsections. Cause it can't investigate anonymous method results =(
func (s *MeService) Manager() *manager.Manager {
	return s.BaseService.Manager()
}

func (s *MeService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/me")
	ws.Doc("Current user management")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("").To(s.info)
	// docs
	r.Doc("info")
	//	r.Notes("This endpoint is available only for authenticated users")
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

	projects, count, err := mgr.Projects.GetByOwner(userId)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
	} else {
		// TODO (m0sth8): create default project on user creation.
		// create one default project
		if count == 0 {
			p, err := mgr.Projects.CreateDefault(userId)
			if err != nil {
				logrus.Error(stackerr.Wrap(err))
				// It might be possible, that default project is already created
				// So, client should repeat request
				resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
				return
			} else {
				projects = append(projects, p)
			}
		}
	}

	info := me.Info{
		User:     u,
		Projects: projects,
	}

	resp.WriteEntity(info)
}
