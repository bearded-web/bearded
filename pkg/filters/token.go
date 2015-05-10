package filters

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/emicklei/go-restful"
)

func getUserByToken(mgr *manager.Manager, authorization string) *user.User {
	if authorization == "" {
		return nil
	}
	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}

	tokenHash := parts[1]

	mgrCopy := mgr.Copy()
	defer mgrCopy.Close()

	token, err := mgrCopy.Tokens.GetByHash(tokenHash)
	if err != nil {
		if !mgrCopy.IsNotFound(err) {
			logrus.Error(err)
		}
		return nil
	}
	u, err := mgrCopy.Users.GetById(token.User)
	if err != nil {
		if !mgrCopy.IsNotFound(err) {
			logrus.Error(err)
		}
		return nil
	}
	return u
}

func AuthTokenFilter(mgr *manager.Manager) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		if u := getUserByToken(mgr, req.Request.Header.Get("Authorization")); u != nil {
			req.SetAttribute(AttrUserKey, u)
		}
		chain.ProcessFilter(req, resp)

	}
}
