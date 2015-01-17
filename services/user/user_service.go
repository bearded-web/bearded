package user

import (
	"net/http"
	"time"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/services"
)

type UserService struct {
	userCol *mgo.Collection
	passCtx *passlib.Context
}

func New(col *mgo.Collection, passCtx *passlib.Context) *UserService {
	return &UserService{
		userCol: col,
		passCtx: passCtx,
	}
}

func (s *UserService) Init() error {
	logrus.Infof("Initialize user indexes")
	s.userCol.EnsureIndex(mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		Background: false,
	})
	return nil
}

func (s *UserService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.
		Path("/api/v1/users").
		Doc("User management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").To(s.list).
		// docs
		Doc("List users").
		Operation("list").
		Writes(user.UserList{}). // on the response
		Do(
		services.Returns(http.StatusOK),
		services.ReturnsE(http.StatusInternalServerError, http.StatusBadRequest),
	))
	ws.Route(ws.POST("").To(s.create).
		// docs
		Doc("Create user").
		Operation("create").
		Writes(user.User{}). // on the response
		Reads(user.User{}).
		Do(
		services.Returns(http.StatusCreated),
		services.ReturnsE(http.StatusConflict, http.StatusInternalServerError),
	))
	ws.Route(ws.GET("{user-id}").To(s.get).
		// docs
		Doc("Get user").
		Operation("get").
		Param(ws.PathParameter("user-id", "")).
		Writes(user.User{}). // on the response
		Do(
		services.Returns(http.StatusOK, http.StatusNotFound),
		services.ReturnsE(http.StatusBadRequest, http.StatusInternalServerError),
	))
	ws.Route(ws.POST("{user-id}/password").To(s.setPassword).
		// docs
		Doc("Set password, only for administrator").
		Operation("setPassword").
		Reads(Password{}).
		Param(ws.PathParameter("user-id", "")).
		Do(
		services.Returns(http.StatusCreated, http.StatusNotFound, http.StatusUnauthorized, http.StatusForbidden),
		services.ReturnsE(http.StatusBadRequest, http.StatusInternalServerError),
	))

	container.Add(ws)
}

// Get user collection with mongo session
func (s *UserService) users(session *mgo.Session) *mgo.Collection {
	return s.userCol.With(session)
}

// ====== service operations

func (s *UserService) list(req *restful.Request, resp *restful.Response) {
	plugins := s.users(filters.GetMongo(req))

	results := []*user.User{}

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

	result := &user.UserList{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *UserService) create(req *restful.Request, resp *restful.Response) {
	users := s.users(filters.GetMongo(req))

	raw := &user.User{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// TODO: add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now()
	raw.Updated = raw.Created
	if err := users.Insert(raw); err != nil {
		if mgo.IsDup(err) {
			logrus.Debug(stackerr.Wrap(err))
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "user with this email is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(raw)
}

func (s *UserService) get(req *restful.Request, resp *restful.Response) {
	users := s.users(filters.GetMongo(req))

	u := &user.User{}

	userId := req.PathParameter("user-id")
	if !bson.IsObjectIdHex(userId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	if err := users.FindId(bson.ObjectIdHex(userId)).One(u); err != nil {
		if err == mgo.ErrNotFound {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteEntity(u)
}

type Password struct {
	Password string `json:"password"`
}

func (s *UserService) setPassword(req *restful.Request, resp *restful.Response) {

	raw := &Password{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	users := s.users(filters.GetMongo(req))

	obj := &user.User{}

	userId := req.PathParameter("user-id")
	if !bson.IsObjectIdHex(userId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	if err := users.FindId(bson.ObjectIdHex(userId)).One(obj); err != nil {
		if err == mgo.ErrNotFound {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	pass, err := s.passCtx.Encrypt(raw.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}
	obj.Password = pass

	if err := users.UpdateId(obj.Id, obj); err != nil {
		if err == mgo.ErrNotFound {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	// resp.WriteHeader(http.StatusCreated) - this method doesn't work if body isn't written
	resp.ResponseWriter.WriteHeader(http.StatusCreated)
}
