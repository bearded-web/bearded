package scan

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const (
	ParamId        = "scan-id"
	SessionParamId = "session-id"
)

type ScanService struct {
	*services.BaseService
}

func New(base *services.BaseService) *ScanService {
	return &ScanService{
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

// Fix for IntelijIdea inpsections. Cause it can't investigate anonymous method results =(
func (s *ScanService) Manager() *manager.Manager {
	return s.BaseService.Manager()
}

func (s *ScanService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/scans")
	ws.Doc("Manage Scans")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthTokenFilter(s.BaseManager()))
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.GET("").To(s.list)
	r.Doc("list")
	r.Operation("list")
	r.Param(ws.QueryParameter("target", "filter by target"))
	r.Param(ws.QueryParameter("status", "filter by status one of [created|queued|working|paused|finished|failed]"))
	addDefaults(r)
	r.Writes(scan.ScanList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	r = ws.POST("").To(s.create)
	r.Doc("create")
	r.Operation("create")
	addDefaults(r)
	r.Writes(scan.Scan{})
	r.Reads(scan.Scan{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(
		http.StatusBadRequest,
		http.StatusConflict,
	))
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeScan(s.get))
	r.Doc("get")
	r.Operation("get")
	addDefaults(r)
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(scan.Scan{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	//	r = ws.PUT(fmt.Sprintf("{%s}", ParamId)).To(s.TakeScan(s.update))
	//	r.Doc("update")
	//	r.Operation("update")
	//	r.Param(ws.PathParameter(ParamId, ""))
	//	r.Writes(scan.Scan{})
	//	r.Reads(scan.Scan{})
	//	r.Do(services.Returns(
	//		http.StatusOK,
	//		http.StatusNotFound))
	//	r.Do(services.ReturnsE(http.StatusBadRequest))
	//	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}", ParamId)).To(s.TakeScan(s.delete))
	r.Doc("delete")
	r.Operation("delete")
	addDefaults(r)
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusNoContent,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	// sessions

	r = ws.GET(fmt.Sprintf("{%s}/sessions/{%s}", ParamId, SessionParamId)).To(s.TakeScan(s.TakeSession(s.sessionGet)))
	r.Doc("sessionGet")
	r.Operation("sessionGet")
	//	addDefaults(r)
	r.Param(ws.PathParameter(ParamId, ""))
	r.Param(ws.PathParameter(SessionParamId, ""))
	r.Writes(scan.Session{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.PUT(fmt.Sprintf("{%s}/sessions/{%s}", ParamId, SessionParamId)).To(s.TakeScan(s.TakeSession(s.sessionUpdate)))
	r.Doc("sessionUpdate")
	r.Operation("sessionUpdate")
	//	addDefaults(r)
	r.Param(ws.PathParameter(ParamId, ""))
	r.Param(ws.PathParameter(SessionParamId, ""))
	r.Reads(SessionUpdateEntity{})
	r.Writes(scan.Session{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	container.Add(ws)
}

// ====== service operations

func (s *ScanService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions
	raw := &scan.Scan{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	u := filters.GetUser(req)

	mgr := s.Manager()
	defer mgr.Close()

	// TODO (m0sth8): check project and target permissions for this user

	// validations
	project, err := mgr.Projects.GetById(raw.Project)
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest,
			services.NewBadReq("project not found"))
		return
	}

	target, err := mgr.Targets.GetById(raw.Target)
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest,
			services.NewBadReq("target not found"))
		return
	}
	if target.Project != project.Id {
		resp.WriteServiceError(http.StatusBadRequest,
			services.NewBadReq("this target is not from this project"))
		return
	}

	plan, err := mgr.Plans.GetById(raw.Plan)
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest,
			services.NewBadReq("plan not found"))
		return
	}
	if plan.TargetType != target.Type {
		resp.WriteServiceError(http.StatusBadRequest,
			services.NewBadReq("target.type and plan.targetType is not compatible"))
		return
	}

	sc := &scan.Scan{
		Status:  scan.StatusCreated,
		Owner:   u.Id,
		Plan:    raw.Plan,
		Project: project.Id,
		Target:  target.Id,
		Conf: scan.ScanConf{
			Target: target.Addr(),
		},
		Sessions: []*scan.Session{},
	}
	now := time.Now()
	// Add session from plans workflow steps
	for _, step := range plan.Workflow {
		// TODO (m0sth8): Take latest plugin or search by version, extract this logic
		plNameVersion := strings.Split(step.Plugin, ":")
		plugin, err := mgr.Plugins.GetByNameVersion(plNameVersion[0], plNameVersion[1])
		if err != nil {
			if mgr.IsNotFound(err) {
				resp.WriteServiceError(http.StatusBadRequest,
					services.NewBadReq(fmt.Sprintf("plugin %s is not found", step.Plugin)))
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}
		// TODO (m0sth8): extract template execution
		if step.Conf != nil {
			if command := step.Conf.CommandArgs; command != "" {
				t, err := template.New("").Parse(command)
				if err != nil {
					logrus.Error(stackerr.Wrap(err))
					resp.WriteServiceError(http.StatusInternalServerError, services.NewAppErr("Wrong command args template"))
					return
				}
				buf := bytes.NewBuffer(nil)
				err = t.Execute(buf, sc.Conf)
				if err != nil {
					logrus.Error(stackerr.Wrap(err))
					resp.WriteServiceError(http.StatusInternalServerError, services.NewAppErr("Wrong command args template"))
					return
				}
				step.Conf.CommandArgs = buf.String()
			}
		}

		sess := scan.Session{
			Id:     mgr.NewId(),
			Step:   step,
			Plugin: plugin.Id,
			Status: scan.StatusCreated,
			Dates: scan.Dates{
				Created: &now,
				Updated: &now,
			},
		}
		sc.Sessions = append(sc.Sessions, &sess)
	}

	obj, err := mgr.Scans.Create(sc)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	// put scan to queue
	s.Scheduler().AddScan(obj)

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *ScanService) list(req *restful.Request, resp *restful.Response) {
	mgr := s.Manager()
	defer mgr.Close()

	fltr := mgr.Scans.Fltr()

	if p := req.QueryParameter("target"); p != "" {
		if !s.IsId(p) {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("target should be bson uuid in hex form"))
			return
		}
		fltr.Target = mgr.ToId(p)
	}
	if p := req.QueryParameter("status"); p != "" {
		fltr.Status = scan.ScanStatus(p)
	}

	results, count, err := mgr.Scans.FilterBy(fltr)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &scan.ScanList{
		Meta:    pagination.Meta{count, "", ""},
		Results: results,
	}
	resp.WriteEntity(result)
}

func (s *ScanService) get(_ *restful.Request, resp *restful.Response, pl *scan.Scan) {
	resp.WriteEntity(pl)
}

// disabled
func (s *ScanService) update(req *restful.Request, resp *restful.Response, pl *scan.Scan) {
	// TODO (m0sth8): Check permissions

	raw := &scan.Scan{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	mgr := s.Manager()
	defer mgr.Close()

	raw.Id = pl.Id

	if err := mgr.Scans.Update(raw); err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "scan with this name and version is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(raw)
}

func (s *ScanService) delete(_ *restful.Request, resp *restful.Response, obj *scan.Scan) {
	mgr := s.Manager()
	defer mgr.Close()

	mgr.Scans.Remove(obj)
	resp.WriteHeader(http.StatusNoContent)
}

// Sessions

func (s *ScanService) sessionGet(_ *restful.Request, resp *restful.Response, _ *scan.Scan, sess *scan.Session) {
	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(sess)
}

func (s *ScanService) sessionUpdate(req *restful.Request, resp *restful.Response, sc *scan.Scan, sess *scan.Session) {
	// TODO (m0sth8): Check permissions

	raw := &SessionUpdateEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Warn(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	if !(raw.Status == scan.StatusWorking || raw.Status == scan.StatusFinished || raw.Status == scan.StatusFailed) {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("status should be one of [working|finished|failed]"))
		return
	}


	mgr := s.Manager()
	defer mgr.Close()

	logrus.Debugf("Update session %s status from %s to %s", mgr.FromId(sess.Id), sess.Status, raw.Status)

	sess.Status = raw.Status
	if err := mgr.Scans.UpdateSession(sc, sess); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	s.Scheduler().UpdateScan(sc)

	resp.WriteEntity(sess)
}

// Helpers

type ScanFunction func(*restful.Request, *restful.Response, *scan.Scan)
type SessionFunction func(*restful.Request, *restful.Response, *scan.Scan, *scan.Session)

// Decorate ScanFunction. Look for session in scan by SessionParamId
// and add session object in the end. If session is not found then return Not Found.
func (s *ScanService) TakeSession(fn SessionFunction) ScanFunction {
	return func(req *restful.Request, resp *restful.Response, sc *scan.Scan) {
		id := req.PathParameter(SessionParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		objId := manager.ToId(id)

		for _, sess := range sc.Sessions {
			if sess.Id == objId {
				fn(req, resp, sc, sess)
				return
			}
		}
		resp.WriteErrorString(http.StatusNotFound, "Not found")
		return
	}
}

// Decorate restful.RouteFunction. Look for scan by ParamId
// and add scan object in the end. If scan is not found then return Not Found.
func (s *ScanService) TakeScan(fn ScanFunction) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Scans.GetById(id)
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
