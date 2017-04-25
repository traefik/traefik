package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"text/template"

	"github.com/containous/flaeg"
	"github.com/mvdan/xurls"
)

var (
	bugtracker  = "https://github.com/containous/traefik/issues/new"
	bugTemplate = `<!--
How to write a good issue?

Read https://github.com/containous/traefik/blob/master/.github/how-to-write-a-good-issue.md

And the contributing guide: https://github.com/containous/traefik/blob/master/.github/CONTRIBUTING.md
-->

### Do you want to request a *feature* or report a *bug*?


### What did you do?



### What did you expect to see?



### What did you see instead?



### Output of ` + "`" + `traefik version` + "`" + `: (_What version of Traefik are you using?_)

` + "```" + `
{{.Version}}
` + "```" + `

### What is your environment & configuration (arguments, toml, provider, platform, ...)?

` + "```" + `toml
{{.Configuration}}
` + "```" + `

<!--
Add more configuration information here.
-->

### If applicable, please paste the log output in debug mode (` + "`" + `--debug` + "`" + ` switch)

` + "```" + `
(paste your output here)
` + "```" + `

`
)

// newBugCmd builds a new Bug command
func newBugCmd(traefikConfiguration interface{}, traefikPointersConfiguration interface{}) *flaeg.Command {

	//version Command init
	return &flaeg.Command{
		Name:                  "bug",
		Description:           `Report an issue on Traefik bugtracker`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Run: func() error {
			var version bytes.Buffer
			if err := getVersionPrint(&version); err != nil {
				return err
			}

			tmpl, err := template.New("").Parse(bugTemplate)
			if err != nil {
				return err
			}

			configJSON, err := json.MarshalIndent(traefikConfiguration, "", " ")
			if err != nil {
				return err
			}

			v := struct {
				Version       string
				Configuration string
			}{
				Version:       version.String(),
				Configuration: anonymize(string(configJSON)),
			}

			var bug bytes.Buffer
			if err := tmpl.Execute(&bug, v); err != nil {
				return err
			}

			body := bug.String()
			URL := bugtracker + "?body=" + url.QueryEscape(body)
			if err := openBrowser(URL); err != nil {
				fmt.Print("Please file a new issue at " + bugtracker + " using this template:\n\n")
				fmt.Print(body)
			}

			return nil
		},
		Metadata: map[string]string{
			"parseAllSources": "true",
		},
	}
}

func openBrowser(URL string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", URL).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", URL).Start()
	case "darwin":
		err = exec.Command("open", URL).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func anonymize(input string) string {
	replace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	mailExp := regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}"`)
	return xurls.Relaxed.ReplaceAllString(mailExp.ReplaceAllString(input, replace), replace)
}
