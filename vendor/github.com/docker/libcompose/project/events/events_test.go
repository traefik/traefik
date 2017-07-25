package events

import (
	"fmt"
	"testing"
)

func TestEventEquality(t *testing.T) {
	if fmt.Sprintf("%s", ServiceStart) != "Started" ||
		fmt.Sprintf("%v", ServiceStart) != "Started" {
		t.Fatalf("EventServiceStart String() doesn't work: %s %v", ServiceStart, ServiceStart)
	}

	if fmt.Sprintf("%s", ServiceStart) != fmt.Sprintf("%s", ServiceUp) {
		t.Fatal("Event messages do not match")
	}

	if ServiceStart == ServiceUp {
		t.Fatal("Events match")
	}
}
