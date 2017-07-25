package command

import (
	"testing"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func testOperatorRaftCommand(t *testing.T) (*cli.MockUi, *OperatorRaftCommand) {
	ui := new(cli.MockUi)
	return ui, &OperatorRaftCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}
}

func TestOperator_Raft_Implements(t *testing.T) {
	var _ cli.Command = &OperatorRaftCommand{}
}
