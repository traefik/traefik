package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"text/template"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/version"
)

var versionTemplate = `Version:      {{.Version}}
Codename:     {{.Codename}}
Go version:   {{.GoVersion}}
Built:        {{.BuildTime}}
OS/Arch:      {{.Os}}/{{.Arch}}`

// newVersionCmd builds a new Version command
func newVersionCmd() *flaeg.Command {

	//version Command init
	return &flaeg.Command{
		Name:                  "version",
		Description:           `Print version`,
		Config:                struct{}{},
		DefaultPointersConfig: struct{}{},
		Run: func() error {
			if err := getVersionPrint(os.Stdout); err != nil {
				return err
			}
			fmt.Printf("\n")
			return nil

		},
	}
}

func getVersionPrint(wr io.Writer) error {
	tmpl, err := template.New("").Parse(versionTemplate)
	if err != nil {
		return err
	}

	v := struct {
		Version   string
		Codename  string
		GoVersion string
		BuildTime string
		Os        string
		Arch      string
	}{
		Version:   version.Version,
		Codename:  version.Codename,
		GoVersion: runtime.Version(),
		BuildTime: version.BuildDate,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	return tmpl.Execute(wr, v)
}
