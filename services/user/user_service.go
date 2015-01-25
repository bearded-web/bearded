package user

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

type UserService struct {
	*services.BaseService
}

func New(base *services.BaseService) *UserService {
	return &UserService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusInternalServerError,
	))
}

func (s *UserService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/users")
	ws.Doc("User management")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	r.Doc("list")
	r.Operation("list")
	r.Writes(user.UserList{})
	r.Do(services.Returns(http.StatusOK))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	r = ws.POST("").To(s.create)
	r.Doc("create")
	r.Operation("create")
	r.Writes(user.User{}) // on the response
	r.Reads(user.User{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(http.StatusConflict))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET("{user-id}").To(s.get)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter("user-id", ""))
	r.Writes(user.User{}) // on the response
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	r = ws.POST("{user-id}/password").To(s.setPassword)
	r.Doc("setPassword")
	r.Operation("setPassword")
	r.Reads(passwordEntity{})
	r.Param(ws.PathParameter("user-id", ""))
	r.Do(services.Returns(
		http.StatusCreated,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	r.Notes("Authorization required. This method available only for administrator")

	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *UserService) list(_ *restful.Request, resp *restful.Response) {
	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Users.All()
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
	// TODO (m0sth8): Check permissions
	raw := &user.User{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	obj, err := mgr.Users.Create(raw)
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

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *UserService) get(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	userId := req.PathParameter("user-id")
	if !s.IsId(userId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
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

	resp.WriteEntity(u)
}

func (s *UserService) setPassword(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	userId := req.PathParameter("user-id")
	if !s.IsId(userId) {
		resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
		return
	}

	raw := &passwordEntity{}
	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
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

	pass, err := s.PassCtx().Encrypt(raw.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}
	u.Password = pass

	if err := mgr.Users.Update(u); err != nil {
		if mgr.IsNotFound(err) {
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
