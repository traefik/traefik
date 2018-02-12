package egoscale

// https://github.com/apache/cloudstack/blob/master/api/src/main/java/com/cloud/configuration/Resource.java

// ResourceTypeName represents the name of a resource type (for limits)
type ResourceTypeName string

const (
	// VirtualMachineTypeName is the resource type name of a VM
	VirtualMachineTypeName ResourceTypeName = "user_vm"
	// IPAddressTypeName is the resource type name of an IP address
	IPAddressTypeName = "public_ip"
	// VolumeTypeName is the resource type name of a volume
	VolumeTypeName = "volume"
	// SnapshotTypeName is the resource type name of a snapshot
	SnapshotTypeName = "snapshot"
	// TemplateTypeName is the resource type name of a template
	TemplateTypeName = "template"
	// ProjectTypeName is the resource type name of a project
	ProjectTypeName = "project"
	// NetworkTypeName is the resource type name of a network
	NetworkTypeName = "network"
	// VPCTypeName is the resource type name of a VPC
	VPCTypeName = "vpc"
	// CPUTypeName is the resource type name of a CPU
	CPUTypeName = "cpu"
	// MemoryTypeName is the resource type name of Memory
	MemoryTypeName = "memory"
	// PrimaryStorageTypeName is the resource type name of primary storage
	PrimaryStorageTypeName = "primary_storage"
	// SecondaryStorageTypeName is the resource type name of secondary storage
	SecondaryStorageTypeName = "secondary_storage"
)

// ResourceType represents the ID of a resource type (for limits)
type ResourceType int64

const (
	// VirtualMachineType is the resource type ID of a VM
	VirtualMachineType ResourceType = iota
	// IPAddressType is the resource type ID of an IP address
	IPAddressType
	// VolumeType is the resource type ID of a volume
	VolumeType
	// SnapshotType is the resource type ID of a snapshot
	SnapshotType
	// TemplateType is the resource type ID of a template
	TemplateType
	// ProjectType is the resource type ID of a project
	ProjectType
	// NetworkType is the resource type ID of a network
	NetworkType
	// VPCType is the resource type ID of a VPC
	VPCType
	// CPUType is the resource type ID of a CPU
	CPUType
	// MemoryType is the resource type ID of Memory
	MemoryType
	// PrimaryStorageType is the resource type ID of primary storage
	PrimaryStorageType
	// SecondaryStorageType is the resource type ID of secondary storage
	SecondaryStorageType
)

// ListResourceLimits lists the resource limits
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listResourceLimits.html
type ListResourceLimits struct {
	Account          string           `json:"account,omittempty"`
	DomainID         string           `json:"domainid,omitempty"`
	ID               string           `json:"id,omitempty"`
	IsRecursive      bool             `json:"isrecursive,omitempty"`
	Keyword          string           `json:"keyword,omitempty"`
	ListAll          bool             `json:"listall,omitempty"`
	Page             int              `json:"page,omitempty"`
	PageSize         int              `json:"pagesize,omitempty"`
	ProjectID        string           `json:"projectid,omitempty"`
	ResourceType     ResourceType     `json:"resourcetype,omitempty"`
	ResourceTypeName ResourceTypeName `json:"resourcetypename,omitempty"`
}

func (*ListResourceLimits) name() string {
	return "listResourceLimits"
}

func (*ListResourceLimits) response() interface{} {
	return new(ListResourceLimitsResponse)
}

// ListResourceLimitsResponse represents a list of resource limits
type ListResourceLimitsResponse struct {
	Count         int             `json:"count"`
	ResourceLimit []ResourceLimit `json:"resourcelimit"`
}

// ResourceLimit represents the limit on a particular resource
type ResourceLimit struct {
	Account          string           `json:"account,omitempty"`
	Domain           string           `json:"domain,omitempty"`
	DomainID         string           `json:"domainid,omitempty"`
	Max              int64            `json:"max,omitempty"` // -1 means the sky is the limit
	Project          string           `json:"project,omitempty"`
	ProjectID        string           `json:"projectid,omitempty"`
	ResourceType     ResourceType     `json:"resourcetype,omitempty"`
	ResourceTypeName ResourceTypeName `json:"resourcetypename,omitempty"`
}
