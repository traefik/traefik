package types

import (
	"net"
	"os"
	"time"

	"github.com/traefik/paerser/types"
)

// Metrics provides options to expose and send Traefik metrics to different third party monitoring systems.
type Metrics struct {
	AddInternals bool `description:"Enables metrics for internal services (ping, dashboard, etc...)." json:"addInternals,omitempty" toml:"addInternals,omitempty" yaml:"addInternals,omitempty" export:"true"`

	Prometheus *Prometheus `description:"Prometheus metrics exporter type." json:"prometheus,omitempty" toml:"prometheus,omitempty" yaml:"prometheus,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Datadog    *Datadog    `description:"Datadog metrics exporter type." json:"datadog,omitempty" toml:"datadog,omitempty" yaml:"datadog,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	StatsD     *Statsd     `description:"StatsD metrics exporter type." json:"statsD,omitempty" toml:"statsD,omitempty" yaml:"statsD,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	InfluxDB2  *InfluxDB2  `description:"InfluxDB v2 metrics exporter type." json:"influxDB2,omitempty" toml:"influxDB2,omitempty" yaml:"influxDB2,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	OTLP       *OTLP       `description:"OpenTelemetry metrics exporter type." json:"otlp,omitempty" toml:"otlp,omitempty" yaml:"otlp,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// Prometheus can contain specific configuration used by the Prometheus Metrics exporter.
type Prometheus struct {
	Buckets              []float64         `description:"Buckets for latency metrics." json:"buckets,omitempty" toml:"buckets,omitempty" yaml:"buckets,omitempty" export:"true"`
	AddEntryPointsLabels bool              `description:"Enable metrics on entry points." json:"addEntryPointsLabels,omitempty" toml:"addEntryPointsLabels,omitempty" yaml:"addEntryPointsLabels,omitempty" export:"true"`
	AddRoutersLabels     bool              `description:"Enable metrics on routers." json:"addRoutersLabels,omitempty" toml:"addRoutersLabels,omitempty" yaml:"addRoutersLabels,omitempty" export:"true"`
	AddServicesLabels    bool              `description:"Enable metrics on services." json:"addServicesLabels,omitempty" toml:"addServicesLabels,omitempty" yaml:"addServicesLabels,omitempty" export:"true"`
	EntryPoint           string            `description:"EntryPoint" json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
	ManualRouting        bool              `description:"Manual routing" json:"manualRouting,omitempty" toml:"manualRouting,omitempty" yaml:"manualRouting,omitempty" export:"true"`
	HeaderLabels         map[string]string `description:"Defines the extra labels for the requests_total metrics, and for each of them, the request header containing the value for this label." json:"headerLabels,omitempty" toml:"headerLabels,omitempty" yaml:"headerLabels,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (p *Prometheus) SetDefaults() {
	p.Buckets = []float64{0.1, 0.3, 1.2, 5}
	p.AddEntryPointsLabels = true
	p.AddServicesLabels = true
	p.EntryPoint = "traefik"
}

// Datadog contains address and metrics pushing interval configuration.
type Datadog struct {
	Address              string         `description:"Datadog's address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	PushInterval         types.Duration `description:"Datadog push interval." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
	AddEntryPointsLabels bool           `description:"Enable metrics on entry points." json:"addEntryPointsLabels,omitempty" toml:"addEntryPointsLabels,omitempty" yaml:"addEntryPointsLabels,omitempty" export:"true"`
	AddRoutersLabels     bool           `description:"Enable metrics on routers." json:"addRoutersLabels,omitempty" toml:"addRoutersLabels,omitempty" yaml:"addRoutersLabels,omitempty" export:"true"`
	AddServicesLabels    bool           `description:"Enable metrics on services." json:"addServicesLabels,omitempty" toml:"addServicesLabels,omitempty" yaml:"addServicesLabels,omitempty" export:"true"`
	Prefix               string         `description:"Prefix to use for metrics collection." json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (d *Datadog) SetDefaults() {
	host, ok := os.LookupEnv("DD_AGENT_HOST")
	if !ok {
		host = "localhost"
	}

	port, ok := os.LookupEnv("DD_DOGSTATSD_PORT")
	if !ok {
		port = "8125"
	}
	d.Address = net.JoinHostPort(host, port)
	d.PushInterval = types.Duration(10 * time.Second)
	d.AddEntryPointsLabels = true
	d.AddServicesLabels = true
	d.Prefix = "traefik"
}

// Statsd contains address and metrics pushing interval configuration.
type Statsd struct {
	Address              string         `description:"StatsD address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	PushInterval         types.Duration `description:"StatsD push interval." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
	AddEntryPointsLabels bool           `description:"Enable metrics on entry points." json:"addEntryPointsLabels,omitempty" toml:"addEntryPointsLabels,omitempty" yaml:"addEntryPointsLabels,omitempty" export:"true"`
	AddRoutersLabels     bool           `description:"Enable metrics on routers." json:"addRoutersLabels,omitempty" toml:"addRoutersLabels,omitempty" yaml:"addRoutersLabels,omitempty" export:"true"`
	AddServicesLabels    bool           `description:"Enable metrics on services." json:"addServicesLabels,omitempty" toml:"addServicesLabels,omitempty" yaml:"addServicesLabels,omitempty" export:"true"`
	Prefix               string         `description:"Prefix to use for metrics collection." json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (s *Statsd) SetDefaults() {
	s.Address = "localhost:8125"
	s.PushInterval = types.Duration(10 * time.Second)
	s.AddEntryPointsLabels = true
	s.AddServicesLabels = true
	s.Prefix = "traefik"
}

// InfluxDB2 contains address, token and metrics pushing interval configuration.
type InfluxDB2 struct {
	Address              string            `description:"InfluxDB v2 address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Token                string            `description:"InfluxDB v2 access token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	PushInterval         types.Duration    `description:"InfluxDB v2 push interval." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
	Org                  string            `description:"InfluxDB v2 org ID." json:"org,omitempty" toml:"org,omitempty" yaml:"org,omitempty" export:"true"`
	Bucket               string            `description:"InfluxDB v2 bucket ID." json:"bucket,omitempty" toml:"bucket,omitempty" yaml:"bucket,omitempty" export:"true"`
	AddEntryPointsLabels bool              `description:"Enable metrics on entry points." json:"addEntryPointsLabels,omitempty" toml:"addEntryPointsLabels,omitempty" yaml:"addEntryPointsLabels,omitempty" export:"true"`
	AddRoutersLabels     bool              `description:"Enable metrics on routers." json:"addRoutersLabels,omitempty" toml:"addRoutersLabels,omitempty" yaml:"addRoutersLabels,omitempty" export:"true"`
	AddServicesLabels    bool              `description:"Enable metrics on services." json:"addServicesLabels,omitempty" toml:"addServicesLabels,omitempty" yaml:"addServicesLabels,omitempty" export:"true"`
	AdditionalLabels     map[string]string `description:"Additional labels (influxdb tags) on all metrics" json:"additionalLabels,omitempty" toml:"additionalLabels,omitempty" yaml:"additionalLabels,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (i *InfluxDB2) SetDefaults() {
	i.Address = "http://localhost:8086"
	i.PushInterval = types.Duration(10 * time.Second)
	i.AddEntryPointsLabels = true
	i.AddServicesLabels = true
}

// OTLP contains specific configuration used by the OpenTelemetry Metrics exporter.
type OTLP struct {
	GRPC *OTelGRPC `description:"gRPC configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	HTTP *OTelHTTP `description:"HTTP configuration for the OpenTelemetry collector." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	AddEntryPointsLabels bool              `description:"Enable metrics on entry points." json:"addEntryPointsLabels,omitempty" toml:"addEntryPointsLabels,omitempty" yaml:"addEntryPointsLabels,omitempty" export:"true"`
	AddRoutersLabels     bool              `description:"Enable metrics on routers." json:"addRoutersLabels,omitempty" toml:"addRoutersLabels,omitempty" yaml:"addRoutersLabels,omitempty" export:"true"`
	AddServicesLabels    bool              `description:"Enable metrics on services." json:"addServicesLabels,omitempty" toml:"addServicesLabels,omitempty" yaml:"addServicesLabels,omitempty" export:"true"`
	ExplicitBoundaries   []float64         `description:"Boundaries for latency metrics." json:"explicitBoundaries,omitempty" toml:"explicitBoundaries,omitempty" yaml:"explicitBoundaries,omitempty" export:"true"`
	PushInterval         types.Duration    `description:"Period between calls to collect a checkpoint." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
	ServiceName          string            `description:"Defines the service name resource attribute." json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty" export:"true"`
	ResourceAttributes   map[string]string `description:"Defines additional resource attributes (key:value)." json:"resourceAttributes,omitempty" toml:"resourceAttributes,omitempty" yaml:"resourceAttributes,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (o *OTLP) SetDefaults() {
	o.HTTP = &OTelHTTP{}
	o.HTTP.SetDefaults()

	o.AddEntryPointsLabels = true
	o.AddServicesLabels = true
	o.ExplicitBoundaries = []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10}
	o.PushInterval = types.Duration(10 * time.Second)
	o.ServiceName = OTelTraefikServiceName
}

// Statistics provides options for monitoring request and response stats.
type Statistics struct {
	RecentErrors int `description:"Number of recent errors logged." json:"recentErrors,omitempty" toml:"recentErrors,omitempty" yaml:"recentErrors,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (s *Statistics) SetDefaults() {
	s.RecentErrors = 10
}
