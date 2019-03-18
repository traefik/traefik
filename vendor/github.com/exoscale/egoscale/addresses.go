package egoscale

import (
	"context"
	"fmt"
	"net"
)

// IPAddress represents an IP Address
type IPAddress struct {
	Allocated                 string        `json:"allocated,omitempty" doc:"date the public IP address was acquired"`
	Associated                string        `json:"associated,omitempty" doc:"date the public IP address was associated"`
	AssociatedNetworkID       *UUID         `json:"associatednetworkid,omitempty" doc:"the ID of the Network associated with the IP address"`
	AssociatedNetworkName     string        `json:"associatednetworkname,omitempty" doc:"the name of the Network associated with the IP address"`
	ForVirtualNetwork         bool          `json:"forvirtualnetwork,omitempty" doc:"the virtual network for the IP address"`
	ID                        *UUID         `json:"id,omitempty" doc:"public IP address id"`
	IPAddress                 net.IP        `json:"ipaddress,omitempty" doc:"public IP address"`
	IsElastic                 bool          `json:"iselastic,omitempty" doc:"is an elastic ip"`
	IsPortable                bool          `json:"isportable,omitempty" doc:"is public IP portable across the zones"`
	IsSourceNat               bool          `json:"issourcenat,omitempty" doc:"true if the IP address is a source nat address, false otherwise"`
	IsStaticNat               *bool         `json:"isstaticnat,omitempty" doc:"true if this ip is for static nat, false otherwise"`
	IsSystem                  bool          `json:"issystem,omitempty" doc:"true if this ip is system ip (was allocated as a part of deployVm or createLbRule)"`
	NetworkID                 *UUID         `json:"networkid,omitempty" doc:"the ID of the Network where ip belongs to"`
	PhysicalNetworkID         *UUID         `json:"physicalnetworkid,omitempty" doc:"the physical network this belongs to"`
	Purpose                   string        `json:"purpose,omitempty" doc:"purpose of the IP address. In Acton this value is not null for Ips with isSystem=true, and can have either StaticNat or LB value"`
	ReverseDNS                []ReverseDNS  `json:"reversedns,omitempty" doc:"the list of PTR record(s) associated with the ip address"`
	State                     string        `json:"state,omitempty" doc:"State of the ip address. Can be: Allocatin, Allocated and Releasing"`
	Tags                      []ResourceTag `json:"tags,omitempty" doc:"the list of resource tags associated with ip address"`
	VirtualMachineDisplayName string        `json:"virtualmachinedisplayname,omitempty" doc:"virtual machine display name the ip address is assigned to (not null only for static nat Ip)"`
	VirtualMachineID          *UUID         `json:"virtualmachineid,omitempty" doc:"virtual machine id the ip address is assigned to (not null only for static nat Ip)"`
	VirtualMachineName        string        `json:"virtualmachinename,omitempty" doc:"virtual machine name the ip address is assigned to (not null only for static nat Ip)"`
	VlanID                    *UUID         `json:"vlanid,omitempty" doc:"the ID of the VLAN associated with the IP address. This parameter is visible to ROOT admins only"`
	VlanName                  string        `json:"vlanname,omitempty" doc:"the VLAN associated with the IP address"`
	VMIPAddress               net.IP        `json:"vmipaddress,omitempty" doc:"virtual machine (dnat) ip address (not null only for static nat Ip)"`
	ZoneID                    *UUID         `json:"zoneid,omitempty" doc:"the ID of the zone the public IP address belongs to"`
	ZoneName                  string        `json:"zonename,omitempty" doc:"the name of the zone the public IP address belongs to"`
}

// ResourceType returns the type of the resource
func (IPAddress) ResourceType() string {
	return "PublicIpAddress"
}

// ListRequest builds the ListAdresses request
func (ipaddress IPAddress) ListRequest() (ListCommand, error) {
	req := &ListPublicIPAddresses{
		AssociatedNetworkID: ipaddress.AssociatedNetworkID,
		ID:                  ipaddress.ID,
		IPAddress:           ipaddress.IPAddress,
		PhysicalNetworkID:   ipaddress.PhysicalNetworkID,
		VlanID:              ipaddress.VlanID,
		ZoneID:              ipaddress.ZoneID,
	}
	if ipaddress.IsElastic {
		req.IsElastic = &ipaddress.IsElastic
	}
	if ipaddress.IsSourceNat {
		req.IsSourceNat = &ipaddress.IsSourceNat
	}
	if ipaddress.ForVirtualNetwork {
		req.ForVirtualNetwork = &ipaddress.ForVirtualNetwork
	}

	return req, nil
}

// Delete removes the resource
func (ipaddress IPAddress) Delete(ctx context.Context, client *Client) error {
	if ipaddress.ID == nil {
		return fmt.Errorf("an IPAddress may only be deleted using ID")
	}

	return client.BooleanRequestWithContext(ctx, &DisassociateIPAddress{
		ID: ipaddress.ID,
	})
}

// AssociateIPAddress (Async) represents the IP creation
type AssociateIPAddress struct {
	IsPortable *bool `json:"isportable,omitempty" doc:"should be set to true if public IP is required to be transferable across zones, if not specified defaults to false"`
	NetworkdID *UUID `json:"networkid,omitempty" doc:"The network this ip address should be associated to."`
	ZoneID     *UUID `json:"zoneid,omitempty" doc:"the ID of the availability zone you want to acquire an public IP address from"`
	_          bool  `name:"associateIpAddress" description:"Acquires and associates a public IP to an account."`
}

// Response returns the struct to unmarshal
func (AssociateIPAddress) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (AssociateIPAddress) AsyncResponse() interface{} {
	return new(IPAddress)
}

// DisassociateIPAddress (Async) represents the IP deletion
type DisassociateIPAddress struct {
	ID *UUID `json:"id" doc:"the id of the public ip address to disassociate"`
	_  bool  `name:"disassociateIpAddress" description:"Disassociates an ip address from the account."`
}

// Response returns the struct to unmarshal
func (DisassociateIPAddress) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (DisassociateIPAddress) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

// UpdateIPAddress (Async) represents the IP modification
type UpdateIPAddress struct {
	ID       *UUID `json:"id" doc:"the id of the public ip address to update"`
	CustomID *UUID `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	_        bool  `name:"updateIpAddress" description:"Updates an ip address"`
}

// Response returns the struct to unmarshal
func (UpdateIPAddress) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (UpdateIPAddress) AsyncResponse() interface{} {
	return new(IPAddress)
}

//go:generate go run generate/main.go -interface=Listable ListPublicIPAddresses

// ListPublicIPAddresses represents a search for public IP addresses
type ListPublicIPAddresses struct {
	AllocatedOnly       *bool         `json:"allocatedonly,omitempty" doc:"limits search results to allocated public IP addresses"`
	AssociatedNetworkID *UUID         `json:"associatednetworkid,omitempty" doc:"lists all public IP addresses associated to the network specified"`
	ForLoadBalancing    *bool         `json:"forloadbalancing,omitempty" doc:"list only ips used for load balancing"`
	ForVirtualNetwork   *bool         `json:"forvirtualnetwork,omitempty" doc:"the virtual network for the IP address"`
	ID                  *UUID         `json:"id,omitempty" doc:"lists ip address by id"`
	IPAddress           net.IP        `json:"ipaddress,omitempty" doc:"lists the specified IP address"`
	IsElastic           *bool         `json:"iselastic,omitempty" doc:"list only elastic ip addresses"`
	IsSourceNat         *bool         `json:"issourcenat,omitempty" doc:"list only source nat ip addresses"`
	IsStaticNat         *bool         `json:"isstaticnat,omitempty" doc:"list only static nat ip addresses"`
	Keyword             string        `json:"keyword,omitempty" doc:"List by keyword"`
	Page                int           `json:"page,omitempty"`
	PageSize            int           `json:"pagesize,omitempty"`
	PhysicalNetworkID   *UUID         `json:"physicalnetworkid,omitempty" doc:"lists all public IP addresses by physical network id"`
	Tags                []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	VlanID              *UUID         `json:"vlanid,omitempty" doc:"lists all public IP addresses by VLAN ID"`
	ZoneID              *UUID         `json:"zoneid,omitempty" doc:"lists all public IP addresses by Zone ID"`
	_                   bool          `name:"listPublicIpAddresses" description:"Lists all public ip addresses"`
}

// ListPublicIPAddressesResponse represents a list of public IP addresses
type ListPublicIPAddressesResponse struct {
	Count           int         `json:"count"`
	PublicIPAddress []IPAddress `json:"publicipaddress"`
}
