package utils

import (
	"time"

	"github.com/facebookgo/stackerr"
	"golang.org/x/net/context"
)

func JustTimeout(parent context.Context, duration time.Duration) context.Context {
	ctx, _ := context.WithTimeout(parent, duration)
	return ctx
}

func IsCanceled(err error) bool {
	if err == context.Canceled {
		return true
	}
	if sErr, casted := err.(*stackerr.Error); casted {
		if sErr.Underlying() == context.Canceled {
			return true
		}
	}
	return false
}
