package dns

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/ns1/ns1-go.v2/rest/model/data"
)

// Answer wraps the values of a Record's "filters" attribute
type Answer struct {
	Meta *data.Meta `json:"meta,omitempty"`

	// Answer response data. eg:
	// Av4: ["1.1.1.1"]
	// Av6: ["2001:db8:85a3::8a2e:370:7334"]
	// MX:  [10, "2.2.2.2"]
	Rdata []string `json:"answer"`

	// Region(grouping) that answer belongs to.
	RegionName string `json:"region,omitempty"`
}

func (a Answer) String() string {
	return strings.Trim(fmt.Sprint(a.Rdata), "[]")
}

// SetRegion associates a region with this answer.
func (a *Answer) SetRegion(name string) {
	a.RegionName = name
}

// NewAnswer creates a generic Answer with given rdata.
func NewAnswer(rdata []string) *Answer {
	return &Answer{
		Meta:  &data.Meta{},
		Rdata: rdata,
	}
}

// NewAv4Answer creates an Answer for A record.
func NewAv4Answer(host string) *Answer {
	return &Answer{
		Meta:  &data.Meta{},
		Rdata: []string{host},
	}
}

// NewAv6Answer creates an Answer for AAAA record.
func NewAv6Answer(host string) *Answer {
	return &Answer{
		Meta:  &data.Meta{},
		Rdata: []string{host},
	}
}

// NewALIASAnswer creates an Answer for ALIAS record.
func NewALIASAnswer(host string) *Answer {
	return &Answer{
		Meta:  &data.Meta{},
		Rdata: []string{host},
	}
}

// NewCNAMEAnswer creates an Answer for CNAME record.
func NewCNAMEAnswer(name string) *Answer {
	return &Answer{
		Meta:  &data.Meta{},
		Rdata: []string{name},
	}
}

// NewTXTAnswer creates an Answer for TXT record.
func NewTXTAnswer(text string) *Answer {
	return &Answer{
		Meta:  &data.Meta{},
		Rdata: []string{text},
	}
}

// NewMXAnswer creates an Answer for MX record.
func NewMXAnswer(pri int, host string) *Answer {
	return &Answer{
		Meta:  &data.Meta{},
		Rdata: []string{strconv.Itoa(pri), host},
	}
}

// NewSRVAnswer creates an Answer for SRV record.
func NewSRVAnswer(priority, weight, port int, target string) *Answer {
	return &Answer{
		Meta: &data.Meta{},
		Rdata: []string{
			strconv.Itoa(priority),
			strconv.Itoa(weight),
			strconv.Itoa(port),
			target,
		},
	}
}
