package configuration

import (
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/old/api"
	"github.com/containous/traefik/old/middlewares/tracing"
	"github.com/containous/traefik/old/provider/file"
	"github.com/containous/traefik/old/types"
	"github.com/containous/traefik/ping"
	"github.com/containous/traefik/provider"
	file2 "github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/tracing/datadog"
	"github.com/containous/traefik/tracing/jaeger"
	"github.com/containous/traefik/tracing/zipkin"
	types2 "github.com/containous/traefik/types"
)

// ConvertStaticConf FIXME sugar
// Deprecated
func ConvertStaticConf(globalConfiguration GlobalConfiguration) static.Configuration {
	staticConfiguration := static.Configuration{}

	staticConfiguration.EntryPoints = make(static.EntryPoints)

	if globalConfiguration.EntryPoints != nil {
		for name, ep := range globalConfiguration.EntryPoints {
			staticConfiguration.EntryPoints[name] = &static.EntryPoint{
				Address: ep.Address,
			}
		}
	}

	if globalConfiguration.Ping != nil {
		old := globalConfiguration.Ping
		staticConfiguration.Ping = &ping.Handler{
			EntryPoint: old.EntryPoint,
		}
	}

	staticConfiguration.API = convertAPI(globalConfiguration.API)
	staticConfiguration.Providers.File = convertFile(globalConfiguration.File)
	staticConfiguration.Metrics = ConvertMetrics(globalConfiguration.Metrics)
	staticConfiguration.AccessLog = ConvertAccessLog(globalConfiguration.AccessLog)
	staticConfiguration.Tracing = ConvertTracing(globalConfiguration.Tracing)
	staticConfiguration.HostResolver = ConvertHostResolverConfig(globalConfiguration.HostResolver)

	return staticConfiguration
}

// ConvertAccessLog FIXME sugar
// Deprecated
func ConvertAccessLog(old *types.AccessLog) *types2.AccessLog {
	if old == nil {
		return nil
	}

	accessLog := &types2.AccessLog{
		FilePath:      old.FilePath,
		Format:        old.Format,
		BufferingSize: old.BufferingSize,
	}

	if old.Filters != nil {
		accessLog.Filters = &types2.AccessLogFilters{
			StatusCodes:   types2.StatusCodes(old.Filters.StatusCodes),
			RetryAttempts: old.Filters.RetryAttempts,
			MinDuration:   old.Filters.MinDuration,
		}
	}

	if old.Fields != nil {
		accessLog.Fields = &types2.AccessLogFields{
			DefaultMode: old.Fields.DefaultMode,
			Names:       types2.FieldNames(old.Fields.Names),
		}

		if old.Fields.Headers != nil {
			accessLog.Fields.Headers = &types2.FieldHeaders{
				DefaultMode: old.Fields.Headers.DefaultMode,
				Names:       types2.FieldHeaderNames(old.Fields.Headers.Names),
			}
		}
	}

	return accessLog
}

// ConvertMetrics FIXME sugar
// Deprecated
func ConvertMetrics(old *types.Metrics) *types2.Metrics {
	if old == nil {
		return nil
	}

	metrics := &types2.Metrics{}

	if old.Prometheus != nil {
		metrics.Prometheus = &types2.Prometheus{
			EntryPoint: old.Prometheus.EntryPoint,
			Buckets:    types2.Buckets(old.Prometheus.Buckets),
		}
	}

	if old.Datadog != nil {
		metrics.Datadog = &types2.Datadog{
			Address:      old.Datadog.Address,
			PushInterval: old.Datadog.PushInterval,
		}
	}

	if old.StatsD != nil {
		metrics.StatsD = &types2.Statsd{
			Address:      old.StatsD.Address,
			PushInterval: old.StatsD.PushInterval,
		}
	}
	if old.InfluxDB != nil {
		metrics.InfluxDB = &types2.InfluxDB{
			Address:         old.InfluxDB.Address,
			Protocol:        old.InfluxDB.Protocol,
			PushInterval:    old.InfluxDB.PushInterval,
			Database:        old.InfluxDB.Database,
			RetentionPolicy: old.InfluxDB.RetentionPolicy,
			Username:        old.InfluxDB.Username,
			Password:        old.InfluxDB.Password,
		}
	}

	return metrics
}

// ConvertTracing FIXME sugar
// Deprecated
func ConvertTracing(old *tracing.Tracing) *static.Tracing {
	if old == nil {
		return nil
	}

	tra := &static.Tracing{
		Backend:       old.Backend,
		ServiceName:   old.ServiceName,
		SpanNameLimit: old.SpanNameLimit,
	}

	if old.Jaeger != nil {
		tra.Jaeger = &jaeger.Config{
			SamplingServerURL:  old.Jaeger.SamplingServerURL,
			SamplingType:       old.Jaeger.SamplingType,
			SamplingParam:      old.Jaeger.SamplingParam,
			LocalAgentHostPort: old.Jaeger.LocalAgentHostPort,
			Gen128Bit:          old.Jaeger.Gen128Bit,
			Propagation:        old.Jaeger.Propagation,
		}
	}

	if old.Zipkin != nil {
		tra.Zipkin = &zipkin.Config{
			HTTPEndpoint: old.Zipkin.HTTPEndpoint,
			SameSpan:     old.Zipkin.SameSpan,
			ID128Bit:     old.Zipkin.ID128Bit,
			Debug:        old.Zipkin.Debug,
		}
	}

	if old.DataDog != nil {
		tra.DataDog = &datadog.Config{
			LocalAgentHostPort: old.DataDog.LocalAgentHostPort,
			GlobalTag:          old.DataDog.GlobalTag,
			Debug:              old.DataDog.Debug,
			PrioritySampling:   old.DataDog.PrioritySampling,
		}
	}

	return tra
}

func convertAPI(old *api.Handler) *static.API {
	if old == nil {
		return nil
	}

	api := &static.API{
		EntryPoint:      old.EntryPoint,
		Dashboard:       old.Dashboard,
		DashboardAssets: old.DashboardAssets,
	}

	if old.Statistics != nil {
		api.Statistics = &types2.Statistics{
			RecentErrors: old.Statistics.RecentErrors,
		}
	}

	return api
}

func convertConstraints(oldConstraints types.Constraints) types2.Constraints {
	constraints := types2.Constraints{}
	for _, value := range oldConstraints {
		constraint := &types2.Constraint{
			Key:       value.Key,
			MustMatch: value.MustMatch,
			Regex:     value.Regex,
		}

		constraints = append(constraints, constraint)
	}
	return constraints
}

func convertFile(old *file.Provider) *file2.Provider {
	if old == nil {
		return nil
	}

	f := &file2.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    old.Watch,
			Filename: old.Filename,
			Trace:    old.Trace,
		},
		Directory:   old.Directory,
		TraefikFile: old.TraefikFile,
	}
	f.DebugLogGeneratedTemplate = old.DebugLogGeneratedTemplate
	f.Constraints = convertConstraints(old.Constraints)

	return f
}

// ConvertHostResolverConfig FIXME
// Deprecated
func ConvertHostResolverConfig(oldconfig *HostResolverConfig) *static.HostResolverConfig {
	if oldconfig == nil {
		return nil
	}

	return &static.HostResolverConfig{
		CnameFlattening: oldconfig.CnameFlattening,
		ResolvConfig:    oldconfig.ResolvConfig,
		ResolvDepth:     oldconfig.ResolvDepth,
	}
}
