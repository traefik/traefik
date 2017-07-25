package command

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/serf/coordinate"
	"github.com/mitchellh/cli"
)

func testRTTCommand(t *testing.T) (*cli.MockUi, *RTTCommand) {
	ui := new(cli.MockUi)
	return ui, &RTTCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetClientHTTP,
		},
	}
}

func TestRTTCommand_Implements(t *testing.T) {
	var _ cli.Command = &RTTCommand{}
}

func TestRTTCommand_Run_BadArgs(t *testing.T) {
	_, c := testRTTCommand(t)

	if code := c.Run([]string{}); code != 1 {
		t.Fatalf("expected return code 1, got %d", code)
	}

	if code := c.Run([]string{"node1", "node2", "node3"}); code != 1 {
		t.Fatalf("expected return code 1, got %d", code)
	}

	if code := c.Run([]string{"-wan", "node1", "node2"}); code != 1 {
		t.Fatalf("expected return code 1, got %d", code)
	}

	if code := c.Run([]string{"-wan", "node1.dc1", "node2"}); code != 1 {
		t.Fatalf("expected return code 1, got %d", code)
	}

	if code := c.Run([]string{"-wan", "node1", "node2.dc1"}); code != 1 {
		t.Fatalf("expected return code 1, got %d", code)
	}
}

func TestRTTCommand_Run_LAN(t *testing.T) {
	updatePeriod := 10 * time.Millisecond
	a := testAgentWithConfig(t, func(c *agent.Config) {
		c.ConsulConfig.CoordinateUpdatePeriod = updatePeriod
	})
	defer a.Shutdown()
	waitForLeader(t, a.httpAddr)

	// Inject some known coordinates.
	c1 := coordinate.NewCoordinate(coordinate.DefaultConfig())
	c2 := c1.Clone()
	c2.Vec[0] = 0.123
	dist_str := fmt.Sprintf("%.3f ms", c1.DistanceTo(c2).Seconds()*1000.0)
	{
		req := structs.CoordinateUpdateRequest{
			Datacenter: a.config.Datacenter,
			Node:       a.config.NodeName,
			Coord:      c1,
		}
		var reply struct{}
		if err := a.agent.RPC("Coordinate.Update", &req, &reply); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		req := structs.RegisterRequest{
			Datacenter: a.config.Datacenter,
			Node:       "dogs",
			Address:    "127.0.0.2",
		}
		var reply struct{}
		if err := a.agent.RPC("Catalog.Register", &req, &reply); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		var reply struct{}
		req := structs.CoordinateUpdateRequest{
			Datacenter: a.config.Datacenter,
			Node:       "dogs",
			Coord:      c2,
		}
		if err := a.agent.RPC("Coordinate.Update", &req, &reply); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Ask for the RTT of two known nodes
	ui, c := testRTTCommand(t)
	args := []string{
		"-http-addr=" + a.httpAddr,
		a.config.NodeName,
		"dogs",
	}

	// Wait for the updates to get flushed to the data store.
	if err := testutil.WaitForResult(func() (bool, error) {
		code := c.Run(args)
		if code != 0 {
			return false, fmt.Errorf("bad: %d: %#v", code, ui.ErrorWriter.String())
		}

		// Make sure the proper RTT was reported in the output.
		expected := fmt.Sprintf("rtt: %s", dist_str)
		if !strings.Contains(ui.OutputWriter.String(), expected) {
			return false, fmt.Errorf("bad: %#v", ui.OutputWriter.String())
		}

		return true, nil
	}); err != nil {
		t.Fatal(err)
	}

	// Default to the agent's node.
	{
		ui, c := testRTTCommand(t)
		args := []string{
			"-http-addr=" + a.httpAddr,
			"dogs",
		}
		code := c.Run(args)
		if code != 0 {
			t.Fatalf("bad: %d: %#v", code, ui.ErrorWriter.String())
		}

		// Make sure the proper RTT was reported in the output.
		expected := fmt.Sprintf("rtt: %s", dist_str)
		if !strings.Contains(ui.OutputWriter.String(), expected) {
			t.Fatalf("bad: %#v", ui.OutputWriter.String())
		}
	}

	// Try an unknown node.
	{
		ui, c := testRTTCommand(t)
		args := []string{
			"-http-addr=" + a.httpAddr,
			a.config.NodeName,
			"nope",
		}
		code := c.Run(args)
		if code != 1 {
			t.Fatalf("bad: %d: %#v", code, ui.ErrorWriter.String())
		}
	}
}

func TestRTTCommand_Run_WAN(t *testing.T) {
	a := testAgent(t)
	defer a.Shutdown()
	waitForLeader(t, a.httpAddr)

	node := fmt.Sprintf("%s.%s", a.config.NodeName, a.config.Datacenter)

	// We can't easily inject WAN coordinates, so we will just query the
	// node with itself.
	{
		ui, c := testRTTCommand(t)
		args := []string{
			"-wan",
			"-http-addr=" + a.httpAddr,
			node,
			node,
		}
		code := c.Run(args)
		if code != 0 {
			t.Fatalf("bad: %d: %#v", code, ui.ErrorWriter.String())
		}

		// Make sure there was some kind of RTT reported in the output.
		if !strings.Contains(ui.OutputWriter.String(), "rtt: ") {
			t.Fatalf("bad: %#v", ui.OutputWriter.String())
		}
	}

	// Default to the agent's node.
	{
		ui, c := testRTTCommand(t)
		args := []string{
			"-wan",
			"-http-addr=" + a.httpAddr,
			node,
		}
		code := c.Run(args)
		if code != 0 {
			t.Fatalf("bad: %d: %#v", code, ui.ErrorWriter.String())
		}

		// Make sure there was some kind of RTT reported in the output.
		if !strings.Contains(ui.OutputWriter.String(), "rtt: ") {
			t.Fatalf("bad: %#v", ui.OutputWriter.String())
		}
	}

	// Try an unknown node.
	{
		ui, c := testRTTCommand(t)
		args := []string{
			"-wan",
			"-http-addr=" + a.httpAddr,
			node,
			"dc1.nope",
		}
		code := c.Run(args)
		if code != 1 {
			t.Fatalf("bad: %d: %#v", code, ui.ErrorWriter.String())
		}
	}
}
