package user

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/pkg/validate"
	"github.com/bearded-web/bearded/services"
)

type UserService struct {
	*services.BaseService
	sorter *fltr.Sorter
}

func New(base *services.BaseService) *UserService {
	return &UserService{
		BaseService: base,
		sorter:      fltr.NewSorter("created", "updated", "email", "nickname"),
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
	ws.Filter(filters.AuthTokenFilter(s.BaseManager()))
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	r.Doc("list")
	r.Operation("list")
	r.Writes(user.UserList{})
	s.SetParams(r, fltr.GetParams(ws, manager.UserFltr{}))
	r.Param(s.sorter.Param())
	r.Param(s.Paginator.SkipParam())
	r.Param(s.Paginator.LimitParam())
	r.Do(services.Returns(http.StatusOK))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	r = ws.POST("").To(s.create)
	r.Doc("create")
	r.Operation("create")
	r.Writes(user.User{}) // on the response
	r.Reads(userEntity{})
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

func (s *UserService) list(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): do not show emails and other fields for everyone
	// TODO (m0sth8): filter by email for admin only
	query, err := fltr.FromRequest(req, manager.UserFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	skip, limit := s.Paginator.Parse(req)
	opt := manager.Opts{
		Sort:  s.sorter.Parse(req),
		Limit: limit,
		Skip:  skip,
	}
	results, count, err := mgr.Users.FilterByQuery(query, opt)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	previous, next := s.Paginator.Urls(req, skip, limit, count)
	result := &user.UserList{
		Meta:    pagination.Meta{Count: count, Previous: previous, Next: next},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *UserService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	raw := &userEntity{}
	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	u := filters.GetUser(req)
	if !mgr.Permission.IsAdmin(u) {
		logrus.Warnf("User %s try to create user without admin permission", u)
		resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
		return
	}

	// check email
	if valid, err := govalidator.ValidateStruct(raw); !valid {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}
	// check password
	if valid, reason := validate.Password(raw.Password); !valid {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Password %s", reason))
		return
	}
	// hash password
	pass, err := s.PassCtx().Encrypt(raw.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}

	obj, err := mgr.Users.Create(&user.User{
		Nickname: raw.Nickname,
		Email:    raw.Email,
		Password: pass,
	})
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

	u, err := mgr.Users.GetById(mgr.ToId(userId))
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
	// TODO (m0sth8): Check permissions for admins
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

	u, err := mgr.Users.GetById(mgr.ToId(userId))
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	currentUser := filters.GetUser(req)
	if !mgr.Permission.IsAdmin(currentUser) || currentUser.Id != u.Id {
		logrus.Warnf("User %s try to set password for user %s", currentUser, u)
		resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
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
