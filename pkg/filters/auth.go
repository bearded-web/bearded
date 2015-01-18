package filters

import (
	restful "github.com/emicklei/go-restful"
	"github.com/bearded-web/bearded/services"
	"net/http"
)



func AuthRequiredFilter() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain){
		session := GetSession(req)
		_, existed := session.Get("userId")
		if !existed {
			resp.WriteServiceError(http.StatusUnauthorized, services.AuthReqErr)
			return
		}
		chain.ProcessFilter(req, resp)
	}
}
