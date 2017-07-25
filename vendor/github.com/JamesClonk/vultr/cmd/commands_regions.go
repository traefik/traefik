package cmd

import (
	"fmt"
	"log"

	"github.com/jawher/mow.cli"
)

func regionList(cmd *cli.Cmd) {
	cmd.Action = func() {
		regions, err := GetClient().GetRegions()
		if err != nil {
			log.Fatal(err)
		}

		if len(regions) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{8, 48, 24, 8, 8, 8, 8}
		tabsPrint(columns{"DCID", "NAME", "CONTINENT", "COUNTRY", "STATE", "STORAGE", "CODE"}, lengths)
		for _, region := range regions {
			tabsPrint(columns{
				region.ID, region.Name, region.Continent,
				region.Country, region.State, region.BlockStorage, region.Code,
			}, lengths)
		}
		tabsFlush()
	}
}
