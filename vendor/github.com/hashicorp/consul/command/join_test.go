package command

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func testJoinCommand(t *testing.T) (*cli.MockUi, *JoinCommand) {
	ui := new(cli.MockUi)
	return ui, &JoinCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetClientHTTP,
		},
	}
}

func TestJoinCommand_implements(t *testing.T) {
	var _ cli.Command = &JoinCommand{}
}

func TestJoinCommandRun(t *testing.T) {
	a1 := testAgent(t)
	a2 := testAgent(t)
	defer a1.Shutdown()
	defer a2.Shutdown()

	ui, c := testJoinCommand(t)
	args := []string{
		"-http-addr=" + a1.httpAddr,
		fmt.Sprintf("127.0.0.1:%d", a2.config.Ports.SerfLan),
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if len(a1.agent.LANMembers()) != 2 {
		t.Fatalf("bad: %#v", a1.agent.LANMembers())
	}
}

func TestJoinCommandRun_wan(t *testing.T) {
	a1 := testAgent(t)
	a2 := testAgent(t)
	defer a1.Shutdown()
	defer a2.Shutdown()

	ui, c := testJoinCommand(t)
	args := []string{
		"-http-addr=" + a1.httpAddr,
		"-wan",
		fmt.Sprintf("127.0.0.1:%d", a2.config.Ports.SerfWan),
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if len(a1.agent.WANMembers()) != 2 {
		t.Fatalf("bad: %#v", a1.agent.WANMembers())
	}
}

func TestJoinCommandRun_noAddrs(t *testing.T) {
	ui, c := testJoinCommand(t)
	args := []string{"-http-addr=foo"}

	code := c.Run(args)
	if code != 1 {
		t.Fatalf("bad: %d", code)
	}

	if !strings.Contains(ui.ErrorWriter.String(), "one address") {
		t.Fatalf("bad: %#v", ui.ErrorWriter.String())
	}
}
