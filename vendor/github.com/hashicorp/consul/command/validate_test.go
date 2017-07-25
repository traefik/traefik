package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

func testValidateCommand(t *testing.T) (*cli.MockUi, *ValidateCommand) {
	ui := new(cli.MockUi)
	return ui, &ValidateCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetNone,
		},
	}
}

func TestValidateCommand_implements(t *testing.T) {
	var _ cli.Command = &ValidateCommand{}
}

func TestValidateCommandFailOnEmptyFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "consul")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(tmpFile.Name())

	_, cmd := testValidateCommand(t)

	args := []string{tmpFile.Name()}

	if code := cmd.Run(args); code == 0 {
		t.Fatalf("bad: %d", code)
	}
}

func TestValidateCommandSucceedOnEmptyDir(t *testing.T) {
	td, err := ioutil.TempDir("", "consul")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	ui, cmd := testValidateCommand(t)

	args := []string{td}

	if code := cmd.Run(args); code != 0 {
		t.Fatalf("bad: %d, %s", code, ui.ErrorWriter.String())
	}
}

func TestValidateCommandSucceedOnMinimalConfigFile(t *testing.T) {
	td, err := ioutil.TempDir("", "consul")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	fp := filepath.Join(td, "config.json")
	err = ioutil.WriteFile(fp, []byte(`{}`), 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, cmd := testValidateCommand(t)

	args := []string{fp}

	if code := cmd.Run(args); code != 0 {
		t.Fatalf("bad: %d", code)
	}
}

func TestValidateCommandSucceedOnMinimalConfigDir(t *testing.T) {
	td, err := ioutil.TempDir("", "consul")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	err = ioutil.WriteFile(filepath.Join(td, "config.json"), []byte(`{}`), 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, cmd := testValidateCommand(t)

	args := []string{td}

	if code := cmd.Run(args); code != 0 {
		t.Fatalf("bad: %d", code)
	}
}

func TestValidateCommandSucceedOnConfigDirWithEmptyFile(t *testing.T) {
	td, err := ioutil.TempDir("", "consul")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	err = ioutil.WriteFile(filepath.Join(td, "config.json"), []byte{}, 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, cmd := testValidateCommand(t)

	args := []string{td}

	if code := cmd.Run(args); code != 0 {
		t.Fatalf("bad: %d", code)
	}
}

func TestValidateCommandQuiet(t *testing.T) {
	td, err := ioutil.TempDir("", "consul")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	ui, cmd := testValidateCommand(t)

	args := []string{"-quiet", td}

	if code := cmd.Run(args); code != 0 {
		t.Fatalf("bad: %d, %s", code, ui.ErrorWriter.String())
	}
	if ui.OutputWriter.String() != "<nil>" {
		t.Fatalf("bad: %v", ui.OutputWriter.String())
	}
}
