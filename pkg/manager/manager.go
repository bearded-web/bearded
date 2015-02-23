package manager

import (
	"time"

	"gopkg.in/mgo.v2"
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
	Plans    *PlanManager
	Scans    *ScanManager
	Agents   *AgentManager
	Reports  *ReportManager
	Feed     *FeedManager
	Files    *FileManager
	Comments *CommentManager

	managers []ManagerInterface
}

// Manager contains all available managers for different models
// and hidden all db related operations
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
	m.Plans = &PlanManager{manager: m, col: db.C("plans")}
	m.Scans = &ScanManager{manager: m, col: db.C("scans")}
	m.Agents = &AgentManager{manager: m, col: db.C("agents")}
	m.Reports = &ReportManager{manager: m, col: db.C("reports")}
	m.Feed = &FeedManager{manager: m, col: db.C("feed")}
	m.Files = &FileManager{manager: m, grid: db.GridFS("")}
	m.Comments = &CommentManager{manager: m, col: db.C("comments")}

	m.managers = append(m.managers,
		m.Users,
		m.Plugins,
		m.Projects,
		m.Targets,
		m.Plans,
		m.Scans,
		m.Agents,
		m.Reports,
		m.Feed,
		m.Files,
		m.Comments,
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

// convert string id to ObjectId

func (m *Manager) ToId(id string) bson.ObjectId {
	return ToId(id)
}

// convert ObjectId to string
func (m *Manager) FromId(id bson.ObjectId) string {
	return FromId(id)
}

func (m *Manager) All(col *mgo.Collection, results interface{}) (int, error) {
	return m.FilterBy(col, &bson.M{}, results)
}

func (m *Manager) GetById(col *mgo.Collection, id bson.ObjectId, result interface{}) error {
	return col.FindId(id).One(result)
}

func (m *Manager) GetBy(col *mgo.Collection, query *bson.M, result interface{}, opts ...Opts) error {
	q := col.Find(query)
	for _, opt := range opts {
		if opt.Limit != 0 {
			q.Limit(opt.Limit)
		}
		if opt.Skip != 0 {
			q.Skip(opt.Skip)
		}
		if opt.Sort != nil {
			q.Sort(opt.Sort...)
		}
	}
	return q.One(result)
}

func (m *Manager) NewId() bson.ObjectId {
	return bson.NewObjectId()
}

type Opts struct {
	Limit int
	Skip  int
	Sort  []string
}

func (m *Manager) FilterBy(col *mgo.Collection, query *bson.M, results interface{}, opts ...Opts) (int, error) {
	q := col.Find(query)
	for _, opt := range opts {
		if opt.Limit != 0 {
			q.Limit(opt.Limit)
		}
		if opt.Skip != 0 {
			q.Skip(opt.Skip)
		}
		if opt.Sort != nil {
			q.Sort(opt.Sort...)
		}
	}
	if err := q.All(results); err != nil {
		return 0, err
	}
	q.Limit(0)
	q.Skip(0)
	count, err := q.Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *Manager) FilterAndSortBy(col *mgo.Collection, query *bson.M, sort []string, results interface{}) (int, error) {
	q := col.Find(query)
	if sort != nil && len(sort) > 0 {
		q.Sort(sort...)
	}
	if err := q.All(results); err != nil {
		return 0, err
	}
	count, err := q.Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *Manager) Opts(skip, limit int, sort []string) Opts {
	return GetOpts(skip, limit, sort)
}

// helpers

func TimeP(t time.Time) *time.Time {
	return &t
}

func ToId(id string) bson.ObjectId {
	return bson.ObjectIdHex(id)
}

func FromId(id bson.ObjectId) string {
	return id.Hex()
}

func GetOpts(skip, limit int, sort []string) Opts {
	return Opts{
		Skip:  skip,
		Limit: limit,
		Sort:  sort,
	}
}

func Or(q bson.M) bson.M {
	results := make([]bson.M, 0, len(q))
	for key, value := range q {
		results = append(results, bson.M{key: value})
	}
	return bson.M{"$or": results}
}
