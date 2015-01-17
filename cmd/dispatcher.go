package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
	"github.com/codegangsta/negroni"
	restful "github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"

	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/services"
	"github.com/bearded-web/bearded/services/auth"
	"github.com/bearded-web/bearded/services/plugin"
	"github.com/bearded-web/bearded/services/user"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/services/me"
)

var Dispatcher = cli.Command{
	Name:   "dispatcher",
	Usage:  "Start Dispatcher",
	Action: dispatcherAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "bind-addr",
			Value:  "127.0.0.1:3003",
			EnvVar: "BEARDED_BIND_ADDR",
			Usage:  "http address for binding api server",
		},
		cli.StringFlag{
			Name:   "mongo-addr",
			Value:  "127.0.0.1",
			EnvVar: "BEARDED_MONGO_ADDR",
			Usage:  MongoUsage,
		},
		cli.StringFlag{
			Name:   "mongo-db",
			Value:  "bearded",
			EnvVar: "BEARDED_MONGO_DB",
			Usage:  "Mongodb database",
		},
		cli.StringFlag{
			Name:	"frontend",
			Value:  "../front/dist/",
			EnvVar: "BEARDED_FRONTEND",
			Usage:  "path to frontend to serve static",
		},
		cli.BoolFlag{
			Name:	"frontend-off",
			EnvVar: "BEARDED_FRONTEND_OFF",
			Usage:	"do not serve frontend files",
		},
	},
}

func init() {
	Dispatcher.Flags = append(Dispatcher.Flags, swaggerFlags()...)
}

func initServices(wsContainer *restful.Container, db *mgo.Database) error {
	// collections
	users := db.C("users")
	plugins := db.C("plugins")

	// password manager for generation and verification passwords
	passCtx := passlib.NewContext()

	// services
	authService := auth.New(users, passCtx)
	pluginService := plugin.New(plugins)
	userService := user.New(users, passCtx)
	meService := me.New(users, passCtx)

	// initialize services: set up indexes
	authService.Init()
	userService.Init()
	pluginService.Init()
	meService.Init()

	// register services in container
	authService.Register(wsContainer)
	userService.Register(wsContainer)
	pluginService.Register(wsContainer)
	meService.Register(wsContainer)

	return nil
}

func dispatcherAction(ctx *cli.Context) {

	// initialize mongodb session
	mongoAddr := ctx.String("mongo-addr")
	logrus.Infof("Init mongodb on %s", mongoAddr)
	session, err := mgo.Dial(mongoAddr)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	logrus.Infof("Successfull")
	dbName := ctx.String("mongo-db")
	logrus.Infof("Set mongo database %s", dbName)

	if ctx.GlobalBool("debug") {
		// see what happens inside the package restful
		// TODO (m0sth8): set output to logrus
		restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	}

	// Create container and initialize services
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})        // CurlyRouter is the faster routing alternative for restful

	// setup session
	cookieOpts := &filters.CookieOpts{
		Path: "/api/",
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
	initServices(wsContainer, session.DB(dbName))

	// Swagger should be initialized after services registration
	if !ctx.Bool("swagger-disabled") {
		services.Swagger(wsContainer,
			ctx.String("swagger-api-path"),
			ctx.String("swagger-path"),
			ctx.String("swagger-filepath"))
	}

	// We user negroni as middleware framework.
	app := negroni.New()
	recovery := negroni.NewRecovery() // TODO (m0sth8): create recovery with ServiceError response

	if ctx.GlobalBool("debug") {
		app.Use(negroni.NewLogger())
		// TODO (m0sth8): set output to logrus
		// existed middleware https://github.com/meatballhat/negroni-logrus
	} else {
		recovery.PrintStack = false // do not print stack to response
	}
	app.Use(recovery)

	// TODO (m0sth8): add secure middleware

	if !ctx.Bool("frontend-off") {
		logrus.Infof("Frontend served from %s directory", ctx.String("frontend"))
		app.Use(negroni.NewStatic(http.Dir(ctx.String("frontend"))))
	}

	app.UseHandler(wsContainer) // set wsContainer as main handler

	// Start negroini middleware with our restful container
	bindAddr := ctx.String("bind-addr")
	server := &http.Server{Addr: bindAddr, Handler: app}
	logrus.Infof("Listening on %s", bindAddr)
	logrus.Fatal(server.ListenAndServe())

}
