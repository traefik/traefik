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
	defaultSecretTableFormat           = "table {{.ID}}\t{{.Name}}\t{{.CreatedAt}}\t{{.UpdatedAt}}"
	secretIDHeader                     = "ID"
	secretCreatedHeader                = "CREATED"
	secretUpdatedHeader                = "UPDATED"
	secretInspectPrettyTemplate Format = `ID:			{{.ID}}
Name:			{{.Name}}
{{- if .Labels }}
Labels:
{{- range $k, $v := .Labels }}
 - {{ $k }}{{if $v }}={{ $v }}{{ end }}
{{- end }}{{ end }}
Created at:            	{{.CreatedAt}}
Updated at:            	{{.UpdatedAt}}`
)

// NewSecretFormat returns a Format for rendering using a secret Context
func NewSecretFormat(source string, quiet bool) Format {
	switch source {
	case PrettyFormatKey:
		return secretInspectPrettyTemplate
	case TableFormatKey:
		if quiet {
			return defaultQuietFormat
		}
		return defaultSecretTableFormat
	}
	return Format(source)
}

// SecretWrite writes the context
func SecretWrite(ctx Context, secrets []swarm.Secret) error {
	render := func(format func(subContext subContext) error) error {
		for _, secret := range secrets {
			secretCtx := &secretContext{s: secret}
			if err := format(secretCtx); err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(newSecretContext(), render)
}

func newSecretContext() *secretContext {
	sCtx := &secretContext{}

	sCtx.header = map[string]string{
		"ID":        secretIDHeader,
		"Name":      nameHeader,
		"CreatedAt": secretCreatedHeader,
		"UpdatedAt": secretUpdatedHeader,
		"Labels":    labelsHeader,
	}
	return sCtx
}

type secretContext struct {
	HeaderContext
	s swarm.Secret
}

func (c *secretContext) MarshalJSON() ([]byte, error) {
	return marshalJSON(c)
}

func (c *secretContext) ID() string {
	return c.s.ID
}

func (c *secretContext) Name() string {
	return c.s.Spec.Annotations.Name
}

func (c *secretContext) CreatedAt() string {
	return units.HumanDuration(time.Now().UTC().Sub(c.s.Meta.CreatedAt)) + " ago"
}

func (c *secretContext) UpdatedAt() string {
	return units.HumanDuration(time.Now().UTC().Sub(c.s.Meta.UpdatedAt)) + " ago"
}

func (c *secretContext) Labels() string {
	mapLabels := c.s.Spec.Annotations.Labels
	if mapLabels == nil {
		return ""
	}
	var joinLabels []string
	for k, v := range mapLabels {
		joinLabels = append(joinLabels, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(joinLabels, ",")
}

func (c *secretContext) Label(name string) string {
	if c.s.Spec.Annotations.Labels == nil {
		return ""
	}
	return c.s.Spec.Annotations.Labels[name]
}

// SecretInspectWrite renders the context for a list of secrets
func SecretInspectWrite(ctx Context, refs []string, getRef inspect.GetRefFunc) error {
	if ctx.Format != secretInspectPrettyTemplate {
		return inspect.Inspect(ctx.Output, refs, string(ctx.Format), getRef)
	}
	render := func(format func(subContext subContext) error) error {
		for _, ref := range refs {
			secretI, _, err := getRef(ref)
			if err != nil {
				return err
			}
			secret, ok := secretI.(swarm.Secret)
			if !ok {
				return fmt.Errorf("got wrong object to inspect :%v", ok)
			}
			if err := format(&secretInspectContext{Secret: secret}); err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(&secretInspectContext{}, render)
}

type secretInspectContext struct {
	swarm.Secret
	subContext
}

func (ctx *secretInspectContext) ID() string {
	return ctx.Secret.ID
}

func (ctx *secretInspectContext) Name() string {
	return ctx.Secret.Spec.Name
}

func (ctx *secretInspectContext) Labels() map[string]string {
	return ctx.Secret.Spec.Labels
}

func (ctx *secretInspectContext) CreatedAt() string {
	return command.PrettyPrint(ctx.Secret.CreatedAt)
}

func (ctx *secretInspectContext) UpdatedAt() string {
	return command.PrettyPrint(ctx.Secret.UpdatedAt)
}
