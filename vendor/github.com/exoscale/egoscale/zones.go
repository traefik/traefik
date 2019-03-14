package egoscale

import (
	"net"
)

// Zone represents a data center
//
// TODO: represent correctly the capacity field.
type Zone struct {
	AllocationState       string            `json:"allocationstate,omitempty" doc:"the allocation state of the cluster"`
	Description           string            `json:"description,omitempty" doc:"Zone description"`
	DhcpProvider          string            `json:"dhcpprovider,omitempty" doc:"the dhcp Provider for the Zone"`
	DisplayText           string            `json:"displaytext,omitempty" doc:"the display text of the zone"`
	DNS1                  net.IP            `json:"dns1,omitempty" doc:"the first DNS for the Zone"`
	DNS2                  net.IP            `json:"dns2,omitempty" doc:"the second DNS for the Zone"`
	GuestCIDRAddress      *CIDR             `json:"guestcidraddress,omitempty" doc:"the guest CIDR address for the Zone"`
	ID                    *UUID             `json:"id,omitempty" doc:"Zone id"`
	InternalDNS1          net.IP            `json:"internaldns1,omitempty" doc:"the first internal DNS for the Zone"`
	InternalDNS2          net.IP            `json:"internaldns2,omitempty" doc:"the second internal DNS for the Zone"`
	IP6DNS1               net.IP            `json:"ip6dns1,omitempty" doc:"the first IPv6 DNS for the Zone"`
	IP6DNS2               net.IP            `json:"ip6dns2,omitempty" doc:"the second IPv6 DNS for the Zone"`
	LocalStorageEnabled   *bool             `json:"localstorageenabled,omitempty" doc:"true if local storage offering enabled, false otherwise"`
	Name                  string            `json:"name,omitempty" doc:"Zone name"`
	NetworkType           string            `json:"networktype,omitempty" doc:"the network type of the zone; can be Basic or Advanced"`
	ResourceDetails       map[string]string `json:"resourcedetails,omitempty" doc:"Meta data associated with the zone (key/value pairs)"`
	SecurityGroupsEnabled *bool             `json:"securitygroupsenabled,omitempty" doc:"true if security groups support is enabled, false otherwise"`
	Tags                  []ResourceTag     `json:"tags,omitempty" doc:"the list of resource tags associated with zone."`
	Vlan                  string            `json:"vlan,omitempty" doc:"the vlan range of the zone"`
	ZoneToken             string            `json:"zonetoken,omitempty" doc:"Zone Token"`
}

// ListRequest builds the ListZones request
func (zone Zone) ListRequest() (ListCommand, error) {
	req := &ListZones{
		ID:   zone.ID,
		Name: zone.Name,
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListZones

// ListZones represents a query for zones
type ListZones struct {
	Available      *bool         `json:"available,omitempty" doc:"true if you want to retrieve all available Zones. False if you only want to return the Zones from which you have at least one VM. Default is false."`
	ID             *UUID         `json:"id,omitempty" doc:"the ID of the zone"`
	Keyword        string        `json:"keyword,omitempty" doc:"List by keyword"`
	Name           string        `json:"name,omitempty" doc:"the name of the zone"`
	Page           int           `json:"page,omitempty"`
	PageSize       int           `json:"pagesize,omitempty"`
	ShowCapacities *bool         `json:"showcapacities,omitempty" doc:"flag to display the capacity of the zones"`
	Tags           []ResourceTag `json:"tags,omitempty" doc:"List zones by resource tags (key/value pairs)"`
	_              bool          `name:"listZones" description:"Lists zones"`
}

// ListZonesResponse represents a list of zones
type ListZonesResponse struct {
	Count int    `json:"count"`
	Zone  []Zone `json:"zone"`
}
