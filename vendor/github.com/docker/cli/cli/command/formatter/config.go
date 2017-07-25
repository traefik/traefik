package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/inspect"
	"github.com/docker/docker/api/types/swarm"
	units "github.com/docker/go-units"
)

const (
	defaultConfigTableFormat           = "table {{.ID}}\t{{.Name}}\t{{.CreatedAt}}\t{{.UpdatedAt}}"
	configIDHeader                     = "ID"
	configCreatedHeader                = "CREATED"
	configUpdatedHeader                = "UPDATED"
	configInspectPrettyTemplate Format = `ID:			{{.ID}}
Name:			{{.Name}}
{{- if .Labels }}
Labels:
{{- range $k, $v := .Labels }}
 - {{ $k }}{{if $v }}={{ $v }}{{ end }}
{{- end }}{{ end }}
Created at:            	{{.CreatedAt}}
Updated at:            	{{.UpdatedAt}}
Data:
{{.Data}}`
)

// NewConfigFormat returns a Format for rendering using a config Context
func NewConfigFormat(source string, quiet bool) Format {
	switch source {
	case PrettyFormatKey:
		return configInspectPrettyTemplate
	case TableFormatKey:
		if quiet {
			return defaultQuietFormat
		}
		return defaultConfigTableFormat
	}
	return Format(source)
}

// ConfigWrite writes the context
func ConfigWrite(ctx Context, configs []swarm.Config) error {
	render := func(format func(subContext subContext) error) error {
		for _, config := range configs {
			configCtx := &configContext{c: config}
			if err := format(configCtx); err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(newConfigContext(), render)
}

func newConfigContext() *configContext {
	cCtx := &configContext{}

	cCtx.header = map[string]string{
		"ID":        configIDHeader,
		"Name":      nameHeader,
		"CreatedAt": configCreatedHeader,
		"UpdatedAt": configUpdatedHeader,
		"Labels":    labelsHeader,
	}
	return cCtx
}

type configContext struct {
	HeaderContext
	c swarm.Config
}

func (c *configContext) MarshalJSON() ([]byte, error) {
	return marshalJSON(c)
}

func (c *configContext) ID() string {
	return c.c.ID
}

func (c *configContext) Name() string {
	return c.c.Spec.Annotations.Name
}

func (c *configContext) CreatedAt() string {
	return units.HumanDuration(time.Now().UTC().Sub(c.c.Meta.CreatedAt)) + " ago"
}

func (c *configContext) UpdatedAt() string {
	return units.HumanDuration(time.Now().UTC().Sub(c.c.Meta.UpdatedAt)) + " ago"
}

func (c *configContext) Labels() string {
	mapLabels := c.c.Spec.Annotations.Labels
	if mapLabels == nil {
		return ""
	}
	var joinLabels []string
	for k, v := range mapLabels {
		joinLabels = append(joinLabels, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(joinLabels, ",")
}

func (c *configContext) Label(name string) string {
	if c.c.Spec.Annotations.Labels == nil {
		return ""
	}
	return c.c.Spec.Annotations.Labels[name]
}

// ConfigInspectWrite renders the context for a list of configs
func ConfigInspectWrite(ctx Context, refs []string, getRef inspect.GetRefFunc) error {
	if ctx.Format != configInspectPrettyTemplate {
		return inspect.Inspect(ctx.Output, refs, string(ctx.Format), getRef)
	}
	render := func(format func(subContext subContext) error) error {
		for _, ref := range refs {
			configI, _, err := getRef(ref)
			if err != nil {
				return err
			}
			config, ok := configI.(swarm.Config)
			if !ok {
				return fmt.Errorf("got wrong object to inspect :%v", ok)
			}
			if err := format(&configInspectContext{Config: config}); err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(&configInspectContext{}, render)
}

type configInspectContext struct {
	swarm.Config
	subContext
}

func (ctx *configInspectContext) ID() string {
	return ctx.Config.ID
}

func (ctx *configInspectContext) Name() string {
	return ctx.Config.Spec.Name
}

func (ctx *configInspectContext) Labels() map[string]string {
	return ctx.Config.Spec.Labels
}

func (ctx *configInspectContext) CreatedAt() string {
	return command.PrettyPrint(ctx.Config.CreatedAt)
}

func (ctx *configInspectContext) UpdatedAt() string {
	return command.PrettyPrint(ctx.Config.UpdatedAt)
}

func (ctx *configInspectContext) Data() string {
	if ctx.Config.Spec.Data == nil {
		return ""
	}
	return string(ctx.Config.Spec.Data)
}
