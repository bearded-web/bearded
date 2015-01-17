package services

import (
	restful "github.com/emicklei/go-restful"
)

type ServiceInterface interface {
	Init() error
	Register(container *restful.Container)
}
