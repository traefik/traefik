package command

import (
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/mitchellh/cli"
)

func TestOperator_Autopilot_Set_Implements(t *testing.T) {
	var _ cli.Command = &OperatorAutopilotSetCommand{}
}

func TestOperator_Autopilot_Set(t *testing.T) {
	a1 := testAgent(t)
	defer a1.Shutdown()
	waitForLeader(t, a1.httpAddr)

	ui := new(cli.MockUi)
	c := OperatorAutopilotSetCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}
	args := []string{
		"-http-addr=" + a1.httpAddr,
		"-cleanup-dead-servers=false",
		"-max-trailing-logs=99",
		"-last-contact-threshold=123ms",
		"-server-stabilization-time=123ms",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}
	output := strings.TrimSpace(ui.OutputWriter.String())
	if !strings.Contains(output, "Configuration updated") {
		t.Fatalf("bad: %s", output)
	}

	req := structs.DCSpecificRequest{
		Datacenter: "dc1",
	}
	var reply structs.AutopilotConfig
	if err := a1.agent.RPC("Operator.AutopilotGetConfiguration", &req, &reply); err != nil {
		t.Fatalf("err: %v", err)
	}

	if reply.CleanupDeadServers {
		t.Fatalf("bad: %#v", reply)
	}
	if reply.MaxTrailingLogs != 99 {
		t.Fatalf("bad: %#v", reply)
	}
	if reply.LastContactThreshold != 123*time.Millisecond {
		t.Fatalf("bad: %#v", reply)
	}
	if reply.ServerStabilizationTime != 123*time.Millisecond {
		t.Fatalf("bad: %#v", reply)
	}
}
