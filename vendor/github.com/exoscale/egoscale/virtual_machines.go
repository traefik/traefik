package egoscale

import "net"

// VirtualMachine reprents a virtual machine
type VirtualMachine struct {
	ID                    string            `json:"id,omitempty"`
	Account               string            `json:"account,omitempty"`
	ClusterID             string            `json:"clusterid,omitempty"`
	ClusterName           string            `json:"clustername,omitempty"`
	CPUNumber             int64             `json:"cpunumber,omitempty"`
	CPUSpeed              int64             `json:"cpuspeed,omitempty"`
	CPUUsed               string            `json:"cpuused,omitempty"`
	Created               string            `json:"created,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	DiskIoRead            int64             `json:"diskioread,omitempty"`
	DiskIoWrite           int64             `json:"diskiowrite,omitempty"`
	DiskKbsRead           int64             `json:"diskkbsread,omitempty"`
	DiskKbsWrite          int64             `json:"diskkbswrite,omitempty"`
	DiskOfferingID        string            `json:"diskofferingid,omitempty"`
	DiskOfferingName      string            `json:"diskofferingname,omitempty"`
	DisplayName           string            `json:"displayname,omitempty"`
	DisplayVM             bool              `json:"displayvm,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainID              string            `json:"domainid,omitempty"`
	ForVirtualNetwork     bool              `json:"forvirtualnetwork,omitempty"`
	Group                 string            `json:"group,omitempty"`
	GroupID               string            `json:"groupid,omitempty"`
	GuestOsID             string            `json:"guestosid,omitempty"`
	HaEnable              bool              `json:"haenable,omitempty"`
	HostID                string            `json:"hostid,omitempty"`
	HostName              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	InstanceName          string            `json:"instancename,omitempty"` // root only
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty"`
	IsoDisplayText        string            `json:"isodisplaytext,omitempty"`
	IsoID                 string            `json:"isoid,omitempty"`
	IsoName               string            `json:"isoname,omitempty"`
	KeyPair               string            `json:"keypair,omitempty"`
	Memory                int64             `json:"memory,omitempty"`
	MemoryIntFreeKbs      int64             `json:"memoryintfreekbs,omitempty"`
	MemoryKbs             int64             `json:"memorykbs,omitempty"`
	MemoryTargetKbs       int64             `json:"memorytargetkbs,omitempty"`
	Name                  string            `json:"name,omitempty"`
	NetworkKbsRead        int64             `json:"networkkbsread,omitempty"`
	NetworkKbsWrite       int64             `json:"networkkbswrite,omitempty"`
	OsCategoryID          string            `json:"oscategoryid,omitempty"`
	OsTypeID              int64             `json:"ostypeid,omitempty"`
	Password              string            `json:"password,omitempty"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty"`
	PCIDevices            string            `json:"pcidevices,omitempty"` // not in the doc
	PodID                 string            `json:"podid,omitempty"`
	PodName               string            `json:"podname,omitempty"`
	Project               string            `json:"project,omitempty"`
	ProjectID             string            `json:"projectid,omitempty"`
	PublicIP              string            `json:"publicip,omitempty"`
	PublicIPID            string            `json:"publicipid,omitempty"`
	RootDeviceID          int64             `json:"rootdeviceid,omitempty"`
	RootDeviceType        string            `json:"rootdevicetype,omitempty"`
	ServiceOfferingID     string            `json:"serviceofferingid,omitempty"`
	ServiceOfferingName   string            `json:"serviceofferingname,omitempty"`
	ServiceState          string            `json:"servicestate,omitempty"`
	State                 string            `json:"state,omitempty"`
	TemplateDisplayText   string            `json:"templatedisplaytext,omitempty"`
	TemplateID            string            `json:"templateid,omitempty"`
	TemplateName          string            `json:"templatename,omitempty"`
	UserID                string            `json:"userid,omitempty"`   // not in the doc
	UserName              string            `json:"username,omitempty"` // not in the doc
	Vgpu                  string            `json:"vgpu,omitempty"`     // not in the doc
	ZoneID                string            `json:"zoneid,omitempty"`
	ZoneName              string            `json:"zonename,omitempty"`
	AffinityGroup         []AffinityGroup   `json:"affinitygroup,omitempty"`
	Nic                   []Nic             `json:"nic,omitempty"`
	SecurityGroup         []SecurityGroup   `json:"securitygroup,omitempty"`
	Tags                  []ResourceTag     `json:"tags,omitempty"`
	JobID                 string            `json:"jobid,omitempty"`
	JobStatus             JobStatusType     `json:"jobstatus,omitempty"`
}

// ResourceType returns the type of the resource
func (*VirtualMachine) ResourceType() string {
	return "UserVM"
}

// NicsByType returns the corresponding interfaces base on the given type
func (vm *VirtualMachine) NicsByType(nicType string) []Nic {
	nics := make([]Nic, 0)
	for _, nic := range vm.Nic {
		if nic.Type == nicType {
			// XXX The CloudStack API forgets to specify it
			nic.VirtualMachineID = vm.ID
			nics = append(nics, nic)
		}
	}
	return nics
}

// NicByNetworkID returns the corresponding interface based on the given NetworkID
func (vm *VirtualMachine) NicByNetworkID(networkID string) *Nic {
	for _, nic := range vm.Nic {
		if nic.NetworkID == networkID {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}
	return nil
}

// NicByID returns the corresponding interface base on its ID
func (vm *VirtualMachine) NicByID(nicID string) *Nic {
	for _, nic := range vm.Nic {
		if nic.ID == nicID {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}

	return nil
}

// IPToNetwork represents a mapping between ip and networks
type IPToNetwork struct {
	IP        string `json:"ip,omitempty"`
	IPV6      string `json:"ipv6,omitempty"`
	NetworkID string `json:"networkid,omitempty"`
}

// VirtualMachineResponse represents a generic Virtual Machine response
type VirtualMachineResponse struct {
	VirtualMachine VirtualMachine `json:"virtualmachine"`
}

// DeployVirtualMachine (Async) represents the machine creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/deployVirtualMachine.html
type DeployVirtualMachine struct {
	ServiceOfferingID  string            `json:"serviceofferingid"`
	TemplateID         string            `json:"templateid"`
	ZoneID             string            `json:"zoneid"`
	Account            string            `json:"account,omitempty"`
	AffinityGroupIDs   []string          `json:"affinitygroupids,omitempty"`
	AffinityGroupNames []string          `json:"affinitygroupnames,omitempty"`
	CustomID           string            `json:"customid,omitempty"`          // root only
	DeploymentPlanner  string            `json:"deploymentplanner,omitempty"` // root only
	Details            map[string]string `json:"details,omitempty"`
	DiskOfferingID     string            `json:"diskofferingid,omitempty"`
	DisplayName        string            `json:"displayname,omitempty"`
	DisplayVM          bool              `json:"displayvm,omitempty"`
	DomainID           string            `json:"domainid,omitempty"`
	Group              string            `json:"group,omitempty"`
	HostID             string            `json:"hostid,omitempty"`
	Hypervisor         string            `json:"hypervisor,omitempty"`
	IP6Address         net.IP            `json:"ip6address,omitempty"`
	IPAddress          net.IP            `json:"ipaddress,omitempty"`
	IPToNetworkList    []IPToNetwork     `json:"iptonetworklist,omitempty"`
	Keyboard           string            `json:"keyboard,omitempty"`
	KeyPair            string            `json:"keypair,omitempty"`
	Name               string            `json:"name,omitempty"`
	NetworkIDs         []string          `json:"networkids,omitempty"` // mutually exclusive with IPToNetworkList
	ProjectID          string            `json:"projectid,omitempty"`
	RootDiskSize       int64             `json:"rootdisksize,omitempty"` // in GiB
	SecurityGroupIDs   []string          `json:"securitygroupids,omitempty"`
	SecurityGroupNames []string          `json:"securitygroupnames,omitempty"` // does nothing, mutually exclusive
	Size               string            `json:"size,omitempty"`               // mutually exclusive with DiskOfferingID
	StartVM            bool              `json:"startvm,omitempty"`
	UserData           string            `json:"userdata,omitempty"` // the client is responsible to base64/gzip it
}

func (*DeployVirtualMachine) name() string {
	return "deployVirtualMachine"
}

func (*DeployVirtualMachine) asyncResponse() interface{} {
	return new(DeployVirtualMachineResponse)
}

// DeployVirtualMachineResponse represents a deployed VM instance
type DeployVirtualMachineResponse VirtualMachineResponse

// StartVirtualMachine (Async) represents the creation of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/startVirtualMachine.html
type StartVirtualMachine struct {
	ID                string `json:"id"`
	DeploymentPlanner string `json:"deploymentplanner,omitempty"` // root only
	HostID            string `json:"hostid,omitempty"`            // root only
}

func (*StartVirtualMachine) name() string {
	return "startVirtualMachine"
}
func (*StartVirtualMachine) asyncResponse() interface{} {
	return new(StartVirtualMachineResponse)
}

// StartVirtualMachineResponse represents a started VM instance
type StartVirtualMachineResponse VirtualMachineResponse

// StopVirtualMachine (Async) represents the stopping of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/stopVirtualMachine.html
type StopVirtualMachine struct {
	ID     string `json:"id"`
	Forced bool   `json:"forced,omitempty"`
}

func (*StopVirtualMachine) name() string {
	return "stopVirtualMachine"
}

func (*StopVirtualMachine) asyncResponse() interface{} {
	return new(StopVirtualMachineResponse)
}

// StopVirtualMachineResponse represents a stopped VM instance
type StopVirtualMachineResponse VirtualMachineResponse

// RebootVirtualMachine (Async) represents the rebooting of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/rebootVirtualMachine.html
type RebootVirtualMachine struct {
	ID string `json:"id"`
}

func (*RebootVirtualMachine) name() string {
	return "rebootVirtualMachine"
}

func (*RebootVirtualMachine) asyncResponse() interface{} {
	return new(RebootVirtualMachineResponse)
}

// RebootVirtualMachineResponse represents a rebooted VM instance
type RebootVirtualMachineResponse VirtualMachineResponse

// RestoreVirtualMachine (Async) represents the restoration of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/restoreVirtualMachine.html
type RestoreVirtualMachine struct {
	VirtualMachineID string `json:"virtualmachineid"`
	TemplateID       string `json:"templateid,omitempty"`
}

func (*RestoreVirtualMachine) name() string {
	return "restoreVirtualMachine"
}

func (*RestoreVirtualMachine) asyncResponse() interface{} {
	return new(RestoreVirtualMachineResponse)
}

// RestoreVirtualMachineResponse represents a restored VM instance
type RestoreVirtualMachineResponse VirtualMachineResponse

// RecoverVirtualMachine represents the restoration of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/recoverVirtualMachine.html
type RecoverVirtualMachine struct {
	ID string `json:"virtualmachineid"`
}

func (*RecoverVirtualMachine) name() string {
	return "recoverVirtualMachine"
}

func (*RecoverVirtualMachine) response() interface{} {
	return new(RecoverVirtualMachineResponse)
}

// RecoverVirtualMachineResponse represents a recovered VM instance
type RecoverVirtualMachineResponse VirtualMachineResponse

// DestroyVirtualMachine (Async) represents the destruction of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/destroyVirtualMachine.html
type DestroyVirtualMachine struct {
	ID      string `json:"id"`
	Expunge bool   `json:"expunge,omitempty"`
}

func (*DestroyVirtualMachine) name() string {
	return "destroyVirtualMachine"
}

func (*DestroyVirtualMachine) asyncResponse() interface{} {
	return new(DestroyVirtualMachineResponse)
}

// DestroyVirtualMachineResponse represents a destroyed VM instance
type DestroyVirtualMachineResponse VirtualMachineResponse

// UpdateVirtualMachine represents the update of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/updateVirtualMachine.html
type UpdateVirtualMachine struct {
	ID                    string            `json:"id"`
	CustomID              string            `json:"customid,omitempty"` // root only
	Details               map[string]string `json:"details,omitempty"`
	DisplayName           string            `json:"displayname,omitempty"`
	DisplayVM             bool              `json:"displayvm,omitempty"`
	Group                 string            `json:"group,omitempty"`
	HAEnable              bool              `json:"haenable,omitempty"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty"`
	Name                  string            `json:"name,omitempty"` // must reboot
	OsTypeID              int64             `json:"ostypeid,omitempty"`
	SecurityGroupIDs      []string          `json:"securitygroupids,omitempty"`
	UserData              string            `json:"userdata,omitempty"`
}

func (*UpdateVirtualMachine) name() string {
	return "updateVirtualMachine"
}

func (*UpdateVirtualMachine) response() interface{} {
	return new(UpdateVirtualMachineResponse)
}

// UpdateVirtualMachineResponse represents an updated VM instance
type UpdateVirtualMachineResponse VirtualMachineResponse

// ExpungeVirtualMachine represents the annihilation of a VM
type ExpungeVirtualMachine struct {
	ID string `json:"id"`
}

func (*ExpungeVirtualMachine) name() string {
	return "expungeVirtualMachine"
}

func (*ExpungeVirtualMachine) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ScaleVirtualMachine (Async) represents the scaling of a VM
//
// ChangeServiceForVirtualMachine does the same thing but returns the
// new Virtual Machine which is more consistent with the rest of the API.
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/scaleVirtualMachine.html
type ScaleVirtualMachine struct {
	ID                string            `json:"id"`
	ServiceOfferingID string            `json:"serviceofferingid"`
	Details           map[string]string `json:"details,omitempty"`
}

func (*ScaleVirtualMachine) name() string {
	return "scaleVirtualMachine"
}

func (*ScaleVirtualMachine) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ChangeServiceForVirtualMachine represents the scaling of a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/changeServiceForVirtualMachine.html
type ChangeServiceForVirtualMachine ScaleVirtualMachine

func (*ChangeServiceForVirtualMachine) name() string {
	return "changeServiceForVirtualMachine"
}

func (*ChangeServiceForVirtualMachine) response() interface{} {
	return new(ChangeServiceForVirtualMachineResponse)
}

// ChangeServiceForVirtualMachineResponse represents an changed VM instance
type ChangeServiceForVirtualMachineResponse VirtualMachineResponse

// ResetPasswordForVirtualMachine (Async) represents the scaling of a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/resetPasswordForVirtualMachine.html
type ResetPasswordForVirtualMachine ScaleVirtualMachine

func (*ResetPasswordForVirtualMachine) name() string {
	return "resetPasswordForVirtualMachine"
}

func (*ResetPasswordForVirtualMachine) asyncResponse() interface{} {
	return new(ResetPasswordForVirtualMachineResponse)
}

// ResetPasswordForVirtualMachineResponse represents the updated vm
type ResetPasswordForVirtualMachineResponse VirtualMachineResponse

// GetVMPassword asks for an encrypted password
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/getVMPassword.html
type GetVMPassword struct {
	ID string `json:"id"`
}

func (*GetVMPassword) name() string {
	return "getVMPassword"
}

func (*GetVMPassword) response() interface{} {
	return new(GetVMPasswordResponse)
}

// GetVMPasswordResponse represents the encrypted password
type GetVMPasswordResponse struct {
	// Base64 encrypted password for the VM
	EncryptedPassword string `json:"encryptedpassword"`
}

// ListVirtualMachines represents a search for a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listVirtualMachine.html
type ListVirtualMachines struct {
	Account           string            `json:"account,omitempty"`
	AffinityGroupID   string            `json:"affinitygroupid,omitempty"`
	Details           map[string]string `json:"details,omitempty"`
	DisplayVM         bool              `json:"displayvm,omitempty"` // root only
	DomainID          string            `json:"domainid,omitempty"`
	ForVirtualNetwork bool              `json:"forvirtualnetwork,omitempty"`
	GroupID           string            `json:"groupid,omitempty"`
	HostID            string            `json:"hostid,omitempty"`
	Hypervisor        string            `json:"hypervisor,omitempty"`
	ID                string            `json:"id,omitempty"`
	IDs               []string          `json:"ids,omitempty"` // mutually exclusive with id
	IsoID             string            `json:"isoid,omitempty"`
	IsRecursive       bool              `json:"isrecursive,omitempty"`
	KeyPair           string            `json:"keypair,omitempty"`
	Keyword           string            `json:"keyword,omitempty"`
	ListAll           bool              `json:"listall,omitempty"`
	Name              string            `json:"name,omitempty"`
	NetworkID         string            `json:"networkid,omitempty"`
	Page              int               `json:"page,omitempty"`
	PageSize          int               `json:"pagesize,omitempty"`
	PodID             string            `json:"podid,omitempty"`
	ProjectID         string            `json:"projectid,omitempty"`
	ServiceOfferindID string            `json:"serviceofferingid,omitempty"`
	State             string            `json:"state,omitempty"` // Running, Stopped, Present, ...
	StorageID         string            `json:"storageid,omitempty"`
	Tags              []ResourceTag     `json:"tags,omitempty"`
	TemplateID        string            `json:"templateid,omitempty"`
	UserID            string            `json:"userid,omitempty"`
	VpcID             string            `json:"vpcid,omitempty"`
	ZoneID            string            `json:"zoneid,omitempty"`
}

func (*ListVirtualMachines) name() string {
	return "listVirtualMachines"
}

func (*ListVirtualMachines) response() interface{} {
	return new(ListVirtualMachinesResponse)
}

// ListVirtualMachinesResponse represents a list of virtual machines
type ListVirtualMachinesResponse struct {
	Count          int              `json:"count"`
	VirtualMachine []VirtualMachine `json:"virtualmachine"`
}

// AddNicToVirtualMachine (Async) adds a NIC to a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/addNicToVirtualMachine.html
type AddNicToVirtualMachine struct {
	NetworkID        string `json:"networkid"`
	VirtualMachineID string `json:"virtualmachineid"`
	IPAddress        net.IP `json:"ipaddress,omitempty"`
}

func (*AddNicToVirtualMachine) name() string {
	return "addNicToVirtualMachine"
}

func (*AddNicToVirtualMachine) asyncResponse() interface{} {
	return new(AddNicToVirtualMachineResponse)
}

// AddNicToVirtualMachineResponse represents the modified VM
type AddNicToVirtualMachineResponse VirtualMachineResponse

// RemoveNicFromVirtualMachine (Async) removes a NIC from a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/removeNicFromVirtualMachine.html
type RemoveNicFromVirtualMachine struct {
	NicID            string `json:"nicid"`
	VirtualMachineID string `json:"virtualmachineid"`
}

func (*RemoveNicFromVirtualMachine) name() string {
	return "removeNicFromVirtualMachine"
}

func (*RemoveNicFromVirtualMachine) asyncResponse() interface{} {
	return new(RemoveNicFromVirtualMachineResponse)
}

// RemoveNicFromVirtualMachineResponse represents the modified VM
type RemoveNicFromVirtualMachineResponse VirtualMachineResponse

// UpdateDefaultNicForVirtualMachine (Async) adds a NIC to a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateDefaultNicForVirtualMachine.html
type UpdateDefaultNicForVirtualMachine struct {
	NetworkID        string `json:"networkid"`
	VirtualMachineID string `json:"virtualmachineid"`
	IPAddress        net.IP `json:"ipaddress,omitempty"`
}

func (*UpdateDefaultNicForVirtualMachine) name() string {
	return "updateDefaultNicForVirtualMachine"
}

func (*UpdateDefaultNicForVirtualMachine) asyncResponse() interface{} {
	return new(UpdateDefaultNicForVirtualMachineResponse)
}

// UpdateDefaultNicForVirtualMachineResponse represents the modified VM
type UpdateDefaultNicForVirtualMachineResponse VirtualMachineResponse
