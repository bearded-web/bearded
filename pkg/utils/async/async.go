package async

import (
	"strings"

	"golang.org/x/net/context"
)

type Async interface {
	Cancel()
	Result() <-chan error
}

type asyncImpl struct {
	promise <-chan error
	cancel  func()
}

func (a *asyncImpl) Cancel() {
	a.cancel()
}

func (a *asyncImpl) Result() <-chan error {
	return a.promise
}

func New(parent context.Context, f func(ctx context.Context) error) Async {
	ctx, cancel := context.WithCancel(parent)
	return &asyncImpl{
		promise: Promise(func() error {
			return f(ctx)
		}),
		cancel: cancel,
	}
}

type GroupError struct {
	Errs []error
}

func (e *GroupError) Error() string {
	v := make([]string, 0, len(e.Errs))
	for _, err := range e.Errs {
		v = append(v, err.Error())
	}
	return strings.Join(v, "; ")
}

// group asyncs and return err in Result for the first async with error
func Any(asyncs ...Async) Async {
	return &asyncImpl{
		cancel: func() {
			for _, a := range asyncs {
				a.Cancel()
			}
		},
		promise: Promise(func() error {
			// TODO (m0sth8): rewrite this with select for every async
			for _, a := range asyncs {
				err := <-a.Result()
				if err != nil {
					return err
				}
			}
			return nil
		}),
	}
}

// group asyncs and return grouped err after all async is finished
func All(asyncs ...Async) Async {
	return &asyncImpl{
		cancel: func() {
			for _, a := range asyncs {
				a.Cancel()
			}
		},
		promise: Promise(func() error {
			errs := make([]error, 0)
			for _, a := range asyncs {
				err := <-a.Result()
				if err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) > 0 {
				return &GroupError{Errs: errs}
			}
			return nil
		}),
	}
}
