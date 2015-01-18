package services

import (
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/passlib"
	restful "github.com/emicklei/go-restful"
)

type BaseService struct {
	manager *manager.Manager
	passCtx *passlib.Context
}

func New(mgr *manager.Manager, passCtx *passlib.Context) *BaseService {
	return &BaseService{
		manager: mgr,
		passCtx: passCtx,
	}
}

// Get copy of the manager, don't forget to close it
func (s *BaseService) Manager() *manager.Manager {
	return s.manager.Copy()
}

// Get the original manager, don't close it
func (s *BaseService) BaseManager() *manager.Manager {
	return s.manager
}

func (s *BaseService) PassCtx() *passlib.Context {
	return s.passCtx
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
