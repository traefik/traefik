package command

import (
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
	"strings"
	"testing"
)

func TestEventCommand_implements(t *testing.T) {
	var _ cli.Command = &EventCommand{}
}

func TestEventCommandRun(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	ui := new(cli.MockUi)
	c := &EventCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetClientHTTP,
		},
	}
	args := []string{"-http-addr=" + a1.httpAddr, "-name=cmd"}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if !strings.Contains(ui.OutputWriter.String(), "Event ID: ") {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}
