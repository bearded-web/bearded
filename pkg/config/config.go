package config

import "github.com/bearded-web/bearded/pkg/utils"

type Dispatcher struct {
	Debug bool `flag:"-"`

	Frontend Frontend
	Agent    Agent
	Swagger  Swagger
	Mongo    Mongo
	Email    Email
	Api      Api
	//	Log      Log
	TplsPath string `desc:path to templates files`
}

type Api struct {
	BindAddr string `desc:"http address for binding api server"`
	Host     string `desc:"host for website, required for building urls"`

	ResetPasswordSecret   string `flag:"-" desc:"secret required for reset token generation"`
	ResetPasswordDuration int    `desc:"lifetime for reset token in seconds"`

	SystemEmail string `desc:"for sending system emails, like password reseting"`
}

type Frontend struct {
	Disable bool   `desc:"disable serving frontend files"`
	Path    string `desc:"path to frontend to serve static"`
}

type Agent struct {
	Enable bool `desc:"run agent inside the dispatcher" env:"-"`
}

type Email struct {
	Backend  string `desc:"one of: [console|smtp]"`
	Smtp     Smtp
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
		TplsPath: "./extra/templates",
	}
}
