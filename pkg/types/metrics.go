package types

import (
	"time"
)

// Metrics provides options to expose and send Traefik metrics to different third party monitoring systems
type Metrics struct {
	Prometheus *Prometheus `description:"Prometheus metrics exporter type." json:"prometheus,omitempty" toml:"prometheus,omitempty" yaml:"prometheus,omitempty" export:"true" label:"allowEmpty"`
	DataDog    *DataDog    `description:"DataDog metrics exporter type." json:"dataDog,omitempty" toml:"dataDog,omitempty" yaml:"dataDog,omitempty" export:"true" label:"allowEmpty"`
	StatsD     *Statsd     `description:"StatsD metrics exporter type." json:"statsD,omitempty" toml:"statsD,omitempty" yaml:"statsD,omitempty" export:"true" label:"allowEmpty"`
	InfluxDB   *InfluxDB   `description:"InfluxDB metrics exporter type." json:"influxDB,omitempty" toml:"influxDB,omitempty" yaml:"influxDB,omitempty" label:"allowEmpty"`
}

// Prometheus can contain specific configuration used by the Prometheus Metrics exporter
type Prometheus struct {
	Buckets       []float64 `description:"Buckets for latency metrics." json:"buckets,omitempty" toml:"buckets,omitempty" yaml:"buckets,omitempty" export:"true"`
	EntryPoint    string    `description:"EntryPoint." json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
	Middlewares   []string  `description:"Middlewares." json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	OnEntryPoints bool      `description:"Enable metrics on entry points." json:"onEntryPoints,omitempty" toml:"onEntryPoints,omitempty" yaml:"onEntryPoints,omitempty" export:"true"`
	OnServices    bool      `description:"Enable metrics on services." json:"onServices,omitempty" toml:"onServices,omitempty" yaml:"onServices,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (p *Prometheus) SetDefaults() {
	p.Buckets = []float64{0.1, 0.3, 1.2, 5}
	p.EntryPoint = "traefik"
	p.OnEntryPoints = true
	p.OnServices = true
}

// DataDog contains address and metrics pushing interval configuration
type DataDog struct {
	Address       string   `description:"DataDog's address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	PushInterval  Duration `description:"DataDog push interval." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
	OnEntryPoints bool     `description:"Enable metrics on entry points." json:"onEntryPoints,omitempty" toml:"onEntryPoints,omitempty" yaml:"onEntryPoints,omitempty" export:"true"`
	OnServices    bool     `description:"Enable metrics on services." json:"onServices,omitempty" toml:"onServices,omitempty" yaml:"onServices,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (d *DataDog) SetDefaults() {
	d.Address = "localhost:8125"
	d.PushInterval = Duration(10 * time.Second)
	d.OnEntryPoints = true
	d.OnServices = true
}

// Statsd contains address and metrics pushing interval configuration
type Statsd struct {
	OnEntryPoints bool     `description:"Enable metrics on entry points." json:"onEntryPoints,omitempty" toml:"onEntryPoints,omitempty" yaml:"onEntryPoints,omitempty" export:"true"`
	OnServices    bool     `description:"Enable metrics on services." json:"onServices,omitempty" toml:"onServices,omitempty" yaml:"onServices,omitempty" export:"true"`
	Address       string   `description:"StatsD address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	PushInterval  Duration `description:"StatsD push interval." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (s *Statsd) SetDefaults() {
	s.Address = "localhost:8125"
	s.PushInterval = Duration(10 * time.Second)
	s.OnEntryPoints = true
	s.OnServices = true
}

// InfluxDB contains address, login and metrics pushing interval configuration
type InfluxDB struct {
	Address         string   `description:"InfluxDB address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Protocol        string   `description:"InfluxDB address protocol (udp or http)." json:"protocol,omitempty" toml:"protocol,omitempty" yaml:"protocol,omitempty"`
	PushInterval    Duration `description:"InfluxDB push interval." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
	Database        string   `description:"InfluxDB database used when protocol is http." json:"database,omitempty" toml:"database,omitempty" yaml:"database,omitempty" export:"true"`
	RetentionPolicy string   `description:"InfluxDB retention policy used when protocol is http." json:"retentionPolicy,omitempty" toml:"retentionPolicy,omitempty" yaml:"retentionPolicy,omitempty" export:"true"`
	Username        string   `description:"InfluxDB username (only with http)." json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" export:"true"`
	Password        string   `description:"InfluxDB password (only with http)." json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" export:"true"`
	OnEntryPoints   bool     `description:"Enable metrics on entry points." json:"onEntryPoints,omitempty" toml:"onEntryPoints,omitempty" yaml:"onEntryPoints,omitempty" export:"true"`
	OnServices      bool     `description:"Enable metrics on services." json:"onServices,omitempty" toml:"onServices,omitempty" yaml:"onServices,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (i *InfluxDB) SetDefaults() {
	i.Address = "localhost:8089"
	i.Protocol = "udp"
	i.PushInterval = Duration(10 * time.Second)
	i.OnEntryPoints = true
	i.OnServices = true
}

// Statistics provides options for monitoring request and response stats
type Statistics struct {
	RecentErrors int `description:"Number of recent errors logged." json:"recentErrors,omitempty" toml:"recentErrors,omitempty" yaml:"recentErrors,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (s *Statistics) SetDefaults() {
	s.RecentErrors = 10
}
