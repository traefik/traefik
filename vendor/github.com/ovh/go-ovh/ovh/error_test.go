package ovh

import (
	"fmt"
	"net/http"
	"testing"
)

func TestErrorString(t *testing.T) {
	err := &APIError{
		Code:    http.StatusBadRequest,
		Message: "Bad request",
	}

	expected := `Error 400: "Bad request"`
	got := fmt.Sprintf("%s", err)

	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
