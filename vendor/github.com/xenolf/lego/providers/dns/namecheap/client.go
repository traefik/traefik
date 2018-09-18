package namecheap

import "encoding/xml"

// host describes a DNS record returned by the Namecheap DNS gethosts API.
// Namecheap uses the term "host" to refer to all DNS records that include
// a host field (A, AAAA, CNAME, NS, TXT, URL).
type host struct {
	Type    string `xml:",attr"`
	Name    string `xml:",attr"`
	Address string `xml:",attr"`
	MXPref  string `xml:",attr"`
	TTL     string `xml:",attr"`
}

// apierror describes an error record in a namecheap API response.
type apierror struct {
	Number      int    `xml:",attr"`
	Description string `xml:",innerxml"`
}

type setHostsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Status  string     `xml:"Status,attr"`
	Errors  []apierror `xml:"Errors>Error"`
	Result  struct {
		IsSuccess string `xml:",attr"`
	} `xml:"CommandResponse>DomainDNSSetHostsResult"`
}

type getHostsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Status  string     `xml:"Status,attr"`
	Errors  []apierror `xml:"Errors>Error"`
	Hosts   []host     `xml:"CommandResponse>DomainDNSGetHostsResult>host"`
}

type getTldsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Errors  []apierror `xml:"Errors>Error"`
	Result  []struct {
		Name string `xml:",attr"`
	} `xml:"CommandResponse>Tlds>Tld"`
}
