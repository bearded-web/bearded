package manager

import (
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/user"
	"gopkg.in/fatih/set.v0"
)

type PermissionManager struct {
	manager *Manager

	admins set.Interface
}

func (m *PermissionManager) Init() error {
	m.SetAdmins(nil)
	return nil
}

func (m *PermissionManager) HasProjectAccess(p *project.Project, u *user.User) bool {
	admin := m.IsAdmin(u)
	return !(!admin && p.Owner != u.Id && p.GetMember(u.Id) == nil)
}

func (m *PermissionManager) IsAdmin(u *user.User) bool {
	return m.IsAdminEmail(u.Email)
}

func (m *PermissionManager) IsAdminEmail(email string) bool {
	if m.admins == nil {
		return false
	}
	return m.admins.Has(email)
}

func (m *PermissionManager) SetAdmins(emails []string) {
	m.admins = set.New(AgentEmail)
	for _, email := range emails {
		m.admins.Add(email)
	}
}

func (m *PermissionManager) Copy(new *PermissionManager) {
	new.admins = m.admins
}
