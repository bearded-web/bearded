package scan

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/report"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/fltr"
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
		http.StatusForbidden,
		http.StatusInternalServerError,
	))
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
	s.SetParams(r, fltr.GetParams(ws, manager.ScanFltr{}))
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

	r = ws.GET(fmt.Sprintf("{%s}/reports", ParamId)).To(s.TakeScan(s.reports))
	r.Doc("reports")
	r.Operation("reports")
	r.Param(ws.PathParameter(ParamId, ""))
	addDefaults(r)
	r.Writes(report.ReportList{})
	r.Do(services.Returns(http.StatusOK))
	ws.Route(r)

	s.RegisterSessions(ws)

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
		if mgr.IsNotFound(err) {
			resp.WriteServiceError(http.StatusBadRequest,
				services.NewBadReq("target not found"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	if target.Project != project.Id {
		resp.WriteServiceError(http.StatusBadRequest,
			services.NewBadReq("this target is not from this project"))
		return
	}

	planObj, err := mgr.Plans.GetById(raw.Plan)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteServiceError(http.StatusBadRequest,
				services.NewBadReq("plan not found"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	if planObj.TargetType != target.Type {
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
	now := time.Now().UTC()
	// Add session from plans workflow steps
	for _, step := range planObj.Workflow {
		plugin, err := mgr.Plugins.GetByName(step.Plugin)
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
			if formData := step.Conf.FormData; formData != "" {
				t, err := template.New("").Parse(formData)
				if err != nil {
					logrus.Error(stackerr.Wrap(err))
					resp.WriteServiceError(http.StatusInternalServerError, services.NewAppErr("Wrong form data template"))
					return
				}
				buf := bytes.NewBuffer(nil)
				err = t.Execute(buf, sc.Conf)
				if err != nil {
					logrus.Error(stackerr.Wrap(err))
					resp.WriteServiceError(http.StatusInternalServerError, services.NewAppErr("Wrong form data template"))
					return
				}
				step.Conf.FormData = buf.String()
			}
			if target := step.Conf.Target; target == "" {
				step.Conf.Target = sc.Conf.Target
			}
		} else {
			step.Conf = &plan.Conf{
				Target: sc.Conf.Target,
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
	if _, err := mgr.Feed.AddScan(obj); err != nil {
		logrus.Error(stackerr.Wrap(err))
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *ScanService) list(req *restful.Request, resp *restful.Response) {
	query, err := fltr.FromRequest(req, manager.ScanFltr{})
	if err != nil {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	results, count, err := mgr.Scans.FilterByQuery(query)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	result := &scan.ScanList{
		Meta:    pagination.Meta{Count: count},
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
	// TODO (m0sth8): Forbid changes in scan after finished status

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
	// TODO (m0sth8): Forbid to remove scan after queued status

	mgr := s.Manager()
	defer mgr.Close()

	mgr.Scans.Remove(obj)
	resp.WriteHeader(http.StatusNoContent)
}

func (s *ScanService) reports(_ *restful.Request, resp *restful.Response, sc *scan.Scan) {

	mgr := s.Manager()
	defer mgr.Close()

	results := []*report.Report{}

	results, count, err := mgr.Reports.FilterBySessions(sc.GetAllSessions())
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	reportList := report.ReportList{
		Meta:    pagination.Meta{Count: count},
		Results: results,
	}

	resp.WriteEntity(reportList)
}

// Helpers

type ScanFunction func(*restful.Request, *restful.Response, *scan.Scan)

// Decorate restful.RouteFunction. Look for scan by ParamId
// and add scan object in the end. If scan is not found then return Not Found.
// Also check project permission for current user.
func (s *ScanService) TakeScan(fn ScanFunction) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		id := req.PathParameter(ParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Scans.GetById(mgr.ToId(id))
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
