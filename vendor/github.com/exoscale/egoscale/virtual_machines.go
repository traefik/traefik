package egoscale

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
)

// VirtualMachineState holds the state of the instance
//
// https://github.com/apache/cloudstack/blob/master/api/src/main/java/com/cloud/vm/VirtualMachine.java
type VirtualMachineState string

const (
	// VirtualMachineStarting VM is being started. At this state, you should find host id filled which means it's being started on that host
	VirtualMachineStarting VirtualMachineState = "Starting"
	// VirtualMachineRunning VM is running. host id has the host that it is running on
	VirtualMachineRunning VirtualMachineState = "Running"
	// VirtualMachineStopping VM is being stopped. host id has the host that it is being stopped on
	VirtualMachineStopping VirtualMachineState = "Stopping"
	// VirtualMachineStopped VM is stopped. host id should be null
	VirtualMachineStopped VirtualMachineState = "Stopped"
	// VirtualMachineDestroyed VM is marked for destroy
	VirtualMachineDestroyed VirtualMachineState = "Destroyed"
	// VirtualMachineExpunging "VM is being expunged
	VirtualMachineExpunging VirtualMachineState = "Expunging"
	// VirtualMachineMigrating VM is being migrated. host id holds to from host
	VirtualMachineMigrating VirtualMachineState = "Migrating"
	// VirtualMachineError VM is in error
	VirtualMachineError VirtualMachineState = "Error"
	// VirtualMachineUnknown VM state is unknown
	VirtualMachineUnknown VirtualMachineState = "Unknown"
	// VirtualMachineShutdowned VM is shutdowned from inside
	VirtualMachineShutdowned VirtualMachineState = "Shutdowned"
)

// VirtualMachine represents a virtual machine
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html
type VirtualMachine struct {
	Account               string            `json:"account,omitempty" doc:"the account associated with the virtual machine"`
	AccountID             *UUID             `json:"accountid,omitempty" doc:"the account ID associated with the virtual machine"`
	AffinityGroup         []AffinityGroup   `json:"affinitygroup,omitempty" doc:"list of affinity groups associated with the virtual machine"`
	ClusterID             *UUID             `json:"clusterid,omitempty" doc:"the ID of the vm's cluster"`
	ClusterName           string            `json:"clustername,omitempty" doc:"the name of the vm's cluster"`
	CPUNumber             int               `json:"cpunumber,omitempty" doc:"the number of cpu this virtual machine is running with"`
	CPUSpeed              int               `json:"cpuspeed,omitempty" doc:"the speed of each cpu"`
	CPUUsed               string            `json:"cpuused,omitempty" doc:"the amount of the vm's CPU currently used"`
	Created               string            `json:"created,omitempty" doc:"the date when this virtual machine was created"`
	Details               map[string]string `json:"details,omitempty" doc:"Vm details in key/value pairs."`
	DiskIoRead            int64             `json:"diskioread,omitempty" doc:"the read (io) of disk on the vm"`
	DiskIoWrite           int64             `json:"diskiowrite,omitempty" doc:"the write (io) of disk on the vm"`
	DiskKbsRead           int64             `json:"diskkbsread,omitempty" doc:"the read (bytes) of disk on the vm"`
	DiskKbsWrite          int64             `json:"diskkbswrite,omitempty" doc:"the write (bytes) of disk on the vm"`
	DiskOfferingID        *UUID             `json:"diskofferingid,omitempty" doc:"the ID of the disk offering of the virtual machine"`
	DiskOfferingName      string            `json:"diskofferingname,omitempty" doc:"the name of the disk offering of the virtual machine"`
	DisplayName           string            `json:"displayname,omitempty" doc:"user generated name. The name of the virtual machine is returned if no displayname exists."`
	DisplayVM             bool              `json:"displayvm,omitempty" doc:"an optional field whether to the display the vm to the end user or not."`
	Domain                string            `json:"domain,omitempty" doc:"the name of the domain in which the virtual machine exists"`
	DomainID              *UUID             `json:"domainid,omitempty" doc:"the ID of the domain in which the virtual machine exists"`
	ForVirtualNetwork     bool              `json:"forvirtualnetwork,omitempty" doc:"the virtual network for the service offering"`
	Group                 string            `json:"group,omitempty" doc:"the group name of the virtual machine"`
	GroupID               *UUID             `json:"groupid,omitempty" doc:"the group ID of the virtual machine"`
	HAEnable              bool              `json:"haenable,omitempty" doc:"true if high-availability is enabled, false otherwise"`
	HostID                *UUID             `json:"hostid,omitempty" doc:"the ID of the host for the virtual machine"`
	HostName              string            `json:"hostname,omitempty" doc:"the name of the host for the virtual machine"`
	Hypervisor            string            `json:"hypervisor,omitempty" doc:"the hypervisor on which the template runs"`
	ID                    *UUID             `json:"id,omitempty" doc:"the ID of the virtual machine"`
	InstanceName          string            `json:"instancename,omitempty" doc:"instance name of the user vm; this parameter is returned to the ROOT admin only"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty" doc:"true if vm contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory."`
	IsoDisplayText        string            `json:"isodisplaytext,omitempty" doc:"an alternate display text of the ISO attached to the virtual machine"`
	IsoID                 *UUID             `json:"isoid,omitempty" doc:"the ID of the ISO attached to the virtual machine"`
	IsoName               string            `json:"isoname,omitempty" doc:"the name of the ISO attached to the virtual machine"`
	KeyPair               string            `json:"keypair,omitempty" doc:"ssh key-pair"`
	Memory                int               `json:"memory,omitempty" doc:"the memory allocated for the virtual machine"`
	Name                  string            `json:"name,omitempty" doc:"the name of the virtual machine"`
	NetworkKbsRead        int64             `json:"networkkbsread,omitempty" doc:"the incoming network traffic on the vm"`
	NetworkKbsWrite       int64             `json:"networkkbswrite,omitempty" doc:"the outgoing network traffic on the host"`
	Nic                   []Nic             `json:"nic,omitempty" doc:"the list of nics associated with vm"`
	OSCategoryID          *UUID             `json:"oscategoryid,omitempty" doc:"Os category ID of the virtual machine"`
	OSCategoryName        string            `json:"oscategoryname,omitempty" doc:"Os category name of the virtual machine"`
	Password              string            `json:"password,omitempty" doc:"the password (if exists) of the virtual machine"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty" doc:"true if the password rest feature is enabled, false otherwise"`
	PCIDevices            []PCIDevice       `json:"pcidevices,omitempty" doc:"list of PCI devices"`
	PodID                 *UUID             `json:"podid,omitempty" doc:"the ID of the vm's pod"`
	PodName               string            `json:"podname,omitempty" doc:"the name of the vm's pod"`
	PublicIP              string            `json:"publicip,omitempty" doc:"public IP address id associated with vm via Static nat rule"`
	PublicIPID            *UUID             `json:"publicipid,omitempty" doc:"public IP address id associated with vm via Static nat rule"`
	RootDeviceID          int64             `json:"rootdeviceid,omitempty" doc:"device ID of the root volume"`
	RootDeviceType        string            `json:"rootdevicetype,omitempty" doc:"device type of the root volume"`
	SecurityGroup         []SecurityGroup   `json:"securitygroup,omitempty" doc:"list of security groups associated with the virtual machine"`
	ServiceOfferingID     *UUID             `json:"serviceofferingid,omitempty" doc:"the ID of the service offering of the virtual machine"`
	ServiceOfferingName   string            `json:"serviceofferingname,omitempty" doc:"the name of the service offering of the virtual machine"`
	ServiceState          string            `json:"servicestate,omitempty" doc:"State of the Service from LB rule"`
	State                 string            `json:"state,omitempty" doc:"the state of the virtual machine"`
	Tags                  []ResourceTag     `json:"tags,omitempty" doc:"the list of resource tags associated with vm"`
	TemplateDisplayText   string            `json:"templatedisplaytext,omitempty" doc:"an alternate display text of the template for the virtual machine"`
	TemplateID            *UUID             `json:"templateid,omitempty" doc:"the ID of the template for the virtual machine. A -1 is returned if the virtual machine was created from an ISO file."`
	TemplateName          string            `json:"templatename,omitempty" doc:"the name of the template for the virtual machine"`
	ZoneID                *UUID             `json:"zoneid,omitempty" doc:"the ID of the availablility zone for the virtual machine"`
	ZoneName              string            `json:"zonename,omitempty" doc:"the name of the availability zone for the virtual machine"`
}

// ResourceType returns the type of the resource
func (VirtualMachine) ResourceType() string {
	return "UserVM"
}

// Delete destroys the VM
func (vm VirtualMachine) Delete(ctx context.Context, client *Client) error {
	_, err := client.RequestWithContext(ctx, &DestroyVirtualMachine{
		ID: vm.ID,
	})

	return err
}

// ListRequest builds the ListVirtualMachines request
func (vm VirtualMachine) ListRequest() (ListCommand, error) {
	// XXX: AffinityGroupID, SecurityGroupID, Tags

	req := &ListVirtualMachines{
		Account:    vm.Account,
		DomainID:   vm.DomainID,
		GroupID:    vm.GroupID,
		ID:         vm.ID,
		Name:       vm.Name,
		State:      vm.State,
		TemplateID: vm.TemplateID,
		ZoneID:     vm.ZoneID,
	}

	nic := vm.DefaultNic()
	if nic != nil {
		req.IPAddress = nic.IPAddress
	}

	return req, nil
}

// DefaultNic returns the default nic
func (vm VirtualMachine) DefaultNic() *Nic {
	for _, nic := range vm.Nic {
		if nic.IsDefault {
			return &nic
		}
	}

	return nil
}

// IP returns the default nic IP address
func (vm VirtualMachine) IP() *net.IP {
	nic := vm.DefaultNic()
	if nic != nil {
		ip := nic.IPAddress
		return &ip
	}

	return nil
}

// NicsByType returns the corresponding interfaces base on the given type
func (vm VirtualMachine) NicsByType(nicType string) []Nic {
	nics := make([]Nic, 0)
	for _, nic := range vm.Nic {
		if nic.Type == nicType {
			// XXX The API forgets to specify it
			nic.VirtualMachineID = vm.ID
			nics = append(nics, nic)
		}
	}
	return nics
}

// NicByNetworkID returns the corresponding interface based on the given NetworkID
//
// A VM cannot be connected twice to a same network.
func (vm VirtualMachine) NicByNetworkID(networkID UUID) *Nic {
	for _, nic := range vm.Nic {
		if nic.NetworkID.Equal(networkID) {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}
	return nil
}

// NicByID returns the corresponding interface base on its ID
func (vm VirtualMachine) NicByID(nicID UUID) *Nic {
	for _, nic := range vm.Nic {
		if nic.ID.Equal(nicID) {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}

	return nil
}

// IPToNetwork represents a mapping between ip and networks
type IPToNetwork struct {
	IP        net.IP `json:"ip,omitempty"`
	Ipv6      net.IP `json:"ipv6,omitempty"`
	NetworkID *UUID  `json:"networkid,omitempty"`
}

// PCIDevice represents a PCI card present in the host
type PCIDevice struct {
	PCIVendorName     string `json:"pcivendorname,omitempty" doc:"Device vendor name of pci card"`
	DeviceID          *UUID  `json:"deviceid,omitempty" doc:"Device model ID of pci card"`
	RemainingCapacity int    `json:"remainingcapacity,omitempty" doc:"Remaining capacity in terms of no. of more VMs that can be deployped with this vGPU type"`
	MaxCapacity       int    `json:"maxcapacity,omitempty" doc:"Maximum vgpu can be created with this vgpu type on the given pci group"`
	PCIVendorID       *UUID  `json:"pcivendorid,omitempty" doc:"Device vendor ID of pci card"`
	PCIDeviceName     string `json:"pcidevicename,omitempty" doc:"Device model name of pci card"`
}

// Password represents an encrypted password
//
// TODO: method to decrypt it, https://cwiki.apache.org/confluence/pages/viewpage.action?pageId=34014652
type Password struct {
	EncryptedPassword string `json:"encryptedpassword"`
}

// VirtualMachineUserData represents the base64 encoded user-data
type VirtualMachineUserData struct {
	UserData         string `json:"userdata,omitempty" doc:"Base 64 encoded VM user data"`
	VirtualMachineID *UUID  `json:"virtualmachineid,omitempty" doc:"the ID of the virtual machine"`
}

// Decode decodes as a readable string the content of the user-data (base64 · gzip)
func (userdata VirtualMachineUserData) Decode() (string, error) {
	data, err := base64.StdEncoding.DecodeString(userdata.UserData)
	if err != nil {
		return "", err
	}
	// 0x1f8b is the magic number for gzip
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return string(data), nil
	}
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer gr.Close() // nolint: errcheck

	str, err := ioutil.ReadAll(gr)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

// DeployVirtualMachine (Async) represents the machine creation
//
// Regarding the UserData field, the client is responsible to base64 (and probably gzip) it. Doing it within this library would make the integration with other tools, e.g. Terraform harder.
type DeployVirtualMachine struct {
	Account            string            `json:"account,omitempty" doc:"an optional account for the virtual machine. Must be used with domainId."`
	AffinityGroupIDs   []UUID            `json:"affinitygroupids,omitempty" doc:"comma separated list of affinity groups id that are going to be applied to the virtual machine. Mutually exclusive with affinitygroupnames parameter"`
	AffinityGroupNames []string          `json:"affinitygroupnames,omitempty" doc:"comma separated list of affinity groups names that are going to be applied to the virtual machine.Mutually exclusive with affinitygroupids parameter"`
	CustomID           *UUID             `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	DeploymentPlanner  string            `json:"deploymentplanner,omitempty" doc:"Deployment planner to use for vm allocation. Available to ROOT admin only"`
	Details            map[string]string `json:"details,omitempty" doc:"used to specify the custom parameters."`
	DiskOfferingID     *UUID             `json:"diskofferingid,omitempty" doc:"the ID of the disk offering for the virtual machine. If the template is of ISO format, the diskOfferingId is for the root disk volume. Otherwise this parameter is used to indicate the offering for the data disk volume. If the templateId parameter passed is from a Template object, the diskOfferingId refers to a DATA Disk Volume created. If the templateId parameter passed is from an ISO object, the diskOfferingId refers to a ROOT Disk Volume created."`
	DisplayName        string            `json:"displayname,omitempty" doc:"an optional user generated name for the virtual machine"`
	DisplayVM          *bool             `json:"displayvm,omitempty" doc:"an optional field, whether to the display the vm to the end user or not."`
	DomainID           *UUID             `json:"domainid,omitempty" doc:"an optional domainId for the virtual machine. If the account parameter is used, domainId must also be used."`
	Group              string            `json:"group,omitempty" doc:"an optional group for the virtual machine"`
	HostID             *UUID             `json:"hostid,omitempty" doc:"destination Host ID to deploy the VM to - parameter available for root admin only"`
	Hypervisor         string            `json:"hypervisor,omitempty" doc:"the hypervisor on which to deploy the virtual machine"`
	IP4                *bool             `json:"ip4,omitempty" doc:"True to set an IPv4 to the default interface"`
	IP6                *bool             `json:"ip6,omitempty" doc:"True to set an IPv6 to the default interface"`
	IP6Address         net.IP            `json:"ip6address,omitempty" doc:"the ipv6 address for default vm's network"`
	IPAddress          net.IP            `json:"ipaddress,omitempty" doc:"the ip address for default vm's network"`
	IPToNetworkList    []IPToNetwork     `json:"iptonetworklist,omitempty" doc:"ip to network mapping. Can't be specified with networkIds parameter. Example: iptonetworklist[0].ip=10.10.10.11&iptonetworklist[0].ipv6=fc00:1234:5678::abcd&iptonetworklist[0].networkid=uuid - requests to use ip 10.10.10.11 in network id=uuid"`
	Keyboard           string            `json:"keyboard,omitempty" doc:"an optional keyboard device type for the virtual machine. valid value can be one of de,de-ch,es,fi,fr,fr-be,fr-ch,is,it,jp,nl-be,no,pt,uk,us"`
	KeyPair            string            `json:"keypair,omitempty" doc:"name of the ssh key pair used to login to the virtual machine"`
	Name               string            `json:"name,omitempty" doc:"host name for the virtual machine"`
	NetworkIDs         []UUID            `json:"networkids,omitempty" doc:"list of network ids used by virtual machine. Can't be specified with ipToNetworkList parameter"`
	RootDiskSize       int64             `json:"rootdisksize,omitempty" doc:"Optional field to resize root disk on deploy. Value is in GB. Only applies to template-based deployments. Analogous to details[0].rootdisksize, which takes precedence over this parameter if both are provided"`
	SecurityGroupIDs   []UUID            `json:"securitygroupids,omitempty" doc:"comma separated list of security groups id that going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupnames parameter"`
	SecurityGroupNames []string          `json:"securitygroupnames,omitempty" doc:"comma separated list of security groups names that going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupids parameter"`
	ServiceOfferingID  *UUID             `json:"serviceofferingid" doc:"the ID of the service offering for the virtual machine"`
	Size               int64             `json:"size,omitempty" doc:"the arbitrary size for the DATADISK volume. Mutually exclusive with diskOfferingId"`
	StartVM            *bool             `json:"startvm,omitempty" doc:"true if start vm after creating. Default value is true"`
	TemplateID         *UUID             `json:"templateid" doc:"the ID of the template for the virtual machine"`
	UserData           string            `json:"userdata,omitempty" doc:"an optional binary data that can be sent to the virtual machine upon a successful deployment. This binary data must be base64 encoded before adding it to the request. Using HTTP GET (via querystring), you can send up to 2KB of data after base64 encoding. Using HTTP POST(via POST body), you can send up to 32K of data after base64 encoding."`
	ZoneID             *UUID             `json:"zoneid" doc:"availability zone for the virtual machine"`
	_                  bool              `name:"deployVirtualMachine" description:"Creates and automatically starts a virtual machine based on a service offering, disk offering, and template."`
}

func (req DeployVirtualMachine) onBeforeSend(params url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) > 0 && len(req.AffinityGroupNames) > 0 {
		return fmt.Errorf("either AffinityGroupIDs or AffinityGroupNames must be set")
	}

	// Either SecurityGroupIDs or SecurityGroupNames must be set
	if len(req.SecurityGroupIDs) > 0 && len(req.SecurityGroupNames) > 0 {
		return fmt.Errorf("either SecurityGroupIDs or SecurityGroupNames must be set")
	}

	return nil
}

func (DeployVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (DeployVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// StartVirtualMachine (Async) represents the creation of the virtual machine
type StartVirtualMachine struct {
	ID                *UUID  `json:"id" doc:"The ID of the virtual machine"`
	DeploymentPlanner string `json:"deploymentplanner,omitempty" doc:"Deployment planner to use for vm allocation. Available to ROOT admin only"`
	HostID            *UUID  `json:"hostid,omitempty" doc:"destination Host ID to deploy the VM to - parameter available for root admin only"`
	_                 bool   `name:"startVirtualMachine" description:"Starts a virtual machine."`
}

func (StartVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (StartVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// StopVirtualMachine (Async) represents the stopping of the virtual machine
type StopVirtualMachine struct {
	ID     *UUID `json:"id" doc:"The ID of the virtual machine"`
	Forced *bool `json:"forced,omitempty" doc:"Force stop the VM (vm is marked as Stopped even when command fails to be send to the backend).  The caller knows the VM is stopped."`
	_      bool  `name:"stopVirtualMachine" description:"Stops a virtual machine."`
}

func (StopVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (StopVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// RebootVirtualMachine (Async) represents the rebooting of the virtual machine
type RebootVirtualMachine struct {
	ID *UUID `json:"id" doc:"The ID of the virtual machine"`
	_  bool  `name:"rebootVirtualMachine" description:"Reboots a virtual machine."`
}

func (RebootVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (RebootVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// RestoreVirtualMachine (Async) represents the restoration of the virtual machine
type RestoreVirtualMachine struct {
	VirtualMachineID *UUID `json:"virtualmachineid" doc:"Virtual Machine ID"`
	TemplateID       *UUID `json:"templateid,omitempty" doc:"an optional template Id to restore vm from the new template. This can be an ISO id in case of restore vm deployed using ISO"`
	RootDiskSize     int64 `json:"rootdisksize,omitempty" doc:"Optional field to resize root disk on restore. Value is in GB. Only applies to template-based deployments."`
	_                bool  `name:"restoreVirtualMachine" description:"Restore a VM to original template/ISO or new template/ISO"`
}

func (RestoreVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (RestoreVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// RecoverVirtualMachine represents the restoration of the virtual machine
type RecoverVirtualMachine struct {
	ID *UUID `json:"id" doc:"The ID of the virtual machine"`
	_  bool  `name:"recoverVirtualMachine" description:"Recovers a virtual machine."`
}

func (RecoverVirtualMachine) response() interface{} {
	return new(VirtualMachine)
}

// DestroyVirtualMachine (Async) represents the destruction of the virtual machine
type DestroyVirtualMachine struct {
	ID      *UUID `json:"id" doc:"The ID of the virtual machine"`
	Expunge *bool `json:"expunge,omitempty" doc:"If true is passed, the vm is expunged immediately. False by default."`
	_       bool  `name:"destroyVirtualMachine" description:"Destroys a virtual machine."`
}

func (DestroyVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (DestroyVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// UpdateVirtualMachine represents the update of the virtual machine
type UpdateVirtualMachine struct {
	ID                    *UUID             `json:"id" doc:"The ID of the virtual machine"`
	CustomID              *UUID             `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	Details               map[string]string `json:"details,omitempty" doc:"Details in key/value pairs."`
	DisplayName           string            `json:"displayname,omitempty" doc:"user generated name"`
	DisplayVM             *bool             `json:"displayvm,omitempty" doc:"an optional field, whether to the display the vm to the end user or not."`
	Group                 string            `json:"group,omitempty" doc:"group of the virtual machine"`
	HAEnable              *bool             `json:"haenable,omitempty" doc:"true if high-availability is enabled for the virtual machine, false otherwise"`
	IsDynamicallyScalable *bool             `json:"isdynamicallyscalable,omitempty" doc:"true if VM contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	Name                  string            `json:"name,omitempty" doc:"new host name of the vm. The VM has to be stopped/started for this update to take affect"`
	SecurityGroupIDs      []UUID            `json:"securitygroupids,omitempty" doc:"list of security group ids to be applied on the virtual machine."`
	UserData              string            `json:"userdata,omitempty" doc:"an optional binary data that can be sent to the virtual machine upon a successful deployment. This binary data must be base64 encoded before adding it to the request. Using HTTP GET (via querystring), you can send up to 2KB of data after base64 encoding. Using HTTP POST(via POST body), you can send up to 32K of data after base64 encoding."`
	_                     bool              `name:"updateVirtualMachine" description:"Updates properties of a virtual machine. The VM has to be stopped and restarted for the new properties to take effect. UpdateVirtualMachine does not first check whether the VM is stopped. Therefore, stop the VM manually before issuing this call."`
}

func (UpdateVirtualMachine) response() interface{} {
	return new(VirtualMachine)
}

// ExpungeVirtualMachine represents the annihilation of a VM
type ExpungeVirtualMachine struct {
	ID *UUID `json:"id" doc:"The ID of the virtual machine"`
	_  bool  `name:"expungeVirtualMachine" description:"Expunge a virtual machine. Once expunged, it cannot be recoverd."`
}

func (ExpungeVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (ExpungeVirtualMachine) asyncResponse() interface{} {
	return new(booleanResponse)
}

// ScaleVirtualMachine (Async) scales the virtual machine to a new service offering.
//
// ChangeServiceForVirtualMachine does the same thing but returns the
// new Virtual Machine which is more consistent with the rest of the API.
type ScaleVirtualMachine struct {
	ID                *UUID             `json:"id" doc:"The ID of the virtual machine"`
	ServiceOfferingID *UUID             `json:"serviceofferingid" doc:"the ID of the service offering for the virtual machine"`
	Details           map[string]string `json:"details,omitempty" doc:"name value pairs of custom parameters for cpu,memory and cpunumber. example details[i].name=value"`
	_                 bool              `name:"scaleVirtualMachine" description:"Scales the virtual machine to a new service offering."`
}

func (ScaleVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (ScaleVirtualMachine) asyncResponse() interface{} {
	return new(booleanResponse)
}

// ChangeServiceForVirtualMachine changes the service offering for a virtual machine. The virtual machine must be in a "Stopped" state for this command to take effect.
type ChangeServiceForVirtualMachine struct {
	ID                *UUID             `json:"id" doc:"The ID of the virtual machine"`
	ServiceOfferingID *UUID             `json:"serviceofferingid" doc:"the service offering ID to apply to the virtual machine"`
	Details           map[string]string `json:"details,omitempty" doc:"name value pairs of custom parameters for cpu, memory and cpunumber. example details[i].name=value"`
	_                 bool              `name:"changeServiceForVirtualMachine" description:"Changes the service offering for a virtual machine. The virtual machine must be in a \"Stopped\" state for this command to take effect."`
}

func (ChangeServiceForVirtualMachine) response() interface{} {
	return new(VirtualMachine)
}

// ResetPasswordForVirtualMachine resets the password for virtual machine. The virtual machine must be in a "Stopped" state...
type ResetPasswordForVirtualMachine struct {
	ID *UUID `json:"id" doc:"The ID of the virtual machine"`
	_  bool  `name:"resetPasswordForVirtualMachine" description:"Resets the password for virtual machine. The virtual machine must be in a \"Stopped\" state and the template must already support this feature for this command to take effect."`
}

func (ResetPasswordForVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}
func (ResetPasswordForVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// GetVMPassword asks for an encrypted password
type GetVMPassword struct {
	ID *UUID `json:"id" doc:"The ID of the virtual machine"`
	_  bool  `name:"getVMPassword" description:"Returns an encrypted password for the VM"`
}

func (GetVMPassword) response() interface{} {
	return new(Password)
}

// ListVirtualMachines represents a search for a VM
type ListVirtualMachines struct {
	Account           string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	AffinityGroupID   *UUID         `json:"affinitygroupid,omitempty" doc:"list vms by affinity group"`
	Details           []string      `json:"details,omitempty" doc:"comma separated list of host details requested, value can be a list of [all, group, nics, stats, secgrp, tmpl, servoff, diskoff, iso, volume, min, affgrp]. If no parameter is passed in, the details will be defaulted to all"`
	DisplayVM         *bool         `json:"displayvm,omitempty" doc:"list resources by display flag; only ROOT admin is eligible to pass this parameter"`
	DomainID          *UUID         `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ForVirtualNetwork *bool         `json:"forvirtualnetwork,omitempty" doc:"list by network type; true if need to list vms using Virtual Network, false otherwise"`
	GroupID           *UUID         `json:"groupid,omitempty" doc:"the group ID"`
	HostID            *UUID         `json:"hostid,omitempty" doc:"the host ID"`
	Hypervisor        string        `json:"hypervisor,omitempty" doc:"the target hypervisor for the template"`
	ID                *UUID         `json:"id,omitempty" doc:"the ID of the virtual machine"`
	IDs               []string      `json:"ids,omitempty" doc:"the IDs of the virtual machines, mutually exclusive with id"`
	IPAddress         net.IP        `json:"ipaddress,omitempty" doc:"an IP address to filter the result"`
	IsoID             *UUID         `json:"isoid,omitempty" doc:"list vms by iso"`
	IsRecursive       *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword           string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll           *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name              string        `json:"name,omitempty" doc:"name of the virtual machine"`
	NetworkID         *UUID         `json:"networkid,omitempty" doc:"list by network id"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	ServiceOfferindID *UUID         `json:"serviceofferingid,omitempty" doc:"list by the service offering"`
	State             string        `json:"state,omitempty" doc:"state of the virtual machine"`
	Tags              []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	TemplateID        *UUID         `json:"templateid,omitempty" doc:"list vms by template"`
	ZoneID            *UUID         `json:"zoneid,omitempty" doc:"the availability zone ID"`
	_                 bool          `name:"listVirtualMachines" description:"List the virtual machines owned by the account."`
}

// ListVirtualMachinesResponse represents a list of virtual machines
type ListVirtualMachinesResponse struct {
	Count          int              `json:"count"`
	VirtualMachine []VirtualMachine `json:"virtualmachine"`
}

func (ListVirtualMachines) response() interface{} {
	return new(ListVirtualMachinesResponse)
}

// SetPage sets the current page
func (ls *ListVirtualMachines) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVirtualMachines) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (ListVirtualMachines) each(resp interface{}, callback IterateItemFunc) {
	vms, ok := resp.(*ListVirtualMachinesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListVirtualMachinesResponse expected, got %T", resp))
		return
	}

	for i := range vms.VirtualMachine {
		if !callback(&vms.VirtualMachine[i], nil) {
			break
		}
	}
}

// AddNicToVirtualMachine (Async) adds a NIC to a VM
type AddNicToVirtualMachine struct {
	NetworkID        *UUID  `json:"networkid" doc:"Network ID"`
	VirtualMachineID *UUID  `json:"virtualmachineid" doc:"Virtual Machine ID"`
	IPAddress        net.IP `json:"ipaddress,omitempty" doc:"IP Address for the new network"`
	_                bool   `name:"addNicToVirtualMachine" description:"Adds VM to specified network by creating a NIC"`
}

func (AddNicToVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (AddNicToVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// RemoveNicFromVirtualMachine (Async) removes a NIC from a VM
type RemoveNicFromVirtualMachine struct {
	NicID            *UUID `json:"nicid" doc:"NIC ID"`
	VirtualMachineID *UUID `json:"virtualmachineid" doc:"Virtual Machine ID"`
	_                bool  `name:"removeNicFromVirtualMachine" description:"Removes VM from specified network by deleting a NIC"`
}

func (RemoveNicFromVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (RemoveNicFromVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// UpdateDefaultNicForVirtualMachine (Async) adds a NIC to a VM
type UpdateDefaultNicForVirtualMachine struct {
	NicID            *UUID `json:"nicid" doc:"NIC ID"`
	VirtualMachineID *UUID `json:"virtualmachineid" doc:"Virtual Machine ID"`
	_                bool  `name:"updateDefaultNicForVirtualMachine" description:"Changes the default NIC on a VM"`
}

func (UpdateDefaultNicForVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}
func (UpdateDefaultNicForVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// GetVirtualMachineUserData returns the user-data of the given VM
type GetVirtualMachineUserData struct {
	VirtualMachineID *UUID `json:"virtualmachineid" doc:"The ID of the virtual machine"`
	_                bool  `name:"getVirtualMachineUserData" description:"Returns user data associated with the VM"`
}

func (GetVirtualMachineUserData) response() interface{} {
	return new(VirtualMachineUserData)
}

// Decode decodes the base64 / gzipped encoded user data

// MigrateVirtualMachine (Async) attempts migration of a VM to a different host or Root volume of the vm to a different storage pool
type MigrateVirtualMachine struct {
	HostID           *UUID `json:"hostid,omitempty" doc:"Destination Host ID to migrate VM to. Required for live migrating a VM from host to host"`
	StorageID        *UUID `json:"storageid,omitempty" doc:"Destination storage pool ID to migrate VM volumes to. Required for migrating the root disk volume"`
	VirtualMachineID *UUID `json:"virtualmachineid" doc:"the ID of the virtual machine"`
	_                bool  `name:"migrateVirtualMachine" description:"Attempts Migration of a VM to a different host or Root volume of the vm to a different storage pool"`
}

func (MigrateVirtualMachine) response() interface{} {
	return new(AsyncJobResult)
}

func (MigrateVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}
