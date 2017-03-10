package main

import (
	"os"

	"github.com/JamesClonk/vultr/cmd"
)

func main() {
	cli := cmd.NewCLI()
	cli.RegisterCommands()
	cli.Run(os.Args)
}
