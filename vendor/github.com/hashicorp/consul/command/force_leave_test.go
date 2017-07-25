package command

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/serf/serf"
	"github.com/mitchellh/cli"
)

func testForceLeaveCommand(t *testing.T) (*cli.MockUi, *ForceLeaveCommand) {
	ui := new(cli.MockUi)
	return ui, &ForceLeaveCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetClientHTTP,
		},
	}
}

func TestForceLeaveCommand_implements(t *testing.T) {
	var _ cli.Command = &ForceLeaveCommand{}
}

func TestForceLeaveCommandRun(t *testing.T) {
	a1 := testAgent(t)
	a2 := testAgent(t)
	defer a1.Shutdown()
	defer a2.Shutdown()

	addr := fmt.Sprintf("127.0.0.1:%d", a2.config.Ports.SerfLan)
	_, err := a1.agent.JoinLAN([]string{addr})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Forcibly shutdown a2 so that it appears "failed" in a1
	a2.Shutdown()

	ui, c := testForceLeaveCommand(t)
	args := []string{
		"-http-addr=" + a1.httpAddr,
		a2.config.NodeName,
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	m := a1.agent.LANMembers()
	if len(m) != 2 {
		t.Fatalf("should have 2 members: %#v", m)
	}

	if err := testutil.WaitForResult(func() (bool, error) {
		m = a1.agent.LANMembers()
		success := m[1].Status == serf.StatusLeft
		return success, errors.New(m[1].Status.String())
	}); err != nil {
		t.Fatalf("member status is %v, should be left", err)
	}
}

func TestForceLeaveCommandRun_noAddrs(t *testing.T) {
	ui := new(cli.MockUi)
	ui, c := testForceLeaveCommand(t)
	args := []string{"-http-addr=foo"}

	code := c.Run(args)
	if code != 1 {
		t.Fatalf("bad: %d", code)
	}

	if !strings.Contains(ui.ErrorWriter.String(), "node name") {
		t.Fatalf("bad: %#v", ui.ErrorWriter.String())
	}
}
