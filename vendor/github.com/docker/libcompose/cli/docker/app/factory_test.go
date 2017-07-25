package app

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/libcompose/project"
	"github.com/urfave/cli"
)

func TestProjectFactoryProjectNameIsNormalized(t *testing.T) {
	projects := []struct {
		name     string
		expected string
	}{
		{
			name:     "example",
			expected: "example",
		},
		{
			name:     "example-test",
			expected: "exampletest",
		},
		{
			name:     "aW3Ird_Project_with_$$",
			expected: "aw3irdprojectwith",
		},
	}

	tmpDir, err := ioutil.TempDir("", "project-factory-test")
	if err != nil {
		t.Fatal(err)
	}
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	ioutil.WriteFile(composeFile, []byte(`hello:
    image: busybox`), 0700)

	for _, projectCase := range projects {
		globalSet := flag.NewFlagSet("test", 0)
		// Set the project-name flag
		globalSet.String("project-name", projectCase.name, "doc")
		// Set the compose file flag
		globalSet.Var(&cli.StringSlice{composeFile}, "file", "doc")
		c := cli.NewContext(nil, globalSet, nil)
		factory := &ProjectFactory{}
		p, err := factory.Create(c)
		if err != nil {
			t.Fatal(err)
		}

		if p.(*project.Project).Name != projectCase.expected {
			t.Fatalf("expected %s, got %s", projectCase.expected, p.(*project.Project).Name)
		}
	}
}

func TestProjectFactoryFileArgMayContainMultipleFiles(t *testing.T) {
	sep := string(os.PathListSeparator)
	fileCases := []struct {
		requested []string
		available []string
		expected  []string
	}{
		{
			requested: []string{},
			available: []string{"docker-compose.yml"},
			expected:  []string{"docker-compose.yml"},
		},
		{
			requested: []string{},
			available: []string{"docker-compose.yml", "docker-compose.override.yml"},
			expected:  []string{"docker-compose.yml", "docker-compose.override.yml"},
		},
		{
			requested: []string{"one.yml"},
			available: []string{"one.yml"},
			expected:  []string{"one.yml"},
		},
		{
			requested: []string{"one.yml"},
			available: []string{"docker-compose.yml", "one.yml"},
			expected:  []string{"one.yml"},
		},
		{
			requested: []string{"one.yml", "two.yml", "three.yml"},
			available: []string{"one.yml", "two.yml", "three.yml"},
			expected:  []string{"one.yml", "two.yml", "three.yml"},
		},
		{
			requested: []string{"one.yml" + sep + "two.yml" + sep + "three.yml"},
			available: []string{"one.yml", "two.yml", "three.yml"},
			expected:  []string{"one.yml", "two.yml", "three.yml"},
		},
		{
			requested: []string{"one.yml" + sep + "two.yml", "three.yml" + sep + "four.yml"},
			available: []string{"one.yml", "two.yml", "three.yml", "four.yml"},
			expected:  []string{"one.yml", "two.yml", "three.yml", "four.yml"},
		},
		{
			requested: []string{"one.yml", "two.yml" + sep + "three.yml"},
			available: []string{"one.yml", "two.yml", "three.yml"},
			expected:  []string{"one.yml", "two.yml", "three.yml"},
		},
	}

	for _, fileCase := range fileCases {
		tmpDir, err := ioutil.TempDir("", "project-factory-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)
		if err = os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		for _, file := range fileCase.available {
			ioutil.WriteFile(file, []byte(`hello:
    image: busybox`), 0700)
		}
		globalSet := flag.NewFlagSet("test", 0)
		// Set the project-name flag
		globalSet.String("project-name", "example", "doc")
		// Set the compose file flag
		fcr := cli.StringSlice(fileCase.requested)
		globalSet.Var(&fcr, "file", "doc")
		c := cli.NewContext(nil, globalSet, nil)
		factory := &ProjectFactory{}
		p, err := factory.Create(c)
		if err != nil {
			t.Fatal(err)
		}

		for i, v := range p.(*project.Project).Files {
			if v != fileCase.expected[i] {
				t.Fatalf("requested %s, available %s, expected %s, got %s",
					fileCase.requested, fileCase.available, fileCase.expected, p.(*project.Project).Files)
			}
		}
	}
}
