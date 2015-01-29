package me

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

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

func addDefaults(r *restful.RouteBuilder) {
	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

func (s *MeService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/me")
	ws.Doc("Current user management")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.info)
	// docs
	r.Doc("info")
	//	r.Notes("This endpoint is available only for authenticated users")
	r.Operation("info")
	r.Writes(me.Info{}) // on the response
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	container.Add(ws)
}

func (s *MeService) info(req *restful.Request, resp *restful.Response) {
	mgr := s.Manager()
	defer mgr.Close()

	u := filters.GetUser(req)

	projects, count, err := mgr.Projects.FilterBy(&manager.ProjectFltr{Owner: u.Id})
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
	} else {
		// TODO (m0sth8): create default project when user on create is triggered.
		// create one default project
		if count == 0 {
			p, err := mgr.Projects.CreateDefault(u.Id)
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
