package dns

import (
	"fmt"
	"strings"

	"gopkg.in/ns1/ns1-go.v2/rest/model/data"
	"gopkg.in/ns1/ns1-go.v2/rest/model/filter"
)

// Record wraps an NS1 /zone/{zone}/{domain}/{type} resource
type Record struct {
	Meta *data.Meta `json:"meta,omitempty"`

	ID              string `json:"id,omitempty"`
	Zone            string `json:"zone"`
	Domain          string `json:"domain"`
	Type            string `json:"type"`
	Link            string `json:"link,omitempty"`
	TTL             int    `json:"ttl,omitempty"`
	UseClientSubnet *bool  `json:"use_client_subnet,omitempty"`

	// Answers must all be of the same type as the record.
	Answers []*Answer `json:"answers"`
	// The records' filter chain.
	Filters []*filter.Filter `json:"filters,omitempty"`
	// The records' regions.
	Regions data.Regions `json:"regions,omitempty"`
}

func (r Record) String() string {
	return fmt.Sprintf("%s %s", r.Domain, r.Type)
}

// NewRecord takes a zone, domain and record type t and creates a *Record with
// UseClientSubnet: true & empty Answers.
func NewRecord(zone string, domain string, t string) *Record {
	if !strings.HasSuffix(domain, zone) {
		domain = fmt.Sprintf("%s.%s", domain, zone)
	}
	return &Record{
		Meta:    &data.Meta{},
		Zone:    zone,
		Domain:  domain,
		Type:    t,
		Answers: []*Answer{},
		Regions: data.Regions{},
	}
}

// LinkTo sets a Record Link to an FQDN.
// to is the FQDN of the target record whose config should be used. Does
// not have to be in the same zone.
func (r *Record) LinkTo(to string) {
	r.Meta = nil
	r.Answers = []*Answer{}
	r.Link = to
}

// AddAnswer adds an answer to the record.
func (r *Record) AddAnswer(ans *Answer) {
	if r.Answers == nil {
		r.Answers = []*Answer{}
	}

	r.Answers = append(r.Answers, ans)
}

// AddFilter adds a filter to the records' filter chain(ordering of filters matters).
func (r *Record) AddFilter(fil *filter.Filter) {
	if r.Filters == nil {
		r.Filters = []*filter.Filter{}
	}

	r.Filters = append(r.Filters, fil)
}
