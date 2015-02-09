package filters

import (
	"strings"

	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/emicklei/go-restful"
)

func AuthTokenFilter(mgr *manager.Manager) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		// TODO (m0sth8): there should be an oauth token verification
		func() {
			// Temporary hack for agents, will be deleted after oauth integration
			authorization := req.Request.Header.Get("Authorization")
			if authorization == "" {
				return
			}
			parts := strings.Split(authorization, " ")
			if len(parts) != 2 || parts[0] != "token" {
				return
			}

			if parts[1] != "agent-token" {
				return
			}

			agentUser, err := mgr.Users.GetByEmail("agent@barbudo.net")
			if err != nil {
				return
			}
			req.SetAttribute(AttrUserKey, agentUser)
		}()
		chain.ProcessFilter(req, resp)

	}
}
