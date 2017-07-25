package cmd

import (
	"fmt"
	"log"

	"github.com/jawher/mow.cli"
)

func snapshotsCreate(cmd *cli.Cmd) {
	cmd.Spec = "SUBID [-d]"

	id := cmd.StringArg("SUBID", "", "SUBID of virtual machine (see <servers>)")
	description := cmd.StringOpt("d description", "", "Description of snapshot")

	cmd.Action = func() {
		snapshot, err := GetClient().CreateSnapshot(*id, *description)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Snapshot created\n\n")
		lengths := []int{16, 48}
		tabsPrint(columns{"SNAPSHOTID", "DESCRIPTION"}, lengths)
		tabsPrint(columns{snapshot.ID, snapshot.Description}, lengths)
		tabsFlush()
	}
}

func snapshotsDelete(cmd *cli.Cmd) {
	id := cmd.StringArg("SNAPSHOTID", "", "SNAPSHOTID of snapshot to delete (see <snapshots>)")
	cmd.Action = func() {
		if err := GetClient().DeleteSnapshot(*id); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Snapshot deleted")
	}
}

func snapshotsList(cmd *cli.Cmd) {
	cmd.Action = func() {
		snapshots, err := GetClient().GetSnapshots()
		if err != nil {
			log.Fatal(err)
		}

		if len(snapshots) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{16, 40, 16, 16, 24}
		tabsPrint(columns{"SNAPSHOTID", "DESCRIPTION", "SIZE", "STATUS", "DATE"}, lengths)
		for _, snapshot := range snapshots {
			tabsPrint(columns{snapshot.ID, snapshot.Description, snapshot.Size, snapshot.Status, snapshot.Created}, lengths)
		}
		tabsFlush()
	}
}
