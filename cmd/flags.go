package cmd

import (
	//	"github.com/codegangsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
)

func swaggerFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:   "swagger-disable",
			EnvVar: "BEARDED_SWAGGER_DISABLE",
			Usage:  "Set to disable swagger api",
		},
		cli.StringFlag{
			Name:   "swagger-api-path",
			Value:  "/apidocs.json",
			EnvVar: "BEARDED_SWAGGER_API_PATH",
			Usage:  "path where the JSON api is avaiable , e.g. /apidocs",
		},
		cli.StringFlag{
			Name:   "swagger-path",
			Value:  "/apidocs/",
			EnvVar: "BEARDED_SWAGGER_PATH",
			Usage:  "path where the swagger UI will be served, e.g. /swagger",
		},
		cli.StringFlag{
			Name:   "swagger-filepath",
			Value:  "./extra/swagger-ui/dist",
			EnvVar: "BEARDED_SWAGGER_FILEPATH",
			Usage:  "location of folder containing Swagger HTML5 application index.html",
		},
	}
}
