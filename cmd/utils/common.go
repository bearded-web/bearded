package utils

import (
	"time"

	//	"github.com/codegangpsta/cli"
	"code.google.com/p/go.net/context"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed

	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/utils"
)

var ApiFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "api-addr",
		Value:  "http://127.0.0.1:3003/api/",
		EnvVar: "BEARDED_API_ADDR",
		Usage:  "http address for connection to the api server",
	},
	cli.IntFlag{
		Name:   "api-timeout",
		Value:  10,
		EnvVar: "BEARDED_API_TIMEOUT",
		Usage:  "timeout for client requests in seconds",
	},
}

type Timeout func() context.Context

func takeApi(fn func(*cli.Context, *client.Client, Timeout)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		timeout := func() context.Context {
			return utils.JustTimeout(context.Background(), time.Duration(ctx.Int("api-timeout"))*time.Second)
		}
		api := client.NewClient(ctx.String("api-addr"), nil)
		if ctx.GlobalBool("debug") {
			api.Debug = true
		}
		fn(ctx, api, timeout)
	}
}
