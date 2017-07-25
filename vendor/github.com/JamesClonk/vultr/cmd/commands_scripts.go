package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	vultr "github.com/JamesClonk/vultr/lib"
	"github.com/jawher/mow.cli"
)

func scriptsCreate(cmd *cli.Cmd) {
	cmd.Spec = "-n (-s | -f) [-t]"

	name := cmd.StringOpt("n name", "", "Name of the new startup script")
	content := cmd.StringOpt("s script", "", "Startup script contents")
	file := cmd.StringOpt("f file", "", "Startup script file to upload")
	stype := cmd.StringOpt("t type", "boot", "Type of startup script (boot or pxe)")

	cmd.Action = func() {
		if *file != "" {
			data, err := ioutil.ReadFile(*file)
			if err != nil {
				log.Fatal(err)
			}
			*content = string(data)
		}

		script, err := GetClient().CreateStartupScript(*name, *content, *stype)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Startup script created\n\n")
		lengths := []int{12, 32, 8, 64}
		tabsPrint(columns{"SCRIPTID", "NAME", "TYPE", "SCRIPT"}, lengths)
		tabsPrint(columns{script.ID, script.Name, script.Type, script.Content}, lengths)
		tabsFlush()
	}
}

func scriptsUpdate(cmd *cli.Cmd) {
	cmd.Spec = "SCRIPTID [-n] [(-s | -f)]"

	id := cmd.StringArg("SCRIPTID", "", "SCRIPTID of script to update (see <scripts>)")
	name := cmd.StringOpt("n name", "", "New name for the startup script")
	content := cmd.StringOpt("s script", "", "New contents for startup script")
	file := cmd.StringOpt("f file", "", "New file for startup script to upload")

	cmd.Action = func() {
		if *file != "" {
			data, err := ioutil.ReadFile(*file)
			if err != nil {
				log.Fatal(err)
			}
			*content = string(data)
		}

		script := vultr.StartupScript{
			ID:      *id,
			Name:    *name,
			Content: *content,
		}
		if err := GetClient().UpdateStartupScript(script); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Startup script updated")
	}
}

func scriptsDelete(cmd *cli.Cmd) {
	id := cmd.StringArg("SCRIPTID", "", "SCRIPTID of script to delete (see <scripts>)")
	cmd.Action = func() {
		if err := GetClient().DeleteStartupScript(*id); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Startup script deleted")
	}
}

func scriptsList(cmd *cli.Cmd) {
	cmd.Action = func() {
		scripts, err := GetClient().GetStartupScripts()
		if err != nil {
			log.Fatal(err)
		}

		if len(scripts) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{12, 32, 8, 64}
		tabsPrint(columns{"SCRIPTID", "NAME", "TYPE", "SCRIPT"}, lengths)
		for _, script := range scripts {
			tabsPrint(columns{script.ID, script.Name, script.Type, script.Content}, lengths)
		}
		tabsFlush()
	}
}

func scriptsShow(cmd *cli.Cmd) {
	cmd.Spec = "SCRIPTID"

	id := cmd.StringArg("SCRIPTID", "", "SCRIPTID of startup script (see <scripts>)")

	cmd.Action = func() {
		script, err := GetClient().GetStartupScript(*id)
		if err != nil {
			log.Fatal(err)
		}

		if script.ID == "" {
			fmt.Printf("No startup script with SUBID %v found!\n", *id)
			return
		}

		fmt.Println(script.Content)
	}
}
