package command

import (
	"fmt"
	"time"

	"github.com/buger/goterm"
	"github.com/codegangsta/cli"
	"github.com/vulcand/vulcand/engine"
)

func NewTopCommand(cmd *Command) cli.Command {
	return cli.Command{
		Name:  "top",
		Usage: "Show vulcan status and configuration in top-style mode",
		Flags: []cli.Flag{
			cli.IntFlag{Name: "limit", Usage: "How many top entries to show", Value: 20},
			cli.IntFlag{Name: "refresh", Usage: "How often refresh (in seconds), if 0 - will display only once", Value: 1},
			cli.StringFlag{Name: "backend, b", Usage: "Filter frontends and servers by backend id", Value: ""},
		},
		Action: cmd.topAction,
	}
}

func (cmd *Command) topAction(c *cli.Context) {
	cmd.overviewAction(c.String("backend"), c.Int("refresh"), c.Int("limit"))
}

func (cmd *Command) overviewAction(backendId string, watch int, limit int) {
	var bk *engine.BackendKey
	if backendId != "" {
		bk = &engine.BackendKey{Id: backendId}
	}
	for {
		frontends, err := cmd.client.TopFrontends(bk, limit)
		if err != nil {
			cmd.printError(err)
			frontends = []engine.Frontend{}
		}

		servers, err := cmd.client.TopServers(bk, limit)
		if err != nil {
			cmd.printError(err)
			servers = []engine.Server{}
		}
		t := time.Now()
		if watch != 0 {
			goterm.Clear()
			goterm.MoveCursor(1, 1)
			goterm.Flush()
			fmt.Fprintf(cmd.out, "%s Every %d seconds. Top %d entries\n\n", t.Format("2006-01-02 15:04:05"), watch, limit)
		}
		cmd.printOverview(frontends, servers)
		if watch != 0 {
			goterm.Flush()
		} else {
			return
		}
		time.Sleep(time.Second * time.Duration(watch))
	}
}
