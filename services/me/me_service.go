package me

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/models/me"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/services"
	"github.com/bearded-web/bearded/pkg/filters"

)

type MeService struct {
	userCol *mgo.Collection
	passCtx *passlib.Context
}

func New(col *mgo.Collection, passCtx *passlib.Context) *MeService {
	return &MeService{
		userCol: col,
		passCtx: passCtx,
	}
}

func (s *MeService) Init() error {
	return nil
}

func (s *MeService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.
		Path("/api/v1/me").
		Doc("Current user management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").To(s.info).
		// docs
		Doc("Information").
		Operation("info").
		Writes(me.Info{}). // on the response
		Do(services.Returns(http.StatusOK), services.ReturnsE(http.StatusInternalServerError, http.StatusUnauthorized)))

	container.Add(ws)
}

// Get user collection with mongo session
func (s *MeService) users(session *mgo.Session) *mgo.Collection {
	return s.userCol.With(session)
}

func (s *MeService) info(req *restful.Request, resp *restful.Response) {
	users := s.users(filters.GetMongo(req))
	session := filters.GetSession(req)

	userId, existed := session.Get("userId");
	if !existed {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
		return
	}
	if !bson.IsObjectIdHex(userId) {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
		// TODO (m0sth8): logout here
		return
	}

	u := &user.User{}
	if err := users.FindId(bson.ObjectIdHex(userId)).One(u); err != nil {
		if err == mgo.ErrNotFound {
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
