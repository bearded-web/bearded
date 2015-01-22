package utils

import (
	"code.google.com/p/go.net/context"
	"time"
)

func JustTimeout(parent context.Context, duration time.Duration) context.Context {
	ctx, _ := context.WithTimeout(parent, duration)
	return ctx
}
