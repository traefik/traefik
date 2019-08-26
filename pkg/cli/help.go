package cli

import (
	"io"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/containous/traefik/v2/pkg/config/flag"
	"github.com/containous/traefik/v2/pkg/config/generator"
	"github.com/containous/traefik/v2/pkg/config/parser"
)

const tmplHelp = `{{ .Cmd.Name }}	{{ .Cmd.Description }}

Usage: {{ .Cmd.Name }} [command] [flags] [arguments]

Use "{{ .Cmd.Name }} [command] --help" for help on any command.
{{if .SubCommands }}
Commands:
{{- range $i, $subCmd := .SubCommands }}
{{ if not $subCmd.Hidden }}	{{ $subCmd.Name }}	{{ $subCmd.Description }}{{end}}{{end}}
{{end}}
{{- if .Flags }}
Flag's usage: {{ .Cmd.Name }} [--flag=flag_argument] [-f [flag_argument]]	# set flag_argument to flag(s)
          or: {{ .Cmd.Name }} [--flag[=true|false| ]] [-f [true|false| ]]	# set true/false to boolean flag(s)

Flags:
{{- range $i, $flag := .Flags }}
	--{{ SliceIndexN $flag.Name }}  {{if ne $flag.Name "global.sendanonymoususage"}}(Default: "{{ $flag.Default}}"){{end}}
{{if $flag.Description }}		{{ wrapWith 80 "\n\t\t" $flag.Description }}
{{else}}
{{- end}}
{{- end}}
{{- end}}
`

func isHelp(args []string) bool {
	for _, name := range args {
		if name == "--help" || name == "-help" || name == "-h" {
			return true
		}
	}
	return false
}

// PrintHelp prints the help for the command given as argument.
func PrintHelp(w io.Writer, cmd *Command) error {
	var flags []parser.Flat
	if cmd.Configuration != nil {
		generator.Generate(cmd.Configuration)

		var err error
		flags, err = flag.Encode(cmd.Configuration)
		if err != nil {
			return err
		}
	}

	model := map[string]interface{}{
		"Cmd":         cmd,
		"Flags":       flags,
		"SubCommands": cmd.subCommands,
	}

	funcs := sprig.TxtFuncMap()
	funcs["SliceIndexN"] = sliceIndexN

	tmpl, err := template.New("flags").
		Funcs(funcs).
		Parse(tmplHelp)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 4, 0, 4, ' ', 0)

	err = tmpl.Execute(tw, model)
	if err != nil {
		return err
	}

	return tw.Flush()
}

func sliceIndexN(flag string) string {
	return strings.ReplaceAll(flag, "[0]", "[n]")
}
