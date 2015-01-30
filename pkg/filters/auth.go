package filters

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/services"
)

const (
	SessionUserKey = "__user"
	AttrUserKey    = "__user"
)

func AuthRequiredFilter(mgr *manager.Manager) restful.FilterFunction {
	// TODO (m0sth8): It's not a good solution to make db request on every http request. Fix it.
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		if req.Attribute(AttrUserKey) != nil {
			// user is already set in attributes
			chain.ProcessFilter(req, resp)
			return
		}
		session := GetSession(req)
		userId, existed := session.Get(SessionUserKey)
		if !existed {
			resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
			return
		}
		mgrCopy := mgr.Copy()
		defer mgrCopy.Close() // if something goes wrong
		user, err := mgr.Users.GetById(userId)
		mgrCopy.Close() // manually close manager here, because defer will be triggered too late

		if err != nil {
			if mgr.IsNotFound(err) {
				// it seem's like this user was deleted, so logout him forcibly
				session.Del(SessionUserKey)
				resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
				return
			}
			logrus.Error(stackerr.Wrap(err))
			resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
			return
		}
		// save user to restful attributes
		req.SetAttribute(AttrUserKey, user)
		chain.ProcessFilter(req, resp)
	}
}

// Get user from restful.Request attribute or panic
func GetUser(req *restful.Request) *user.User {
	raw := req.Attribute(AttrUserKey)
	if raw == nil {
		panic("GetUser attribute is nil")
	}
	u, ok := raw.(*user.User)
	if !ok {
		panic(fmt.Sprintf("GetUser attribute isn't a user type, but %#v", raw))
	}
	return u
}
