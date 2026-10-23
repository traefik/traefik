package externalaction

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type execTarget struct {
	enrichments []*compiledEnrichment
	command     string
	args        []*compiledTemplate
	env         map[string]*compiledTemplate
	stdin       *compiledTemplate
}

func newExecTarget(rule dynamic.ExternalActionRule) (target, error) {
	if !filepath.IsAbs(rule.Exec.Command) {
		return nil, errors.New("exec command must be an absolute path")
	}
	if strings.ContainsAny(rule.Exec.Command, "\r\n\x00") {
		return nil, errors.New("exec command contains invalid characters")
	}
	value := &execTarget{
		command: rule.Exec.Command,
		env:     make(map[string]*compiledTemplate),
	}
	for index, argument := range rule.Exec.Args {
		parsed, err := newTemplate(fmt.Sprintf("arg-%d", index), argument)
		if err != nil {
			return nil, fmt.Errorf("compiling argument %d: %w", index, err)
		}
		value.args = append(value.args, parsed)
	}
	for name, variable := range rule.Exec.Env {
		parsed, err := newTemplate("env-"+name, variable)
		if err != nil {
			return nil, fmt.Errorf("compiling environment variable %q: %w", name, err)
		}
		value.env[name] = parsed
	}
	stdin, err := newTemplate("stdin", rule.Exec.Stdin)
	if err != nil {
		return nil, fmt.Errorf("compiling stdin: %w", err)
	}
	value.stdin = stdin
	return value, nil
}

func (e *execTarget) Run(ctx context.Context, data *actionData) error {
	stdin, err := e.stdin.render(data)
	if err != nil {
		return fmt.Errorf("rendering stdin: %w", err)
	}
	enriched, stdin, err := enrichPayload(ctx, data, stdin, e.enrichments)
	if err != nil {
		return err
	}
	arguments := make([]string, len(e.args))
	for index, argument := range e.args {
		value, err := argument.render(enriched)
		if err != nil {
			return fmt.Errorf("rendering argument %d: %w", index, err)
		}
		arguments[index] = value
	}
	// The executable is a literal absolute operator-configured path. Arguments
	// stay separate: request data is never concatenated into a shell command.
	command := exec.CommandContext(ctx, e.command, arguments...)
	command.Env = os.Environ()
	for name, variable := range e.env {
		value, err := variable.render(enriched)
		if err != nil {
			return fmt.Errorf("rendering environment variable %q: %w", name, err)
		}
		command.Env = append(command.Env, name+"="+value)
	}
	command.Stdin = strings.NewReader(stdin)

	stdout := &limitedBuffer{remaining: 64 * 1024}
	command.Stdout = stdout
	command.Stderr = stdout
	if err := command.Run(); err != nil {
		return fmt.Errorf("command failed: %w: %s", err, strings.TrimSpace(stdout.String()))
	}
	return nil
}

type limitedBuffer struct {
	value     bytes.Buffer
	remaining int
	truncated bool
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	original := len(value)
	if len(value) > b.remaining {
		value = value[:max(b.remaining, 0)]
		b.truncated = true
	}
	if len(value) > 0 {
		_, _ = b.value.Write(value)
		b.remaining -= len(value)
	}
	return original, nil
}

func (b *limitedBuffer) String() string {
	if b.truncated {
		return b.value.String() + "…"
	}
	return b.value.String()
}
