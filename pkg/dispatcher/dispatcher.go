package dispatcher

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/emicklei/go-restful"
	"gopkg.in/mgo.v2"

	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/template"
	"github.com/bearded-web/bearded/services"
	"github.com/bearded-web/bearded/services/agent"
	"github.com/bearded-web/bearded/services/auth"
	configService "github.com/bearded-web/bearded/services/config"
	"github.com/bearded-web/bearded/services/feed"
	"github.com/bearded-web/bearded/services/file"
	"github.com/bearded-web/bearded/services/issue"
	"github.com/bearded-web/bearded/services/me"
	"github.com/bearded-web/bearded/services/plan"
	"github.com/bearded-web/bearded/services/plugin"
	"github.com/bearded-web/bearded/services/project"
	"github.com/bearded-web/bearded/services/scan"
	"github.com/bearded-web/bearded/services/target"
	"github.com/bearded-web/bearded/services/token"
	"github.com/bearded-web/bearded/services/user"
	"github.com/bearded-web/bearded/services/vulndb"
)

func initServices(wsContainer *restful.Container, cfg *config.Dispatcher,
	mgr *manager.Manager, mailer email.Mailer, tmpl *template.Template) error {

	// password manager for generation and verification passwords
	passCtx := passlib.NewContext()

	sch := scheduler.NewMemoryScheduler(mgr.Copy())

	// services
	base := services.New(mgr, passCtx, sch, mailer, cfg.Api)
	base.Template = tmpl
	all := []services.ServiceInterface{
		auth.New(base),
		plugin.New(base),
		plan.New(base),
		user.New(base),
		project.New(base),
		target.New(base),
		scan.New(base),
		me.New(base),
		agent.New(base),
		feed.New(base),
		file.New(base),
		issue.New(base),
		vulndb.New(base),
		configService.New(base),
		token.New(base),
	}

	// initialize services
	for _, s := range all {
		if err := s.Init(); err != nil {
			return err
		}
	}
	// register services in container
	for _, s := range all {
		s.Register(wsContainer)
	}

	return nil
}

type MgoLogger struct {
}

func (m *MgoLogger) Output(calldepth int, s string) error {
	logrus.Debug(s)
	return nil
}

func getManager(cfg config.Mongo) (*manager.Manager, error) {
	// initialize mongodb session
	logrus.Infof("Init mongodb on %s", cfg.Addr)
	session, err := mgo.Dial(cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to mongodb: %s", err.Error())
	}
	logrus.Infof("Successfull")
	logrus.Infof("Set mongo database %s", cfg.Database)
	mgr := manager.New(session.DB(cfg.Database))
	// Initialize db indexes
	if err := mgr.Init(); err != nil {
		return nil, fmt.Errorf("Cannot initilize models: %s", err.Error())
	}
	return mgr, nil
}

func getRestContainer(cfg *config.Dispatcher) *restful.Container {
	// Create container and initialize services
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{}) // CurlyRouter is the faster routing alternative for restful

	// setup session
	cookieOpts := &filters.CookieOpts{
		Path:     "/api/",
		HttpOnly: true,
		//		Secure: true,
	}
	// TODO (m0sth8): extract keys to configuration file
	hashKey := []byte("12345678901234567890123456789012")
	encKey := []byte("12345678901234567890123456789012")
	wsContainer.Filter(filters.SessionCookieFilter("bearded-sss", cookieOpts, hashKey, encKey))

	// Disable recovering in restful cause we recover all panics in negroni
	wsContainer.DoNotRecover(true)
	return wsContainer

}

func getNegroniApp(cfg *config.Dispatcher) *negroni.Negroni {
	// Use negroni as middleware framework.
	app := negroni.New()
	// TODO (m0sth8): create recovery with ServiceError response
	recovery := negroni.NewRecovery()

	if cfg.Debug {
		app.Use(negroni.NewLogger())
		// TODO (m0sth8): set output to logrus
		// existed middleware https://github.com/meatballhat/negroni-logrus
	} else {
		recovery.PrintStack = false // do not print stack to response
	}
	app.Use(recovery)

	// TODO (m0sth8): add secure middleware
	if !cfg.Frontend.Disable {
		logrus.Infof("Frontend served from %s directory", cfg.Frontend.Path)
		app.Use(negroni.NewStatic(http.Dir(cfg.Frontend.Path)))
	}
	return app
}

func runInternalAgent(mgr *manager.Manager, app *negroni.Negroni, cfg config.InternalAgent) {
	if !cfg.Enable {
		return
	}
	if tkn, err := getAgentToken(mgr); err != nil {
		logrus.Errorf("Can't get agent token: %s", err)
	} else {
		if err := RunInternalAgent(app, tkn, &cfg.Agent); err != nil {
			logrus.Errorf("Can't run agent: %s", err)
		}
	}
}

func Serve(cfg *config.Dispatcher) error {
	if cfg.Debug {
		logrus.Info("Debug mode is enabled")
	}
	// TODO (m0sth8): validate config
	logrus.Infof("Template path: %v", cfg.Template.Path)
	tmpl := template.New(&template.Opts{Directory: cfg.Template.Path})

	mgr, err := getManager(cfg.Mongo)
	if err != nil {
		return err
	}
	defer mgr.Close()

	mgr.Permission.SetAdmins(cfg.Api.Admins)

	// initialize mailer
	mailer, err := email.New(cfg.Email)
	if err != nil {
		return fmt.Errorf("Cannot initialize mailer: %s", err.Error())
	}

	if cfg.Debug {
		mgo.SetLogger(&MgoLogger{})
		mgo.SetDebug(true)
		// see what happens inside the package restful
		// TODO (m0sth8): set output to logrus
		restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	}

	wsContainer := getRestContainer(cfg)
	// Initialize and register services in container
	err = initServices(wsContainer, cfg, mgr, mailer, tmpl)
	if err != nil {
		return fmt.Errorf("Cannot initialize services: %s", err.Error())
	}

	// Swagger should be initialized after services registration
	if cfg.Swagger.Enable {
		services.Swagger(wsContainer, cfg.Swagger)
	}

	app := getNegroniApp(cfg)
	app.UseHandler(wsContainer) // set wsContainer as main handler

	runInternalAgent(mgr, app, cfg.Agent)

	// Start negroni middleware with our restful container
	bindAddr := cfg.Api.BindAddr
	server := &http.Server{Addr: bindAddr, Handler: app}
	logrus.Infof("Listening on %s", bindAddr)
	return server.ListenAndServe()
}
