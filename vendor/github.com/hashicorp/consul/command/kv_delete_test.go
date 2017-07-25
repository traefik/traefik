package command

import (
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func testKVDeleteCommand(t *testing.T) (*cli.MockUi, *KVDeleteCommand) {
	ui := new(cli.MockUi)
	return ui, &KVDeleteCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}
}

func TestKVDeleteCommand_implements(t *testing.T) {
	var _ cli.Command = &KVDeleteCommand{}
}

func TestKVDeleteCommand_noTabs(t *testing.T) {
	assertNoTabs(t, new(KVDeleteCommand))
}

func TestKVDeleteCommand_Validation(t *testing.T) {
	ui, c := testKVDeleteCommand(t)

	cases := map[string]struct {
		args   []string
		output string
	}{
		"-cas and -recurse": {
			[]string{"-cas", "-modify-index", "2", "-recurse", "foo"},
			"Cannot specify both",
		},
		"-cas no -modify-index": {
			[]string{"-cas", "foo"},
			"Must specify -modify-index",
		},
		"-modify-index no -cas": {
			[]string{"-modify-index", "2", "foo"},
			"Cannot specify -modify-index without",
		},
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

func TestKVDeleteCommand_Run(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVDeleteCommand(t)

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

	pair, _, err = client.KV().Get("foo", nil)
	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if pair != nil {
		t.Fatalf("bad: %#v", pair)
	}
}

func TestKVDeleteCommand_Recurse(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVDeleteCommand(t)

	keys := []string{"foo/a", "foo/b", "food"}

	for _, k := range keys {
		pair := &api.KVPair{
			Key:   k,
			Value: []byte("bar"),
		}
		_, err := client.KV().Put(pair, nil)
		if err != nil {
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

	for _, k := range keys {
		pair, _, err := client.KV().Get(k, nil)
		if err != nil {
			t.Fatalf("err: %#v", err)
		}
		if pair != nil {
			t.Fatalf("bad: %#v", pair)
		}
	}
}

func TestKVDeleteCommand_CAS(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testKVDeleteCommand(t)

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
		"-cas",
		"-modify-index", "1",
		"foo",
	}

	code := c.Run(args)
	if code == 0 {
		t.Fatalf("bad: expected error")
	}

	data, _, err := client.KV().Get("foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Reset buffers
	ui.OutputWriter.Reset()
	ui.ErrorWriter.Reset()

	args = []string{
		"-http-addr=" + srv.httpAddr,
		"-cas",
		"-modify-index", strconv.FormatUint(data.ModifyIndex, 10),
		"foo",
	}

	code = c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	data, _, err = client.KV().Get("foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Fatalf("bad: %#v", data)
	}
}
