package filters

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	restful "github.com/emicklei/go-restful"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	"sync"
)

var SessionKey = "__session"

type Session struct {
	store    map[string]string
	modified bool
	m        sync.RWMutex
}

func NewSession() *Session {
	return &Session{
		store: map[string]string{},
	}
}

func (s *Session) Set(key, val string) {
	s.m.Lock()
	s.modified = true
	s.store[key] = val
	s.m.Unlock()
}

func (s *Session) Get(key string) (string, bool) {
	s.m.RLock()
	val, ok := s.store[key]
	s.m.RUnlock()
	return val, ok
}

func (s *Session) Del(key string) {
	if _, ok := s.Get(key); ok {
		s.m.Lock()
		s.modified = true
		delete(s.store, key)
		s.m.Unlock()
	}
}

func (s *Session) IsModified() bool {
	return s.modified
}

// Options stores configuration for a session or session store.
//
// Fields are a subset of http.Cookie fields.
type CookieOpts struct {
	Path   string
	Domain string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

func SessionCookieFilter(cookieName string, opts *CookieOpts, keyPairs ...[]byte) restful.FilterFunction {
	codecs := securecookie.CodecsFromPairs(keyPairs...)

	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

		session := NewSession()
		if cookie, err := req.Request.Cookie(cookieName); err == nil {
			if err = securecookie.DecodeMulti(cookieName, cookie.Value, &session.store, codecs...); err == nil {

			} else {
				logrus.Warn(err)
			}
		} else {
			if err != http.ErrNoCookie {
				logrus.Warn(err)
			}
		}
		req.SetAttribute(SessionKey, session)

		// I don't know how to write cookie in restful, so I use underneath negroni before hook
		resp.ResponseWriter.(negroni.ResponseWriter).Before(func(rw negroni.ResponseWriter) {
			if !session.IsModified() {
				return
			}
			if encoded, err := securecookie.EncodeMulti(cookieName, session.store, codecs...); err == nil {
				cookie := NewCookie(cookieName, encoded, opts)
				http.SetCookie(rw, cookie)
			}
		})

		chain.ProcessFilter(req, resp)
	}
}

// Get session from restful.Request attribute or panic
func GetSession(req *restful.Request) *Session {
	m := req.Attribute(SessionKey)
	if m == nil {
		panic("GetSession attribute is nil")
	}
	session, ok := m.(*Session)
	if !ok {
		panic(fmt.Sprintf("GetSession attribute isn't a session type, but %#v", m))
	}
	return session
}

// NewCookie returns an http.Cookie with the options set. It also sets
// the Expires field calculated based on the MaxAge value, for Internet
// Explorer compatibility.
func NewCookie(name, value string, options *CookieOpts) *http.Cookie {
	if options == nil {
		options = &CookieOpts{}
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
	if options.MaxAge > 0 {
		d := time.Duration(options.MaxAge) * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if options.MaxAge < 0 {
		// Set it to the past to expire now.
		cookie.Expires = time.Unix(1, 0)
	}
	return cookie
}
