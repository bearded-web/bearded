package agent

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/models/agent"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "agent-id"

type AgentService struct {
	*services.BaseService
}

func New(base *services.BaseService) *AgentService {
	return &AgentService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	//	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		//		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

// Fix for IntelijIdea inpsections. Cause it can't investigate anonymous method results =(
func (s *AgentService) Manager() *manager.Manager {
	return s.BaseService.Manager()
}

func (s *AgentService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/agents")
	ws.Doc("Manage Agents")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("").To(s.list)
	addDefaults(r)
	r.Doc("list")
	r.Operation("list")
	r.Param(ws.QueryParameter("name", "Agent name"))
	r.Param(ws.QueryParameter("type", "Agent type, one of [system]"))
	r.Writes(agent.AgentList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	r = ws.POST("").To(s.create)
	addDefaults(r)
	r.Doc("create")
	r.Operation("create")
	r.Writes(agent.Agent{})
	r.Reads(agent.Agent{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
		http.StatusConflict,
	))
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeAgent(s.get))
	addDefaults(r)
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(agent.Agent{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.TakeAgent(s.update))
	// docs
	r.Doc("update")
	r.Operation("update")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(agent.Agent{})
	r.Reads(agent.Agent{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r) //disable put

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.TakeAgent(s.delete))
	// docs
	r.Doc("delete")
	r.Operation("delete")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusNoContent,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	// actions
	r = ws.POST(fmt.Sprintf("{%s}/approve", ParamId)).To(s.TakeAgent(s.approve))
	addDefaults(r)
	r.Doc("approve")
	r.Operation("approve")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(http.StatusOK))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *AgentService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	raw := &agent.Agent{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	raw.Type = agent.System
	raw.Status = agent.Registered

	obj, err := mgr.Agents.Create(raw)
	if err != nil {
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

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *AgentService) list(req *restful.Request, resp *restful.Response) {
	mgr := s.Manager()
	defer mgr.Close()
	fltr := mgr.Agents.Fltr()

	if p := req.QueryParameter("name"); p != "" {
		fltr.Name = p
	}
	if p := req.QueryParameter("type"); p != "" {
		fltr.Type = agent.Type(p)
	}

	results, count, err := mgr.Agents.FilterBy(fltr)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &agent.AgentList{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *AgentService) get(_ *restful.Request, resp *restful.Response, pl *agent.Agent) {
	resp.WriteEntity(pl)
}

func (s *AgentService) update(req *restful.Request, resp *restful.Response, pl *agent.Agent) {
	// TODO (m0sth8): Check permissions

	raw := &agent.Agent{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	mgr := s.Manager()
	defer mgr.Close()

	raw.Id = pl.Id

	if err := s.updateAgent(resp, raw); err != nil {
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(raw)
}

func (s *AgentService) delete(_ *restful.Request, resp *restful.Response, obj *agent.Agent) {
	// TODO (m0sth8): Check permissions

	mgr := s.Manager()
	defer mgr.Close()

	mgr.Agents.Remove(obj)
	resp.WriteHeader(http.StatusNoContent)
}

func (s *AgentService) approve(_ *restful.Request, resp *restful.Response, ag *agent.Agent) {
	// TODO (m0sth8): Check permissions

	if ag.Status == agent.Registered {
		ag.Status = agent.Approved
		s.updateAgent(resp, ag)
	}
	return
}

// helpers

func (s *AgentService) updateAgent(resp *restful.Response, ag *agent.Agent) error {
	mgr := s.Manager()
	defer mgr.Close()

	if err := mgr.Agents.Update(ag); err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return err
		}
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.DuplicateErr)
			return err
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return err
	}
	return nil
}

func (s *AgentService) TakeAgent(fn func(*restful.Request,
	*restful.Response, *agent.Agent)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Agents.GetById(mgr.ToId(id))
		mgr.Close()
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteErrorString(http.StatusNotFound, "Not found")
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}
		fn(req, resp, obj)
	}
}
