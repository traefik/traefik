package dns

import "encoding/json"
import "gopkg.in/ns1/ns1-go.v2/rest/model/data"

// Zone wraps an NS1 /zone resource
type Zone struct {
	// Zones have metadata tables, but no filters act on 'zone-level' meta.
	Meta *data.Meta `json:"meta,omitempty"`

	// Read-only fields
	DNSServers   []string `json:"dns_servers,omitempty"`
	NetworkPools []string `json:"network_pools,omitempty"`
	Pool         string   `json:"pool,omitempty"` // Deprecated

	ID   string `json:"id,omitempty"`
	Zone string `json:"zone,omitempty"`

	TTL        int    `json:"ttl,omitempty"`
	NxTTL      int    `json:"nx_ttl,omitempty"`
	Retry      int    `json:"retry,omitempty"`
	Serial     int    `json:"serial,omitempty"`
	Refresh    int    `json:"refresh,omitempty"`
	Expiry     int    `json:"expiry,omitempty"`
	Hostmaster string `json:"hostmaster,omitempty"`

	// If this is a linked zone, Link points to an existing standard zone,
	// reusing its configuration and records. Link is a zones' domain name.
	Link *string `json:"link,omitempty"`

	// Networks contains the network ids the zone is available. Most zones
	// will be in the NSONE Global Network(which is id 0).
	NetworkIDs []int         `json:"networks,omitempty"`
	Records    []*ZoneRecord `json:"records,omitempty"`

	// Primary contains info to enable slaving of the zone by third party dns servers.
	Primary *ZonePrimary `json:"primary,omitempty"`
	// Secondary contains info for slaving the zone to a primary dns server.
	Secondary *ZoneSecondary `json:"secondary,omitempty"`
}

func (z Zone) String() string {
	return z.Zone
}

// ZoneRecord wraps Zone's "records" attribute
type ZoneRecord struct {
	Domain   string      `json:"Domain,omitempty"`
	ID       string      `json:"id,omitempty"`
	Link     string      `json:"link,omitempty"`
	ShortAns []string    `json:"short_answers,omitempty"`
	Tier     json.Number `json:"tier,omitempty"`
	TTL      int         `json:"ttl,omitempty"`
	Type     string      `json:"type,omitempty"`
}

// ZonePrimary wraps a Zone's "primary" attribute
type ZonePrimary struct {
	// Enabled determines whether AXFR queries (and optionally NOTIFY messages)
	// will be enabled for the zone.
	Enabled     bool                  `json:"enabled"`
	Secondaries []ZoneSecondaryServer `json:"secondaries"`
}

// ZoneSecondaryServer wraps elements of a Zone's "primary.secondary" attribute
type ZoneSecondaryServer struct {
	// Read-Only
	NetworkIDs []int `json:"networks,omitempty"`

	IP     string `json:"ip"`
	Port   int    `json:"port,omitempty"`
	Notify bool   `json:"notify"`
}

// ZoneSecondary wraps a Zone's "secondary" attribute
type ZoneSecondary struct {
	// Read-Only fields
	Expired bool    `json:"expired,omitempty"`
	LastXfr int     `json:"last_xfr,omitempty"`
	Status  string  `json:"status,omitempty"`
	Error   *string `json:"error"`

	PrimaryIP   string `json:"primary_ip,omitempty"`
	PrimaryPort int    `json:"primary_port,omitempty"`
	Enabled     bool   `json:"enabled"`

	TSIG *TSIG `json:"tsig"`
}

// TSIG is a zones transaction signature.
type TSIG struct {
	// Key is the encrypted TSIG key(read-only)
	Key string `json:"key,omitempty"`

	// Whether TSIG is enabled for a secondary zone.
	Enabled bool `json:"enabled,omitempty"`
	// Which hashing algorithm
	Hash string `json:"hash,omitempty"`
	// Name of the TSIG key
	Name string `json:"name,omitempty"`
}

// NewZone takes a zone domain name and creates a new zone.
func NewZone(zone string) *Zone {
	z := Zone{
		Zone: zone,
	}
	return &z
}

// MakePrimary enables Primary, disables Secondary, and sets primary's
// Secondaries to all provided ZoneSecondaryServers
func (z *Zone) MakePrimary(secondaries ...ZoneSecondaryServer) {
	z.Secondary = nil
	z.Primary = &ZonePrimary{
		Enabled:     true,
		Secondaries: secondaries,
	}
	if z.Primary.Secondaries == nil {
		z.Primary.Secondaries = make([]ZoneSecondaryServer, 0)
	}
}

// MakeSecondary enables Secondary, disables Primary, and sets secondary's
// Primary_ip to provided ip.
func (z *Zone) MakeSecondary(ip string) {
	z.Secondary = &ZoneSecondary{
		Enabled:     true,
		PrimaryIP:   ip,
		PrimaryPort: 53,
	}
	z.Primary = &ZonePrimary{
		Enabled:     false,
		Secondaries: make([]ZoneSecondaryServer, 0),
	}
}

// LinkTo sets Link to a target zone domain name and unsets all other configuration properties.
// No other zone configuration properties (such as refresh, retry, etc) may be specified,
// since they are all pulled from the target zone. Linked zones, once created, cannot be
// configured at all and cannot have records added to them. They may only be deleted, which
// does not affect the target zone at all.
func (z *Zone) LinkTo(to string) {
	z.Meta = nil
	z.TTL = 0
	z.NxTTL = 0
	z.Retry = 0
	z.Refresh = 0
	z.Expiry = 0
	z.Primary = nil
	z.DNSServers = nil
	z.NetworkIDs = nil
	z.NetworkPools = nil
	z.Hostmaster = ""
	z.Pool = ""
	z.Secondary = nil
	z.Link = &to
}
