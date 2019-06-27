package types

import (
	"time"
)

// Metrics provides options to expose and send Traefik metrics to different third party monitoring systems
type Metrics struct {
	Prometheus *Prometheus `description:"Prometheus metrics exporter type." export:"true" label:"allowEmpty"`
	Datadog    *Datadog    `description:"DataDog metrics exporter type." export:"true" label:"allowEmpty"`
	StatsD     *Statsd     `description:"StatsD metrics exporter type." export:"true" label:"allowEmpty"`
	InfluxDB   *InfluxDB   `description:"InfluxDB metrics exporter type." label:"allowEmpty"`
}

// Prometheus can contain specific configuration used by the Prometheus Metrics exporter
type Prometheus struct {
	Buckets     []float64 `description:"Buckets for latency metrics." export:"true"`
	EntryPoint  string    `description:"EntryPoint." export:"true"`
	Middlewares []string  `description:"Middlewares." export:"true"`
}

// SetDefaults sets the default values.
func (p *Prometheus) SetDefaults() {
	p.Buckets = []float64{0.1, 0.3, 1.2, 5}
	p.EntryPoint = "traefik"
}

// Datadog contains address and metrics pushing interval configuration
type Datadog struct {
	Address      string   `description:"DataDog's address."`
	PushInterval Duration `description:"DataDog push interval." export:"true"`
}

// SetDefaults sets the default values.
func (d *Datadog) SetDefaults() {
	d.Address = "localhost:8125"
	d.PushInterval = Duration(10 * time.Second)
}

// Statsd contains address and metrics pushing interval configuration
type Statsd struct {
	Address      string   `description:"StatsD address."`
	PushInterval Duration `description:"StatsD push interval." export:"true"`
}

// SetDefaults sets the default values.
func (s *Statsd) SetDefaults() {
	s.Address = "localhost:8125"
	s.PushInterval = Duration(10 * time.Second)
}

// InfluxDB contains address, login and metrics pushing interval configuration
type InfluxDB struct {
	Address         string   `description:"InfluxDB address."`
	Protocol        string   `description:"InfluxDB address protocol (udp or http)."`
	PushInterval    Duration `description:"InfluxDB push interval." export:"true"`
	Database        string   `description:"InfluxDB database used when protocol is http." export:"true"`
	RetentionPolicy string   `description:"InfluxDB retention policy used when protocol is http." export:"true"`
	Username        string   `description:"InfluxDB username (only with http)." export:"true"`
	Password        string   `description:"InfluxDB password (only with http)." export:"true"`
}

// SetDefaults sets the default values.
func (i *InfluxDB) SetDefaults() {
	i.Address = "localhost:8089"
	i.Protocol = "udp"
	i.PushInterval = Duration(10 * time.Second)
}

// Statistics provides options for monitoring request and response stats
type Statistics struct {
	RecentErrors int `description:"Number of recent errors logged." export:"true"`
}

// SetDefaults sets the default values.
func (s *Statistics) SetDefaults() {
	s.RecentErrors = 10
}
