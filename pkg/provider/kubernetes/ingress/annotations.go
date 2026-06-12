package ingress

import (
	"regexp"
	"slices"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/label"
)

const (
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set
	annotationsPrefix = "traefik.ingress.kubernetes.io/"
)

var annotationsRegex = regexp.MustCompile(`(.+)\.(\w+)\.(\d+)\.(.+)`)

// RouterConfig is the router's root configuration from annotations.
type RouterConfig struct {
	Router *RouterIng `json:"router,omitempty"`
}

// RouterIng is the router's configuration from annotations.
type RouterIng struct {
	PathMatcher   string                             `json:"pathMatcher,omitempty"`
	EntryPoints   []string                           `json:"entryPoints,omitempty"`
	Middlewares   []string                           `json:"middlewares,omitempty"`
	Priority      int                                `json:"priority,omitempty"`
	RuleSyntax    string                             `json:"ruleSyntax,omitempty"`
	TLS           *dynamic.RouterTLSConfig           `json:"tls,omitempty" label:"allowEmpty"`
	Observability *dynamic.RouterObservabilityConfig `json:"observability,omitempty" label:"allowEmpty"`
}

func (r *RouterConfig) DeepCopy() *RouterConfig {
	if r == nil {
		return nil
	}

	out := new(RouterConfig)
	if r.Router != nil {
		out.Router = r.Router.DeepCopy()
	}

	return out
}

func (r *RouterIng) DeepCopy() *RouterIng {
	if r == nil {
		return nil
	}

	out := *r
	out.EntryPoints = slices.Clone(r.EntryPoints)
	out.Middlewares = slices.Clone(r.Middlewares)
	out.TLS = r.TLS.DeepCopy()
	out.Observability = r.Observability.DeepCopy()

	return &out
}

// SetDefaults sets the default values.
func (r *RouterIng) SetDefaults() {
	r.PathMatcher = defaultPathMatcher
}

// ServiceConfig is the service's root configuration from annotations.
type ServiceConfig struct {
	Service *ServiceIng `json:"service,omitempty"`
}

// ServiceIng is the service's configuration from annotations.
type ServiceIng struct {
	ServersScheme    string          `json:"serversScheme,omitempty"`
	ServersTransport string          `json:"serversTransport,omitempty"`
	PassHostHeader   *bool           `json:"passHostHeader"`
	Middlewares      []string        `json:"middlewares,omitempty"`
	Sticky           *dynamic.Sticky `json:"sticky,omitempty" label:"allowEmpty"`
	NativeLB         *bool           `json:"nativeLB,omitempty"`
	NodePortLB       bool            `json:"nodePortLB,omitempty"`
}

// SetDefaults sets the default values.
func (s *ServiceIng) SetDefaults() {
	s.PassHostHeader = func(v bool) *bool { return &v }(true)
}

func parseRouterConfig(annotations map[string]string) (*RouterConfig, error) {
	labels := convertAnnotations(annotations)
	if len(labels) == 0 {
		return nil, nil
	}

	cfg := &RouterConfig{}

	err := label.Decode(labels, cfg, "traefik.router.")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseServiceConfig(annotations map[string]string) (*ServiceConfig, error) {
	labels := convertAnnotations(annotations)
	if len(labels) == 0 {
		return nil, nil
	}

	cfg := &ServiceConfig{}

	err := label.Decode(labels, cfg, "traefik.service.")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func convertAnnotations(annotations map[string]string) map[string]string {
	if len(annotations) == 0 {
		return nil
	}

	result := make(map[string]string)

	for key, value := range annotations {
		if !strings.HasPrefix(key, annotationsPrefix) {
			continue
		}

		newKey := strings.ReplaceAll(key, "ingress.kubernetes.io/", "")

		if annotationsRegex.MatchString(newKey) {
			newKey = annotationsRegex.ReplaceAllString(newKey, "$1.$2[$3].$4")
		}

		result[newKey] = value
	}

	return result
}
