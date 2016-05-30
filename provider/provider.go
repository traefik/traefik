package provider

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/autogen"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to traefik
	// using the given configuration channel.
	Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints []*types.Constraint) error
}

// BaseProvider should be inherited by providers
type BaseProvider struct {
	Watch    bool   `description:"Watch provider"`
	Filename string `description:"Override default configuration template. For advanced users :)"`
	Constraints []*types.Constraint `description:"Filter services by constraint, matching with Traefik tags."`
}

// MatchConstraints must match with EVERY single contraint
// returns first constraint that do not match or nil
// returns errors for future use (regex)
func (p *BaseProvider) MatchConstraints(tags []string) (bool, *types.Constraint, error) {
	// if there is no tags and no contraints, filtering is disabled
	if len(tags) == 0 && len(p.Constraints) == 0 {
		return true, nil, nil
	}

	for _, constraint := range p.Constraints {
		if ok := constraint.MatchConstraintWithAtLeastOneTag(tags); xor(ok == true, constraint.MustMatch == true) {
			return false, constraint, nil
		}
	}

	// If no constraint or every constraints matching
	return true, nil, nil
>>>>>>> e844462... feat(constraints): Implementation of constraints (cmd + toml + matching functions), implementation proposal with consul
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

// golang does not support ^ operator
func xor(cond1 bool, cond2 bool) bool {
	return cond1 != cond2
}
