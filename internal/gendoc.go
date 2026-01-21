package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/paerser/generator"
	"github.com/traefik/traefik/v3/cmd"
	"github.com/traefik/traefik/v3/pkg/collector/hydratation"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"gopkg.in/yaml.v3"
)

var commentGenerated = `## CODE GENERATED AUTOMATICALLY
## THIS FILE MUST NOT BE EDITED BY HAND
`

func main() {
	genRoutingConfDoc()
	genInstallConfDoc()
	genAnchors()
}

// Generate the Routing Configuration YAML and TOML files.
func genRoutingConfDoc() {
	logger := log.With().Logger()

	dynConf := &dynamic.Configuration{}

	err := hydratation.Hydrate(dynConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	dynConf.HTTP.Models = map[string]*dynamic.Model{}
	clean(dynConf.HTTP.Middlewares)
	clean(dynConf.TCP.Middlewares)
	clean(dynConf.HTTP.Services)
	clean(dynConf.TCP.Services)
	clean(dynConf.UDP.Services)

	err = tomlWrite("./docs/content/reference/routing-configuration/other-providers/file.toml", dynConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	err = yamlWrite("./docs/content/reference/routing-configuration/other-providers/file.yaml", dynConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
}

func yamlWrite(outputFile string, element any) error {
	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the comment at the beginning of the file.
	if _, err := file.WriteString(commentGenerated); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	err = encoder.Encode(element)
	if err != nil {
		return err
	}

	_, err = file.Write(buf.Bytes())
	return err
}

func tomlWrite(outputFile string, element any) error {
	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the comment at the beginning of the file.
	if _, err := file.WriteString(commentGenerated); err != nil {
		return err
	}

	return toml.NewEncoder(file).Encode(element)
}

func clean(element any) {
	valSvcs := reflect.ValueOf(element)

	key := valSvcs.MapKeys()[0]
	valueSvcRoot := valSvcs.MapIndex(key).Elem()

	var svcFieldNames []string
	for i := range valueSvcRoot.NumField() {
		field := valueSvcRoot.Type().Field(i)
		// do not create empty node for hidden config.
		if field.Tag.Get("file") == "-" && field.Tag.Get("kv") == "-" && field.Tag.Get("label") == "-" {
			continue
		}

		svcFieldNames = append(svcFieldNames, field.Name)
	}

	sort.Strings(svcFieldNames)

	for i, fieldName := range svcFieldNames {
		v := reflect.New(valueSvcRoot.Type())
		v.Elem().FieldByName(fieldName).Set(valueSvcRoot.FieldByName(fieldName))

		valSvcs.SetMapIndex(reflect.ValueOf(fmt.Sprintf("%s%.2d", valueSvcRoot.Type().Name(), i+1)), v)
	}

	valSvcs.SetMapIndex(reflect.ValueOf(fmt.Sprintf("%s0", valueSvcRoot.Type().Name())), reflect.Value{})
	valSvcs.SetMapIndex(reflect.ValueOf(fmt.Sprintf("%s1", valueSvcRoot.Type().Name())), reflect.Value{})
}

// Generate the Install Configuration in a table.
func genInstallConfDoc() {
	outputFile := "./docs/content/reference/install-configuration/configuration-options.md"
	logger := log.With().Str("file", outputFile).Logger()

	element := &cmd.NewTraefikConfiguration().Configuration

	generator.Generate(element)

	flats, err := flag.Encode(element)
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
	w.writeln(`# Install Configuration Options`)
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

func (ew *errWriter) writeln(a ...any) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintln(ew.w, a...)
}
