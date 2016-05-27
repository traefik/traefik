package provider

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/autogen"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"unicode"
)

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to traefik
	// using the given configuration channel.
	Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error
}

// BaseProvider should be inherited by providers
type BaseProvider struct {
	Watch    bool   `description:"Watch provider"`
	Filename string `description:"Override default configuration template. For advanced users :)"`
}

func (p *BaseProvider) getConfiguration(defaultTemplateFile string, funcMap template.FuncMap, templateObjects interface{}) (*types.Configuration, error) {
	var (
		buf []byte
		err error
	)
	configuration := new(types.Configuration)
	tmpl := template.New(p.Filename).Funcs(funcMap)
	if len(p.Filename) > 0 {
		buf, err = ioutil.ReadFile(p.Filename)
		if err != nil {
			return nil, err
		}
	} else {
		buf, err = autogen.Asset(defaultTemplateFile)
		if err != nil {
			return nil, err
		}
	}
	_, err = tmpl.Parse(string(buf))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(buffer.String(), configuration); err != nil {
		return nil, err
	}
	return configuration, nil
}

func replace(s1 string, s2 string, s3 string) string {
	return strings.Replace(s3, s1, s2, -1)
}

// Escape beginning slash "/", convert all others to dash "-"
func getEscapedName(name string) string {
	return strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1)
}

func normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}
