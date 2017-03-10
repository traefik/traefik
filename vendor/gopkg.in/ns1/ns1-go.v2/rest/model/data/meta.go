package data

// FeedPtr represents the dynamic metadata value in which a feed is providing the value.
type FeedPtr struct {
	FeedID string `json:"feed,omitempty"`
}

// Meta contains information on an entities metadata table. Metadata key/value
// pairs are used by a records' filter pipeline during a dns query.
// All values can be a feed id as well, indicating real-time updates of these values.
// Structure/Precendence of metadata tables:
//  - Record
//    - Meta <- lowest precendence in filter
//    - Region(s)
//      - Meta <- middle precedence in filter chain
//      - ...
//    - Answer(s)
//      - Meta <- highest precedence in filter chain
//      - ...
//    - ...
type Meta struct {
	// STATUS

	// Indicates whether or not entity is considered 'up'
	// bool or FeedPtr.
	Up interface{} `json:"up,omitempty"`

	// Indicates the number of active connections.
	// Values must be positive.
	// int or FeedPtr.
	Connections interface{} `json:"connections,omitempty"`

	// Indicates the number of active requests (HTTP or otherwise).
	// Values must be positive.
	// int or FeedPtr.
	Requests interface{} `json:"requests,omitempty"`

	// Indicates the "load average".
	// Values must be positive, and will be rounded to the nearest tenth.
	// float64 or FeedPtr.
	LoadAvg interface{} `json:"loadavg,omitempty"`

	// The Job ID of a Pulsar telemetry gathering job and routing granularities
	// to associate with.
	// string or FeedPtr.
	Pulsar interface{} `json:"pulsar,omitempty"`

	// GEOGRAPHICAL

	// Must be between -180.0 and +180.0 where negative
	// indicates South and positive indicates North.
	// e.g., the longitude of the datacenter where a server resides.
	// float64 or FeedPtr.
	Latitude interface{} `json:"latitude,omitempty"`

	// Must be between -180.0 and +180.0 where negative
	// indicates West and positive indicates East.
	// e.g., the longitude of the datacenter where a server resides.
	// float64 or FeedPtr.
	Longitude interface{} `json:"longitude,omitempty"`

	// Valid geographic regions are: 'US-EAST', 'US-CENTRAL', 'US-WEST',
	// 'EUROPE', 'ASIAPAC', 'SOUTH-AMERICA', 'AFRICA'.
	// e.g., the rough geographic location of the Datacenter where a server resides.
	// []string or FeedPtr.
	Georegion interface{} `json:"georegion,omitempty"`

	// Countr(ies) must be specified as ISO3166 2-character country code(s).
	// []string or FeedPtr.
	Country interface{} `json:"country,omitempty"`

	// State(s) must be specified as standard 2-character state code(s).
	// []string or FeedPtr.
	USState interface{} `json:"us_state,omitempty"`

	// Canadian Province(s) must be specified as standard 2-character province
	// code(s).
	// []string or FeedPtr.
	CAProvince interface{} `json:"ca_province,omitempty"`

	// INFORMATIONAL

	// Notes to indicate any necessary details for operators.
	// Up to 256 characters in length.
	// string or FeedPtr.
	Note interface{} `json:"note,omitempty"`

	// NETWORK

	// IP (v4 and v6) prefixes in CIDR format ("a.b.c.d/mask").
	// May include up to 1000 prefixes.
	// e.g., "1.2.3.4/24"
	// []string or FeedPtr.
	IPPrefixes interface{} `json:"ip_prefixes,omitempty"`

	// Autonomous System (AS) number(s).
	// May include up to 1000 AS numbers.
	// []string or FeedPtr.
	ASN interface{} `json:"asn,omitempty"`

	// TRAFFIC

	// Indicates the "priority tier".
	// Lower values indicate higher priority.
	// Values must be positive.
	// int or FeedPtr.
	Priority interface{} `json:"priority,omitempty"`

	// Indicates a weight.
	// Filters that use weights normalize them.
	// Any positive values are allowed.
	// Values between 0 and 100 are recommended for simplicity's sake.
	// float64 or FeedPtr.
	Weight interface{} `json:"weight,omitempty"`

	// Indicates a "low watermark" to use for load shedding.
	// The value should depend on the metric used to determine
	// load (e.g., loadavg, connections, etc).
	// int or FeedPtr.
	LowWatermark interface{} `json:"low_watermark,omitempty"`

	// Indicates a "high watermark" to use for load shedding.
	// The value should depend on the metric used to determine
	// load (e.g., loadavg, connections, etc).
	// int or FeedPtr.
	HighWatermark interface{} `json:"high_watermark,omitempty"`
}
