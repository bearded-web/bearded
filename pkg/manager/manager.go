package manager

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ManagerInterface interface {
	Init() error
}

type Manager struct {
	db *mgo.Database

	Users    *UserManager
	Plugins  *PluginManager
	Projects *ProjectManager
	Targets  *TargetManager

	managers []ManagerInterface
}

func New(db *mgo.Database) *Manager {
	m := &Manager{
		db:       db,
		managers: []ManagerInterface{},
	}

	// initialize different managers
	m.Users = &UserManager{manager: m, col: db.C("users")}
	m.Plugins = &PluginManager{manager: m, col: db.C("plugins")}
	m.Projects = &ProjectManager{manager: m, col: db.C("projects")}
	m.Targets = &TargetManager{manager: m, col: db.C("targets")}

	m.managers = append(m.managers,
		m.Users,
		m.Plugins,
		m.Projects,
		m.Targets,
	)

	return m
}

// Initialize all managers. Ensure indexes.
func (m *Manager) Init() error {
	for _, manager := range m.managers {
		if err := manager.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Get copy of manager with copied session, don't forget to call Close after
func (m *Manager) Copy() *Manager {
	sess := m.db.Session.Copy()
	return New(m.db.With(sess))
}

// Clone works just like Copy, but also reuses the same socket as the original
// session, in case it had already reserved one due to its consistency
// guarantees.  This behavior ensures that writes performed in the old session
// are necessarily observed when using the new session, as long as it was a
// strong or monotonic session.  That said, it also means that long operations
// may cause other goroutines using the original session to wait.
func (m *Manager) Clone() *Manager {
	sess := m.db.Session.Clone()
	return New(m.db.With(sess))
}

// Close terminates the session.  It's a runtime error to use a session
// after it has been closed.
func (m *Manager) Close() {
	m.db.Session.Close()
}

// Different methods which help to hide all database things

// Return true if object is not found
func (m *Manager) IsNotFound(err error) bool {
	return err == mgo.ErrNotFound
}

// IsDup returns whether err informs of a duplicate key error because
// a primary key index or a secondary unique index already has an entry
// with the given value.
func (m *Manager) IsDup(err error) bool {
	return mgo.IsDup(err)
}

// IsId returns whether id is in a valid form (bson ObjectId hex format)
func (m *Manager) IsId(id string) bool {
	return bson.IsObjectIdHex(id)
}

func (m *Manager) ToId(id string) bson.ObjectId {
	return bson.ObjectIdHex(id)
}

func (m *Manager) FromId(id bson.ObjectId) string {
	return id.Hex()
}
