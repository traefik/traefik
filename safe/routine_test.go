package safe

import (
	"github.com/cenk/backoff"
	"testing"
)

func TestOperationWithRecover(t *testing.T) {
	operation := func() error {
		panic("BOOM")
	}
	err := backoff.Retry(OperationWithRecover(operation), &backoff.StopBackOff{})
	if err == nil {
		t.Fatalf("Error in OperationWithRecover: %s", err)
	}
}
