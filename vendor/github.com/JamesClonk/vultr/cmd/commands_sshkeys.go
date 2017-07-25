package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	vultr "github.com/JamesClonk/vultr/lib"
	"github.com/jawher/mow.cli"
)

func sshKeysCreate(cmd *cli.Cmd) {
	cmd.Spec = "-n (-k | -f)"

	name := cmd.StringOpt("n name", "", "Name of the SSH key")
	key := cmd.StringOpt("k key", "", "SSH public key (in authorized_keys format)")
	file := cmd.StringOpt("f file", "", "SSH public key file to upload")

	cmd.Action = func() {
		if *file != "" {
			data, err := ioutil.ReadFile(*file)
			if err != nil {
				log.Fatal(err)
			}
			*key = string(data)
		}

		sshkey, err := GetClient().CreateSSHKey(*name, *key)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("SSH key created\n\n")
		lengths := []int{24, 32, 64}
		tabsPrint(columns{"SSHKEYID", "NAME", "KEY"}, lengths)
		tabsPrint(columns{sshkey.ID, sshkey.Name, sshkey.Key}, lengths)
		tabsFlush()
	}
}

func sshKeysUpdate(cmd *cli.Cmd) {
	cmd.Spec = "SSHKEYID [-n] [(-k | -f)]"

	id := cmd.StringArg("SSHKEYID", "", "SSHKEYID of key to update (see <sshkeys>)")
	name := cmd.StringOpt("n name", "", "New name for the SSH key")
	key := cmd.StringOpt("k key", "", "New SSH key contents")
	file := cmd.StringOpt("f file", "", "New SSH public key file to upload")

	cmd.Action = func() {
		if *file != "" {
			data, err := ioutil.ReadFile(*file)
			if err != nil {
				log.Fatal(err)
			}
			*key = string(data)
		}

		sshkey := vultr.SSHKey{
			ID:   *id,
			Name: *name,
			Key:  *key,
		}
		if err := GetClient().UpdateSSHKey(sshkey); err != nil {
			log.Fatal(err)
		}
		fmt.Println("SSH key updated")
	}
}

func sshKeysDelete(cmd *cli.Cmd) {
	id := cmd.StringArg("SSHKEYID", "", "SSHKEYID of key to delete (see <sshkeys>)")
	cmd.Action = func() {
		if err := GetClient().DeleteSSHKey(*id); err != nil {
			log.Fatal(err)
		}
		fmt.Println("SSH key deleted")
	}
}

func sshKeysList(cmd *cli.Cmd) {
	cmd.Spec = "[-f]"

	full := cmd.BoolOpt("f full", false, "Display full length of SSH key")

	cmd.Action = func() {
		keys, err := GetClient().GetSSHKeys()
		if err != nil {
			log.Fatal(err)
		}

		if len(keys) == 0 {
			fmt.Println()
			return
		}

		keyLength := 64
		if *full {
			keyLength = 8192
		}
		lengths := []int{24, 32, keyLength}

		tabsPrint(columns{"SSHKEYID", "NAME", "KEY"}, lengths)
		for _, key := range keys {
			tabsPrint(columns{key.ID, key.Name, key.Key}, lengths)
		}
		tabsFlush()
	}
}
