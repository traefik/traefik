package command

import (
	"strings"
	"testing"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func TestWatchCommand_implements(t *testing.T) {
	var _ cli.Command = &WatchCommand{}
}

func TestWatchCommandRun(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	ui := new(cli.MockUi)
	c := &WatchCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}
	args := []string{"-http-addr=" + a1.httpAddr, "-type=nodes"}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if !strings.Contains(ui.OutputWriter.String(), a1.config.NodeName) {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}
