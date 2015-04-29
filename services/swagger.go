package services

import (
	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
)

func init() {
	swagger.LogInfo = logrus.Infof
}

func Swagger(wsContainer *restful.Container, cfg config.Swagger) {
	config := swagger.Config{
		WebServices: wsContainer.RegisteredWebServices(), // you control what services are visible
		ApiPath:     cfg.ApiPath,

		// Optionally, specifiy where the UI is located
		SwaggerPath:     cfg.Path,
		SwaggerFilePath: cfg.FilePath,
	}
	swagger.RegisterSwaggerService(config, wsContainer)
}
