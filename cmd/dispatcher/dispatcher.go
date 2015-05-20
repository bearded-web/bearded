package dispatcher

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
	"golang.org/x/net/context"

	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/config/flags"
	"github.com/bearded-web/bearded/pkg/dispatcher"
	"github.com/bearded-web/bearded/pkg/utils"
	"github.com/bearded-web/bearded/pkg/utils/async"
	"github.com/bearded-web/bearded/pkg/utils/load"
)

func New() cli.Command {
	cmd := cli.Command{
		Name:   "dispatcher",
		Usage:  "Start Dispatcher",
		Action: dispatcherAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "config",
				EnvVar: "BEARDED_CONFIG",
				Usage:  "path to config",
			},
			cli.StringFlag{
				Name:   "config-format",
				EnvVar: "BEARDED_CONFIG_FORMAT",
				Usage:  "Specify config format, by default format is taken from ext",
			},
		},
	}
	cfg := config.NewDispatcher()
	cmd.Flags = append(cmd.Flags, flags.GenerateFlags(cfg, flags.Opts{
		EnvPrefix: "BEARDED",
	})...)
	return cmd
}

func dispatcherAction(cliCtx *cli.Context) {
	cfg := config.NewDispatcher()
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
		EnvPrefix: "BEARDED",
	})
	if err != nil {
		logrus.Fatal(err)
	}
	if cfg.Debug = cliCtx.GlobalBool("debug"); cfg.Debug {
		logrus.Info("Debug mode is enabled")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dErr := async.Promise(func() error { return dispatcher.Serve(ctx, cfg) })

	select {
	case <-utils.NotifyInterrupt():
		logrus.Info("Interrupting by signal")
		cancel()
		select {
		case err = <-dErr:
		case <-time.After(time.Second * 20):
			err = fmt.Errorf("Timeout exceeded")
		}
	case err = <-dErr:
	}

	if err != nil {
		logrus.Fatal(err)
	}
}
