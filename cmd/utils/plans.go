package utils

import (
	"encoding/json"
	"fmt"
	"os"

	//	"github.com/codegangpsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed

	"github.com/bearded-web/bearded/cmd"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/utils/load"
)

var Plans = cli.Command{
	Name:  "plans",
	Usage: "Helper to work with plans",
	Flags: cmd.ApiFlags("BEARDED"),
	Subcommands: []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "Show all installed plans",
			Action: cmd.TakeApi(plansListAction),
		},
		cli.Command{
			Name:   "show",
			Usage:  "Show plan by id",
			Action: cmd.TakeApi(plansShowAction),
		},
		cli.Command{
			Name:   "load",
			Usage:  "Load plans from file",
			Action: cmd.TakeApi(plansLoadAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "update",
					Usage: "Update plan if existed",
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

func plansListAction(ctx *cli.Context, api *client.Client, timeout cmd.Timeout) {

	plans, err := api.Plans.List(timeout(), nil)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d plans:\n", plans.Count)
	for _, p := range plans.Results {
		fmt.Println(p)
	}
}

func plansShowAction(ctx *cli.Context, api *client.Client, timeout cmd.Timeout) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set plan id argument: plans show [id]\n")
		os.Exit(1)
	}
	plan, err := api.Plans.Get(timeout(), ctx.Args()[0])
	if err != nil {
		if client.IsNotFound(err) {
			fmt.Println("Plan not found")
			return
		}
		fmt.Print(err)
		return
	}
	fmt.Println("Plan found:")
	data, err := json.MarshalIndent(plan, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	return
}

func plansLoadAction(ctx *cli.Context, api *client.Client, timeout cmd.Timeout) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set filename argument: f.e plans load ./extra/data/plans.json\n")
		os.Exit(1)
	}
	filename := ctx.Args()[0]
	data := ExtraData{}
	load.FromFile(filename, &data, load.Opts{Format: load.Format(ctx.String("format"))})

	update := ctx.Bool("update")
	if update {
		fmt.Println("Autoupdate is enabled")
	}
	fmt.Printf("Found %d plans\n", len(data.Plans))
	for i, p := range data.Plans {
		fmt.Printf("%d) %s\n", i, p)
		fmt.Printf("Creating..\n")

		_, err := api.Plans.Create(timeout(), p)
		if err != nil {
			if client.IsConflicted(err) {
				fmt.Println("Plan with this name is already existed")
				if update {
					fmt.Println("Updating..")
					// retreive existed plan
					planList, err := api.Plans.List(timeout(), &client.PlansListOpts{Name: p.Name})
					if err != nil {
						panic(err)
					}
					if planList.Count != 1 {
						err := fmt.Errorf("Expected 1 plan, but actual is %d", planList.Count)
						panic(err)
					}

					// update it
					p.Id = planList.Results[0].Id
					_, err = api.Plans.Update(timeout(), p)
					if err != nil {
						fmt.Printf("Plugin updating failed, because: %v\n", err)
						continue
					}
				} else {
					continue
				}
			} else {
				fmt.Printf("Plan wasn't created because: %v\n", err)
				continue
			}
		}
		fmt.Println("Successful")
		//		fmt.Printf("%s\n", created)
	}

}
