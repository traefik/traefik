package command

import (
	"strings"
	"testing"

	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func testKeyringCommand(t *testing.T) (*cli.MockUi, *KeyringCommand) {
	ui := new(cli.MockUi)
	return ui, &KeyringCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetClientHTTP,
		},
	}
}

func TestKeyringCommand_implements(t *testing.T) {
	var _ cli.Command = &KeyringCommand{}
}

func TestKeyringCommandRun(t *testing.T) {
	key1 := "HS5lJ+XuTlYKWaeGYyG+/A=="
	key2 := "kZyFABeAmc64UMTrm9XuKA=="

	// Begin with a single key
	a1 := testAgentWithConfig(t, func(c *agent.Config) {
		c.EncryptKey = key1
	})
	defer a1.Shutdown()

	// The LAN and WAN keyrings were initialized with key1
	out := listKeys(t, a1.httpAddr)
	if !strings.Contains(out, "dc1 (LAN):\n  "+key1) {
		t.Fatalf("bad: %#v", out)
	}
	if !strings.Contains(out, "WAN:\n  "+key1) {
		t.Fatalf("bad: %#v", out)
	}
	if strings.Contains(out, key2) {
		t.Fatalf("bad: %#v", out)
	}

	// Install the second key onto the keyring
	installKey(t, a1.httpAddr, key2)

	// Both keys should be present
	out = listKeys(t, a1.httpAddr)
	for _, key := range []string{key1, key2} {
		if !strings.Contains(out, key) {
			t.Fatalf("bad: %#v", out)
		}
	}

	// Rotate to key2, remove key1
	useKey(t, a1.httpAddr, key2)
	removeKey(t, a1.httpAddr, key1)

	// Only key2 is present now
	out = listKeys(t, a1.httpAddr)
	if !strings.Contains(out, "dc1 (LAN):\n  "+key2) {
		t.Fatalf("bad: %#v", out)
	}
	if !strings.Contains(out, "WAN:\n  "+key2) {
		t.Fatalf("bad: %#v", out)
	}
	if strings.Contains(out, key1) {
		t.Fatalf("bad: %#v", out)
	}
}

func TestKeyringCommandRun_help(t *testing.T) {
	ui, c := testKeyringCommand(t)
	code := c.Run(nil)
	if code != 1 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	// Test that we didn't actually try to dial the RPC server.
	if !strings.Contains(ui.ErrorWriter.String(), "Usage:") {
		t.Fatalf("bad: %#v", ui.ErrorWriter.String())
	}
}

func TestKeyringCommandRun_failedConnection(t *testing.T) {
	ui, c := testKeyringCommand(t)
	args := []string{"-list", "-http-addr=127.0.0.1:0"}
	code := c.Run(args)
	if code != 1 {
		t.Fatalf("bad: %d, %#v", code, ui.ErrorWriter.String())
	}
	if !strings.Contains(ui.ErrorWriter.String(), "dial") {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}

func TestKeyringCommandRun_invalidRelayFactor(t *testing.T) {
	ui, c := testKeyringCommand(t)

	args := []string{"-list", "-relay-factor=6"}
	code := c.Run(args)
	if code != 1 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}
}

func listKeys(t *testing.T, addr string) string {
	ui, c := testKeyringCommand(t)

	args := []string{"-list", "-http-addr=" + addr}
	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	return ui.OutputWriter.String()
}

func installKey(t *testing.T, addr string, key string) {
	ui, c := testKeyringCommand(t)

	args := []string{"-install=" + key, "-http-addr=" + addr}
	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}
}

func useKey(t *testing.T, addr string, key string) {
	ui, c := testKeyringCommand(t)

	args := []string{"-use=" + key, "-http-addr=" + addr}
	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}
}

func removeKey(t *testing.T, addr string, key string) {
	ui, c := testKeyringCommand(t)

	args := []string{"-remove=" + key, "-http-addr=" + addr}
	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}
}
