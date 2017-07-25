package cmd

import (
	"fmt"
	"log"

	"github.com/jawher/mow.cli"
)

func isoList(cmd *cli.Cmd) {
	cmd.Action = func() {
		iso, err := GetClient().GetISO()
		if err != nil {
			log.Fatal(err)
		}

		if len(iso) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{8, 48, 16, 48, 24}
		tabsPrint(columns{"ISOID", "FILENAME", "SIZE", "MD5SUM", "CREATED DATE"}, lengths)
		for _, iso := range iso {
			tabsPrint(columns{iso.ID, iso.Filename, iso.Size, iso.MD5sum, iso.Created}, lengths)
		}
		tabsFlush()
	}
}
