package agent

import (
	"golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
	//	"github.com/codegangpsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed

	agentServer "github.com/bearded-web/bearded/pkg/agent"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/docker"
	"github.com/bearded-web/bearded/pkg/utils"
)

var Agent = cli.Command{
	Name:  "agent",
	Usage: "Start agent",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "api-addr",
			Value:  "http://127.0.0.1:3003/api/",
			EnvVar: "BEARDED_API_ADDR",
			Usage:  "http address for connection to the api server",
		},
		cli.StringFlag{
			Name:   "name",
			EnvVar: "BEARDED_AGENT_NAME",
			Usage:  "Unique agent name, set to fqdn if empty",
		},
		cli.StringFlag{
			Name:   "token",
			Value:  "agent-token", // TODO (m0sth8): remove test-token
			EnvVar: "BEARDED_AGENT_TOKEN",
			Usage:  "Token, used for request",
		},
	},
	Action: takeApi(agentAction),
}

func agentAction(ctx *cli.Context, api *client.Client) {

	var agentName string
	if agentName = ctx.String("name"); agentName == "" {
		hostname, err := utils.GetHostname()
		if err != nil {
			panic(err)
		}
		agentName = hostname
	}
	err := ServeAgent(agentName, api)
	logrus.Error(err)
}

func ServeAgent(agentName string, api *client.Client) error {
	dclient, err := docker.NewDocker()
	if err != nil {
		panic(err)
	}
	logrus.Infof("Agent name: %s", agentName)
	server, err := agentServer.New(api, dclient, agentName)
	if err != nil {
		panic(err)
	}
	return server.Serve(context.Background())
}

func takeApi(fn func(*cli.Context, *client.Client)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		api := client.NewClient(ctx.String("api-addr"), nil)
		if ctx.GlobalBool("debug") {
			api.Debug = true
		}
		api.Token = ctx.String("token")
		fn(ctx, api)
	}
}
