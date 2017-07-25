package cmd

import (
	"fmt"
	"log"

	"github.com/jawher/mow.cli"
)

func sshServer(cmd *cli.Cmd) {
	cmd.Spec = "SUBID [OPTIONS]"
	id := cmd.StringArg("SUBID", "", "SUBID of virtual machine (see <servers>)")
	user := cmd.StringOpt("u user", "root", "Username for SSH login")
	port := cmd.IntOpt("p port", 22, "Port for SSH server")
	key := cmd.StringOpt("i keyfile", "", "Private keyfile for SSH login")

	cmd.Action = func() {
		// get IP of server
		server, err := GetClient().GetServer(*id)
		if err != nil {
			log.Fatal(err)
		}

		if server.ID == "" {
			fmt.Printf("No virtual machine with SUBID %v found\n", *id)
			return
		}

		if server.MainIP == "" || server.MainIP == "0.0.0.0" {
			fmt.Println("Virtual machine has no valid IP address")
			return
		}

		connectSSH(*user, server.MainIP, *key, *port)
	}
}
