package egoscale

// ServiceOffering corresponds to the Compute Offerings
//
// A service offering correspond to some hardware features (CPU, RAM).
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/service_offerings.html
type ServiceOffering struct {
	Authorized                bool              `json:"authorized,omitempty" doc:"is the account/domain authorized to use this service offering"`
	CPUNumber                 int               `json:"cpunumber,omitempty" doc:"the number of CPU"`
	CPUSpeed                  int               `json:"cpuspeed,omitempty" doc:"the clock rate CPU speed in Mhz"`
	Created                   string            `json:"created,omitempty" doc:"the date this service offering was created"`
	DefaultUse                bool              `json:"defaultuse,omitempty" doc:"is this a  default system vm offering"`
	DeploymentPlanner         string            `json:"deploymentplanner,omitempty" doc:"deployment strategy used to deploy VM."`
	DiskBytesReadRate         int64             `json:"diskBytesReadRate,omitempty" doc:"bytes read rate of the service offering"`
	DiskBytesWriteRate        int64             `json:"diskBytesWriteRate,omitempty" doc:"bytes write rate of the service offering"`
	DiskIopsReadRate          int64             `json:"diskIopsReadRate,omitempty" doc:"io requests read rate of the service offering"`
	DiskIopsWriteRate         int64             `json:"diskIopsWriteRate,omitempty" doc:"io requests write rate of the service offering"`
	Displaytext               string            `json:"displaytext,omitempty" doc:"an alternate display text of the service offering."`
	HostTags                  string            `json:"hosttags,omitempty" doc:"the host tag for the service offering"`
	HypervisorSnapshotReserve int               `json:"hypervisorsnapshotreserve,omitempty" doc:"Hypervisor snapshot reserve space as a percent of a volume (for managed storage using Xen or VMware)"`
	ID                        *UUID             `json:"id" doc:"the id of the service offering"`
	IsCustomized              bool              `json:"iscustomized,omitempty" doc:"is true if the offering is customized"`
	IsCustomizedIops          bool              `json:"iscustomizediops,omitempty" doc:"true if disk offering uses custom iops, false otherwise"`
	IsSystem                  bool              `json:"issystem,omitempty" doc:"is this a system vm offering"`
	IsVolatile                bool              `json:"isvolatile,omitempty" doc:"true if the vm needs to be volatile, i.e., on every reboot of vm from API root disk is discarded and creates a new root disk"`
	LimitCPUUse               bool              `json:"limitcpuuse,omitempty" doc:"restrict the CPU usage to committed service offering"`
	MaxIops                   int64             `json:"maxiops,omitempty" doc:"the max iops of the disk offering"`
	Memory                    int               `json:"memory,omitempty" doc:"the memory in MB"`
	MinIops                   int64             `json:"miniops,omitempty" doc:"the min iops of the disk offering"`
	Name                      string            `json:"name,omitempty" doc:"the name of the service offering"`
	NetworkRate               int               `json:"networkrate,omitempty" doc:"data transfer rate in megabits per second allowed."`
	OfferHA                   bool              `json:"offerha,omitempty" doc:"the ha support in the service offering"`
	Restricted                bool              `json:"restricted,omitempty" doc:"is this offering restricted"`
	ServiceOfferingDetails    map[string]string `json:"serviceofferingdetails,omitempty" doc:"additional key/value details tied with this service offering"`
	StorageType               string            `json:"storagetype,omitempty" doc:"the storage type for this service offering"`
	SystemVMType              string            `json:"systemvmtype,omitempty" doc:"is this a the systemvm type for system vm offering"`
	Tags                      string            `json:"tags,omitempty" doc:"the tags for the service offering"`
}

// ListRequest builds the ListSecurityGroups request
func (so ServiceOffering) ListRequest() (ListCommand, error) {
	// Restricted cannot be applied here because it really has three states
	req := &ListServiceOfferings{
		ID:           so.ID,
		Name:         so.Name,
		SystemVMType: so.SystemVMType,
	}

	if so.IsSystem {
		req.IsSystem = &so.IsSystem
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListServiceOfferings

// ListServiceOfferings represents a query for service offerings
type ListServiceOfferings struct {
	ID               *UUID  `json:"id,omitempty" doc:"ID of the service offering"`
	IsSystem         *bool  `json:"issystem,omitempty" doc:"is this a system vm offering"`
	Keyword          string `json:"keyword,omitempty" doc:"List by keyword"`
	Name             string `json:"name,omitempty" doc:"name of the service offering"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	Restricted       *bool  `json:"restricted,omitempty" doc:"filter by the restriction flag: true to list only the restricted service offerings, false to list non-restricted service offerings, or nothing for all."`
	SystemVMType     string `json:"systemvmtype,omitempty" doc:"the system VM type. Possible types are \"consoleproxy\", \"secondarystoragevm\" or \"domainrouter\"."`
	VirtualMachineID *UUID  `json:"virtualmachineid,omitempty" doc:"the ID of the virtual machine. Pass this in if you want to see the available service offering that a virtual machine can be changed to."`
	_                bool   `name:"listServiceOfferings" description:"Lists all available service offerings."`
}

// ListServiceOfferingsResponse represents a list of service offerings
type ListServiceOfferingsResponse struct {
	Count           int               `json:"count"`
	ServiceOffering []ServiceOffering `json:"serviceoffering"`
}
