package static

import (
	oldapi "github.com/containous/traefik/old/api"
	"github.com/containous/traefik/old/configuration"
	oldtracing "github.com/containous/traefik/old/middlewares/tracing"
	oldfile "github.com/containous/traefik/old/provider/file"
	oldtypes "github.com/containous/traefik/old/types"
	"github.com/containous/traefik/ping"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/tracing/datadog"
	"github.com/containous/traefik/tracing/jaeger"
	"github.com/containous/traefik/tracing/zipkin"
	"github.com/containous/traefik/types"
)

// ConvertStaticConf FIXME sugar
// Deprecated
func ConvertStaticConf(globalConfiguration configuration.GlobalConfiguration) Configuration {
	staticConfiguration := Configuration{}

	staticConfiguration.EntryPoints = &EntryPoints{
		EntryPointList: make(EntryPointList),
		Defaults:       globalConfiguration.DefaultEntryPoints,
	}

	if globalConfiguration.EntryPoints != nil {
		for name, ep := range globalConfiguration.EntryPoints {
			staticConfiguration.EntryPoints.EntryPointList[name] = EntryPoint{
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
	staticConfiguration.Constraints = convertConstraints(globalConfiguration.Constraints)
	staticConfiguration.File = convertFile(globalConfiguration.File)
	staticConfiguration.Metrics = ConvertMetrics(globalConfiguration.Metrics)
	staticConfiguration.AccessLog = ConvertAccessLog(globalConfiguration.AccessLog)
	staticConfiguration.Tracing = ConvertTracing(globalConfiguration.Tracing)
	staticConfiguration.HostResolver = ConvertHostResolverConfig(globalConfiguration.HostResolver)

	return staticConfiguration
}

// ConvertAccessLog FIXME sugar
// Deprecated
func ConvertAccessLog(old *oldtypes.AccessLog) *types.AccessLog {
	if old == nil {
		return nil
	}

	accessLog := &types.AccessLog{
		FilePath:      old.FilePath,
		Format:        old.Format,
		BufferingSize: old.BufferingSize,
	}

	if old.Filters != nil {
		accessLog.Filters = &types.AccessLogFilters{
			StatusCodes:   types.StatusCodes(old.Filters.StatusCodes),
			RetryAttempts: old.Filters.RetryAttempts,
			MinDuration:   old.Filters.MinDuration,
		}
	}

	if old.Fields != nil {
		accessLog.Fields = &types.AccessLogFields{
			DefaultMode: old.Fields.DefaultMode,
			Names:       types.FieldNames(old.Fields.Names),
		}

		if old.Fields.Headers != nil {
			accessLog.Fields.Headers = &types.FieldHeaders{
				DefaultMode: old.Fields.Headers.DefaultMode,
				Names:       types.FieldHeaderNames(old.Fields.Headers.Names),
			}
		}
	}

	return accessLog
}

// ConvertMetrics FIXME sugar
// Deprecated
func ConvertMetrics(old *oldtypes.Metrics) *types.Metrics {
	if old == nil {
		return nil
	}

	metrics := &types.Metrics{}

	if old.Prometheus != nil {
		metrics.Prometheus = &types.Prometheus{
			EntryPoint: old.Prometheus.EntryPoint,
			Buckets:    types.Buckets(old.Prometheus.Buckets),
		}
	}

	if old.Datadog != nil {
		metrics.Datadog = &types.Datadog{
			Address:      old.Datadog.Address,
			PushInterval: old.Datadog.PushInterval,
		}
	}

	if old.StatsD != nil {
		metrics.StatsD = &types.Statsd{
			Address:      old.StatsD.Address,
			PushInterval: old.StatsD.PushInterval,
		}
	}
	if old.InfluxDB != nil {
		metrics.InfluxDB = &types.InfluxDB{
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
func ConvertTracing(old *oldtracing.Tracing) *Tracing {
	if old == nil {
		return nil
	}

	tra := &Tracing{
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
		}
	}

	return tra
}

func convertAPI(old *oldapi.Handler) *API {
	if old == nil {
		return nil
	}

	api := &API{
		EntryPoint:      old.EntryPoint,
		Dashboard:       old.Dashboard,
		DashboardAssets: old.DashboardAssets,
	}

	if old.Statistics != nil {
		api.Statistics = &types.Statistics{
			RecentErrors: old.Statistics.RecentErrors,
		}
	}

	return api
}

func convertConstraints(oldConstraints oldtypes.Constraints) types.Constraints {
	constraints := types.Constraints{}
	for _, value := range oldConstraints {
		constraint := &types.Constraint{
			Key:       value.Key,
			MustMatch: value.MustMatch,
			Regex:     value.Regex,
		}

		constraints = append(constraints, constraint)
	}
	return constraints
}

func convertFile(old *oldfile.Provider) *file.Provider {
	if old == nil {
		return nil
	}

	f := &file.Provider{
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
func ConvertHostResolverConfig(oldconfig *configuration.HostResolverConfig) *HostResolverConfig {
	if oldconfig == nil {
		return nil
	}

	return &HostResolverConfig{
		CnameFlattening: oldconfig.CnameFlattening,
		ResolvConfig:    oldconfig.ResolvConfig,
		ResolvDepth:     oldconfig.ResolvDepth,
	}
}
