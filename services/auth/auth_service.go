package auth

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
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

func addDefaults(r *restful.RouteBuilder) {
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

func (s *AuthService) Register(container *restful.Container) {
	authRequired := filters.AuthRequiredFilter(s.BaseManager())

	ws := &restful.WebService{}
	ws.Path("/api/v1/auth")
	ws.Doc("Authorization management")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.POST("").To(s.login)
	r.Doc("login")
	r.Operation("login")
	r.Reads(authEntity{})
	r.Returns(http.StatusCreated, "Session created", sessionEntity{})
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	r = ws.DELETE("").To(s.logout)
	r.Doc("logout")
	r.Operation("logout")
	r.Filter(authRequired)
	r.Do(services.Returns(http.StatusNoContent))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET("").To(s.status)
	r.Doc("status")
	r.Operation("status")
	r.Notes("Returns 200 ok if user is authenticated")
	r.Filter(authRequired)
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	// registration actions
	r = ws.POST("register").To(s.register)
	r.Doc("register")
	r.Operation("register")
	r.Reads(registerEntity{})
	r.Returns(http.StatusCreated, "User registered", sessionEntity{})
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
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
	// users without password can't login
	if u.Password == "" {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
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

	// TODO (m0sth8): extract auth methods, like login or logout.
	// set user id to session
	session.Set(filters.SessionUserKey, u.Id.Hex())
	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(sessionEntity{Token: "not ready"})
}

func (s *AuthService) status(_ *restful.Request, _ *restful.Response) {
	// do nothing, just return 200 ok, cause authorization was checked in filter
}

func (s *AuthService) logout(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)
	session.Del(filters.SessionUserKey)
	resp.WriteHeader(http.StatusNoContent)
}

func (s *AuthService) register(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)

	raw := &registerEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	// TODO (m0sth8): validate email and password
	// TODO (m0sth8): add captcha support

	mgr := s.Manager()
	defer mgr.Close()

	pass, err := s.PassCtx().Encrypt(raw.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}

	u := &user.User{
		Email:    raw.Email,
		Password: pass,
	}

	u, err = mgr.Users.Create(u)
	if err != nil {
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "user with this email is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	session.Set(filters.SessionUserKey, u.Id.Hex())
	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(sessionEntity{Token: "not ready"})
}
