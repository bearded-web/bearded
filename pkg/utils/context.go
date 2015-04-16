package utils

import (
	"time"

	"golang.org/x/net/context"
)

func JustTimeout(parent context.Context, duration time.Duration) context.Context {
	ctx, _ := context.WithTimeout(parent, duration)
	return ctx
}
