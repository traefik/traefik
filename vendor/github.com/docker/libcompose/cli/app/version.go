package app

import (
	"fmt"
	"os"
	"runtime"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/version"
	"github.com/urfave/cli"
)

var versionTemplate = `Version:      {{.Version}} ({{.GitCommit}})
Go version:   {{.GoVersion}}
Built:        {{.BuildTime}}
OS/Arch:      {{.Os}}/{{.Arch}}`

// Version prints the libcompose version number and additionnal informations.
func Version(c *cli.Context) error {
	if c.Bool("short") {
		fmt.Println(version.VERSION)
		return nil
	}

	tmpl, err := template.New("").Parse(versionTemplate)
	if err != nil {
		logrus.Fatal(err)
	}

	v := struct {
		Version   string
		GitCommit string
		GoVersion string
		BuildTime string
		Os        string
		Arch      string
	}{
		Version:   version.VERSION,
		GitCommit: version.GITCOMMIT,
		GoVersion: runtime.Version(),
		BuildTime: version.BUILDTIME,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	if err := tmpl.Execute(os.Stdout, v); err != nil {
		logrus.Fatal(err)
	}
	fmt.Printf("\n")
	return nil
}
