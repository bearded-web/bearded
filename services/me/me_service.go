package me

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/me"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/passlib/reset"
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
	r.Doc("info")
	r.Operation("info")
	r.Writes(me.Info{}) // on the response
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	r = ws.PUT("/password").To(s.changePassword)
	r.Doc("changePassword")
	r.Operation("changePassword")
	r.Reads(ChangePasswordEntity{})
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	container.Add(ws)
}

func (s *MeService) info(req *restful.Request, resp *restful.Response) {
	mgr := s.Manager()
	defer mgr.Close()

	u := filters.GetUser(req)

	query := manager.Or(fltr.GetQuery(&manager.ProjectFltr{Owner: u.Id, Member: u.Id}))

	projects, count, err := mgr.Projects.FilterByQuery(query)
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

func (s *MeService) changePassword(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): add captcha support
	raw := &ChangePasswordEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	u := filters.GetUser(req)

	if len(raw.Token) > 0 {
		// TODO (m0sth8: take variables from config
		secret := []byte("12345678910")

		getUser := func(email string) ([]byte, error) {
			if email != u.Email {
				return nil, reset.ErrWrongSignature
			}
			return []byte(u.Password), nil
		}

		_, err := reset.VerifyToken(raw.Token, getUser, secret)
		if err != nil {
			if err == reset.ErrExpiredToken {
				resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Token expired, try again"))

			}
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Wrong token, try again"))
			return
		}
	} else {
		// verify old password
		verified, err := s.PassCtx().Verify(raw.Old, u.Password)
		if err != nil {
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
			return
		}
		if !verified {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("old password is incorrect"))
			return
		}
		// TODO (m0sth8): validate new password (length, symbols etc); extract
		if len(raw.New) < 7 {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("new password must be more than 6 symbols"))
			return
		}
	}

	pass, err := s.PassCtx().Encrypt(raw.New)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}
	u.Password = pass

	mgr := s.Manager()
	defer mgr.Close()

	err = mgr.Users.Update(u)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	resp.WriteHeader(http.StatusOK)
}
