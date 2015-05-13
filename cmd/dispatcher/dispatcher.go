package dispatcher

import (
	"github.com/Sirupsen/logrus"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed

	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/config/flags"
	"github.com/bearded-web/bearded/pkg/dispatcher"
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

func dispatcherAction(ctx *cli.Context) {
	cfg := config.NewDispatcher()
	if cfgPath := ctx.String("config"); cfgPath != "" {
		logrus.Infof("Load config from %s", cfgPath)
		err := load.FromFile(cfgPath, cfg)
		if err != nil {
			logrus.Fatalf("Couldn't load config: %s", err)
			return
		}
	}

	logrus.Info("Load config from flags")
	err := flags.ParseFlags(cfg, ctx, flags.Opts{
		EnvPrefix: "BEARDED",
	})
	if err != nil {
		logrus.Fatal(err)
	}
	if cfg.Debug = ctx.GlobalBool("debug"); cfg.Debug {
		logrus.Info("Debug mode is enabled")
	}
	logrus.Fatal(dispatcher.Serve(cfg))
}
