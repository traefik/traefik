package egoscale

// NetworkOffering corresponds to the Compute Offerings
type NetworkOffering struct {
	Availability             string            `json:"availability,omitempty" doc:"availability of the network offering"`
	ConserveMode             bool              `json:"conservemode,omitempty" doc:"true if network offering is ip conserve mode enabled"`
	Created                  string            `json:"created,omitempty" doc:"the date this network offering was created"`
	Details                  map[string]string `json:"details,omitempty" doc:"additional key/value details tied with network offering"`
	DisplayText              string            `json:"displaytext,omitempty" doc:"an alternate display text of the network offering."`
	EgressDefaultPolicy      bool              `json:"egressdefaultpolicy,omitempty" doc:"true if guest network default egress policy is allow; false if default egress policy is deny"`
	GuestIPType              string            `json:"guestiptype,omitempty" doc:"guest type of the network offering, can be Shared or Isolated"`
	ID                       *UUID             `json:"id,omitempty" doc:"the id of the network offering"`
	IsDefault                bool              `json:"isdefault,omitempty" doc:"true if network offering is default, false otherwise"`
	IsPersistent             bool              `json:"ispersistent,omitempty" doc:"true if network offering supports persistent networks, false otherwise"`
	MaxConnections           int               `json:"maxconnections,omitempty" doc:"maximum number of concurrents connections to be handled by lb"`
	Name                     string            `json:"name,omitempty" doc:"the name of the network offering"`
	NetworkRate              int               `json:"networkrate,omitempty" doc:"data transfer rate in megabits per second allowed."`
	Service                  []Service         `json:"service,omitempty" doc:"the list of supported services"`
	ServiceOfferingID        *UUID             `json:"serviceofferingid,omitempty" doc:"the ID of the service offering used by virtual router provider"`
	SpecifyIPRanges          bool              `json:"specifyipranges,omitempty" doc:"true if network offering supports specifying ip ranges, false otherwise"`
	SpecifyVlan              bool              `json:"specifyvlan,omitempty" doc:"true if network offering supports vlans, false otherwise"`
	State                    string            `json:"state,omitempty" doc:"state of the network offering. Can be Disabled/Enabled/Inactive"`
	SupportsStrechedL2Subnet bool              `json:"supportsstrechedl2subnet,omitempty" doc:"true if network offering supports network that span multiple zones"`
	Tags                     string            `json:"tags,omitempty" doc:"the tags for the network offering"`
	TrafficType              string            `json:"traffictype,omitempty" doc:"the traffic type for the network offering, supported types are Public, Management, Control, Guest, Vlan or Storage."`
}

// ListRequest builds the ListNetworkOfferings request
//
// This doesn't take into account the IsDefault flag as the default value is true.
func (no NetworkOffering) ListRequest() (ListCommand, error) {
	req := &ListNetworkOfferings{
		Availability: no.Availability,
		ID:           no.ID,
		Name:         no.Name,
		State:        no.State,
		TrafficType:  no.TrafficType,
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListNetworkOfferings

// ListNetworkOfferings represents a query for network offerings
type ListNetworkOfferings struct {
	Availability       string    `json:"availability,omitempty" doc:"the availability of network offering. Default value is Required"`
	DisplayText        string    `json:"displaytext,omitempty" doc:"list network offerings by display text"`
	GuestIPType        string    `json:"guestiptype,omitempty" doc:"list network offerings by guest type: Shared or Isolated"`
	ID                 *UUID     `json:"id,omitempty" doc:"list network offerings by id"`
	IsDefault          *bool     `json:"isdefault,omitempty" doc:"true if need to list only default network offerings. Default value is false"`
	IsTagged           *bool     `json:"istagged,omitempty" doc:"true if offering has tags specified"`
	Keyword            string    `json:"keyword,omitempty" doc:"List by keyword"`
	Name               string    `json:"name,omitempty" doc:"list network offerings by name"`
	NetworkID          *UUID     `json:"networkid,omitempty" doc:"the ID of the network. Pass this in if you want to see the available network offering that a network can be changed to."`
	Page               int       `json:"page,omitempty"`
	PageSize           int       `json:"pagesize,omitempty"`
	SourceNATSupported *bool     `json:"sourcenatsupported,omitempty" doc:"true if need to list only netwok offerings where source nat is supported, false otherwise"`
	SpecifyIPRanges    *bool     `json:"specifyipranges,omitempty" doc:"true if need to list only network offerings which support specifying ip ranges"`
	SpecifyVlan        *bool     `json:"specifyvlan,omitempty" doc:"the tags for the network offering."`
	State              string    `json:"state,omitempty" doc:"list network offerings by state"`
	SupportedServices  []Service `json:"supportedservices,omitempty" doc:"list network offerings supporting certain services"`
	Tags               string    `json:"tags,omitempty" doc:"list network offerings by tags"`
	TrafficType        string    `json:"traffictype,omitempty" doc:"list by traffic type"`
	ZoneID             *UUID     `json:"zoneid,omitempty" doc:"list network offerings available for network creation in specific zone"`
	_                  bool      `name:"listNetworkOfferings" description:"Lists all available network offerings."`
}

// ListNetworkOfferingsResponse represents a list of service offerings
type ListNetworkOfferingsResponse struct {
	Count           int               `json:"count"`
	NetworkOffering []NetworkOffering `json:"networkoffering"`
}

// UpdateNetworkOffering represents a modification of a network offering
type UpdateNetworkOffering struct {
	Availability     string `json:"availability,omitempty" doc:"the availability of network offering. Default value is Required for Guest Virtual network offering; Optional for Guest Direct network offering"`
	DisplayText      string `json:"displaytext,omitempty" doc:"the display text of the network offering"`
	ID               *UUID  `json:"id,omitempty" doc:"the id of the network offering"`
	KeepAliveEnabled *bool  `json:"keepaliveenabled,omitempty" doc:"if true keepalive will be turned on in the loadbalancer. At the time of writing this has only an effect on haproxy; the mode http and httpclose options are unset in the haproxy conf file."`
	MaxConnections   int    `json:"maxconnections,omitempty" doc:"maximum number of concurrent connections supported by the network offering"`
	Name             string `json:"name,omitempty" doc:"the name of the network offering"`
	SortKey          int    `json:"sortkey,omitempty" doc:"sort key of the network offering, integer"`
	State            string `json:"state,omitempty" doc:"update state for the network offering"`
	_                bool   `name:"updateNetworkOffering" description:"Updates a network offering."`
}

// Response returns the struct to unmarshal
func (UpdateNetworkOffering) Response() interface{} {
	return new(NetworkOffering)
}
