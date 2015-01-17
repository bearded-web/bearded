package auth

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/services"
)

type AuthService struct {
	*services.BaseService
}

func New(base *services.BaseService) *AuthService {
	return &AuthService{
		BaseService: base,
	}
}

func (s *AuthService) Init() error {
	return nil
}

func (s *AuthService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/auth")
	ws.Doc("Authorization management")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.POST("").To(s.login)
	r.Doc("Login")
	r.Operation("login")
	r.Reads(authEntity{})
	r.Returns(http.StatusCreated, "Session created", sessionEntity{})
	r.Do(services.ReturnsE(
		http.StatusInternalServerError,
		http.StatusUnauthorized,
		http.StatusBadRequest))
	ws.Route(r)

	r = ws.DELETE("").To(s.logout)
	r.Doc("Logout")
	r.Operation("logout")
	r.Do(services.Returns(http.StatusNoContent))
	r.Do(services.ReturnsE(
		http.StatusInternalServerError,
		http.StatusBadRequest))
	ws.Route(r)

	r = ws.GET("").To(s.status)
	r.Doc("Status")
	r.Operation("status")
	r.Do(services.Returns(http.StatusOK))
	r.Do(services.ReturnsE(
		http.StatusInternalServerError,
		http.StatusUnauthorized,
		http.StatusBadRequest))
	ws.Route(r)

	container.Add(ws)
}

func (s *AuthService) login(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)

	raw := &authEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// TODO (m0sth8): extract user login and logout, this helps to login in other services
	// TODO (m0sth8): validate password and email, type, max length etc
	if raw.Email == "" {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("password shouldn't be empty"))
		return
	}
	if raw.Password == "" {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("password shouldn't be empty"))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	// get user
	u, err := mgr.Users.GetByEmail(raw.Email)
	if err != nil {
		if mgr.IsNotFound(err) {
			// TODO (m0sth8): add captcha to protect against bruteforce
			resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	// verify password
	verified, err := s.PassCtx().Verify(raw.Password, u.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}
	if !verified {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
		return
	}

	// set userId to session
	session.Set("userId", u.Id.Hex())
	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(sessionEntity{Token: "not ready"})
}

func (s *AuthService) status(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)
	if _, ok := session.Get("userId"); !ok {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
		return
	}
}

func (s *AuthService) logout(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)
	session.Del("userId")
	resp.ResponseWriter.WriteHeader(http.StatusNoContent)
}
