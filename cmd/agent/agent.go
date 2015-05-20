package agent

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	//	"github.com/codegangpsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
	"golang.org/x/net/context"

	"github.com/bearded-web/bearded/cmd"
	agentServer "github.com/bearded-web/bearded/pkg/agent"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/config/flags"
	"github.com/bearded-web/bearded/pkg/utils"
	"github.com/bearded-web/bearded/pkg/utils/load"
)

const EnvPrefix = "BEARDED_AGENT"

func New() cli.Command {
	agent := cli.Command{
		Name:  "agent",
		Usage: "Start agent",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "config",
				EnvVar: fmt.Sprintf("%s_CONFIG", EnvPrefix),
				Usage:  "path to config",
			},
		},
		Action: cmd.TakeApi(agentAction),
	}
	cfg := config.NewAgent()
	agent.Flags = append(agent.Flags, flags.GenerateFlags(cfg, flags.Opts{
		EnvPrefix: EnvPrefix,
	})...)
	agent.Flags = append(agent.Flags, cmd.ApiFlags("BEARDED")...)
	return agent
}

func agentAction(cliCtx *cli.Context, api *client.Client, timeout cmd.Timeout) {
	cfg := config.NewAgent()
	if cfgPath := cliCtx.String("config"); cfgPath != "" {
		logrus.Infof("Load config from %s", cfgPath)
		err := load.FromFile(cfgPath, cfg)
		if err != nil {
			logrus.Fatalf("Couldn't load config: %s", err)
			return
		}
	}

	logrus.Info("Load config from flags")
	err := flags.ParseFlags(cfg, cliCtx, flags.Opts{
		EnvPrefix: EnvPrefix,
	})
	if err != nil {
		logrus.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aErr := make(chan error)
	go func(ch chan<- error) {
		ch <- agentServer.ServeAgent(ctx, cfg, api)
	}(aErr)

	select {
	case <-utils.NotifyInterrupt():
		logrus.Infof("Interrupting by signal")
		cancel()
		select {
		case err = <-aErr:
		case <-time.After(time.Second * 20):
			err = fmt.Errorf("Timeout exceeded")
		}
	case err = <-aErr:
	}

	if err != nil {
		logrus.Fatal(err)
	}
}
