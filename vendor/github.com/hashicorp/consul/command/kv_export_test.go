package command

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func TestKVExportCommand_Run(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui := new(cli.MockUi)
	c := KVExportCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}

	keys := map[string]string{
		"foo/a": "a",
		"foo/b": "b",
		"foo/c": "c",
		"bar":   "d",
	}
	for k, v := range keys {
		pair := &api.KVPair{Key: k, Value: []byte(v)}
		if _, err := client.KV().Put(pair, nil); err != nil {
			t.Fatalf("err: %#v", err)
		}
	}

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"foo",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	output := ui.OutputWriter.String()

	var exported []*kvExportEntry
	err := json.Unmarshal([]byte(output), &exported)
	if err != nil {
		t.Fatalf("bad: %d", code)
	}

	if len(exported) != 3 {
		t.Fatalf("bad: expected 3, got %d", len(exported))
	}

	for _, entry := range exported {
		if base64.StdEncoding.EncodeToString([]byte(keys[entry.Key])) != entry.Value {
			t.Fatalf("bad: expected %s, got %s", keys[entry.Key], entry.Value)
		}
	}
}
