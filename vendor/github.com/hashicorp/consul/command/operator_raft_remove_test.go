package command

import (
	"strings"
	"testing"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func TestOperator_Raft_RemovePeer_Implements(t *testing.T) {
	var _ cli.Command = &OperatorRaftRemoveCommand{}
}

func TestOperator_Raft_RemovePeer(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()
	waitForLeader(t, a1.httpAddr)

	// Test the legacy mode with 'consul operator raft -remove-peer'
	{
		ui, c := testOperatorRaftCommand(t)
		args := []string{"-http-addr=" + a1.httpAddr, "-remove-peer", "-address=nope"}

		code := c.Run(args)
		if code != 1 {
			t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
		}

		// If we get this error, it proves we sent the address all they through.
		output := strings.TrimSpace(ui.ErrorWriter.String())
		if !strings.Contains(output, "address \"nope\" was not found in the Raft configuration") {
			t.Fatalf("bad: %s", output)
		}
	}

	// Test the remove-peer subcommand directly
	{
		ui := new(cli.MockUi)
		c := OperatorRaftRemoveCommand{
			Command: base.Command{
				Ui:    ui,
				Flags: base.FlagSetHTTP,
			},
		}
		args := []string{"-http-addr=" + a1.httpAddr, "-address=nope"}

		code := c.Run(args)
		if code != 1 {
			t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
		}

		// If we get this error, it proves we sent the address all they through.
		output := strings.TrimSpace(ui.ErrorWriter.String())
		if !strings.Contains(output, "address \"nope\" was not found in the Raft configuration") {
			t.Fatalf("bad: %s", output)
		}
	}

	// Test the remove-peer subcommand with -id
	{
		ui := new(cli.MockUi)
		c := OperatorRaftRemoveCommand{
			Command: base.Command{
				Ui:    ui,
				Flags: base.FlagSetHTTP,
			},
		}
		args := []string{"-http-addr=" + a1.httpAddr, "-id=nope"}

		code := c.Run(args)
		if code != 1 {
			t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
		}

		// If we get this error, it proves we sent the address all they through.
		output := strings.TrimSpace(ui.ErrorWriter.String())
		if !strings.Contains(output, "id \"nope\" was not found in the Raft configuration") {
			t.Fatalf("bad: %s", output)
		}
	}
}
