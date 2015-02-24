package file

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/file"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/services"
)

const ParamId = "file-id"

type FileService struct {
	*services.BaseService
}

func New(base *services.BaseService) *FileService {
	return &FileService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	r.Notes("Authorization required")
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
		http.StatusForbidden,
		http.StatusBadRequest,
	))
}

func (s *FileService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Path("/api/v1/files")
	ws.Doc("Manage Files")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(filters.AuthTokenFilter(s.BaseManager()))
	ws.Filter(filters.AuthRequiredFilter(s.BaseManager()))

	r := ws.POST("").To(s.create)
	r.Doc("create")
	r.Operation("create")
	r.Consumes("multipart/form-data")
	r.Param(ws.FormParameter("file", "file to upload").DataType("File"))
	r.Writes(file.Meta{})
	r.Do(services.Returns(http.StatusCreated))
	r.Do(services.ReturnsE(http.StatusConflict))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}", ParamId)).To(s.TakeFile(s.get))
	r.Doc("get")
	r.Operation("get")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Writes(file.Meta{})
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	ws.Route(r)

	r = ws.GET(fmt.Sprintf("{%s}/download", ParamId)).To(s.TakeFile(s.download))
	r.Doc("download")
	r.Operation("download")
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	ws.Route(r)

	container.Add(ws)
}

func (s *FileService) create(req *restful.Request, resp *restful.Response) {
	// TODO (m0sth8): Check permissions for the user, he is might be blocked or removed

	f, header, err := req.Request.FormFile("file")
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Couldn't read file"))
		return
	}

	contentType := header.Header.Get("Content-Type")

	// TODO (m0sth8): reduce filename length
	meta := &file.Meta{
		Name:        header.Filename,
		ContentType: contentType,
	}

	mgr := s.Manager()
	defer mgr.Close()

	obj, err := mgr.Files.Create(f, meta)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(obj)
}

func (s *FileService) get(_ *restful.Request, resp *restful.Response, obj *file.File) {
	resp.WriteEntity(obj.Meta)
}

func (s *FileService) download(_ *restful.Request, resp *restful.Response, obj *file.File) {
	resp.AddHeader("Content-Type", "application/octet-stream")

	if filename := obj.Meta.Name; filename != "" {
		filename = url.QueryEscape(filename)
		resp.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	io.Copy(resp.ResponseWriter, obj)
}

func (s *FileService) TakeFile(fn func(*restful.Request,
	*restful.Response, *file.File)) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		// TODO (m0sth8): Add token for file to close access to file for everyone
		id := req.PathParameter(ParamId)

		mgr := s.Manager()
		defer mgr.Close()

		obj, err := mgr.Files.GetById(id)
		defer obj.Close()
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
