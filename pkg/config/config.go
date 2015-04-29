package config

type Dispatcher struct {
	Debug bool `flag:"-"`

	Frontend Frontend
	Agent    Agent
	Swagger  Swagger
	Mongo    Mongo
	Email    Email
	Api      Api
	//	Log      Log
}

type Api struct {
	BindAddr string `desc:"http address for binding api server"`
}

type Frontend struct {
	Disable bool   `desc:"disable serving frontend files"`
	Path    string `desc:"path to frontend to serve static"`
}

type Agent struct {
	Enable bool `desc:"run agent inside the dispatcher" env:"-"`
}

type Email struct {
	Smtp Smtp
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
			BindAddr: "127.0.0.1:3003",
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
			Smtp: Smtp{
				Addr: "127.0.0.1",
				Port: 587,
			},
		},
	}
}
