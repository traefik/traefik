package types

import (
	"fmt"
	"strconv"
	"strings"
)

// Metrics provides options to expose and send Traefik metrics to different third party monitoring systems
type Metrics struct {
	Prometheus *Prometheus `description:"Prometheus metrics exporter type" export:"true"`
	Datadog    *Datadog    `description:"DataDog metrics exporter type" export:"true"`
	StatsD     *Statsd     `description:"StatsD metrics exporter type" export:"true"`
	InfluxDB   *InfluxDB   `description:"InfluxDB metrics exporter type"`
}

// Prometheus can contain specific configuration used by the Prometheus Metrics exporter
type Prometheus struct {
	Buckets     Buckets  `description:"Buckets for latency metrics" export:"true"`
	EntryPoint  string   `description:"EntryPoint" export:"true"`
	Middlewares []string `description:"Middlewares" export:"true"`
}

// Datadog contains address and metrics pushing interval configuration
type Datadog struct {
	Address      string `description:"DataDog's address"`
	PushInterval string `description:"DataDog push interval" export:"true"`
}

// Statsd contains address and metrics pushing interval configuration
type Statsd struct {
	Address      string `description:"StatsD address"`
	PushInterval string `description:"StatsD push interval" export:"true"`
}

// InfluxDB contains address, login and metrics pushing interval configuration
type InfluxDB struct {
	Address         string `description:"InfluxDB address"`
	Protocol        string `description:"InfluxDB address protocol (udp or http)"`
	PushInterval    string `description:"InfluxDB push interval" export:"true"`
	Database        string `description:"InfluxDB database used when protocol is http" export:"true"`
	RetentionPolicy string `description:"InfluxDB retention policy used when protocol is http" export:"true"`
	Username        string `description:"InfluxDB username (only with http)" export:"true"`
	Password        string `description:"InfluxDB password (only with http)" export:"true"`
}

// Statistics provides options for monitoring request and response stats
type Statistics struct {
	RecentErrors int `description:"Number of recent errors logged" export:"true"`
}

// Buckets holds Prometheus Buckets
type Buckets []float64

// Set adds strings elem into the the parser
// it splits str on "," and ";" and apply ParseFloat to string
func (b *Buckets) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	for _, bucket := range slice {
		bu, err := strconv.ParseFloat(bucket, 64)
		if err != nil {
			return err
		}
		*b = append(*b, bu)
	}
	return nil
}

// Get []float64
func (b *Buckets) Get() interface{} { return *b }

// String return slice in a string
func (b *Buckets) String() string { return fmt.Sprintf("%v", *b) }

// SetValue sets []float64 into the parser
func (b *Buckets) SetValue(val interface{}) {
	*b = val.(Buckets)
}
