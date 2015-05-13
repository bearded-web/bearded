package services

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stackerr"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/manager"
)

func HasProjectIdPermission(mgr *manager.Manager, u *user.User,
	projectId bson.ObjectId) (bool, *ErrResp) {

	p, err := mgr.Projects.GetById(projectId)
	if err != nil {
		if mgr.IsNotFound(err) {
			return false, &ErrResp{Code: http.StatusBadRequest, Err: NewBadReq("Project not found")}
		}
		logrus.Error(stackerr.Wrap(err))
		return false, &ErrResp{Code: http.StatusInternalServerError, Err: DbErr}
	}

	return HasProjectPermission(mgr, u, p)
}

func HasProjectPermission(mgr *manager.Manager, u *user.User,
	p *project.Project) (bool, *ErrResp) {

	//	current user should have a permission to create issue there
	if !mgr.Permission.HasProjectAccess(p, u) {
		logrus.Warnf("User %s try to access to project %s", u, p)
		return false, nil
	}
	return true, nil
}

func Must(ok bool, sErr *ErrResp) *ErrResp {
	if sErr != nil {
		return sErr
	}
	if !ok {
		return &ErrResp{Code: http.StatusForbidden, Err: AuthForbidErr}
	}
	return nil
}
