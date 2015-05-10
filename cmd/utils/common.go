package utils

import (
	"os"
	"time"

	//	"github.com/codegangpsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
	"golang.org/x/net/context"

	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/plugin"
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
	cli.StringFlag{
		Name:   "api-token",
		EnvVar: "BEARDED_API_TOKEN",
		Usage:  "token for communication with bearded server",
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
		if token := ctx.String("api-token"); token != "" {
			api.Token = token
		} else {
			println("Please set up api-token[$BEARDED_API_TOKEN] flag")
			os.Exit(1)
		}
		fn(ctx, api, timeout)
	}
}

type ExtraData struct {
	Plugins []*plugin.Plugin
	Plans   []*plan.Plan
}
