package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestOperator_Autopilot_Implements(t *testing.T) {
	var _ cli.Command = &OperatorAutopilotCommand{}
}
