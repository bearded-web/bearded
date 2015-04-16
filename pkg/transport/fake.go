package transport

import "golang.org/x/net/context"

// Fake transport helps while developing or for mocking
// if you don't want to implement all Transport interface methods
type Fake struct{}

func (t *Fake) Serve(ctx context.Context, h Handler) error {
	return nil
}

func (t *Fake) Request(ctx context.Context, send, recv interface{}) error {
	return nil
}
