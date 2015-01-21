package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	//	"github.com/codegangpsta/cli"
	"github.com/m0sth8/cli" // use fork until subcommands will be fixed

	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/pkg/client"
)

var Plans = cli.Command{
	Name:  "plans",
	Usage: "Helper to work with plans",
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
			Usage:  "Show all installed plans",
			Action: takeApi(plansListAction),
		},
		cli.Command{
			Name:   "show",
			Usage:  "Show plan by id",
			Action: takeApi(plansShowAction),
		},
		cli.Command{
			Name:   "load",
			Usage:  "Load plans from file",
			Action: takeApi(plansLoadAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "update",
					Usage: "Update plan if existed",
				},
			},
		},
	},
}

// ========= Actions

func plansListAction(ctx *cli.Context, api *client.Client) {
	plans, err := api.Plans.List(nil)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d plans:\n", plans.Count)
	for _, p := range plans.Results {
		fmt.Println(p)
	}
}

func plansShowAction(ctx *cli.Context, api *client.Client) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set plan id argument: plans show [id]\n")
		os.Exit(1)
	}
	plan, err := api.Plans.Get(ctx.Args()[0])
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

func plansLoadAction(ctx *cli.Context, api *client.Client) {
	if len(ctx.Args()) == 0 {
		fmt.Printf("You should set filename argument: f.e plans load ./extra/data/plans.json\n")
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
	plans := []*plan.Plan{}
	if err := json.Unmarshal(data, &plans); err != nil {
		panic(err)
	}
	update := ctx.Bool("update")
	if update {
		fmt.Println("Autoupdate is enabled")
	}
	fmt.Printf("Found %d plans\n", len(plans))
	for i, p := range plans {
		fmt.Printf("%d) %s\n", i, p)
		fmt.Printf("Creating..\n")

		_, err := api.Plans.Create(p)
		if err != nil {
			if client.IsConflicted(err) {
				fmt.Println("Plan with this name is already existed")
				if update {
					fmt.Println("Updating..")
					// retreive existed plan
					planList, err := api.Plans.List(&client.PlansListOpts{Name: p.Name})
					if err != nil {
						panic(err)
					}
					if planList.Count != 1 {
						err := fmt.Errorf("Expected 1 plan, but actual is %d", planList.Count)
						panic(err)
					}

					// update it
					p.Id = planList.Results[0].Id
					_, err = api.Plans.Update(p)
					if err != nil {
						fmt.Printf("Plugin updating failed, because: %v", err)
						continue
					}
				} else {
					continue
				}
			} else {
				fmt.Printf("Plan wasn't created because: %v", err)
				continue
			}
		}
		fmt.Println("Successful")
		//		fmt.Printf("%s\n", created)
	}

}
