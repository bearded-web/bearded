package tech

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"gopkg.in/mgo.v2/bson"
	//	"gopkg.in/validator.v2"

	"github.com/bearded-web/bearded/models/tech"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "techId"

type TechService struct {
	*services.BaseService
	sorter *fltr.Sorter
}

func New(base *services.BaseService) *TechService {
	return &TechService{
		BaseService: base,
		sorter:      fltr.NewSorter("created", "updated", "confidence", "version"),
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

func (s *TechService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/techs")
	ws.Doc("Manage Techs")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthTokenFilter(s.BaseManager()))
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	addDefaults(r)
	r.Doc("list")
	r.Operation("list")
	s.SetParams(r, fltr.GetParams(ws, manager.TechFltr{}))
	r.Param(ws.QueryParameter("search", "search by summary and description"))
	r.Param(s.sorter.Param())
	r.Param(s.Paginator.SkipParam())
	r.Param(s.Paginator.LimitParam())
	r.Writes(tech.TargetTechList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	//	r = ws.POST("").To(s.create)
	//	addDefaults(r)
	//	r.Doc("create")
	//	r.Operation("create")
	//	r.Writes(tech.TargetTech{})
	//	r.Reads(TargetTechEntity{})
	//	r.Do(services.Returns(http.StatusCreated))
	//	r.Do(services.ReturnsE(
	//		http.StatusBadRequest,
	//		http.StatusConflict,
	//	))
	//	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeTech(s.get))
	addDefaults(r)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(tech.TargetTech{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.TakeTech(s.update))
	// docs
	r.Doc("update")
	r.Operation("update")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(tech.TargetTech{})
	r.Reads(TargetTechEntity{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.TakeTech(s.delete))
	// docs
	r.Doc("delete")
	r.Operation("delete")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusNoContent,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.GET("categories").To(s.categories)
	addDefaults(r)
	r.Doc("categories")
	r.Operation("categories")
	r.Writes(tech.CategoryList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations
//
//func (s *TechService) create(req *restful.Request, resp *restful.Response) {
//	// TODO (m0sth8): Check permissions
//	raw := &TargetTechEntity{}
//
//	if err := req.ReadEntity(raw); err != nil {
//		logrus.Error(stackerr.Wrap(err))
//		resp.WriteServiceError(
//			http.StatusBadRequest,
//			services.WrongEntityErr,
//		)
//		return
//	}
//	// check target field, it must be present
//	if !s.IsId(raw.Target) {
//		resp.WriteServiceError(
//			http.StatusBadRequest,
//			services.NewBadReq("Target is wrong"),
//		)
//		return
//	}
//	// validate other fields
//	if err := validator.WithTag("creating").Validate(raw); err != nil {
//		resp.WriteServiceError(
//			http.StatusBadRequest,
//			services.NewBadReq("Validation error: %s", err.Error()),
//		)
//		return
//	}
//
//	mgr := s.Manager()
//	defer mgr.Close()
//
//	// load target and project
//	t, err := mgr.Targets.GetById(mgr.ToId(raw.Target))
//	if err != nil {
//		if mgr.IsNotFound(err) {
//			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Target not found"))
//			return
//		}
//		logrus.Error(stackerr.Wrap(err))
//		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
//		return
//	}
//
//	//	current user should have a permission to create tech there
//	u := filters.GetUser(req)
//
//	if sErr := services.Must(services.HasProjectIdPermission(mgr, u, t.Project)); sErr != nil {
//		sErr.Write(resp)
//		return
//	}
//
//	newObj := &tech.TargetTech{
//		Project: t.Project,
//		Target:  t.Id,
//	}
//	updateTargetTech(raw, newObj)
//	newObj.AddUserReportActivity(u.Id)
//
//	obj, err := mgr.Techs.Create(newObj)
//	if err != nil {
//		if mgr.IsDup(err) {
//			resp.WriteServiceError(
//				http.StatusConflict,
//				services.DuplicateErr)
//			return
//		}
//		logrus.Error(stackerr.Wrap(err))
//		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
//		return
//	}
//	// TODO (m0sth8): extract to worker
//	func(mgr *manager.Manager) {
//		tgt, err := mgr.Targets.GetById(obj.Target)
//		if err != nil {
//			logrus.Error(stackerr.Wrap(err))
//			return
//		}
//		err = mgr.Targets.UpdateSummary(tgt)
//		if err != nil {
//			logrus.Error(stackerr.Wrap(err))
//			return
//		}
//	}(s.Manager())
//	resp.WriteHeader(http.StatusCreated)
//	resp.WriteEntity(obj)
//}

func (s *TechService) list(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): show techs only if user has permissions
	query, err := fltr.FromRequest(req, manager.TechFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	if search := req.QueryParameter("search"); search != "" {
		if mgr.Cfg.TextSearchEnable {
			query["$text"] = &bson.M{"$search": search}
		}
	}

	skip, limit := s.Paginator.Parse(req)

	opt := manager.Opts{
		Sort:  s.sorter.Parse(req),
		Limit: limit,
		Skip:  skip,
	}
	results, count, err := mgr.Techs.FilterByQuery(query, opt)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	previous, next := s.Paginator.Urls(req, skip, limit, count)
	result := &tech.TargetTechList{
		Meta: pagination.Meta{
			Count:    count,
			Previous: previous,
			Next:     next,
		},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *TechService) get(_ *restful.Request, resp *restful.Response, techObj *tech.TargetTech) {
	resp.WriteEntity(techObj)
}

func (s *TechService) update(req *restful.Request, resp *restful.Response, techObj *tech.TargetTech) {
	raw := &TargetTechEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	mgr := s.Manager()
	defer mgr.Close()

	// update tech object from entity
	updateTargetTech(raw, techObj)

	if err := mgr.Techs.Update(techObj); err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.DuplicateErr)
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(techObj)
}

func (s *TechService) delete(_ *restful.Request, resp *restful.Response, obj *tech.TargetTech) {
	mgr := s.Manager()
	defer mgr.Close()

	mgr.Techs.Remove(obj)
	resp.WriteHeader(http.StatusNoContent)
}

func (s *TechService) categories(_ *restful.Request, resp *restful.Response) {
	result := &tech.CategoryList{
		Meta: pagination.Meta{
			Count: len(tech.Categories),
		},
		Results: tech.Categories,
	}
	resp.WriteEntity(result)
}

// Helpers

func (s *TechService) TakeTech(fn func(*restful.Request,
	*restful.Response, *tech.TargetTech)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Techs.GetById(mgr.ToId(id))
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteErrorString(http.StatusNotFound, "Not found")
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}

		sErr := services.Must(services.HasProjectIdPermission(mgr, filters.GetUser(req), obj.Project))
		if sErr != nil {
			sErr.Write(resp)
			return
		}

		mgr.Close()

		fn(req, resp, obj)
	}
}
