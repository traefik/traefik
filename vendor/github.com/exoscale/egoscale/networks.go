package egoscale

import (
	"net"
	"net/url"
)

// Network represents a network
type Network struct {
	ID                          string        `json:"id"`
	Account                     string        `json:"account"`
	ACLID                       string        `json:"aclid,omitempty"`
	ACLType                     string        `json:"acltype,omitempty"`
	BroadcastDomainType         string        `json:"broadcastdomaintype,omitempty"`
	BroadcastURI                string        `json:"broadcasturi,omitempty"`
	CanUseForDeploy             bool          `json:"canusefordeploy,omitempty"`
	Cidr                        string        `json:"cidr,omitempty"`
	DisplayNetwork              bool          `json:"diplaynetwork,omitempty"`
	DisplayText                 string        `json:"displaytext"`
	DNS1                        net.IP        `json:"dns1,omitempty"`
	DNS2                        net.IP        `json:"dns2,omitempty"`
	Domain                      string        `json:"domain,omitempty"`
	DomainID                    string        `json:"domainid,omitempty"`
	Gateway                     net.IP        `json:"gateway,omitempty"`
	IP6Cidr                     string        `json:"ip6cidr,omitempty"`
	IP6Gateway                  net.IP        `json:"ip6gateway,omitempty"`
	IsDefault                   bool          `json:"isdefault,omitempty"`
	IsPersistent                bool          `json:"ispersistent,omitempty"`
	Name                        string        `json:"name"`
	Netmask                     net.IP        `json:"netmask,omitempty"`
	NetworkCidr                 string        `json:"networkcidr,omitempty"`
	NetworkDomain               string        `json:"networkdomain,omitempty"`
	NetworkOfferingAvailability string        `json:"networkofferingavailability,omitempty"`
	NetworkOfferingConserveMode bool          `json:"networkofferingconservemode,omitempty"`
	NetworkOfferingDisplayText  string        `json:"networkofferingdisplaytext,omitempty"`
	NetworkOfferingID           string        `json:"networkofferingid,omitempty"`
	NetworkOfferingName         string        `json:"networkofferingname,omitempty"`
	PhysicalNetworkID           string        `json:"physicalnetworkid,omitempty"`
	Project                     string        `json:"project,omitempty"`
	ProjectID                   string        `json:"projectid,omitempty"`
	Related                     string        `json:"related,omitempty"`
	ReserveIPRange              string        `json:"reserveiprange,omitempty"`
	RestartRequired             bool          `json:"restartrequired,omitempty"`
	SpecifyIPRanges             bool          `json:"specifyipranges,omitempty"`
	State                       string        `json:"state"`
	StrechedL2Subnet            bool          `json:"strechedl2subnet,omitempty"`
	SubdomainAccess             bool          `json:"subdomainaccess,omitempty"`
	TrafficType                 string        `json:"traffictype"`
	Type                        string        `json:"type"`
	Vlan                        string        `json:"vlan,omitemtpy"` // root only
	VpcID                       string        `json:"vpcid,omitempty"`
	ZoneID                      string        `json:"zoneid,omitempty"`
	ZoneName                    string        `json:"zonename,omitempty"`
	ZonesNetworkSpans           string        `json:"zonesnetworkspans,omitempty"`
	Service                     []Service     `json:"service"`
	Tags                        []ResourceTag `json:"tags"`
}

// ResourceType returns the type of the resource
func (*Network) ResourceType() string {
	return "Network"
}

// Service is a feature of a network
type Service struct {
	Name       string              `json:"name"`
	Capability []ServiceCapability `json:"capability,omitempty"`
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
	ID                           string   `json:"id"`
	CanEnableIndividualService   bool     `json:"canenableindividualservice"`
	DestinationPhysicalNetworkID string   `json:"destinationphysicalnetworkid"`
	Name                         string   `json:"name"`
	PhysicalNetworkID            string   `json:"physicalnetworkid"`
	ServiceList                  []string `json:"servicelist,omitempty"`
}

// NetworkResponse represents a network
type NetworkResponse struct {
	Network Network `json:"network"`
}

// CreateNetwork creates a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createNetwork.html
type CreateNetwork struct {
	DisplayText       string `json:"displaytext,omitempty"`
	Name              string `json:"name,omitempty"`
	NetworkOfferingID string `json:"networkofferingid"`
	ZoneID            string `json:"zoneid"`
	Account           string `json:"account,omitempty"`
	ACLID             string `json:"aclid,omitempty"`
	ACLType           string `json:"acltype,omitempty"`        // Account or Domain
	DisplayNetwork    bool   `json:"displaynetwork,omitempty"` // root only
	DomainID          string `json:"domainid,omitempty"`
	EndIP             net.IP `json:"endip,omitempty"`
	EndIpv6           net.IP `json:"endipv6,omitempty"`
	Gateway           net.IP `json:"gateway,omitempty"`
	IP6Cidr           string `json:"ip6cidr,omitempty"`
	IP6Gateway        net.IP `json:"ip6gateway,omitempty"`
	IsolatedPVlan     string `json:"isolatedpvlan,omitempty"`
	Netmask           net.IP `json:"netmask,omitempty"`
	NetworkDomain     string `json:"networkdomain,omitempty"`
	PhysicalNetworkID string `json:"physicalnetworkid,omitempty"`
	ProjectID         string `json:"projectid,omitempty"`
	StartIP           net.IP `json:"startip,omitempty"`
	StartIpv6         net.IP `json:"startipv6,omitempty"`
	SubdomainAccess   string `json:"subdomainaccess,omitempty"`
	Vlan              string `json:"vlan,omitempty"`
	VpcID             string `json:"vpcid,omitempty"`
}

func (*CreateNetwork) name() string {
	return "createNetwork"
}

func (*CreateNetwork) response() interface{} {
	return new(CreateNetworkResponse)
}

func (req *CreateNetwork) onBeforeSend(params *url.Values) error {
	// Those fields are required but might be empty
	if req.Name == "" {
		params.Set("name", "")
	}
	if req.DisplayText == "" {
		params.Set("displaytext", "")
	}
	return nil
}

// CreateNetworkResponse represents a freshly created network
type CreateNetworkResponse NetworkResponse

// UpdateNetwork updates a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateNetwork.html
type UpdateNetwork struct {
	ID                string `json:"id"`
	ChangeCidr        bool   `json:"changecidr,omitempty"`
	CustomID          string `json:"customid,omitempty"` // root only
	DisplayNetwork    string `json:"displaynetwork,omitempty"`
	DisplayText       string `json:"displaytext,omitempty"`
	Forced            bool   `json:"forced,omitempty"`
	GuestVMCidr       string `json:"guestvmcidr,omitempty"`
	Name              string `json:"name,omitempty"`
	NetworkDomain     string `json:"networkdomain,omitempty"`
	NetworkOfferingID string `json:"networkofferingid,omitempty"`
	UpdateInSequence  bool   `json:"updateinsequence,omitempty"`
}

func (*UpdateNetwork) name() string {
	return "updateNetwork"
}

func (*UpdateNetwork) asyncResponse() interface{} {
	return new(UpdateNetworkResponse)
}

// UpdateNetworkResponse represents a freshly created network
type UpdateNetworkResponse NetworkResponse

// RestartNetwork updates a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/restartNetwork.html
type RestartNetwork struct {
	ID      string `json:"id"`
	Cleanup bool   `json:"cleanup,omitempty"`
}

func (*RestartNetwork) name() string {
	return "restartNetwork"
}

func (*RestartNetwork) asyncResponse() interface{} {
	return new(RestartNetworkResponse)
}

// RestartNetworkResponse represents a freshly created network
type RestartNetworkResponse NetworkResponse

// DeleteNetwork deletes a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteNetwork.html
type DeleteNetwork struct {
	ID     string `json:"id"`
	Forced bool   `json:"forced,omitempty"`
}

func (*DeleteNetwork) name() string {
	return "deleteNetwork"
}

func (*DeleteNetwork) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ListNetworks represents a query to a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listNetworks.html
type ListNetworks struct {
	Account           string        `json:"account,omitempty"`
	ACLType           string        `json:"acltype,omitempty"` // Account or Domain
	CanUseForDeploy   bool          `json:"canusefordeploy,omitempty"`
	DisplayNetwork    bool          `json:"displaynetwork,omitempty"` // root only
	DomainID          string        `json:"domainid,omitempty"`
	ForVpc            string        `json:"forvpc,omitempty"`
	ID                string        `json:"id,omitempty"`
	IsRecursive       bool          `json:"isrecursive,omitempty"`
	IsSystem          bool          `json:"issystem,omitempty"`
	Keyword           string        `json:"keyword,omitempty"`
	ListAll           bool          `json:"listall,omitempty"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	PhysicalNetworkID string        `json:"physicalnetworkid,omitempty"`
	ProjectID         string        `json:"projectid,omitempty"`
	RestartRequired   bool          `json:"restartrequired,omitempty"`
	SpecifyRanges     bool          `json:"specifyranges,omitempty"`
	SupportedServices []Service     `json:"supportedservices,omitempty"`
	Tags              []ResourceTag `json:"resourcetag,omitempty"`
	TrafficType       string        `json:"traffictype,omitempty"`
	Type              string        `json:"type,omitempty"`
	VpcID             string        `json:"vpcid,omitempty"`
	ZoneID            string        `json:"zoneid,omitempty"`
}

func (*ListNetworks) name() string {
	return "listNetworks"
}

func (*ListNetworks) response() interface{} {
	return new(ListNetworksResponse)
}

// ListNetworksResponse represents the list of networks
type ListNetworksResponse struct {
	Count   int       `json:"count"`
	Network []Network `json:"network"`
}
