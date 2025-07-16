package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/paerser/generator"
	"github.com/traefik/paerser/parser"
	"github.com/traefik/traefik/v3/cmd"
)

func main() {
	genStaticConfDoc("./docs/content/reference/install-configuration/configuration-options.md", "", flag.Encode)
}

func genStaticConfDoc(outputFile, prefix string, encodeFn func(interface{}) ([]parser.Flat, error)) {
	logger := log.With().Str("file", outputFile).Logger()

	element := &cmd.NewTraefikConfiguration().Configuration

	generator.Generate(element)

	flats, err := encodeFn(element)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	err = os.RemoveAll(outputFile)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	defer file.Close()

	w := errWriter{w: file}

	w.writeln(`<!--
CODE GENERATED AUTOMATICALLY
THIS FILE MUST NOT BE EDITED BY HAND
-->`)
	w.writeln(`# Install Configuration Options
`)
	w.writeln(`## Configuration Options`)

	w.writeln(`
| Field | Description | Default | 
|:-------|:------------|:-------|`)

	for _, flat := range flats {
		// TODO must be move into the flats creation.
		if flat.Name == "experimental.plugins.<name>" || flat.Name == "TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>" {
			continue
		}

		if strings.HasPrefix(flat.Name, "pilot.") || strings.HasPrefix(flat.Name, "TRAEFIK_PILOT_") {
			continue
		}

		line := "| " + strings.ReplaceAll(strings.ReplaceAll(flat.Name, "<", "_"), ">", "_") + " | " + flat.Description + " | "

		if flat.Default == "" {
			line += "|"
		} else {
			line += flat.Default + " |"
		}

		w.writeln(line)

	}

	if w.err != nil {
		logger.Fatal().Err(err).Send()
	}
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) writeln(a ...interface{}) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintln(ew.w, a...)
}
