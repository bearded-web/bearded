package services

import (
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/template"
	"github.com/emicklei/go-restful"
)

type BaseService struct {
	manager   *manager.Manager
	passCtx   *passlib.Context
	scheduler scheduler.Scheduler
	mailer    email.Mailer
	apiCfg    config.Api
	Template  template.Renderer
}

func New(mgr *manager.Manager, passCtx *passlib.Context,
	sch scheduler.Scheduler, mailer email.Mailer, cfg config.Api) *BaseService {

	return &BaseService{
		manager:   mgr,
		passCtx:   passCtx,
		scheduler: sch,
		mailer:    mailer,
		apiCfg:    cfg,
	}
}

// Get copy of the manager, don't forget to close it
func (s *BaseService) Manager() *manager.Manager {
	return s.manager.Copy()
}

// Get the original manager, don't close it!
func (s *BaseService) BaseManager() *manager.Manager {
	return s.manager
}

func (s *BaseService) PassCtx() *passlib.Context {
	return s.passCtx
}

func (s *BaseService) Scheduler() scheduler.Scheduler {
	return s.scheduler
}

func (s *BaseService) Mailer() email.Mailer {
	return s.mailer
}

func (s *BaseService) ApiCfg() config.Api {
	return s.apiCfg
}

func (s *BaseService) Init() error {
	return nil
}

func (s *BaseService) Register(*restful.Container) {

}

// Check if id is in right format without making a copy of manager
func (s *BaseService) IsId(id string) bool {
	return s.manager.IsId(id)
}

// set multiple params to route
func (s *BaseService) SetParams(r *restful.RouteBuilder, params []*restful.Parameter) {
	for _, p := range params {
		r.Param(p)
	}
}
