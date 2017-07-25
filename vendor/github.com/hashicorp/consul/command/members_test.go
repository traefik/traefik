package command

import (
	"fmt"
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
	"strings"
	"testing"
)

func testMembersCommand(t *testing.T) (*cli.MockUi, *MembersCommand) {
	ui := new(cli.MockUi)
	return ui, &MembersCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetClientHTTP,
		},
	}
}

func TestMembersCommand_implements(t *testing.T) {
	var _ cli.Command = &MembersCommand{}
}

func TestMembersCommandRun(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	ui, c := testMembersCommand(t)
	args := []string{"-http-addr=" + a1.httpAddr}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	// Name
	if !strings.Contains(ui.OutputWriter.String(), a1.config.NodeName) {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}

	// Agent type
	if !strings.Contains(ui.OutputWriter.String(), "server") {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}

	// Datacenter
	if !strings.Contains(ui.OutputWriter.String(), "dc1") {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}

func TestMembersCommandRun_WAN(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	ui, c := testMembersCommand(t)
	args := []string{"-http-addr=" + a1.httpAddr, "-wan"}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if !strings.Contains(ui.OutputWriter.String(), fmt.Sprintf("%d", a1.config.Ports.SerfWan)) {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}

func TestMembersCommandRun_statusFilter(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	ui, c := testMembersCommand(t)
	args := []string{
		"-http-addr=" + a1.httpAddr,
		"-status=a.*e",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if !strings.Contains(ui.OutputWriter.String(), a1.config.NodeName) {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}

func TestMembersCommandRun_statusFilter_failed(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()

	ui, c := testMembersCommand(t)
	args := []string{
		"-http-addr=" + a1.httpAddr,
		"-status=(fail|left)",
	}

	code := c.Run(args)
	if code == 1 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if strings.Contains(ui.OutputWriter.String(), a1.config.NodeName) {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}

	if code != 2 {
		t.Fatalf("bad: %d", code)
	}
}
