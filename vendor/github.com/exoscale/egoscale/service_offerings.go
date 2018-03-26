package egoscale

// ServiceOffering corresponds to the Compute Offerings
type ServiceOffering struct {
	ID                        string            `json:"id"`
	CPUNumber                 int               `json:"cpunumber"`
	CPUSpeed                  int               `json:"cpuspeed"`
	Created                   string            `json:"created"`
	DefaultUse                bool              `json:"defaultuse,omitempty"`
	DeploymentPlanner         string            `json:"deploymentplanner,omitempty"`
	DiskBytesReadRate         int64             `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate        int64             `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate          int64             `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate         int64             `json:"diskIopsWriteRate,omitempty"`
	DisplayText               string            `json:"displaytext,omitempty"`
	Domain                    string            `json:"domain"`
	DomainID                  string            `json:"domainid"`
	HostTags                  string            `json:"hosttags,omitempty"`
	HypervisorSnapshotReserve int               `json:"hypervisorsnapshotreserve,omitempty"`
	IsCustomized              bool              `json:"iscustomized,omitempty"`
	IsCustomizedIops          bool              `json:"iscustomizediops,omitempty"`
	IsSystem                  bool              `json:"issystem,omitempty"`
	IsVolatile                bool              `json:"isvolatile,omitempty"`
	LimitCPUUse               bool              `json:"limitcpuuse,omitempty"`
	MaxIops                   int64             `json:"maxiops,omitempty"`
	Memory                    int               `json:"memory,omitempty"`
	MinIops                   int64             `json:"miniops,omitempty"`
	Name                      string            `json:"name,omitempty"`
	NetworkRate               int               `json:"networkrate,omitempty"`
	OfferHA                   bool              `json:"offerha,omitempty"`
	ServiceOfferingDetails    map[string]string `json:"serviceofferingdetails,omitempty"`
	StorageType               string            `json:"storagetype,omitempty"`
	SystemVMType              string            `json:"systemvmtype,omitempty"`
	Tags                      []ResourceTag     `json:"tags,omitempty"`
}

// ListServiceOfferings represents a query for service offerings
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listServiceOfferings.html
type ListServiceOfferings struct {
	DomainID         string `json:"domainid,omitempty"`
	ID               string `json:"id,omitempty"`
	IsRecursive      bool   `json:"isrecursive,omitempty"`
	IsSystem         bool   `json:"issystem,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	Name             string `json:"name,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	SystemVMType     string `json:"systemvmtype,omitempty"` // consoleproxy, secondarystoragevm, or domainrouter
	VirtualMachineID string `json:"virtualmachineid,omitempty"`
}

func (*ListServiceOfferings) name() string {
	return "listServiceOfferings"
}

func (*ListServiceOfferings) response() interface{} {
	return new(ListServiceOfferingsResponse)
}

// ListServiceOfferingsResponse represents a list of service offerings
type ListServiceOfferingsResponse struct {
	Count           int               `json:"count"`
	ServiceOffering []ServiceOffering `json:"serviceoffering"`
}
