package main

import (
	"github.com/codegangsta/cli"
	"github.com/sirupsen/logrus"

	"github.com/bearded-web/bearded/cmd"
	"os"
)

const (
	Version = "0.0.1"
	Author  = "m0sth8"
	Email   = "m0sth8@gmail.com"
	Name    = "bearded"
)

func BeforeHandler(c *cli.Context) error {
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
		cmd.Run,
	}

	app.Flags = append(app.Flags, []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "debug",
			Usage: "Logger level output [debug|info|warning|error|fatal], debug is default",
		},
	}...)

	app.Before = BeforeHandler

//	app.Run([]string{"bearded", "run"})
	app.Run(os.Args)
}
