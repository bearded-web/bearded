package services

import (
	"github.com/emicklei/go-restful"
)

type ServiceInterface interface {
	Init() error
	Register(container *restful.Container)
}
