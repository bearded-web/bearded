package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/mgo.v2/bson"
	//	"github.com/codegangsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed

	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/pkg/client"
)

var Plugins = cli.Command{
	Name:  "plugins",
	Usage: "Helper to work with plugins",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "api-addr",
			Value:  "http://127.0.0.1:3003/api/",
			EnvVar: "BEARDED_API_ADDR",
			Usage:  "http address for connection to the api server",
		},
	},
	Subcommands: []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "Show all installed plugins",
			Action: takeApi(pluginsListAction),
		},
		cli.Command{
			Name:   "show",
			Usage:  "Show plugin by id",
			Action: takeApi(pluginsShowAction),
		},
		cli.Command{
			Name:   "load",
			Usage:  "Load plugins from file",
			Action: takeApi(pluginsLoadAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "update",
					Usage: "Update plugin if existed",
				},
			},
		},
	},
}

func takeApi(fn func(*cli.Context, *client.Client)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		api := client.NewClient(ctx.String("api-addr"), nil)
		fn(ctx, api)
	}
}

// ========= Actions

func pluginsListAction(ctx *cli.Context, api *client.Client) {
	plugins, err := api.Plugins.List(nil)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d plugins:\n", plugins.Count)
	for _, p := range plugins.Results {
		fmt.Println(p)
	}
}

func pluginsShowAction(ctx *cli.Context, api *client.Client) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set plugin id argument: plugins show [id]\n")
		os.Exit(1)
	}
	plugin, err := api.Plugins.Get(ctx.Args()[0])
	if err != nil {
		if client.IsNotFound(err) {
			fmt.Println("Plugin not found")
			return
		}
		fmt.Print(err)
		return
	}
	fmt.Println("Plugin found:")
	data, err := json.MarshalIndent(plugin, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	return
}

func pluginsLoadAction(ctx *cli.Context, api *client.Client) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set filename argument: f.e plugins load ./extra/plugins/base.json\n")
		os.Exit(1)
	}
	filename := ctx.Args()[0]
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if !os.IsExist(err) {
			fmt.Printf("File %s is not existed\n", filename)
			os.Exit(1)
		}
		panic(err)
	}
	plugins := []*plugin.Plugin{}
	if err := json.Unmarshal(data, &plugins); err != nil {
		panic(err)
	}
	update := ctx.Bool("update")
	if update {
		fmt.Println("Autoupdate is enabled")
	}
	fmt.Printf("Found %d plugins\n", len(plugins))
	for i, p := range plugins {
		fmt.Printf("%d) %s\n", i, p)
		fmt.Printf("Creating..\n")
		_, err := api.Plugins.Create(p)
		if err != nil {
			if client.IsConflicted(err) {
				fmt.Println("Plugin is already existed")
				if update {
					fmt.Println("Updating..")
					p.Id = bson.ObjectIdHex("54b6561ec2806c7576000001")
					_, err = api.Plugins.Update(p)
					if err != nil {
						fmt.Printf("Plugin updating failed, because: %v", err)
						continue
					}
				} else {
					continue
				}
			} else {
				fmt.Printf("Plugin wasn't created because: %v", err)
				continue
			}
		}
		fmt.Println("Successful")
		//		fmt.Printf("%s\n", created)
	}

}
