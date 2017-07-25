package command

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func testKVGetCommand(t *testing.T) (*cli.MockUi, *KVGetCommand) {
	ui := new(cli.MockUi)
	return ui, &KVGetCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}
}

func TestKVGetCommand_implements(t *testing.T) {
	var _ cli.Command = &KVGetCommand{}
}

func TestKVGetCommand_noTabs(t *testing.T) {
	assertNoTabs(t, new(KVGetCommand))
}

func TestKVGetCommand_Validation(t *testing.T) {
	ui, c := testKVGetCommand(t)

	cases := map[string]struct {
		args   []string
		output string
	}{
		"no key": {
			[]string{},
			"Missing KEY argument",
		},
		"extra args": {
			[]string{"foo", "bar", "baz"},
			"Too many arguments",
		},
	}

	for name, tc := range cases {
		// Ensure our buffer is always clear
		if ui.ErrorWriter != nil {
			ui.ErrorWriter.Reset()
		}
		if ui.OutputWriter != nil {
			ui.OutputWriter.Reset()
		}

		code := c.Run(tc.args)
		if code == 0 {
			t.Errorf("%s: expected non-zero exit", name)
		}

		output := ui.ErrorWriter.String()
		if !strings.Contains(output, tc.output) {
			t.Errorf("%s: expected %q to contain %q", name, output, tc.output)
		}
	}
}

func TestKVGetCommand_Run(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVGetCommand(t)

	pair := &api.KVPair{
		Key:   "foo",
		Value: []byte("bar"),
	}
	_, err := client.KV().Put(pair, nil)
	if err != nil {
		t.Fatalf("err: %#v", err)
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
	if !strings.Contains(output, "bar") {
		t.Errorf("bad: %#v", output)
	}
}

func TestKVGetCommand_Missing(t *testing.T) {
	srv, _ := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	_, c := testKVGetCommand(t)

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"not-a-real-key",
	}

	code := c.Run(args)
	if code == 0 {
		t.Fatalf("expected bad code")
	}
}

func TestKVGetCommand_Empty(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVGetCommand(t)

	pair := &api.KVPair{
		Key:   "empty",
		Value: []byte(""),
	}
	_, err := client.KV().Put(pair, nil)
	if err != nil {
		t.Fatalf("err: %#v", err)
	}

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"empty",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}
}

func TestKVGetCommand_Detailed(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVGetCommand(t)

	pair := &api.KVPair{
		Key:   "foo",
		Value: []byte("bar"),
	}
	_, err := client.KV().Put(pair, nil)
	if err != nil {
		t.Fatalf("err: %#v", err)
	}

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"-detailed",
		"foo",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	output := ui.OutputWriter.String()
	for _, key := range []string{
		"CreateIndex",
		"LockIndex",
		"ModifyIndex",
		"Flags",
		"Session",
		"Value",
	} {
		if !strings.Contains(output, key) {
			t.Fatalf("bad %#v, missing %q", output, key)
		}
	}
}

func TestKVGetCommand_Keys(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVGetCommand(t)

	keys := []string{"foo/bar", "foo/baz", "foo/zip"}
	for _, key := range keys {
		if _, err := client.KV().Put(&api.KVPair{Key: key}, nil); err != nil {
			t.Fatalf("err: %#v", err)
		}
	}

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"-keys",
		"foo/",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	output := ui.OutputWriter.String()
	for _, key := range keys {
		if !strings.Contains(output, key) {
			t.Fatalf("bad %#v missing %q", output, key)
		}
	}
}

func TestKVGetCommand_Recurse(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVGetCommand(t)

	keys := map[string]string{
		"foo/a": "a",
		"foo/b": "b",
		"foo/c": "c",
	}
	for k, v := range keys {
		pair := &api.KVPair{Key: k, Value: []byte(v)}
		if _, err := client.KV().Put(pair, nil); err != nil {
			t.Fatalf("err: %#v", err)
		}
	}

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"-recurse",
		"foo",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	output := ui.OutputWriter.String()
	for key, value := range keys {
		if !strings.Contains(output, key+":"+value) {
			t.Fatalf("bad %#v missing %q", output, key)
		}
	}
}

func TestKVGetCommand_RecurseBase64(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVGetCommand(t)

	keys := map[string]string{
		"foo/a": "Hello World 1",
		"foo/b": "Hello World 2",
		"foo/c": "Hello World 3",
	}
	for k, v := range keys {
		pair := &api.KVPair{Key: k, Value: []byte(v)}
		if _, err := client.KV().Put(pair, nil); err != nil {
			t.Fatalf("err: %#v", err)
		}
	}

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"-recurse",
		"-base64",
		"foo",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	output := ui.OutputWriter.String()
	for key, value := range keys {
		if !strings.Contains(output, key+":"+base64.StdEncoding.EncodeToString([]byte(value))) {
			t.Fatalf("bad %#v missing %q", output, key)
		}
	}
}

func TestKVGetCommand_DetailedBase64(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVGetCommand(t)

	pair := &api.KVPair{
		Key:   "foo",
		Value: []byte("bar"),
	}
	_, err := client.KV().Put(pair, nil)
	if err != nil {
		t.Fatalf("err: %#v", err)
	}

	args := []string{
		"-http-addr=" + srv.httpAddr,
		"-detailed",
		"-base64",
		"foo",
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	output := ui.OutputWriter.String()
	for _, key := range []string{
		"CreateIndex",
		"LockIndex",
		"ModifyIndex",
		"Flags",
		"Session",
		"Value",
	} {
		if !strings.Contains(output, key) {
			t.Fatalf("bad %#v, missing %q", output, key)
		}
	}

	if !strings.Contains(output, base64.StdEncoding.EncodeToString([]byte("bar"))) {
		t.Fatalf("bad %#v, value is not base64 encoded", output)
	}
}
