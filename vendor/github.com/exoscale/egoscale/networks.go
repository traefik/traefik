package egoscale

import (
	"net"
	"net/url"
)

// Network represents a network
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html
type Network struct {
	Account                     string        `json:"account,omitempty" doc:"the owner of the network"`
	AccountID                   *UUID         `json:"accountid,omitempty" doc:"the owner ID of the network"`
	BroadcastDomainType         string        `json:"broadcastdomaintype,omitempty" doc:"Broadcast domain type of the network"`
	BroadcastURI                string        `json:"broadcasturi,omitempty" doc:"broadcast uri of the network."`
	CanUseForDeploy             bool          `json:"canusefordeploy,omitempty" doc:"list networks available for vm deployment"`
	CIDR                        *CIDR         `json:"cidr,omitempty" doc:"Cloudstack managed address space, all CloudStack managed VMs get IP address from CIDR"`
	DisplayText                 string        `json:"displaytext,omitempty" doc:"the displaytext of the network"`
	DNS1                        net.IP        `json:"dns1,omitempty" doc:"the first DNS for the network"`
	DNS2                        net.IP        `json:"dns2,omitempty" doc:"the second DNS for the network"`
	EndIP                       net.IP        `json:"endip,omitempty" doc:"the ending IP address in the network IP range. Required for managed networks."`
	Gateway                     net.IP        `json:"gateway,omitempty" doc:"the network's gateway"`
	ID                          *UUID         `json:"id,omitempty" doc:"the id of the network"`
	IP6CIDR                     *CIDR         `json:"ip6cidr,omitempty" doc:"the cidr of IPv6 network"`
	IP6Gateway                  net.IP        `json:"ip6gateway,omitempty" doc:"the gateway of IPv6 network"`
	IsDefault                   bool          `json:"isdefault,omitempty" doc:"true if network is default, false otherwise"`
	IsPersistent                bool          `json:"ispersistent,omitempty" doc:"list networks that are persistent"`
	IsSystem                    bool          `json:"issystem,omitempty" doc:"true if network is system, false otherwise"`
	Name                        string        `json:"name,omitempty" doc:"the name of the network"`
	Netmask                     net.IP        `json:"netmask,omitempty" doc:"the network's netmask"`
	NetworkCIDR                 *CIDR         `json:"networkcidr,omitempty" doc:"the network CIDR of the guest network configured with IP reservation. It is the summation of CIDR and RESERVED_IP_RANGE"`
	NetworkDomain               string        `json:"networkdomain,omitempty" doc:"the network domain"`
	NetworkOfferingAvailability string        `json:"networkofferingavailability,omitempty" doc:"availability of the network offering the network is created from"`
	NetworkOfferingConserveMode bool          `json:"networkofferingconservemode,omitempty" doc:"true if network offering is ip conserve mode enabled"`
	NetworkOfferingDisplayText  string        `json:"networkofferingdisplaytext,omitempty" doc:"display text of the network offering the network is created from"`
	NetworkOfferingID           *UUID         `json:"networkofferingid,omitempty" doc:"network offering id the network is created from"`
	NetworkOfferingName         string        `json:"networkofferingname,omitempty" doc:"name of the network offering the network is created from"`
	PhysicalNetworkID           *UUID         `json:"physicalnetworkid,omitempty" doc:"the physical network id"`
	Related                     string        `json:"related,omitempty" doc:"related to what other network configuration"`
	ReservedIPRange             string        `json:"reservediprange,omitempty" doc:"the network's IP range not to be used by CloudStack guest VMs and can be used for non CloudStack purposes"`
	RestartRequired             bool          `json:"restartrequired,omitempty" doc:"true network requires restart"`
	Service                     []Service     `json:"service,omitempty" doc:"the list of services"`
	SpecifyIPRanges             bool          `json:"specifyipranges,omitempty" doc:"true if network supports specifying ip ranges, false otherwise"`
	StartIP                     net.IP        `json:"startip,omitempty" doc:"the beginning IP address in the network IP range. Required for managed networks."`
	State                       string        `json:"state,omitempty" doc:"state of the network"`
	StrechedL2Subnet            bool          `json:"strechedl2subnet,omitempty" doc:"true if network can span multiple zones"`
	SubdomainAccess             bool          `json:"subdomainaccess,omitempty" doc:"true if users from subdomains can access the domain level network"`
	Tags                        []ResourceTag `json:"tags,omitempty" doc:"the list of resource tags associated with network"`
	TrafficType                 string        `json:"traffictype,omitempty" doc:"the traffic type of the network"`
	Type                        string        `json:"type,omitempty" doc:"the type of the network"`
	Vlan                        string        `json:"vlan,omitemtpy" doc:"The vlan of the network. This parameter is visible to ROOT admins only"`
	ZoneID                      *UUID         `json:"zoneid,omitempty" doc:"zone id of the network"`
	ZoneName                    string        `json:"zonename,omitempty" doc:"the name of the zone the network belongs to"`
	ZonesNetworkSpans           []Zone        `json:"zonesnetworkspans,omitempty" doc:"If a network is enabled for 'streched l2 subnet' then represents zones on which network currently spans"`
}

// ListRequest builds the ListNetworks request
func (network Network) ListRequest() (ListCommand, error) {
	//TODO add tags support
	req := &ListNetworks{
		ID:                network.ID,
		Keyword:           network.Name, // this is a hack as listNetworks doesn't support to search by name.
		PhysicalNetworkID: network.PhysicalNetworkID,
		TrafficType:       network.TrafficType,
		Type:              network.Type,
		ZoneID:            network.ZoneID,
	}

	if network.CanUseForDeploy {
		req.CanUseForDeploy = &network.CanUseForDeploy
	}
	if network.RestartRequired {
		req.RestartRequired = &network.RestartRequired
	}

	return req, nil
}

// ResourceType returns the type of the resource
func (Network) ResourceType() string {
	return "Network"
}

// Service is a feature of a network
type Service struct {
	Capability []ServiceCapability `json:"capability,omitempty"`
	Name       string              `json:"name"`
	Provider   []ServiceProvider   `json:"provider,omitempty"`
}

// ServiceCapability represents optional capability of a service
type ServiceCapability struct {
	CanChooseServiceCapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

// ServiceProvider represents the provider of the service
type ServiceProvider struct {
	CanEnableIndividualService   bool     `json:"canenableindividualservice"`
	DestinationPhysicalNetworkID *UUID    `json:"destinationphysicalnetworkid"`
	ID                           *UUID    `json:"id"`
	Name                         string   `json:"name"`
	PhysicalNetworkID            *UUID    `json:"physicalnetworkid"`
	ServiceList                  []string `json:"servicelist,omitempty"`
}

// CreateNetwork creates a network
type CreateNetwork struct {
	DisplayText       string `json:"displaytext,omitempty" doc:"the display text of the network"` // This field is required but might be empty
	EndIP             net.IP `json:"endip,omitempty" doc:"the ending IP address in the network IP range. Required for managed networks."`
	EndIpv6           net.IP `json:"endipv6,omitempty" doc:"the ending IPv6 address in the IPv6 network range"`
	Gateway           net.IP `json:"gateway,omitempty" doc:"the gateway of the network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	IP6CIDR           *CIDR  `json:"ip6cidr,omitempty" doc:"the CIDR of IPv6 network, must be at least /64"`
	IP6Gateway        net.IP `json:"ip6gateway,omitempty" doc:"the gateway of the IPv6 network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	IsolatedPVlan     string `json:"isolatedpvlan,omitempty" doc:"the isolated private vlan for this network"`
	Name              string `json:"name,omitempty" doc:"the name of the network"` // This field is required but might be empty
	Netmask           net.IP `json:"netmask,omitempty" doc:"the netmask of the network. Required for managed networks."`
	NetworkDomain     string `json:"networkdomain,omitempty" doc:"network domain"`
	NetworkOfferingID *UUID  `json:"networkofferingid" doc:"the network offering id"`
	PhysicalNetworkID *UUID  `json:"physicalnetworkid,omitempty" doc:"the Physical Network ID the network belongs to"`
	StartIP           net.IP `json:"startip,omitempty" doc:"the beginning IP address in the network IP range. Required for managed networks."`
	StartIpv6         net.IP `json:"startipv6,omitempty" doc:"the beginning IPv6 address in the IPv6 network range"`
	Vlan              string `json:"vlan,omitempty" doc:"the ID or VID of the network"`
	ZoneID            *UUID  `json:"zoneid" doc:"the Zone ID for the network"`
	_                 bool   `name:"createNetwork" description:"Creates a network"`
}

// Response returns the struct to unmarshal
func (CreateNetwork) Response() interface{} {
	return new(Network)
}

func (req CreateNetwork) onBeforeSend(params url.Values) error {
	// Those fields are required but might be empty
	if req.Name == "" {
		params.Set("name", "")
	}
	if req.DisplayText == "" {
		params.Set("displaytext", "")
	}
	return nil
}

// UpdateNetwork (Async) updates a network
type UpdateNetwork struct {
	_                 bool   `name:"updateNetwork" description:"Updates a network"`
	ChangeCIDR        *bool  `json:"changecidr,omitempty" doc:"Force update even if cidr type is different"`
	DisplayText       string `json:"displaytext,omitempty" doc:"the new display text for the network"`
	EndIP             net.IP `json:"endip,omitempty" doc:"the ending IP address in the network IP range. Required for managed networks."`
	GuestVMCIDR       *CIDR  `json:"guestvmcidr,omitempty" doc:"CIDR for Guest VMs,Cloudstack allocates IPs to Guest VMs only from this CIDR"`
	ID                *UUID  `json:"id" doc:"the ID of the network"`
	Name              string `json:"name,omitempty" doc:"the new name for the network"`
	Netmask           net.IP `json:"netmask,omitempty" doc:"the netmask of the network. Required for managed networks."`
	NetworkDomain     string `json:"networkdomain,omitempty" doc:"network domain"`
	NetworkOfferingID *UUID  `json:"networkofferingid,omitempty" doc:"network offering ID"`
	StartIP           net.IP `json:"startip,omitempty" doc:"the beginning IP address in the network IP range. Required for managed networks."`
}

// Response returns the struct to unmarshal
func (UpdateNetwork) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (UpdateNetwork) AsyncResponse() interface{} {
	return new(Network)
}

// RestartNetwork (Async) updates a network
type RestartNetwork struct {
	ID      *UUID `json:"id" doc:"The id of the network to restart."`
	Cleanup *bool `json:"cleanup,omitempty" doc:"If cleanup old network elements"`
	_       bool  `name:"restartNetwork" description:"Restarts the network; includes 1) restarting network elements - virtual routers, dhcp servers 2) reapplying all public ips 3) reapplying loadBalancing/portForwarding rules"`
}

// Response returns the struct to unmarshal
func (RestartNetwork) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (RestartNetwork) AsyncResponse() interface{} {
	return new(Network)
}

// DeleteNetwork deletes a network
type DeleteNetwork struct {
	ID     *UUID `json:"id" doc:"the ID of the network"`
	Forced *bool `json:"forced,omitempty" doc:"Force delete a network. Network will be marked as 'Destroy' even when commands to shutdown and cleanup to the backend fails."`
	_      bool  `name:"deleteNetwork" description:"Deletes a network"`
}

// Response returns the struct to unmarshal
func (DeleteNetwork) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (DeleteNetwork) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

//go:generate go run generate/main.go -interface=Listable ListNetworks

// ListNetworks represents a query to a network
type ListNetworks struct {
	CanUseForDeploy   *bool         `json:"canusefordeploy,omitempty" doc:"List networks available for vm deployment"`
	ID                *UUID         `json:"id,omitempty" doc:"List networks by id"`
	IsSystem          *bool         `json:"issystem,omitempty" doc:"true If network is system, false otherwise"`
	Keyword           string        `json:"keyword,omitempty" doc:"List by keyword"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	PhysicalNetworkID *UUID         `json:"physicalnetworkid,omitempty" doc:"List networks by physical network id"`
	RestartRequired   *bool         `json:"restartrequired,omitempty" doc:"List networks by restartRequired"`
	SpecifyIPRanges   *bool         `json:"specifyipranges,omitempty" doc:"True if need to list only networks which support specifying ip ranges"`
	SupportedServices []Service     `json:"supportedservices,omitempty" doc:"List networks supporting certain services"`
	Tags              []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	TrafficType       string        `json:"traffictype,omitempty" doc:"Type of the traffic"`
	Type              string        `json:"type,omitempty" doc:"The type of the network. Supported values are: Isolated and Shared"`
	ZoneID            *UUID         `json:"zoneid,omitempty" doc:"The Zone ID of the network"`
	_                 bool          `name:"listNetworks" description:"Lists all available networks."`
}

// ListNetworksResponse represents the list of networks
type ListNetworksResponse struct {
	Count   int       `json:"count"`
	Network []Network `json:"network"`
}
