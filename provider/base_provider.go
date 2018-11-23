package provider

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig"
	"github.com/containous/traefik/autogen/gentemplates"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

// BaseProvider should be inherited by providers.
type BaseProvider struct {
	Watch                     bool              `description:"Watch provider" export:"true"`
	Filename                  string            `description:"Override default configuration template. For advanced users :)" export:"true"`
	Constraints               types.Constraints `description:"Filter services by constraint, matching with Traefik tags." export:"true"`
	Trace                     bool              `description:"Display additional provider logs (if available)." export:"true"`
	DebugLogGeneratedTemplate bool              `description:"Enable debug logging of generated configuration template." export:"true"`
}

// Init for compatibility reason the BaseProvider implements an empty Init.
func (p *BaseProvider) Init() error {
	return nil
}

// MatchConstraints must match with EVERY single constraint
// returns first constraint that do not match or nil.
func (p *BaseProvider) MatchConstraints(tags []string) (bool, *types.Constraint) {
	// if there is no tags and no constraints, filtering is disabled
	if len(tags) == 0 && len(p.Constraints) == 0 {
		return true, nil
	}

	for _, constraint := range p.Constraints {
		// xor: if ok and constraint.MustMatch are equal, then no tag is currently matching with the constraint
		if ok := constraint.MatchConstraintWithAtLeastOneTag(tags); ok != constraint.MustMatch {
			return false, constraint
		}
	}

	// If no constraint or every constraints matching
	return true, nil
}

// GetConfiguration returns the provider configuration from default template (file or content) or overrode template file.
func (p *BaseProvider) GetConfiguration(defaultTemplate string, funcMap template.FuncMap, templateObjects interface{}) (*config.Configuration, error) {
	tmplContent, err := p.getTemplateContent(defaultTemplate)
	if err != nil {
		return nil, err
	}
	return p.CreateConfiguration(tmplContent, funcMap, templateObjects)
}

// CreateConfiguration creates a provider configuration from content using templating.
func (p *BaseProvider) CreateConfiguration(tmplContent string, funcMap template.FuncMap, templateObjects interface{}) (*config.Configuration, error) {
	var defaultFuncMap = sprig.TxtFuncMap()
	// tolower is deprecated in favor of sprig's lower function
	defaultFuncMap["tolower"] = strings.ToLower
	defaultFuncMap["normalize"] = Normalize
	defaultFuncMap["split"] = split
	for funcID, funcElement := range funcMap {
		defaultFuncMap[funcID] = funcElement
	}

	tmpl := template.New(p.Filename).Funcs(defaultFuncMap)

	_, err := tmpl.Parse(tmplContent)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		return nil, err
	}

	var renderedTemplate = buffer.String()
	if p.DebugLogGeneratedTemplate {
		log.Debugf("Template content: %s", tmplContent)
		log.Debugf("Rendering results: %s", renderedTemplate)
	}
	return p.DecodeConfiguration(renderedTemplate)
}

// DecodeConfiguration Decodes a *types.Configuration from a content.
func (p *BaseProvider) DecodeConfiguration(content string) (*config.Configuration, error) {
	configuration := new(config.Configuration)
	if _, err := toml.Decode(content, configuration); err != nil {
		return nil, err
	}
	return configuration, nil
}

func (p *BaseProvider) getTemplateContent(defaultTemplateFile string) (string, error) {
	if len(p.Filename) > 0 {
		buf, err := ioutil.ReadFile(p.Filename)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}

	if strings.HasSuffix(defaultTemplateFile, ".tmpl") {
		buf, err := gentemplates.Asset(defaultTemplateFile)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}

	return defaultTemplateFile, nil
}

func split(sep, s string) []string {
	return strings.Split(s, sep)
}

// Normalize transforms a string that work with the rest of traefik.
// Replace '.' with '-' in quoted keys because of this issue https://github.com/BurntSushi/toml/issues/78
func Normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}

// ReverseStringSlice inverts the order of the given slice of string.
func ReverseStringSlice(slice *[]string) {
	for i, j := 0, len(*slice)-1; i < j; i, j = i+1, j-1 {
		(*slice)[i], (*slice)[j] = (*slice)[j], (*slice)[i]
	}
}
