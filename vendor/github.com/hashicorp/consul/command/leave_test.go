package command

import (
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
	"strings"
	"testing"
)

func testLeaveCommand(t *testing.T) (*cli.MockUi, *LeaveCommand) {
	ui := new(cli.MockUi)
	return ui, &LeaveCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetClientHTTP,
		},
	}
}

func TestLeaveCommand_implements(t *testing.T) {
	var _ cli.Command = &LeaveCommand{}
}

func TestLeaveCommandRun(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	ui, c := testLeaveCommand(t)
	args := []string{"-http-addr=" + a1.httpAddr}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if !strings.Contains(ui.OutputWriter.String(), "leave complete") {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}

func TestLeaveCommandFailOnNonFlagArgs(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	_, c := testLeaveCommand(t)
	args := []string{"-http-addr=" + a1.httpAddr, "appserver1"}

	code := c.Run(args)
	if code == 0 {
		t.Fatalf("bad: failed to check for unexpected args")
	}
}
