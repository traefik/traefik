package command

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func testSnapshotSaveCommand(t *testing.T) (*cli.MockUi, *SnapshotSaveCommand) {
	ui := new(cli.MockUi)
	return ui, &SnapshotSaveCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}
}

func TestSnapshotSaveCommand_implements(t *testing.T) {
	var _ cli.Command = &SnapshotSaveCommand{}
}

func TestSnapshotSaveCommand_noTabs(t *testing.T) {
	assertNoTabs(t, new(SnapshotSaveCommand))
}

func TestSnapshotSaveCommand_Validation(t *testing.T) {
	ui, c := testSnapshotSaveCommand(t)

	cases := map[string]struct {
		args   []string
		output string
	}{
		"no file": {
			[]string{},
			"Missing FILE argument",
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

func TestSnapshotSaveCommand_Run(t *testing.T) {
	srv, client := testAgentWithAPIClient(t)
	defer srv.Shutdown()
	waitForLeader(t, srv.httpAddr)

	ui, c := testSnapshotSaveCommand(t)

	dir, err := ioutil.TempDir("", "snapshot")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer os.RemoveAll(dir)

	file := path.Join(dir, "backup.tgz")
	args := []string{
		"-http-addr=" + srv.httpAddr,
		file,
	}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer f.Close()

	if err := client.Snapshot().Restore(nil, f); err != nil {
		t.Fatalf("err: %v", err)
	}
}
