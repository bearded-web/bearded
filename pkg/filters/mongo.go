package filters

import (
	"fmt"
	restful "github.com/emicklei/go-restful"
	mgo "gopkg.in/mgo.v2"
)

var MongoKey = "__mongo"

func MongoFilter(session *mgo.Session) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		newSession := session.Copy()
		defer newSession.Close()
		req.SetAttribute(MongoKey, newSession)
		chain.ProcessFilter(req, resp)
	}
}

// Get mongo session from restful.Request attribute or panic
func GetMongo(req *restful.Request) *mgo.Session {
	m := req.Attribute(MongoKey)
	if m == nil {
		panic("GetMongo attribute is nil")
	}
	session, ok := m.(*mgo.Session)
	if !ok {
		panic(fmt.Sprintf("GetMongo attribute isn't a session type, but %#v", m))
	}
	return session
}
