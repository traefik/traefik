package egoscale

import (
	"net"
)

// Nic represents a Network Interface Controller (NIC)
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html#configuring-multiple-ip-addresses-on-a-single-nic
type Nic struct {
	BroadcastURI     string           `json:"broadcasturi,omitempty" doc:"the broadcast uri of the nic"`
	DeviceID         *UUID            `json:"deviceid,omitempty" doc:"device id for the network when plugged into the virtual machine"`
	Gateway          net.IP           `json:"gateway,omitempty" doc:"the gateway of the nic"`
	ID               *UUID            `json:"id,omitempty" doc:"the ID of the nic"`
	IP6Address       net.IP           `json:"ip6address,omitempty" doc:"the IPv6 address of network"`
	IP6CIDR          *CIDR            `json:"ip6cidr,omitempty" doc:"the cidr of IPv6 network"`
	IP6Gateway       net.IP           `json:"ip6gateway,omitempty" doc:"the gateway of IPv6 network"`
	IPAddress        net.IP           `json:"ipaddress,omitempty" doc:"the ip address of the nic"`
	IsDefault        bool             `json:"isdefault,omitempty" doc:"true if nic is default, false otherwise"`
	IsolationURI     string           `json:"isolationuri,omitempty" doc:"the isolation uri of the nic"`
	MACAddress       MACAddress       `json:"macaddress,omitempty" doc:"true if nic is default, false otherwise"`
	Netmask          net.IP           `json:"netmask,omitempty" doc:"the netmask of the nic"`
	NetworkID        *UUID            `json:"networkid,omitempty" doc:"the ID of the corresponding network"`
	NetworkName      string           `json:"networkname,omitempty" doc:"the name of the corresponding network"`
	ReverseDNS       []ReverseDNS     `json:"reversedns,omitempty" doc:"the list of PTR record(s) associated with the virtual machine"`
	SecondaryIP      []NicSecondaryIP `json:"secondaryip,omitempty" doc:"the Secondary ipv4 addr of nic"`
	TrafficType      string           `json:"traffictype,omitempty" doc:"the traffic type of the nic"`
	Type             string           `json:"type,omitempty" doc:"the type of the nic"`
	VirtualMachineID *UUID            `json:"virtualmachineid,omitempty" doc:"Id of the vm to which the nic belongs"`
}

// ListRequest build a ListNics request from the given Nic
func (nic Nic) ListRequest() (ListCommand, error) {
	req := &ListNics{
		VirtualMachineID: nic.VirtualMachineID,
		NicID:            nic.ID,
		NetworkID:        nic.NetworkID,
	}

	return req, nil
}

// NicSecondaryIP represents a link between NicID and IPAddress
type NicSecondaryIP struct {
	ID               *UUID  `json:"id,omitempty" doc:"the ID of the secondary private IP addr"`
	IPAddress        net.IP `json:"ipaddress,omitempty" doc:"Secondary IP address"`
	NetworkID        *UUID  `json:"networkid,omitempty" doc:"the ID of the network"`
	NicID            *UUID  `json:"nicid,omitempty" doc:"the ID of the nic"`
	VirtualMachineID *UUID  `json:"virtualmachineid,omitempty" doc:"the ID of the vm"`
}

//go:generate go run generate/main.go -interface=Listable ListNics

// ListNics represents the NIC search
type ListNics struct {
	Keyword          string `json:"keyword,omitempty" doc:"List by keyword"`
	NetworkID        *UUID  `json:"networkid,omitempty" doc:"list nic of the specific vm's network"`
	NicID            *UUID  `json:"nicid,omitempty" doc:"the ID of the nic to to list IPs"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	VirtualMachineID *UUID  `json:"virtualmachineid,omitempty" doc:"the ID of the vm"`
	_                bool   `name:"listNics" description:"list the vm nics  IP to NIC"`
}

// ListNicsResponse represents a list of templates
type ListNicsResponse struct {
	Count int   `json:"count"`
	Nic   []Nic `json:"nic"`
}

// AddIPToNic (Async) represents the assignation of a secondary IP
type AddIPToNic struct {
	NicID     *UUID  `json:"nicid" doc:"the ID of the nic to which you want to assign private IP"`
	IPAddress net.IP `json:"ipaddress,omitempty" doc:"Secondary IP Address"`
	_         bool   `name:"addIpToNic" description:"Assigns secondary IP to NIC"`
}

// Response returns the struct to unmarshal
func (AddIPToNic) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (AddIPToNic) AsyncResponse() interface{} {
	return new(NicSecondaryIP)
}

// RemoveIPFromNic (Async) represents a deletion request
type RemoveIPFromNic struct {
	ID *UUID `json:"id" doc:"the ID of the secondary ip address to nic"`
	_  bool  `name:"removeIpFromNic" description:"Removes secondary IP from the NIC."`
}

// Response returns the struct to unmarshal
func (RemoveIPFromNic) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (RemoveIPFromNic) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

// ActivateIP6 (Async) activates the IP6 on the given NIC
//
// Exoscale specific API: https://community.exoscale.ch/api/compute/#activateip6_GET
type ActivateIP6 struct {
	NicID *UUID `json:"nicid" doc:"the ID of the nic to which you want to assign the IPv6"`
	_     bool  `name:"activateIp6" description:"Activate the IPv6 on the VM's nic"`
}

// Response returns the struct to unmarshal
func (ActivateIP6) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (ActivateIP6) AsyncResponse() interface{} {
	return new(Nic)
}
