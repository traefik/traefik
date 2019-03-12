package egoscale

// https://github.com/apache/cloudstack/blob/master/api/src/main/java/com/cloud/configuration/Resource.java

// ResourceTypeName represents the name of a resource type (for limits)
type ResourceTypeName string

const (
	// VirtualMachineTypeName is the resource type name of a VM
	VirtualMachineTypeName ResourceTypeName = "user_vm"
	// IPAddressTypeName is the resource type name of an IP address
	IPAddressTypeName ResourceTypeName = "public_ip"
	// VolumeTypeName is the resource type name of a volume
	VolumeTypeName ResourceTypeName = "volume"
	// SnapshotTypeName is the resource type name of a snapshot
	SnapshotTypeName ResourceTypeName = "snapshot"
	// TemplateTypeName is the resource type name of a template
	TemplateTypeName ResourceTypeName = "template"
	// ProjectTypeName is the resource type name of a project
	ProjectTypeName ResourceTypeName = "project"
	// NetworkTypeName is the resource type name of a network
	NetworkTypeName ResourceTypeName = "network"
	// VPCTypeName is the resource type name of a VPC
	VPCTypeName ResourceTypeName = "vpc"
	// CPUTypeName is the resource type name of a CPU
	CPUTypeName ResourceTypeName = "cpu"
	// MemoryTypeName is the resource type name of Memory
	MemoryTypeName ResourceTypeName = "memory"
	// PrimaryStorageTypeName is the resource type name of primary storage
	PrimaryStorageTypeName ResourceTypeName = "primary_storage"
	// SecondaryStorageTypeName is the resource type name of secondary storage
	SecondaryStorageTypeName ResourceTypeName = "secondary_storage"
)

// ResourceType represents the ID of a resource type (for limits)
type ResourceType string

const (
	// VirtualMachineType is the resource type ID of a VM
	VirtualMachineType ResourceType = "0"
	// IPAddressType is the resource type ID of an IP address
	IPAddressType ResourceType = "1"
	// VolumeType is the resource type ID of a volume
	VolumeType ResourceType = "2"
	// SnapshotType is the resource type ID of a snapshot
	SnapshotType ResourceType = "3"
	// TemplateType is the resource type ID of a template
	TemplateType ResourceType = "4"
	// ProjectType is the resource type ID of a project
	ProjectType ResourceType = "5"
	// NetworkType is the resource type ID of a network
	NetworkType ResourceType = "6"
	// VPCType is the resource type ID of a VPC
	VPCType ResourceType = "7"
	// CPUType is the resource type ID of a CPU
	CPUType ResourceType = "8"
	// MemoryType is the resource type ID of Memory
	MemoryType ResourceType = "9"
	// PrimaryStorageType is the resource type ID of primary storage
	PrimaryStorageType ResourceType = "10"
	// SecondaryStorageType is the resource type ID of secondary storage
	SecondaryStorageType ResourceType = "11"
)

// ResourceLimit represents the limit on a particular resource
type ResourceLimit struct {
	Max              int64        `json:"max,omitempty" doc:"the maximum number of the resource. A -1 means the resource currently has no limit."`
	ResourceType     ResourceType `json:"resourcetype,omitempty" doc:"resource type. Values include 0, 1, 2, 3, 4, 6, 7, 8, 9, 10, 11. See the resourceType parameter for more information on these values."`
	ResourceTypeName string       `json:"resourcetypename,omitempty" doc:"resource type name. Values include user_vm, public_ip, volume, snapshot, template, network, cpu, memory, primary_storage, secondary_storage."`
}

// ListRequest builds the ListResourceLimits request
func (limit ResourceLimit) ListRequest() (ListCommand, error) {
	req := &ListResourceLimits{
		ResourceType:     limit.ResourceType,
		ResourceTypeName: limit.ResourceTypeName,
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListResourceLimits

// ListResourceLimits lists the resource limits
type ListResourceLimits struct {
	ID               int64        `json:"id,omitempty" doc:"Lists resource limits by ID."`
	Keyword          string       `json:"keyword,omitempty" doc:"List by keyword"`
	Page             int          `json:"page,omitempty"`
	PageSize         int          `json:"pagesize,omitempty"`
	ResourceType     ResourceType `json:"resourcetype,omitempty" doc:"Type of resource. Values are 0, 1, 2, 3, 4, 6, 8, 9, 10, 11, 12, and 13. 0 - Instance. Number of instances a user can create. 1 - IP. Number of public IP addresses an account can own. 2 - Volume. Number of disk volumes an account can own. 3 - Snapshot. Number of snapshots an account can own. 4 - Template. Number of templates an account can register/create. 6 - Network. Number of networks an account can own. 8 - CPU. Number of CPU an account can allocate for his resources. 9 - Memory. Amount of RAM an account can allocate for his resources. 10 - PrimaryStorage. Total primary storage space (in GiB) a user can use. 11 - SecondaryStorage. Total secondary storage space (in GiB) a user can use. 12 - Elastic IP. Number of public elastic IP addresses an account can own. 13 - SMTP. If the account is allowed SMTP outbound traffic."`
	ResourceTypeName string       `json:"resourcetypename,omitempty" doc:"Type of resource (wins over resourceType if both are provided). Values are: user_vm - Instance. Number of instances a user can create. public_ip - IP. Number of public IP addresses an account can own. volume - Volume. Number of disk volumes an account can own. snapshot - Snapshot. Number of snapshots an account can own. template - Template. Number of templates an account can register/create. network - Network. Number of networks an account can own. cpu - CPU. Number of CPU an account can allocate for his resources. memory - Memory. Amount of RAM an account can allocate for his resources. primary_storage - PrimaryStorage. Total primary storage space (in GiB) a user can use. secondary_storage - SecondaryStorage. Total secondary storage space (in GiB) a user can use. public_elastic_ip - IP. Number of public elastic IP addresses an account can own. smtp - SG. If the account is allowed SMTP outbound traffic."`

	_ bool `name:"listResourceLimits" description:"Lists resource limits."`
}

// ListResourceLimitsResponse represents a list of resource limits
type ListResourceLimitsResponse struct {
	Count         int             `json:"count"`
	ResourceLimit []ResourceLimit `json:"resourcelimit"`
}
