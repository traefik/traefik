package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/paerser/parser"
)

type DeprecationLoader struct{}

func (d DeprecationLoader) Load(args []string, cmd *cli.Command) (bool, error) {
	for i, arg := range args {
		if !strings.Contains(arg, "=") {
			args[i] = arg + "=true"
		}
	}

	labels, err := flag.Parse(args, nil)
	if err != nil {
		return false, err
	}

	if len(labels) > 0 {
		node, err := parser.DecodeToNode(labels, "traefik")
		if err != nil {
			return false, fmt.Errorf("DecodeToNode: %w", err)
		}

		rawConfig := &configuration{}

		err = filterUnknownNodes(rawConfig, node)
		if err != nil {
			return false, fmt.Errorf("filter: %w", err)
		}

		if len(node.Children) > 0 {
			err = parser.AddMetadata(rawConfig, node, parser.MetadataOpts{})
			if err != nil {
				return false, fmt.Errorf("AddMetadata: %w", err)
			}

			err = parser.Fill(rawConfig, node, parser.FillerOpts{})
			if err != nil {
				return false, err
			}
			logger := log.With().Str("loader", "FLAG").Logger()
			if rawConfig.deprecationNotice(logger) {
				return true, errors.New("deprecated field found")
			}
		}
	}

	// FILE
	ref, err := flag.Parse(args, cmd.Configuration)
	if err != nil {
		_ = cmd.PrintHelp(os.Stdout)
		return false, err
	}

	configFileFlag := "traefik.configfile"
	if _, ok := ref["traefik.configFile"]; ok {
		configFileFlag = "traefik.configFile"
	}

	rawConfig := &configuration{}
	_, err = loadConfigFiles(ref[configFileFlag], rawConfig)

	if err == nil {
		logger := log.With().Str("loader", "FILE").Logger()
		if rawConfig.deprecationNotice(logger) {
			return true, errors.New("deprecated field found")
		}
	}

	rawConfig = &configuration{}
	l := EnvLoader{}
	_, err = l.Load(os.Args, &cli.Command{
		Configuration: rawConfig,
	})

	if err == nil {
		logger := log.With().Str("loader", "ENV").Logger()
		if rawConfig.deprecationNotice(logger) {
			return true, errors.New("deprecated field found")
		}
	}

	return false, nil
}

func filterUnknownNodes(element interface{}, node *parser.Node) error {
	if len(node.Children) == 0 {
		return fmt.Errorf("invalid node %s: no child", node.Name)
	}

	if element == nil {
		return errors.New("nil structure")
	}

	rootType := reflect.TypeOf(element)
	browseChildren(rootType, node)
	return nil
}

func browseChildren(fType reflect.Type, node *parser.Node) bool {
	var children []*parser.Node
	for _, child := range node.Children {
		if isValid(fType, child) {
			children = append(children, child)
		}
	}
	node.Children = children
	return len(node.Children) > 0
}

func isValid(rootType reflect.Type, node *parser.Node) bool {
	rType := rootType
	if rootType.Kind() == reflect.Pointer {
		rType = rootType.Elem()
	}

	if rType.Kind() == reflect.Map && rType.Elem().Kind() == reflect.Interface {
		return true
	}

	if rType.Kind() == reflect.Interface {
		return true
	}

	field, b := findTypedField(rType, node)

	if !b {
		return b
	}

	if len(node.Children) > 0 {
		return browseChildren(field.Type, node)
	}

	return true
}

func findTypedField(rType reflect.Type, node *parser.Node) (reflect.StructField, bool) {
	if rType.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}

	for i := 0; i < rType.NumField(); i++ {
		cField := rType.Field(i)

		fieldName := cField.Tag.Get(parser.TagLabelSliceAsStruct)
		if len(fieldName) == 0 {
			fieldName = cField.Name
		}

		if cField.PkgPath == "" {
			if cField.Anonymous {
				if cField.Type.Kind() == reflect.Struct {
					structField, b := findTypedField(cField.Type, node)
					if !b {
						continue
					}
					return structField, true
				}
			}

			if strings.EqualFold(fieldName, node.Name) {
				node.FieldName = cField.Name
				return cField, true
			}
		}
	}

	return reflect.StructField{}, false
}

// configuration holds the static configuration removed/deprecated options.
type configuration struct {
	Experimental *experimental  `json:"experimental,omitempty" toml:"experimental,omitempty" yaml:"experimental,omitempty"`
	Pilot        map[string]any `json:"pilot,omitempty" toml:"pilot,omitempty" yaml:"pilot,omitempty"`
	Providers    *providers     `json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty"`
	Tracing      *tracing       `json:"tracing,omitempty" toml:"tracing,omitempty" yaml:"tracing,omitempty"`
}

func (c *configuration) deprecationNotice(logger zerolog.Logger) bool {
	if c == nil {
		return false
	}

	var incompatible bool
	if c.Pilot != nil {
		incompatible = true
		logger.Error().Msg("Pilot configuration has been removed in v3, please remove all Marathon-related static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#pilot")
	}

	incompatibleExperimental := c.Experimental.deprecationNotice(logger)
	incompatibleProviders := c.Providers.deprecationNotice(logger)
	incompatibleTracing := c.Tracing.deprecationNotice(logger)
	return incompatible || incompatibleExperimental || incompatibleProviders || incompatibleTracing
}

type providers struct {
	Docker        *docker        `json:"docker,omitempty" toml:"docker,omitempty" yaml:"docker,omitempty"`
	Swarm         *swarm         `json:"swarm,omitempty" toml:"swarm,omitempty" yaml:"swarm,omitempty"`
	Consul        *consul        `json:"consul,omitempty" toml:"consul,omitempty" yaml:"consul,omitempty"`
	ConsulCatalog *consulCatalog `json:"consulCatalog,omitempty" toml:"consulCatalog,omitempty" yaml:"consulCatalog,omitempty"`
	Nomad         *nomad         `json:"nomad,omitempty" toml:"nomad,omitempty" yaml:"nomad,omitempty"`
	Marathon      map[string]any `json:"marathon,omitempty" toml:"marathon,omitempty" yaml:"marathon,omitempty"`
	Rancher       map[string]any `json:"rancher,omitempty" toml:"rancher,omitempty" yaml:"rancher,omitempty"`
	ETCD          *etcd          `json:"etcd,omitempty" toml:"etcd,omitempty" yaml:"etcd,omitempty"`
	Redis         *redis         `json:"redis,omitempty" toml:"redis,omitempty" yaml:"redis,omitempty"`
	HTTP          *http          `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty"`
}

func (p *providers) deprecationNotice(logger zerolog.Logger) bool {
	if p == nil {
		return false
	}

	var incompatible bool

	if p.Marathon != nil {
		incompatible = true
		logger.Error().Msg("Marathon provider has been removed in v3, please remove all Marathon-related static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#marathon-provider")
	}

	if p.Rancher != nil {
		incompatible = true
		logger.Error().Msg("Rancher provider has been removed in v3, please remove all Rancher-related static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#rancher-v1-provider")
	}

	dockerIncompatible := p.Docker.deprecationNotice(logger)
	consulIncompatible := p.Consul.deprecationNotice(logger)
	consulCatalogIncompatible := p.ConsulCatalog.deprecationNotice(logger)
	nomadIncompatible := p.Nomad.deprecationNotice(logger)
	swarmIncompatible := p.Swarm.deprecationNotice(logger)
	etcdIncompatible := p.ETCD.deprecationNotice(logger)
	redisIncompatible := p.Redis.deprecationNotice(logger)
	httpIncompatible := p.HTTP.deprecationNotice(logger)
	return incompatible ||
		dockerIncompatible ||
		consulIncompatible ||
		consulCatalogIncompatible ||
		nomadIncompatible ||
		swarmIncompatible ||
		etcdIncompatible ||
		redisIncompatible ||
		httpIncompatible
}

type tls struct {
	CAOptional *bool `json:"caOptional,omitempty" toml:"caOptional,omitempty" yaml:"caOptional,omitempty"`
}

type docker struct {
	SwarmMode *bool `json:"swarmMode,omitempty" toml:"swarmMode,omitempty" yaml:"swarmMode,omitempty"`
	TLS       *tls  `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (d *docker) deprecationNotice(logger zerolog.Logger) bool {
	if d == nil {
		return false
	}

	var incompatible bool

	if d.SwarmMode != nil {
		incompatible = true
		logger.Error().Msg("Docker provider `swarmMode` option has been removed in v3, please use the Swarm Provider instead." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#docker-docker-swarm")
	}

	if d.TLS != nil && d.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Docker provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tlscaoptional")
	}

	return incompatible
}

type swarm struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (s *swarm) deprecationNotice(logger zerolog.Logger) bool {
	if s == nil {
		return false
	}

	var incompatible bool

	if s.TLS != nil && s.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Swarm provider `tls.CAOptional` option does not exist, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start.")
	}

	return incompatible
}

type etcd struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (e *etcd) deprecationNotice(logger zerolog.Logger) bool {
	if e == nil {
		return false
	}

	var incompatible bool

	if e.TLS != nil && e.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("ETCD provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tlscaoptional_3")
	}

	return incompatible
}

type redis struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (r *redis) deprecationNotice(logger zerolog.Logger) bool {
	if r == nil {
		return false
	}

	var incompatible bool

	if r.TLS != nil && r.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Redis provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tlscaoptional_4")
	}

	return incompatible
}

type consul struct {
	Namespace *string `json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	TLS       *tls    `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (c *consul) deprecationNotice(logger zerolog.Logger) bool {
	if c == nil {
		return false
	}

	var incompatible bool

	if c.Namespace != nil {
		incompatible = true
		logger.Error().Msg("Consul provider `namespace` option has been removed, please use the `namespaces` option instead." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#consul-provider")
	}

	if c.TLS != nil && c.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Consul provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tlscaoptional_1")
	}

	return incompatible
}

type consulCatalog struct {
	Namespace *string         `json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	Endpoint  *endpointConfig `json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
}

type endpointConfig struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (c *consulCatalog) deprecationNotice(logger zerolog.Logger) bool {
	if c == nil {
		return false
	}

	var incompatible bool

	if c.Namespace != nil {
		incompatible = true
		logger.Error().Msg("ConsulCatalog provider `namespace` option has been removed, please use the `namespaces` option instead." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#consulcatalog-provider")
	}

	if c.Endpoint != nil && c.Endpoint.TLS != nil && c.Endpoint.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("ConsulCatalog provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#endpointtlscaoptional")
	}

	return incompatible
}

type nomad struct {
	Namespace *string         `json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	Endpoint  *endpointConfig `json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
}

func (n *nomad) deprecationNotice(logger zerolog.Logger) bool {
	if n == nil {
		return false
	}

	var incompatible bool

	if n.Namespace != nil {
		incompatible = true
		logger.Error().Msg("Nomad provider `namespace` option has been removed, please use the `namespaces` option instead." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#nomad-provider")
	}

	if n.Endpoint != nil && n.Endpoint.TLS != nil && n.Endpoint.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Nomad provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#endpointtlscaoptional_1")
	}

	return incompatible
}

type http struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (h *http) deprecationNotice(logger zerolog.Logger) bool {
	if h == nil {
		return false
	}

	var incompatible bool

	if h.TLS != nil && h.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("HTTP provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			"Please remove all occurrences from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tlscaoptional_2")
	}

	return incompatible
}

type experimental struct {
	HTTP3 *bool `json:"http3,omitempty" toml:"http3,omitempty" yaml:"http3,omitempty"`
}

func (e *experimental) deprecationNotice(logger zerolog.Logger) bool {
	if e == nil {
		return false
	}

	if e.HTTP3 != nil {
		logger.Error().Msg("HTTP3 is not an experimental feature in v3 and the associated enablement has been removed." +
			"Please remove its usage from the static configuration for Traefik to start." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#http3-experimental-configuration")

		return true
	}

	return false
}

type tracing struct {
	SpanNameLimit *int           `json:"spanNameLimit,omitempty" toml:"spanNameLimit,omitempty" yaml:"spanNameLimit,omitempty"`
	Jaeger        map[string]any `json:"jaeger,omitempty" toml:"jaeger,omitempty" yaml:"jaeger,omitempty"`
	Zipkin        map[string]any `json:"zipkin,omitempty" toml:"zipkin,omitempty" yaml:"zipkin,omitempty"`
	Datadog       map[string]any `json:"datadog,omitempty" toml:"datadog,omitempty" yaml:"datadog,omitempty"`
	Instana       map[string]any `json:"instana,omitempty" toml:"instana,omitempty" yaml:"instana,omitempty"`
	Haystack      map[string]any `json:"haystack,omitempty" toml:"haystack,omitempty" yaml:"haystack,omitempty"`
	Elastic       map[string]any `json:"elastic,omitempty" toml:"elastic,omitempty" yaml:"elastic,omitempty"`
}

func (t *tracing) deprecationNotice(logger zerolog.Logger) bool {
	if t == nil {
		return false
	}
	var incompatible bool
	if t.SpanNameLimit != nil {
		incompatible = true
		logger.Error().Msg("SpanNameLimit option for Tracing has been removed in v3, as Span names are now of a fixed length." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tracing")
	}

	if t.Jaeger != nil {
		incompatible = true
		logger.Error().Msg("Jaeger Tracing backend has been removed in v3, please remove all Jaeger-related Tracing static configuration for Traefik to start." +
			"In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tracing")
	}

	if t.Zipkin != nil {
		incompatible = true
		logger.Error().Msg("Zipkin Tracing backend has been removed in v3, please remove all Jaeger-related Tracing static configuration for Traefik to start." +
			"In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tracing")
	}

	if t.Datadog != nil {
		incompatible = true
		logger.Error().Msg("Datadog Tracing backend has been removed in v3, please remove all Jaeger-related Tracing static configuration for Traefik to start." +
			"In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tracing")
	}

	if t.Instana != nil {
		incompatible = true
		logger.Error().Msg("Instana Tracing backend has been removed in v3, please remove all Jaeger-related Tracing static configuration for Traefik to start." +
			"In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tracing")
	}

	if t.Haystack != nil {
		incompatible = true
		logger.Error().Msg("Haystack Tracing backend has been removed in v3, please remove all Haystack-related Tracing static configuration for Traefik to start." +
			"In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tracing")
	}

	if t.Elastic != nil {
		incompatible = true
		logger.Error().Msg("Elastic Tracing backend has been removed in v3, please remove all Elastic-related Tracing static configuration for Traefik to start." +
			"In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#tracing")
	}

	return incompatible
}
