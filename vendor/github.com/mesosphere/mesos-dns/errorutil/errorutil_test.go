package errorutil

import (
	"errors"
	"testing"
)

func returnsError() error {
	return errors.New("test")
}
func TestIgnoreHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("IgnoreHandler did not ignore Error")
		}
	}()

	Ignore(returnsError)

}
