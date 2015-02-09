package utils

import (
	"time"

	"code.google.com/p/go.net/context"
)

func JustTimeout(parent context.Context, duration time.Duration) context.Context {
	ctx, _ := context.WithTimeout(parent, duration)
	return ctx
}
