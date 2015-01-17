package main

import (
	"os"

	//	"github.com/codegangsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/cmd"
)

const (
	Version = "0.0.1"
	Author  = "m0sth8"
	Email   = "m0sth8@gmail.com"
	Name    = "bearded"
)

func BeforeHandler(c *cli.Context) error {
	// set log level
	level := logrus.DebugLevel
	switch c.String("log-level") {
	case "info":
		level = logrus.InfoLevel
	case "warning":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	case "fatal":
		level = logrus.FatalLevel
	case "panic":
		level = logrus.PanicLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{})
	return nil
}

func main() {
	app := cli.NewApp()

	app.Version = Version
	app.Author = Author
	app.Email = Email
	app.Name = Name
	app.Commands = []cli.Command{
		cmd.Dispatcher,
		cmd.Plugins,
	}

	app.Flags = append(app.Flags, []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "debug",
			Usage: "Logger level output [debug|info|warning|error|fatal], debug is default",
		},
		cli.BoolFlag{
			Name: "debug",
			Usage: "Enable some debugging features, such as: disable https checking, trace requests etc",
		},
	}...)

	app.Before = BeforeHandler
	app.Run(os.Args)
}
