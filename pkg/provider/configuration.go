package provider

import (
	"bytes"
	"context"
	"maps"
	"slices"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
)

// MakeDefaultRuleTemplate creates the default rule template.
func MakeDefaultRuleTemplate(defaultRule string, funcMap template.FuncMap) (*template.Template, error) {
	defaultFuncMap := sprig.TxtFuncMap()
	defaultFuncMap["normalize"] = Normalize

	maps.Copy(defaultFuncMap, funcMap)

	return template.New("defaultRule").Funcs(defaultFuncMap).Parse(defaultRule)
}

// BuildTCPRouterConfiguration builds a router configuration.
func BuildTCPRouterConfiguration(ctx context.Context, configuration *dynamic.TCPConfiguration) {
	for routerName, router := range configuration.Routers {
		loggerRouter := log.Ctx(ctx).With().Str(logs.RouterName, routerName).Logger()

		if len(router.Rule) == 0 {
			delete(configuration.Routers, routerName)
			loggerRouter.Error().Msg("Empty rule")
			continue
		}

		if router.Service == "" {
			if len(configuration.Services) > 1 {
				delete(configuration.Routers, routerName)
				loggerRouter.Error().
					Msgf("Router %s cannot be linked automatically with multiple Services: %q", routerName, slices.Collect(maps.Keys(configuration.Services)))
				continue
			}

			for serviceName := range configuration.Services {
				router.Service = serviceName
			}
		}
	}
}

// BuildUDPRouterConfiguration builds a router configuration.
func BuildUDPRouterConfiguration(ctx context.Context, configuration *dynamic.UDPConfiguration) {
	for routerName, router := range configuration.Routers {
		loggerRouter := log.Ctx(ctx).With().Str(logs.RouterName, routerName).Logger()

		if router.Service != "" {
			continue
		}

		if len(configuration.Services) > 1 {
			delete(configuration.Routers, routerName)
			loggerRouter.Error().
				Msgf("Router %s cannot be linked automatically with multiple Services: %q", routerName, slices.Collect(maps.Keys(configuration.Services)))
			continue
		}

		for serviceName := range configuration.Services {
			router.Service = serviceName
			break
		}
	}
}

// BuildRouterConfiguration builds a router configuration.
func BuildRouterConfiguration(ctx context.Context, configuration *dynamic.HTTPConfiguration, defaultRouterName string, defaultRuleTpl *template.Template, model any) {
	if len(configuration.Routers) == 0 {
		if len(configuration.Services) > 1 {
			log.Ctx(ctx).Info().Msg("Could not create a router for the container: too many services")
		} else {
			configuration.Routers = make(map[string]*dynamic.Router)
			configuration.Routers[defaultRouterName] = &dynamic.Router{}
		}
	}

	for routerName, router := range configuration.Routers {
		loggerRouter := log.Ctx(ctx).With().Str(logs.RouterName, routerName).Logger()

		if len(router.Rule) == 0 {
			writer := &bytes.Buffer{}
			if err := defaultRuleTpl.Execute(writer, model); err != nil {
				loggerRouter.Error().Err(err).Msg("Error while parsing default rule")
				delete(configuration.Routers, routerName)
				continue
			}

			router.Rule = writer.String()
			if len(router.Rule) == 0 {
				loggerRouter.Error().Msg("Undefined rule")
				delete(configuration.Routers, routerName)
				continue
			}

			// Flag default rule routers to add the denyRouterRecursion middleware.
			router.DefaultRule = true
		}

		if router.Service == "" {
			if len(configuration.Services) > 1 {
				delete(configuration.Routers, routerName)
				loggerRouter.Error().
					Msgf("Router %s cannot be linked automatically with multiple Services: %q", routerName, slices.Collect(maps.Keys(configuration.Services)))
				continue
			}

			for serviceName := range configuration.Services {
				router.Service = serviceName
			}
		}
	}
}

// Normalize replaces all special chars with `-`.
func Normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}
