package websocket

import (
	"net/http"

	"golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/gorilla/websocket"
)

type server struct {
	server            *http.Server
	certFile, keyFile string
	tls               bool
}

func NewServer(addr string) transport.Transport {
	s := server{
		server: &http.Server{Addr: addr},
	}
	return transport.NewLoopTransport(&s)
}

func NewTlsServer(addr, certFile, keyFile string) transport.Transport {
	s := server{
		certFile: certFile,
		keyFile:  keyFile,
		tls:      true,
		server:   &http.Server{Addr: addr},
	}
	return transport.NewLoopTransport(&s)
}

func (s *server) getHttpHandler(ctx context.Context,
	in chan<- *transport.Message, out <-chan *transport.Message) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{}
		ws, err := upgrader.Upgrade(resp, req, http.Header{})
		if err != nil {
			logrus.Error(err)
			http.Error(resp, "Application error", http.StatusInternalServerError)
			return
		}
		defer ws.Close()
		handleConnection(ctx, ws, in, out)
	}
}

func (s *server) Loop(ctx context.Context,
	in chan<- *transport.Message, out <-chan *transport.Message) <-chan error {
	// http server will never stop. It's ok for scripts, because they died after execution (or will be killed =) )
	// TODO (m0sth8): Implement cancel by context here (maybe after https://github.com/golang/go/issues/4674)
	ch := make(chan error, 1)
	s.server.Handler = s.getHttpHandler(ctx, in, out)

	go func(ch chan<- error) {
		var err error
		if s.tls {
			err = s.server.ListenAndServeTLS(s.certFile, s.keyFile)
		} else {
			err = s.server.ListenAndServe()
		}
		ch <- err
	}(ch)
	return ch
}
