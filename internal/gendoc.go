package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/env"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/paerser/generator"
	"github.com/traefik/paerser/parser"
	"github.com/traefik/traefik/v3/cmd"
	"github.com/traefik/traefik/v3/pkg/collector/hydratation"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"gopkg.in/yaml.v3"
)

var commentGenerated = `## CODE GENERATED AUTOMATICALLY
## THIS FILE MUST NOT BE EDITED BY HAND
`

func main() {
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

	err = tomlWrite("./docs/content/reference/dynamic-configuration/file.toml", dynConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	err = yamlWrite("./docs/content/reference/dynamic-configuration/file.yaml", dynConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	err = labelsWrite("./docs/content/reference/dynamic-configuration", dynConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	staticConf := &static.Configuration{}

	err = hydratation.Hydrate(staticConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	delete(staticConf.EntryPoints, "EntryPoint1")

	err = tomlWrite("./docs/content/reference/static-configuration/file.toml", staticConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	err = yamlWrite("./docs/content/reference/static-configuration/file.yaml", staticConf)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	genStaticConfDoc("./docs/content/reference/static-configuration/env-ref.md", "", func(i interface{}) ([]parser.Flat, error) {
		return env.Encode(env.DefaultNamePrefix, i)
	})
	genStaticConfDoc("./docs/content/reference/static-configuration/cli-ref.md", "--", flag.Encode)
	genKVDynConfDoc("./docs/content/reference/dynamic-configuration/kv-ref.md")
}

func labelsWrite(outputDir string, element *dynamic.Configuration) error {
	cleanServers(element)

	etnOpts := parser.EncoderToNodeOpts{OmitEmpty: true, TagName: parser.TagLabel, AllowSliceAsStruct: true}
	node, err := parser.EncodeToNode(element, parser.DefaultRootName, etnOpts)
	if err != nil {
		return err
	}

	metaOpts := parser.MetadataOpts{TagName: parser.TagLabel, AllowSliceAsStruct: true}
	err = parser.AddMetadata(element, node, metaOpts)
	if err != nil {
		return err
	}

	labels := make(map[string]string)
	encodeNode(labels, node.Name, node)

	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	dockerLabels, err := os.Create(filepath.Join(outputDir, "docker-labels.yml"))
	if err != nil {
		return err
	}
	defer dockerLabels.Close()

	// Write the comment at the beginning of the file
	if _, err := dockerLabels.WriteString(commentGenerated); err != nil {
		return err
	}

	for _, k := range keys {
		v := labels[k]
		if v != "" {
			if v == "42000000000" {
				v = "42s"
			}
			fmt.Fprintln(dockerLabels, `- "`+strings.ToLower(k)+`=`+v+`"`)
		}
	}

	return nil
}

func cleanServers(element *dynamic.Configuration) {
	for _, svc := range element.HTTP.Services {
		if svc.LoadBalancer != nil {
			server := svc.LoadBalancer.Servers[0]
			svc.LoadBalancer.Servers = nil
			svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, server)
		}
	}

	for _, svc := range element.TCP.Services {
		if svc.LoadBalancer != nil {
			server := svc.LoadBalancer.Servers[0]
			svc.LoadBalancer.Servers = nil
			svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, server)
		}
	}

	for _, svc := range element.UDP.Services {
		if svc.LoadBalancer != nil {
			server := svc.LoadBalancer.Servers[0]
			svc.LoadBalancer.Servers = nil
			svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, server)
		}
	}
}

func yamlWrite(outputFile string, element any) error {
	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the comment at the beginning of the file
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

	// Write the comment at the beginning of the file
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
	w.writeln()

	for i, flat := range flats {
		// TODO must be move into the flats creation.
		if flat.Name == "experimental.plugins.<name>" || flat.Name == "TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>" {
			continue
		}

		if strings.HasPrefix(flat.Name, "pilot.") || strings.HasPrefix(flat.Name, "TRAEFIK_PILOT_") {
			continue
		}

		if prefix == "" {
			w.writeln("`" + prefix + strings.ReplaceAll(flat.Name, "[0]", "_n") + "`:  ")
		} else {
			w.writeln("`" + prefix + strings.ReplaceAll(flat.Name, "[0]", "[n]") + "`:  ")
		}

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

func genKVDynConfDoc(outputFile string) {
	dynConfPath := "./docs/content/reference/dynamic-configuration/file.toml"
	conf := map[string]interface{}{}
	_, err := toml.DecodeFile(dynConfPath, &conf)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	store := storeWriter{data: map[string]string{}}

	c := client{store: store}
	err = c.load("traefik", conf)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	var keys []string
	for k := range store.data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	_, _ = fmt.Fprintf(file, `<!--
CODE GENERATED AUTOMATICALLY
THIS FILE MUST NOT BE EDITED BY HAND
-->
`)

	_, _ = fmt.Fprintf(file, `
| Key (Path) | Value |
|------------|-------|
`)

	for _, k := range keys {
		_, _ = fmt.Fprintf(file, "| `%s` | `%s` |\n", k, store.data[k])
	}
}

type storeWriter struct {
	data map[string]string
}

func (f storeWriter) Put(key string, value []byte, _ []string) error {
	f.data[key] = string(value)
	return nil
}

type client struct {
	store storeWriter
}

func (c client) load(parentKey string, conf map[string]interface{}) error {
	for k, v := range conf {
		switch entry := v.(type) {
		case map[string]interface{}:
			key := path.Join(parentKey, k)

			if len(entry) == 0 {
				err := c.store.Put(key, nil, nil)
				if err != nil {
					return err
				}
			} else {
				err := c.load(key, entry)
				if err != nil {
					return err
				}
			}
		case []map[string]interface{}:
			for i, o := range entry {
				key := path.Join(parentKey, k, strconv.Itoa(i))

				if err := c.load(key, o); err != nil {
					return err
				}
			}
		case []interface{}:
			for i, o := range entry {
				key := path.Join(parentKey, k, strconv.Itoa(i))

				err := c.store.Put(key, []byte(fmt.Sprintf("%v", o)), nil)
				if err != nil {
					return err
				}
			}
		default:
			key := path.Join(parentKey, k)

			err := c.store.Put(key, []byte(fmt.Sprintf("%v", v)), nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
