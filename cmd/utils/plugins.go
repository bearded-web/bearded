package utils

import (
	"encoding/json"
	"fmt"
	"os"

	//	"github.com/codegangpsta/cli"
	"github.com/bearded-web/bearded/cmd"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/utils/load"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed
)

var Plugins = cli.Command{
	Name:  "plugins",
	Usage: "Helper to work with plugins",
	Flags: cmd.ApiFlags("BEARDED"),
	Subcommands: []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "Show all installed plugins",
			Action: cmd.TakeApi(pluginsListAction),
		},
		cli.Command{
			Name:   "show",
			Usage:  "Show plugin by id",
			Action: cmd.TakeApi(pluginsShowAction),
		},
		cli.Command{
			Name:   "load",
			Usage:  "Load plugins from file",
			Action: cmd.TakeApi(pluginsLoadAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "update",
					Usage: "Update plugin if existed",
				},
				cli.BoolFlag{
					Name:  "disable",
					Usage: "Set plugin as disabled",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "Specify file format, by default format is taken from ext",
				},
			},
		},
	},
}

// ========= Actions

func pluginsListAction(ctx *cli.Context, api *client.Client, timeout cmd.Timeout) {
	plugins, err := api.Plugins.List(timeout(), nil)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d plugins:\n", plugins.Count)
	for _, p := range plugins.Results {
		fmt.Println(p)
	}
}

func pluginsShowAction(ctx *cli.Context, api *client.Client, timeout cmd.Timeout) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set plugin id argument: plugins show [id]\n")
		os.Exit(1)
	}
	plugin, err := api.Plugins.Get(timeout(), ctx.Args()[0])
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

func pluginsLoadAction(ctx *cli.Context, api *client.Client, timeout cmd.Timeout) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set filename argument: f.e plugins load ./extra/data/plugins.json\n")
		os.Exit(1)
	}
	filename := ctx.Args()[0]
	data := ExtraData{}
	load.FromFile(filename, &data, load.Opts{Format: load.Format(ctx.String("format"))})

	update := ctx.Bool("update")
	if update {
		fmt.Println("Autoupdate is enabled")
	}

	fmt.Printf("Found %d plugins\n", len(data.Plugins))
	for i, p := range data.Plugins {
		fmt.Printf("%d) %s\n", i, p)
		fmt.Printf("Creating..\n")
		if !ctx.Bool("disable") {
			p.Enabled = true
		}
		_, err := api.Plugins.Create(timeout(), p)
		if err != nil {
			if client.IsConflicted(err) {
				fmt.Println("Plugin with this version is already existed")
				if update {
					fmt.Println("Updating..")
					// retrieve existed version
					opts := client.PluginsListOpts{Name: p.Name, Version: p.Version}
					pluginList, err := api.Plugins.List(timeout(), &opts)
					if err != nil {
						panic(err)
					}
					if pluginList.Count != 1 {
						err := fmt.Errorf("Expected 1 plugin, but actual is %d", pluginList.Count)
						panic(err)
					}

					// update it
					p.Id = pluginList.Results[0].Id
					_, err = api.Plugins.Update(timeout(), p)
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
	}
}
