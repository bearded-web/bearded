package cmd

import (
	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/bearded-web/bearded/handlers/api"
	"github.com/bearded-web/bearded/modules/dispatcher"
	"github.com/bearded-web/bearded/models/task"
	"github.com/bearded-web/bearded/modules/worker"
	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/modules/docker"
	"log"
	"os"
)

func runHttpApi(disp *dispatcher.Dispatcher, workers *worker.Manager,  endpoint string) {
	logrus.Debug("Run http api")
	a := api.New(disp, workers)
	mux := gin.Default()
	route := mux.Group("/api")

	route.POST("/tasks", a.TaskCreate)
	route.GET("/tasks", a.TaskList)
	{
		taskRoute := route.Group("/tasks/:taskId")
		taskRoute.GET("", a.TakeTask(a.TaskGet))
		taskRoute.DELETE("", a.TakeTask(a.TaskDelete))
		taskRoute.GET("/state", a.TakeTask(a.TaskStateGet))
		taskRoute.GET("/report", a.TakeTask(a.TaskReportGet))
	}

	mux.Run(endpoint)
}

var Run = cli.Command{
	Name:        "run",
	Usage:       "Start Dispatcher",
	Description: ``,
	Action:      run,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "http",
			Value:  "127.0.0.1:3003",
			EnvVar: "HTTP_ADDR",
		},
		cli.IntFlag{
			Name:  "workers",
			Value: 4,
		},
	},
}

func run(ctx *cli.Context) {
	disp := &dispatcher.Dispatcher{
		TaskManager: task.NewMemoryManager(),
	}
	logger := log.New(os.Stderr, "", log.LstdFlags)
	d, err := docker.New(logger)
	if err != nil {
		panic(err)
	}
	workers := worker.NewManager(
		d,
		&worker.ManagerOpts{
			Size: ctx.Int("workers"),
		},
	)
	go workers.Serve(context.Background())
	if httpEndpoint := ctx.String("http"); httpEndpoint != "" {
		runHttpApi(disp, workers, httpEndpoint)
	}

}
