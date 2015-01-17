package auth

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/services"
	"github.com/facebookgo/stackerr"
)

type AuthService struct {
	userCol *mgo.Collection
	passCtx *passlib.Context
}

func New(col *mgo.Collection, passCtx *passlib.Context) *AuthService {
	return &AuthService{
		userCol: col,
		passCtx: passCtx,
	}
}

func (s *AuthService) Init() error {
	return nil
}

func (s *AuthService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.
		Path("/api/v1/auth").
		Doc("Authorization management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("").To(s.login).
		// docs
		Doc("Login").
		Operation("login").
		Reads(authEntity{}).
		Returns(http.StatusCreated, "Session created", sessionEntity{}).
		Do(
		services.ReturnsE(http.StatusInternalServerError, http.StatusUnauthorized, http.StatusBadRequest),
	))
	ws.Route(ws.DELETE("").To(s.logout).
		// docs
		Doc("Logout").
		Operation("login").
		Do(
		services.Returns(http.StatusNoContent),
		services.ReturnsE(http.StatusInternalServerError, http.StatusBadRequest),
	))
	ws.Route(ws.GET("").To(s.status).
		// docs
		Doc("Status").
		Operation("status").
		Do(
		services.Returns(http.StatusOK),
		services.ReturnsE(http.StatusInternalServerError, http.StatusUnauthorized, http.StatusBadRequest),
	))
	container.Add(ws)
}

// Get user collection with mongo session
func (s *AuthService) users(session *mgo.Session) *mgo.Collection {
	return s.userCol.With(session)
}

func (s *AuthService) login(req *restful.Request, resp *restful.Response) {
	users := s.users(filters.GetMongo(req))
	session := filters.GetSession(req)

	raw := &authEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// TODO (m0sth8): validate password and email, type, max length etc
	if raw.Email == "" {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("password shouldn't be empty"))
		return
	}
	if raw.Password == "" {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("password shouldn't be empty"))
		return
	}

	u := &user.User{}
	if err := users.Find(bson.D{{"email", raw.Email}}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			// TODO (m0sth8) add captcha to protect against bruteforce
			resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	verified, err := s.passCtx.Verify(raw.Password, u.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}
	if !verified {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
		return
	}
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
