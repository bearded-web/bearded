package manager

import (
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/user"
)

type PermissionManager struct {
	manager *Manager
}

func (m *PermissionManager) Init() error {
	return nil
}

func (m *PermissionManager) HasProjectAccess(p *project.Project, u *user.User) bool {
	admin := false
	return !(!admin && p.Owner != u.Id && p.GetMember(u.Id) == nil)
}
