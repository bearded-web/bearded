package services

import (
	restful "github.com/emicklei/go-restful"
	"net/http"
)

func Returns500(b *restful.RouteBuilder) {
	ReturnsE(http.StatusInternalServerError)(b)
}

func Returns200(b *restful.RouteBuilder) {
	Returns(http.StatusOK)(b)
}

func Returns201(b *restful.RouteBuilder) {
	Returns(http.StatusCreated)(b)
}

func Returns204(b *restful.RouteBuilder) {
	Returns(http.StatusNoContent)(b)
}

func Returns400(b *restful.RouteBuilder) {
	ReturnsE(http.StatusBadRequest)(b)
}

func Returns404(b *restful.RouteBuilder) {
	Returns(http.StatusNotFound)(b)
}

func Returns409(b *restful.RouteBuilder) {
	ReturnsE(http.StatusConflict)(b)
}

func Returns(codes ...int) func(*restful.RouteBuilder) {
	return func(b *restful.RouteBuilder) {
		for _, code := range codes {
			b.Returns(code, http.StatusText(code), nil)
		}
	}
}

func ReturnsE(codes ...int) func(*restful.RouteBuilder) {
	return func(b *restful.RouteBuilder) {
		for _, code := range codes {
			b.Returns(code, http.StatusText(code), restful.ServiceError{})
		}
	}
}
