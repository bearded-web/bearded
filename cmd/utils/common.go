package utils

import (
	//	"github.com/codegangpsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed

	"github.com/bearded-web/bearded/pkg/client"
)

func takeApi(fn func(*cli.Context, *client.Client)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		api := client.NewClient(ctx.String("api-addr"), nil)
		fn(ctx, api)
	}
}
