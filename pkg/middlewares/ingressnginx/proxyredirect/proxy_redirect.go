package proxyredirect

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
)

const typeName = "ProxyRedirect"

type proxyRedirect struct {
	name  string
	next  http.Handler
	from  string
	to    string
	regex *regexp.Regexp
}

// New creates a new proxy-redirect middleware.
func New(ctx context.Context, next http.Handler, config dynamic.ProxyRedirect, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.From == "" {
		return nil, errors.New("from cannot be empty")
	}

	var re *regexp.Regexp
	switch {
	case strings.HasPrefix(config.From, "~*"):
		pattern, _ := strings.CutPrefix(config.From, "~*")
		var err error
		re, err = regexp.Compile("(?i)" + pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling case-insensitive regular expression %s: %w", pattern, err)
		}
	case strings.HasPrefix(config.From, "~"):
		pattern, _ := strings.CutPrefix(config.From, "~")
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling regular expression %s: %w", pattern, err)
		}
	}

	return &proxyRedirect{
		name:  name,
		next:  next,
		from:  config.From,
		to:    config.To,
		regex: re,
	}, nil
}

func (p *proxyRedirect) GetTracingInformation() (string, string) {
	return p.name, typeName
}

func (p *proxyRedirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.next.ServeHTTP(middlewares.NewResponseModifier(rw, req, p.modifyResponse), req)
}

func (p *proxyRedirect) modifyResponse(res *http.Response) error {
	loc := res.Header.Get("Location")
	refresh := res.Header.Get("Refresh")
	if loc == "" && refresh == "" {
		return nil
	}

	from := p.from
	if p.regex == nil {
		from = ingressnginx.ReplaceVariables(p.from, res.Request, res.Header, nil)
	}

	to := p.to
	if to == "" || to == "default" {
		to = "$scheme://$best_http_host/"
	} else if p.regex == nil && strings.Contains(from, "://") && strings.HasPrefix(to, "/") {
		// NGINX inserts the server's name and port when the replacement omits it.
		to = "$scheme://$best_http_host" + to
	}
	to = ingressnginx.ReplaceVariables(to, res.Request, res.Header, nil)

	if loc != "" {
		res.Header.Set("Location", p.rewrite(loc, from, to))
	}

	if refresh != "" {
		res.Header.Set("Refresh", p.rewriteRefresh(refresh, from, to))
	}

	return nil
}

func (p *proxyRedirect) rewrite(val, from, to string) string {
	if p.regex != nil {
		return p.regex.ReplaceAllString(val, to)
	}

	if after, ok := strings.CutPrefix(val, from); ok {
		return to + after
	}

	return val
}

func (p *proxyRedirect) rewriteRefresh(val, from, to string) string {
	lower := strings.ToLower(val)
	idx := strings.Index(lower, "url=")
	if idx == -1 {
		return p.rewrite(val, from, to)
	}

	prefix := val[:idx+4]
	urlPart := val[idx+4:]
	quote := ""
	if strings.HasPrefix(urlPart, "\"") || strings.HasPrefix(urlPart, "'") {
		quote = string(urlPart[0])
		urlPart = strings.Trim(urlPart, quote)
	}

	rewritten := p.rewrite(urlPart, from, to)
	if quote != "" {
		return prefix + quote + rewritten + quote
	}

	return prefix + rewritten
}
