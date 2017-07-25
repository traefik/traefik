package consul

import (
	"testing"
)

func TestConfig_GetTokenForAgent(t *testing.T) {
	config := DefaultConfig()
	if token := config.GetTokenForAgent(); token != "" {
		t.Fatalf("bad: %s", token)
	}
	config.ACLToken = "hello"
	if token := config.GetTokenForAgent(); token != "hello" {
		t.Fatalf("bad: %s", token)
	}
	config.ACLAgentToken = "world"
	if token := config.GetTokenForAgent(); token != "world" {
		t.Fatalf("bad: %s", token)
	}
}
