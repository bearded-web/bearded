package config

import "github.com/bearded-web/bearded/pkg/utils"

type Dispatcher struct {
	Debug bool `flag:"-"`

	Frontend Frontend
	Agent    InternalAgent
	Worker   InternalWorker
	Swagger  Swagger
	Mongo    Mongo
	Email    Email
	Api      Api
	//	Log      Log
	Template Template
}

type Template struct {
	Path string `desc:"path to template files"`
}

type Api struct {
	BindAddr string   `desc:"http address for binding api server"`
	Host     string   `desc:"host for website, required for building urls"`
	Admins   []string `desc:"email list of users with admin permissions"`

	ResetPasswordSecret   string `flag:"-" desc:"secret required for reset token generation"`
	ResetPasswordDuration int    `desc:"lifetime for reset token in seconds"`

	SystemEmail  string `desc:"for sending system emails, like password reseting"`
	ContactEmail string `desc:"for show in templates, like contact with us"`

	Raven  string `desc:"sentry addr for frontend logging"`
	GA     string `desc:"google analytics id"`
	Signup Signup
	Cookie Cookie
}

type Cookie struct {
	Name     string   `desc:"name for secure cookie"`
	KeyPairs []string `desc:"key pairs for cookie"` // read more http://www.gorillatoolkit.org/pkg/securecookie
	Secure   bool     `desc:"set cookie only for https"`
}

type Signup struct {
	Disable bool `desc:"disable signup"`
}

type Frontend struct {
	Disable bool   `desc:"disable serving frontend files"`
	Path    string `desc:"path to frontend to serve static"`
}

type InternalAgent struct {
	Enable bool `desc:"run agent inside the dispatcher" env:"-"`
	Agent
}

type Agent struct {
	Name string `desc:"Unique agent name, set to fqdn if empty"`
}

type Worker struct {
	Broker          string
	ResultBackend   string
	ResultsExpireIn int
	Exchange        string
	ExchangeType    string
	DefaultQueue    string
	BindingKey      string
}

type InternalWorker struct {
	Enable bool `desc:"run worker inside the dispatcher" env:"-"`
	Worker
}

type Email struct {
	Backend string `desc:"one of: [console|smtp]"`
	Smtp    Smtp
}

type Smtp struct {
	Addr     string `desc:"smtp server addr"`
	Port     int    `desc:"smpt server port"`
	User     string `desc:"username"`
	Password string `desc:"password"`
}

type Swagger struct {
	Enable   bool   `desc:"enable swagger api"`
	Path     string `desc:"path where the swagger UI will be served, e.g. /swagger"`
	ApiPath  string `desc:"path where the JSON api is avaiable , e.g. /apidocs"`
	FilePath string `desc:"path to Swagger index.html"`
}

type Mongo struct {
	// The seed servers must be provided in the following format:
	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options].
	// Read more: http://docs.mongodb.org/manual/reference/connection-string/
	Addr     string `desc:"[mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]"`
	Database string `desc:"database name"`
}

//type Log struct {
//
//}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		Debug: false,
		Api: Api{
			BindAddr:              "127.0.0.1:3003",
			Host:                  "http://127.0.0.1:3003",
			ResetPasswordSecret:   utils.RandomString(32),
			ResetPasswordDuration: 86400,
			SystemEmail:           "admin@localhost",
			ContactEmail:          "admin@localhost",
			Cookie: Cookie{
				Name:     "bearded-sss",
				KeyPairs: []string{utils.RandomString(16), utils.RandomString(16)},
			},
		},
		Swagger: Swagger{
			ApiPath:  "/apidocs.json",
			Path:     "/swagger/",
			FilePath: "./extra/swagger-ui/dist",
		},
		Mongo: Mongo{
			Addr:     "127.0.0.1",
			Database: "bearded",
		},
		Email: Email{
			Backend: "console",
			Smtp: Smtp{
				Addr: "127.0.0.1",
				Port: 587,
			},
		},
		Template: Template{
			Path: "./extra/templates",
		},
	}
}

func NewAgent() *Agent {
	return &Agent{}
}
