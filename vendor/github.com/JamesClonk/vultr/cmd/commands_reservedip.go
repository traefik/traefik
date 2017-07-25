package cmd

import (
	"fmt"
	"log"

	"github.com/jawher/mow.cli"
)

func reservedIPAttach(cmd *cli.Cmd) {
	cmd.Spec = "SUBID IP_ADDRESS"

	serverID := cmd.StringArg("SUBID", "", "SUBID of virtual machine to attach to (see <servers>)")
	ip := cmd.StringArg("IP_ADDRESS", "", "IP address to attach (see <reservedips>)")

	cmd.Action = func() {
		if err := GetClient().AttachReservedIP(*ip, *serverID); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Reserved IP attached")
	}
}

func reservedIPConvert(cmd *cli.Cmd) {
	cmd.Spec = "SUBID IP_ADDRESS"

	serverID := cmd.StringArg("SUBID", "", "SUBID of virtual machine (see <servers>)")
	ip := cmd.StringArg("IP_ADDRESS", "", "IP address to convert to reserved IP")

	cmd.Action = func() {
		id, err := GetClient().ConvertReservedIP(*serverID, *ip)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Reserved IP converted\n\n")
		lengths := []int{12, 48, 12}
		tabsPrint(columns{"ID", "IP_ADDRESS", "ATTACHED_TO"}, lengths)
		tabsPrint(columns{id, *ip, *serverID}, lengths)
		tabsFlush()
	}
}

func reservedIPCreate(cmd *cli.Cmd) {
	cmd.Spec = "[-r -t -l]"

	regionID := cmd.IntOpt("r region", 1, "Region (DCID)")
	ipType := cmd.StringOpt("t type", "v4", "Type of new reserved IP (v4 or v6)")
	label := cmd.StringOpt("l label", "", "Label for new reserved IP")

	cmd.Action = func() {
		id, err := GetClient().CreateReservedIP(*regionID, *ipType, *label)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Reserved IP created\n\n")
		lengths := []int{12, 6, 10, 32}
		tabsPrint(columns{"ID", "TYPE", "DCID", "LABEL"}, lengths)
		tabsPrint(columns{id, *ipType, *regionID, label}, lengths)
		tabsFlush()
	}
}

func reservedIPDestroy(cmd *cli.Cmd) {
	cmd.Spec = "SUBID"

	id := cmd.StringArg("SUBID", "", "SUBID of reserved IP (see <reservedips>)")

	cmd.Action = func() {
		if err := GetClient().DestroyReservedIP(*id); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Reserved IP deleted")
	}
}

func reservedIPDetach(cmd *cli.Cmd) {
	cmd.Spec = "SUBID IP_ADDRESS"

	serverID := cmd.StringArg("SUBID", "", "SUBID of virtual machine to detach from (see <servers>)")
	ip := cmd.StringArg("IP_ADDRESS", "", "IP address to detach (see <reservedips>)")

	cmd.Action = func() {
		if err := GetClient().DetachReservedIP(*ip, *serverID); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Reserved IP detached")
	}
}

func reservedIPList(cmd *cli.Cmd) {
	cmd.Action = func() {
		ips, err := GetClient().ListReservedIP()
		if err != nil {
			log.Fatal(err)
		}

		if len(ips) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{12, 8, 8, 48, 6, 32, 12}
		tabsPrint(columns{"SUBID", "DCID", "IP_TYPE", "SUBNET", "SIZE", "LABEL", "ATTACHED_TO"}, lengths)
		for _, ip := range ips {
			tabsPrint(columns{
				ip.ID,
				ip.RegionID,
				ip.IPType,
				ip.Subnet,
				ip.SubnetSize,
				ip.Label,
				ip.AttachedTo,
			}, lengths)
		}
		tabsFlush()
	}
}
