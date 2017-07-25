package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestOperator_Implements(t *testing.T) {
	var _ cli.Command = &OperatorCommand{}
}
