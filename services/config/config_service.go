package config

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/bearded-web/bearded/services"
)

type ConfigService struct {
	*services.BaseService
}

func New(base *services.BaseService) *ConfigService {
	return &ConfigService{
		BaseService: base,
	}
}

func (s *ConfigService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/config")
	ws.Doc("Configuration options for frontend")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.GET("").To(s.get)
	r.Doc("get")
	r.Operation("get")
	r.Writes(ConfigEntity{})
	r.Do(services.Returns(
		http.StatusOK))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *ConfigService) get(_ *restful.Request, resp *restful.Response) {
	cfg := s.ApiCfg()
	ent := &ConfigEntity{}
	if cfg.Raven != "" {
		ent.Raven.Enable = true
		ent.Raven.Address = cfg.Raven
	}
	if cfg.GA != "" {
		ent.GA.Enable = true
		ent.GA.Id = cfg.GA
	}
	resp.WriteEntity(ent)
}
