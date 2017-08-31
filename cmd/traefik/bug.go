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
DO NOT FILE ISSUES FOR GENERAL SUPPORT QUESTIONS.

The issue tracker is for reporting bugs and feature requests only.
For end-user related support questions, refer to one of the following:

- Stack Overflow (using the "traefik" tag): https://stackoverflow.com/questions/tagged/traefik
- the Traefik community Slack channel: https://traefik.herokuapp.com

-->

### Do you want to request a *feature* or report a *bug*?

(If you intend to ask a support question: **DO NOT FILE AN ISSUE**.
Use [Stack Overflow](https://stackoverflow.com/questions/tagged/traefik)
or [Slack](https://traefik.herokuapp.com) instead.)



### What did you do?

<!--

HOW TO WRITE A GOOD ISSUE?

- Respect the issue template as more as possible.
- If it's possible use the command ` + "`" + "traefik bug" + "`" + `. See https://www.youtube.com/watch?v=Lyz62L8m93I.
- The title must be short and descriptive.
- Explain the conditions which led you to write this issue: the context.
- The context should lead to something, an idea or a problem that you’re facing.
- Remain clear and concise.
- Format your messages to help the reader focus on what matters and understand the structure of your message, use Markdown syntax https://help.github.com/articles/github-flavored-markdown

-->


### What did you expect to see?



### What did you see instead?



### Output of ` + "`" + `traefik version` + "`" + `: (_What version of Traefik are you using?_)

` + "```" + `
{{.Version}}
` + "```" + `

### What is your environment & configuration (arguments, toml, provider, platform, ...)?

` + "```" + `json
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
				fmt.Printf("Please file a new issue at %s using this template:\n\n", bugtracker)
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
	replace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx\""
	mailExp := regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}"`)
	return xurls.Relaxed.ReplaceAllString(mailExp.ReplaceAllString(input, replace), replace)
}
