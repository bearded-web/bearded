package token

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"gopkg.in/validator.v2"

	"github.com/bearded-web/bearded/models/token"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "tokenId"

type TokenService struct {
	*services.BaseService
}

func New(base *services.BaseService) *TokenService {
	return &TokenService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

func (s *TokenService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/tokens")
	ws.Doc("Manage Tokens")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	addDefaults(r)
	r.Doc("list")
	r.Operation("list")
	s.SetParams(r, fltr.GetParams(ws, manager.TokenFltr{}))
	r.Writes(token.TokenList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	r = ws.POST("").To(s.create)
	addDefaults(r)
	r.Doc("create")
	r.Operation("create")
	r.Writes(token.Token{})
	r.Reads(TokenEntity{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
	))
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeToken(s.get))
	addDefaults(r)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(token.Token{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.TakeToken(s.update))
	// docs
	r.Doc("update")
	r.Operation("update")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(token.Token{})
	r.Reads(TokenEntity{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.TakeToken(s.delete))
	// docs
	r.Doc("delete")
	r.Operation("delete")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusNoContent,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *TokenService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	raw := &TokenEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(
			http.StatusBadRequest,
			services.WrongEntityErr,
		)
		return
	}
	// validate fields
	if err := validator.Validate(raw); err != nil {
		resp.WriteServiceError(
			http.StatusBadRequest,
			services.NewBadReq("Validation error: %s", err.Error()),
		)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	u := filters.GetUser(req)

	newObj := &token.Token{
		User: u.Id,
		Name: raw.Name,
	}

	obj, err := mgr.Tokens.Create(newObj)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	obj.HashValue = obj.Hash // show token hash after creation only

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *TokenService) list(req *restful.Request, resp *restful.Response) {
	query, err := fltr.FromRequest(req, manager.TokenFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	u := filters.GetUser(req)
	// non admins can't see removed and tokens from other users
	if !u.IsAdmin() {
		query["user"] = u.Id
		query["removed"] = false
	}

	results, count, err := mgr.Tokens.FilterByQuery(query)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &token.TokenList{
		Meta:    pagination.Meta{Count: count},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *TokenService) get(_ *restful.Request, resp *restful.Response, tokenObj *token.Token) {
	resp.WriteEntity(tokenObj)
}

func (s *TokenService) update(req *restful.Request, resp *restful.Response, tokenObj *token.Token) {
	raw := &TokenEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// validate fields
	if err := validator.Validate(raw); err != nil {
		resp.WriteServiceError(
			http.StatusBadRequest,
			services.NewBadReq("Validation error: %s", err.Error()),
		)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	// update token object from entity
	if raw.Name != "" {
		tokenObj.Name = raw.Name
	}

	if err := mgr.Tokens.Update(tokenObj); err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(tokenObj)
}

func (s *TokenService) delete(_ *restful.Request, resp *restful.Response, obj *token.Token) {
	mgr := s.Manager()
	defer mgr.Close()

	err := mgr.Tokens.Remove(obj)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	resp.WriteHeader(http.StatusNoContent)
}

// Helpers

func (s *TokenService) TakeToken(fn func(*restful.Request,
	*restful.Response, *token.Token)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Tokens.GetById(mgr.ToId(id))
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteErrorString(http.StatusNotFound, "Not found")
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}
		u := filters.GetUser(req)

		if !u.IsAdmin() && (obj.User != u.Id || obj.Removed) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		mgr.Close()
		fn(req, resp, obj)
	}
}
