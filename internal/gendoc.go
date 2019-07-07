package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/containous/traefik/pkg/config/env"
	"github.com/containous/traefik/pkg/config/flag"
	"github.com/containous/traefik/pkg/config/generator"
	"github.com/containous/traefik/pkg/config/parser"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/log"
)

func main() {
	genStaticConfDoc("./docs/content/reference/static-configuration/env-ref.md", "", env.Encode)
	genStaticConfDoc("./docs/content/reference/static-configuration/cli-ref.md", "--", flag.Encode)
}

func genStaticConfDoc(outputFile string, prefix string, encodeFn func(interface{}) ([]parser.Flat, error)) {
	logger := log.WithoutContext().WithField("file", outputFile)

	element := &static.Configuration{}

	generator.Generate(element)

	flats, err := encodeFn(element)
	if err != nil {
		logger.Fatal(err)
	}

	err = os.RemoveAll(outputFile)
	if err != nil {
		logger.Fatal(err)
	}

	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Fatal(err)
	}

	defer file.Close()

	w := errWriter{w: file}

	w.writeln(`<!--
CODE GENERATED AUTOMATICALLY
THIS FILE MUST NOT BE EDITED BY HAND
-->
`)

	for i, flat := range flats {
		w.writeln("`" + prefix + strings.ReplaceAll(flat.Name, "[0]", "[n]") + "`:  ")
		if flat.Default == "" {
			w.writeln(flat.Description)
		} else {
			w.writeln(flat.Description + " (Default: ```" + flat.Default + "```)")
		}

		if i < len(flats)-1 {
			w.writeln()
		}
	}

	if w.err != nil {
		logger.Fatal(err)
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
