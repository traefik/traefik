package acme

import (
	"testing"
	"time"
)

func TestWaitForTimeout(t *testing.T) {
	c := make(chan error)
	go func() {
		err := WaitFor(3*time.Second, 1*time.Second, func() (bool, error) {
			return false, nil
		})
		c <- err
	}()

	timeout := time.After(4 * time.Second)
	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		if err == nil {
			t.Errorf("expected timeout error; got %v", err)
		}
	}
}
