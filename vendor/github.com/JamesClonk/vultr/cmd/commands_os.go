package cmd

import (
	"fmt"
	"log"

	"github.com/jawher/mow.cli"
)

func osList(cmd *cli.Cmd) {
	cmd.Action = func() {
		os, err := GetClient().GetOS()
		if err != nil {
			log.Fatal(err)
		}

		if len(os) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{8, 32, 8, 16, 8}
		tabsPrint(columns{"OSID", "NAME", "ARCH", "FAMILY", "WINDOWS"}, lengths)
		for _, os := range os {
			tabsPrint(columns{os.ID, os.Name, os.Arch, os.Family, os.Windows}, lengths)
		}
		tabsFlush()
	}
}
