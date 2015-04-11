package transport

import (
	"golang.org/x/net/context"
)

type Extractor interface {
	Extract(obj interface{}) error
}

type Handler interface {
	Handle(context.Context, Extractor) (interface{}, error)
}

// Transport interface
type Transport interface {
	Serve(ctx context.Context, h Handler) error
	Request(ctx context.Context, send, recv interface{}) error
}

type handleFunc struct {
	f func(context.Context, Extractor) (interface{}, error)
}

func (h *handleFunc) Handle(c context.Context, e Extractor) (interface{}, error) {
	return h.f(c, e)
}

func Handle(f func(context.Context, Extractor) (interface{}, error)) Handler {
	return &handleFunc{f: f}
}
