package egoscale

import "net"

// Zone represents a data center
type Zone struct {
	ID                    string            `json:"id"`
	AllocationState       string            `json:"allocationstate,omitempty"`
	Capacity              string            `json:"capacity,omitempty"`
	Description           string            `json:"description,omitempty"`
	DhcpProvider          string            `json:"dhcpprovider,omitempty"`
	DisplayText           string            `json:"displaytext,omitempty"`
	DNS1                  net.IP            `json:"dns1,omitempty"`
	DNS2                  net.IP            `json:"dns2,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainID              string            `json:"domainid,omitempty"`
	DomainName            string            `json:"domainname,omitempty"`
	GuestCidrAddress      string            `json:"guestcidraddress,omitempty"`
	InternalDNS1          net.IP            `json:"internaldns1,omitempty"`
	InternalDNS2          net.IP            `json:"internaldns2,omitempty"`
	IP6DNS1               net.IP            `json:"ip6dns1,omitempty"`
	IP6DNS2               net.IP            `json:"ip6dns2,omitempty"`
	LocalStorageEnabled   bool              `json:"localstorageenabled,omitempty"`
	Name                  string            `json:"name,omitempty"`
	NetworkType           string            `json:"networktype,omitempty"`
	ResourceDetails       map[string]string `json:"resourcedetails,omitempty"`
	SecurityGroupsEnabled bool              `json:"securitygroupsenabled,omitempty"`
	Vlan                  string            `json:"vlan,omitempty"`
	ZoneToken             string            `json:"zonetoken,omitempty"`
	Tags                  []ResourceTag     `json:"tags,omitempty"`
}

// ListZones represents a query for zones
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listZones.html
type ListZones struct {
	Available      bool          `json:"available,omitempty"`
	DomainID       string        `json:"domainid,omitempty"`
	ID             string        `json:"id,omitempty"`
	Keyword        string        `json:"keyword,omitempty"`
	Name           string        `json:"name,omitempty"`
	Page           int           `json:"page,omitempty"`
	PageSize       int           `json:"pagesize,omitempty"`
	ShowCapacities bool          `json:"showcapacities,omitempty"`
	Tags           []ResourceTag `json:"tags,omitempty"`
}

func (*ListZones) name() string {
	return "listZones"
}

func (*ListZones) response() interface{} {
	return new(ListZonesResponse)
}

// ListZonesResponse represents a list of zones
type ListZonesResponse struct {
	Count int    `json:"count"`
	Zone  []Zone `json:"zone"`
}
