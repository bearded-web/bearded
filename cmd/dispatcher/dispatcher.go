package dispatcher

import (
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/emicklei/go-restful"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
	"gopkg.in/mgo.v2"

	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/config/flags"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/template"
	"github.com/bearded-web/bearded/pkg/utils/load"
	"github.com/bearded-web/bearded/services"
	"github.com/bearded-web/bearded/services/agent"
	"github.com/bearded-web/bearded/services/auth"
	"github.com/bearded-web/bearded/services/feed"
	"github.com/bearded-web/bearded/services/file"
	"github.com/bearded-web/bearded/services/issue"
	"github.com/bearded-web/bearded/services/me"
	"github.com/bearded-web/bearded/services/plan"
	"github.com/bearded-web/bearded/services/plugin"
	"github.com/bearded-web/bearded/services/project"
	"github.com/bearded-web/bearded/services/scan"
	"github.com/bearded-web/bearded/services/target"
	"github.com/bearded-web/bearded/services/user"
	"github.com/bearded-web/bearded/services/vulndb"
)

func New() cli.Command {
	cmd := cli.Command{
		Name:   "dispatcher",
		Usage:  "Start Dispatcher",
		Action: dispatcherAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "config",
				EnvVar: "BEARDED_CONFIG",
				Usage:  "path to config",
			},
			cli.StringFlag{
				Name:   "config-format",
				EnvVar: "BEARDED_CONFIG_FORMAT",
				Usage:  "Specify config format, by default format is taken from ext",
			},
		},
	}
	cfg := config.NewDispatcher()
	cmd.Flags = append(cmd.Flags, flags.GenerateFlags(cfg, flags.Opts{
		EnvPrefix: "BEARDED",
	})...)
	return cmd
}

func initServices(wsContainer *restful.Container, cfg *config.Dispatcher,
	db *mgo.Database, mailer email.Mailer, tmpl *template.Template) error {

	// manager
	mgr := manager.New(db)
	if err := mgr.Init(); err != nil {
		return err
	}

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

func dispatcherAction(ctx *cli.Context) {
	cfg := config.NewDispatcher()
	if cfgPath := ctx.String("config"); cfgPath != "" {
		logrus.Infof("Load config from %s", cfgPath)
		err := load.FromFile(cfgPath, cfg)
		if err != nil {
			logrus.Fatalf("Couldn't load config: %s", err)
			return
		}
	}

	logrus.Info("Load config from flags")
	err := flags.ParseFlags(cfg, ctx, flags.Opts{
		EnvPrefix: "BEARDED",
	})
	if err != nil {
		logrus.Fatal(err)
	}
	if cfg.Debug = ctx.GlobalBool("debug"); cfg.Debug {
		logrus.Info("Debug mode is enabled")
	}
	// TODO (m0sth8): validate config
	logrus.Infof("Template path: %v", cfg.Template.Path)
	tmpl := template.New(&template.Opts{Directory: cfg.Template.Path})

	// initialize mongodb session
	mongoAddr := cfg.Mongo.Addr
	logrus.Infof("Init mongodb on %s", mongoAddr)
	session, err := mgo.Dial(mongoAddr)
	if err != nil {
		logrus.Fatalf("Cannot connect to mongodb: %s", err.Error())
		return
	}
	defer session.Close()
	logrus.Infof("Successfull")
	dbName := cfg.Mongo.Database
	logrus.Infof("Set mongo database %s", dbName)

	// initialize mailer
	mailer, err := email.New(cfg.Email)
	if err != nil {
		logrus.Fatalf("Cannot initialize mailer: %s", err.Error())
		return
	}

	if cfg.Debug {
		mgo.SetLogger(&MgoLogger{})
		mgo.SetDebug(true)
		// see what happens inside the package restful
		// TODO (m0sth8): set output to logrus
		restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	}

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

	wsContainer.Filter(filters.MongoFilter(session)) // Add mongo session copy to context on every request
	wsContainer.DoNotRecover(true)                   // Disable recovering in restful cause we recover all panics in negroni

	// Initialize and register services in container
	err = initServices(wsContainer, cfg, session.DB(dbName), mailer, tmpl)
	if err != nil {
		logrus.Fatal(err)
	}

	// Swagger should be initialized after services registration
	if cfg.Swagger.Enable {
		services.Swagger(wsContainer, cfg.Swagger)
	}

	// Use negroni as middleware framework.
	app := negroni.New()
	recovery := negroni.NewRecovery() // TODO (m0sth8): create recovery with ServiceError response

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

	app.UseHandler(wsContainer) // set wsContainer as main handler

	if cfg.Agent.Enable {
		if err := RunInternalAgent(app); err != nil {
			logrus.Error(err)
		}
	}

	// Start negroni middleware with our restful container
	bindAddr := cfg.Api.BindAddr
	server := &http.Server{Addr: bindAddr, Handler: app}
	logrus.Infof("Listening on %s", bindAddr)
	logrus.Fatal(server.ListenAndServe())
}
