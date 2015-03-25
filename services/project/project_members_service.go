package project

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/pagination"
	"github.com/bearded-web/bearded/services"
)

const (
	MemberParamId = "user-id"
)

func (s *ProjectService) RegisterMembers(ws *restful.WebService) {
	r := ws.GET(fmt.Sprintf("{%s}/members", ParamId)).To(s.TakeProject(s.members))
	r.Doc("members")
	r.Operation("members")
	addDefaults(r)
	r.Writes(project.MemberList{})
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusOK,
		http.StatusNotFound))
	ws.Route(r)

	r = ws.POST(fmt.Sprintf("{%s}/members", ParamId)).To(s.TakeProject(s.membersCreate))
	r.Doc("membersCreate")
	r.Operation("membersCreate")
	addDefaults(r)
	r.Reads(project.Member{})
	r.Writes(project.Member{})
	r.Param(ws.PathParameter(ParamId, ""))
	r.Do(services.Returns(
		http.StatusCreated,
		http.StatusNotFound))
	r.Do(services.ReturnsE(http.StatusBadRequest))
	ws.Route(r)

	r = ws.DELETE(fmt.Sprintf("{%s}/members/{%s}", ParamId, MemberParamId)).To(s.TakeProject(s.TakeMember(s.membersDelete)))
	r.Doc("membersDelete")
	r.Operation("membersDelete")
	addDefaults(r)
	r.Param(ws.PathParameter(ParamId, ""))
	r.Param(ws.PathParameter(MemberParamId, ""))
	r.Do(services.Returns(
		http.StatusNoContent,
		http.StatusNotFound))
	ws.Route(r)

}

func (s *ProjectService) members(_ *restful.Request, resp *restful.Response, p *project.Project) {
	result := &project.MemberList{
		Meta:    pagination.Meta{Count: len(p.Members)},
		Results: p.Members,
	}
	resp.WriteEntity(result)
}

func (s *ProjectService) membersCreate(req *restful.Request, resp *restful.Response, p *project.Project) {
	raw := &project.Member{}
	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	if raw.User == "" {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("user is required"))
		return
	}

	u := filters.GetUser(req)
	if p.Owner != u.Id {
		resp.WriteServiceError(http.StatusForbidden, services.AuthForbidErr)
		return
	}

	for _, member := range p.Members {
		if member.User == raw.User {
			resp.WriteServiceError(http.StatusConflict, services.NewError(services.CodeDuplicate, "User is already member"))
			return
		}
	}

	mgr := s.Manager()
	defer mgr.Close()

	mUser, err := mgr.Users.GetById(raw.User)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "User not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	member := &project.Member{User: mUser.Id}
	p.Members = append(p.Members, member)

	err = mgr.Projects.Update(p)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(member)
}

func (s *ProjectService) membersDelete(req *restful.Request, resp *restful.Response, p *project.Project, m *project.Member) {
	members := make([]*project.Member, 0, len(p.Members)-1)
	for _, member := range p.Members {
		if member.User != m.User {
			members = append(members, member)
		}
	}
	p.Members = members

	mgr := s.Manager()
	defer mgr.Close()

	err := mgr.Projects.Update(p)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	resp.ResponseWriter.WriteHeader(http.StatusNoContent)
}

// Helpers

type MemberFunction func(*restful.Request, *restful.Response, *project.Project, *project.Member)

// Decorate ScanFunction. Look for session in scan by SessionParamId
// and add session object in the end. If session is not found then return Not Found.
func (s *ProjectService) TakeMember(fn MemberFunction) ProjectFunction {
	return func(req *restful.Request, resp *restful.Response, p *project.Project) {
		id := req.PathParameter(MemberParamId)
		if !s.IsId(id) {
			resp.WriteServiceError(http.StatusBadRequest, services.IdHexErr)
			return
		}

		objId := manager.ToId(id)

		member := p.GetMember(objId)
		if member == nil {
			resp.WriteErrorString(http.StatusNotFound, "Not found")
			return
		}
		fn(req, resp, p, member)
	}
}
