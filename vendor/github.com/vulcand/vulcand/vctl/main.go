package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/vulcand/vulcand/plugin/registry"
	"github.com/vulcand/vulcand/vctl/command"
)

var vulcanUrl string

func main() {
	cmd := command.NewCommand(registry.GetRegistry())
	err := cmd.Run(os.Args)
	if err != nil {
		log.Errorf("error: %s\n", err)
	}
}
