package egoscale

import (
	"fmt"
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
	DisplayNetwork              bool          `json:"displaynetwork,omitempty" doc:"an optional field, whether to the display the network to the end user or not."`
	DisplayText                 string        `json:"displaytext,omitempty" doc:"the displaytext of the network"`
	DNS1                        net.IP        `json:"dns1,omitempty" doc:"the first DNS for the network"`
	DNS2                        net.IP        `json:"dns2,omitempty" doc:"the second DNS for the network"`
	Domain                      string        `json:"domain,omitempty" doc:"the domain name of the network owner"`
	DomainID                    *UUID         `json:"domainid,omitempty" doc:"the domain id of the network owner"`
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
		Account:           network.Account,
		DomainID:          network.DomainID,
		ID:                network.ID,
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
	Account           string `json:"account,omitempty" doc:"account who will own the network"`
	DisplayNetwork    *bool  `json:"displaynetwork,omitempty" doc:"an optional field, whether to the display the network to the end user or not."`
	DisplayText       string `json:"displaytext,omitempty" doc:"the display text of the network"` // This field is required but might be empty
	DomainID          *UUID  `json:"domainid,omitempty" doc:"domain ID of the account owning a network"`
	EndIP             net.IP `json:"endip,omitempty" doc:"the ending IP address in the network IP range. If not specified, will be defaulted to startIP"`
	EndIpv6           net.IP `json:"endipv6,omitempty" doc:"the ending IPv6 address in the IPv6 network range"`
	Gateway           net.IP `json:"gateway,omitempty" doc:"the gateway of the network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	IP6CIDR           *CIDR  `json:"ip6cidr,omitempty" doc:"the CIDR of IPv6 network, must be at least /64"`
	IP6Gateway        net.IP `json:"ip6gateway,omitempty" doc:"the gateway of the IPv6 network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	IsolatedPVlan     string `json:"isolatedpvlan,omitempty" doc:"the isolated private vlan for this network"`
	Name              string `json:"name,omitempty" doc:"the name of the network"` // This field is required but might be empty
	Netmask           net.IP `json:"netmask,omitempty" doc:"the netmask of the network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	NetworkDomain     string `json:"networkdomain,omitempty" doc:"network domain"`
	NetworkOfferingID *UUID  `json:"networkofferingid" doc:"the network offering id"`
	PhysicalNetworkID *UUID  `json:"physicalnetworkid,omitempty" doc:"the Physical Network ID the network belongs to"`
	StartIP           net.IP `json:"startip,omitempty" doc:"the beginning IP address in the network IP range"`
	StartIpv6         net.IP `json:"startipv6,omitempty" doc:"the beginning IPv6 address in the IPv6 network range"`
	SubdomainAccess   *bool  `json:"subdomainaccess,omitempty" doc:"Defines whether to allow subdomains to use networks dedicated to their parent domain(s). Should be used with aclType=Domain, defaulted to allow.subdomain.network.access global config if not specified"`
	Vlan              string `json:"vlan,omitempty" doc:"the ID or VID of the network"`
	ZoneID            *UUID  `json:"zoneid" doc:"the Zone ID for the network"`
	_                 bool   `name:"createNetwork" description:"Creates a network"`
}

func (CreateNetwork) response() interface{} {
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
	ID                *UUID  `json:"id" doc:"the ID of the network"`
	ChangeCIDR        *bool  `json:"changecidr,omitempty" doc:"Force update even if cidr type is different"`
	CustomID          *UUID  `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	DisplayNetwork    *bool  `json:"displaynetwork,omitempty" doc:"an optional field, whether to the display the network to the end user or not."`
	DisplayText       string `json:"displaytext,omitempty" doc:"the new display text for the network"`
	GuestVMCIDR       *CIDR  `json:"guestvmcidr,omitempty" doc:"CIDR for Guest VMs,Cloudstack allocates IPs to Guest VMs only from this CIDR"`
	Name              string `json:"name,omitempty" doc:"the new name for the network"`
	NetworkDomain     string `json:"networkdomain,omitempty" doc:"network domain"`
	NetworkOfferingID *UUID  `json:"networkofferingid,omitempty" doc:"network offering ID"`
	_                 bool   `name:"updateNetwork" description:"Updates a network"`
}

func (UpdateNetwork) response() interface{} {
	return new(AsyncJobResult)
}

func (UpdateNetwork) asyncResponse() interface{} {
	return new(Network)
}

// RestartNetwork (Async) updates a network
type RestartNetwork struct {
	ID      *UUID `json:"id" doc:"The id of the network to restart."`
	Cleanup *bool `json:"cleanup,omitempty" doc:"If cleanup old network elements"`
	_       bool  `name:"restartNetwork" description:"Restarts the network; includes 1) restarting network elements - virtual routers, dhcp servers 2) reapplying all public ips 3) reapplying loadBalancing/portForwarding rules"`
}

func (RestartNetwork) response() interface{} {
	return new(AsyncJobResult)
}

func (RestartNetwork) asyncResponse() interface{} {
	return new(Network)
}

// DeleteNetwork deletes a network
type DeleteNetwork struct {
	ID     *UUID `json:"id" doc:"the ID of the network"`
	Forced *bool `json:"forced,omitempty" doc:"Force delete a network. Network will be marked as 'Destroy' even when commands to shutdown and cleanup to the backend fails."`
	_      bool  `name:"deleteNetwork" description:"Deletes a network"`
}

func (DeleteNetwork) response() interface{} {
	return new(AsyncJobResult)
}

func (DeleteNetwork) asyncResponse() interface{} {
	return new(booleanResponse)
}

// ListNetworks represents a query to a network
type ListNetworks struct {
	Account           string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	CanUseForDeploy   *bool         `json:"canusefordeploy,omitempty" doc:"list networks available for vm deployment"`
	DisplayNetwork    *bool         `json:"displaynetwork,omitempty" doc:"list resources by display flag; only ROOT admin is eligible to pass this parameter"`
	DomainID          *UUID         `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID                *UUID         `json:"id,omitempty" doc:"list networks by id"`
	IsRecursive       *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	IsSystem          *bool         `json:"issystem,omitempty" doc:"true if network is system, false otherwise"`
	Keyword           string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll           *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	PhysicalNetworkID *UUID         `json:"physicalnetworkid,omitempty" doc:"list networks by physical network id"`
	RestartRequired   *bool         `json:"restartrequired,omitempty" doc:"list networks by restartRequired"`
	SpecifyIPRanges   *bool         `json:"specifyipranges,omitempty" doc:"true if need to list only networks which support specifying ip ranges"`
	SupportedServices []Service     `json:"supportedservices,omitempty" doc:"list networks supporting certain services"`
	Tags              []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	TrafficType       string        `json:"traffictype,omitempty" doc:"type of the traffic"`
	Type              string        `json:"type,omitempty" doc:"the type of the network. Supported values are: Isolated and Shared"`
	ZoneID            *UUID         `json:"zoneid,omitempty" doc:"the Zone ID of the network"`
	_                 bool          `name:"listNetworks" description:"Lists all available networks."`
}

// ListNetworksResponse represents the list of networks
type ListNetworksResponse struct {
	Count   int       `json:"count"`
	Network []Network `json:"network"`
}

func (ListNetworks) response() interface{} {
	return new(ListNetworksResponse)
}

// SetPage sets the current page
func (listNetwork *ListNetworks) SetPage(page int) {
	listNetwork.Page = page
}

// SetPageSize sets the page size
func (listNetwork *ListNetworks) SetPageSize(pageSize int) {
	listNetwork.PageSize = pageSize
}

func (ListNetworks) each(resp interface{}, callback IterateItemFunc) {
	networks, ok := resp.(*ListNetworksResponse)
	if !ok {
		callback(nil, fmt.Errorf("type error: ListNetworksResponse expected, got %T", resp))
		return
	}

	for i := range networks.Network {
		if !callback(&networks.Network[i], nil) {
			break
		}
	}
}
