package cli

import (
	"encoding/json"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type rawConfiguration map[string]interface{}

// deprecationNotice prints warns and hints if deprecated/removed static option are in use,
// and returns whether at least one of these options is incompatible with the current version.
func (r *rawConfiguration) deprecationNotice(logger zerolog.Logger) bool {
	if r == nil {
		return false
	}

	marshal, err := json.Marshal(r)
	if err != nil {
		log.Error().Err(err).Send()
	}
	config := &configuration{}

	err = json.Unmarshal(marshal, config)
	if err != nil {
		log.Error().Err(err).Send()
	}

	return config.deprecationNotice(logger)
}

// configuration holds the static configuration removed/deprecated options.
type configuration struct {
	Experimental *experimental `json:"experimental,omitempty" toml:"experimental,omitempty" yaml:"experimental,omitempty"`
	Pilot        *interface{}  `json:"pilot,omitempty" toml:"pilot,omitempty" yaml:"pilot,omitempty"`
	Providers    *providers    `json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty"`
	Tracing      *tracing      `json:"tracing,omitempty" toml:"tracing,omitempty" yaml:"tracing,omitempty"`
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

	// not incompatible as HTTP3 option is only deprecated and not removed.
	c.Experimental.deprecationNotice(logger)

	return incompatible ||
		c.Providers.deprecationNotice(logger) ||
		c.Tracing.deprecationNotice(logger)
}

type providers struct {
	Docker        *docker        `json:"docker,omitempty" toml:"docker,omitempty" yaml:"docker,omitempty"`
	Swarm         *swarm         `json:"swarm,omitempty" toml:"swarm,omitempty" yaml:"swarm,omitempty"`
	Consul        *consul        `json:"consul,omitempty" toml:"consul,omitempty" yaml:"consul,omitempty"`
	ConsulCatalog *consulCatalog `json:"consulCatalog,omitempty" toml:"consulCatalog,omitempty" yaml:"consulCatalog,omitempty"`
	Nomad         *nomad         `json:"nomad,omitempty" toml:"nomad,omitempty" yaml:"nomad,omitempty"`
	Marathon      *interface{}   `json:"marathon,omitempty" toml:"marathon,omitempty" yaml:"marathon,omitempty"`
	Rancher       *interface{}   `json:"rancher,omitempty" toml:"rancher,omitempty" yaml:"rancher,omitempty"`
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

	return incompatible ||
		p.Docker.deprecationNotice(logger) ||
		p.Consul.deprecationNotice(logger) ||
		p.ConsulCatalog.deprecationNotice(logger) ||
		p.Nomad.deprecationNotice(logger) ||
		p.Swarm.deprecationNotice(logger) ||
		p.ETCD.deprecationNotice(logger) ||
		p.Redis.deprecationNotice(logger) ||
		p.HTTP.deprecationNotice(logger)
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
	Endpoint  *endpointConfig `json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
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
	Endpoint  *endpointConfig `json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
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

func (e *experimental) deprecationNotice(logger zerolog.Logger) {
	if e == nil {
		return
	}

	// As long as HTTP3 is still a field of the experimental static configuration,
	// there will be no parsing error, so we just want to warn here.
	if e.HTTP3 != nil {
		logger.Warn().Msg("HTTP3 is not an experimental feature in v3 and the associated enablement option will be remove in a future major release." +
			"This option has now no effect, but we recommend to stop using it to avoid hassle considering future major release upgrade." +
			"For more information please read the migration guide: https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#http3-experimental-configuration")
	}
}

type tracing struct {
	SpanNameLimit *int         `json:"spanNameLimit,omitempty" toml:"spanNameLimit,omitempty" yaml:"spanNameLimit,omitempty" export:"true"`
	Jaeger        *interface{} `json:"jaeger,omitempty" toml:"jaeger,omitempty" yaml:"jaeger,omitempty"`
	Zipkin        *interface{} `json:"zipkin,omitempty" toml:"zipkin,omitempty" yaml:"zipkin,omitempty"`
	Datadog       *interface{} `json:"datadog,omitempty" toml:"datadog,omitempty" yaml:"datadog,omitempty"`
	Instana       *interface{} `json:"instana,omitempty" toml:"instana,omitempty" yaml:"instana,omitempty"`
	Haystack      *interface{} `json:"haystack,omitempty" toml:"haystack,omitempty" yaml:"haystack,omitempty"`
	Elastic       *interface{} `json:"elastic,omitempty" toml:"elastic,omitempty" yaml:"elastic,omitempty"`
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
