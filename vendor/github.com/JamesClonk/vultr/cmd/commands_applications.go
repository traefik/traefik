package cmd

import (
	"fmt"
	"log"

	"github.com/jawher/mow.cli"
)

func appList(cmd *cli.Cmd) {
	cmd.Action = func() {
		apps, err := GetClient().GetApplications()
		if err != nil {
			log.Fatal(err)
		}

		if len(apps) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{8, 32, 24, 32, 12}
		tabsPrint(columns{"APPID", "NAME", "SHORT_NAME", "DEPLOY_NAME", "SURCHARGE"}, lengths)
		for _, app := range apps {
			tabsPrint(columns{app.ID, app.Name, app.ShortName, app.DeployName, app.Surcharge}, lengths)
		}
		tabsFlush()
	}
}
