package egoscale

import (
	"net"
)

// Host represents the Hypervisor
type Host struct {
	Capabilities             string      `json:"capabilities,omitempty" doc:"capabilities of the host"`
	ClusterID                *UUID       `json:"clusterid,omitempty" doc:"the cluster ID of the host"`
	ClusterName              string      `json:"clustername,omitempty" doc:"the cluster name of the host"`
	ClusterType              string      `json:"clustertype,omitempty" doc:"the cluster type of the cluster that host belongs to"`
	CPUAllocated             int64       `json:"cpuallocated,omitempty" doc:"the amount of the host's CPU currently allocated"`
	CPUNumber                int         `json:"cpunumber,omitempty" doc:"the CPU number of the host"`
	CPUSockets               int         `json:"cpusockets,omitempty" doc:"the number of CPU sockets on the host"`
	CPUSpeed                 int64       `json:"cpuspeed,omitempty" doc:"the CPU speed of the host"`
	CPUUsed                  int64       `json:"cpuused,omitempty" doc:"the amount of the host's CPU currently used"`
	CPUWithOverProvisioning  int64       `json:"cpuwithoverprovisioning,omitempty" doc:"the amount of the host's CPU after applying the cpu.overprovisioning.factor"`
	Created                  string      `json:"created,omitempty" doc:"the date and time the host was created"`
	Disconnected             string      `json:"disconnected,omitempty" doc:"true if the host is disconnected. False otherwise."`
	DiskSizeAllocated        int64       `json:"disksizeallocated,omitempty" doc:"the host's or host storage pool's currently allocated disk size"`
	DiskSizeTotal            int64       `json:"disksizetotal,omitempty" doc:"the total disk size of the host or host storage pool"`
	DiskSizeUsed             int64       `json:"disksizeused,omitempty" doc:"the host's or host storage pool's currently used disk size"`
	DiskWithOverProvisioning int64       `json:"diskwithoverprovisioning,omitempty" doc:"the total disk size of the host or host storage pool with over provisioning factor"`
	Events                   string      `json:"events,omitempty" doc:"events available for the host"`
	HAHost                   *bool       `json:"hahost,omitempty" doc:"true if the host is Ha host (dedicated to vms started by HA process; false otherwise"`
	HostTags                 string      `json:"hosttags,omitempty" doc:"comma-separated list of tags for the host"`
	Hypervisor               string      `json:"hypervisor,omitempty" doc:"the host hypervisor"`
	HypervisorVersion        string      `json:"hypervisorversion,omitempty" doc:"the hypervisor version"`
	ID                       *UUID       `json:"id,omitempty" doc:"the ID of the host"`
	IPAddress                net.IP      `json:"ipaddress,omitempty" doc:"the IP address of the host"`
	IsLocalstorageActive     *bool       `json:"islocalstorageactive,omitempty" doc:"true if local storage is active, false otherwise"`
	LastPinged               string      `json:"lastpinged,omitempty" doc:"the date and time the host was last pinged"`
	ManagementServerID       *UUID       `json:"managementserverid,omitempty" doc:"the management server ID of the host"`
	MemoryAllocated          int64       `json:"memoryallocated,omitempty" doc:"the amount of VM's memory allocated onto the host"`
	MemoryPhysical           int64       `json:"memoryphysical,omitempty" doc:"the total physical memory of the host"`
	MemoryReserved           int64       `json:"memoryreserved,omitempty" doc:"the amount of the host's memory reserved"`
	MemoryTotal              int64       `json:"memorytotal,omitempty" doc:"the total memory of the host available (must be physical - reserved)"`
	MemoryUsed               int64       `json:"memoryused,omitempty" doc:"the amount of the host's memory used by running vm"`
	Name                     string      `json:"name,omitempty" doc:"the name of the host"`
	NetworkKbsRead           int64       `json:"networkkbsread,omitempty" doc:"the incoming network traffic on the host"`
	NetworkKbsWrite          int64       `json:"networkkbswrite,omitempty" doc:"the outgoing network traffic on the host"`
	OSCategoryID             *UUID       `json:"oscategoryid,omitempty" doc:"the OS category ID of the host"`
	OSCategoryName           string      `json:"oscategoryname,omitempty" doc:"the OS category name of the host"`
	PCIDevices               []PCIDevice `json:"pcidevices,omitempty" doc:"PCI cards present in the host"`
	PodID                    *UUID       `json:"podid,omitempty" doc:"the Pod ID of the host"`
	PodName                  string      `json:"podname,omitempty" doc:"the Pod name of the host"`
	Removed                  string      `json:"removed,omitempty" doc:"the date and time the host was removed"`
	ResourceState            string      `json:"resourcestate,omitempty" doc:"the resource state of the host"`
	State                    string      `json:"state,omitempty" doc:"the state of the host"`
	StorageID                *UUID       `json:"storageid,omitempty" doc:"the host's storage pool id"`
	Type                     string      `json:"type,omitempty" doc:"the host type"`
	Version                  string      `json:"version,omitempty" doc:"the host version"`
	ZoneID                   *UUID       `json:"zoneid,omitempty" doc:"the Zone ID of the host"`
	ZoneName                 string      `json:"zonename,omitempty" doc:"the Zone name of the host"`
}

// ListHosts lists hosts
type ListHosts struct {
	ClusterID     *UUID    `json:"clusterid,omitempty" doc:"lists hosts existing in particular cluster"`
	Details       []string `json:"details,omitempty" doc:"comma separated list of host details requested, value can be a list of [ min, all, capacity, events, stats]"`
	HAHost        *bool    `json:"hahost,omitempty" doc:"if true, list only hosts dedicated to HA"`
	Hypervisor    string   `json:"hypervisor,omitempty" doc:"hypervisor type of host: KVM,Simulator"`
	ID            *UUID    `json:"id,omitempty" doc:"the id of the host"`
	Keyword       string   `json:"keyword,omitempty" doc:"List by keyword"`
	Name          string   `json:"name,omitempty" doc:"the name of the host"`
	Page          int      `json:"page,omitempty"`
	PageSize      int      `json:"pagesize,omitempty"`
	PodID         *UUID    `json:"podid,omitempty" doc:"the Pod ID for the host"`
	ResourceState string   `json:"resourcestate,omitempty" doc:"list hosts by resource state. Resource state represents current state determined by admin of host, value can be one of [Enabled, Disabled, Unmanaged, PrepareForMaintenance, ErrorInMaintenance, Maintenance, Error]"`
	State         string   `json:"state,omitempty" doc:"the state of the host"`
	Type          string   `json:"type,omitempty" doc:"the host type"`
	ZoneID        *UUID    `json:"zoneid,omitempty" doc:"the Zone ID for the host"`
	_             bool     `name:"listHosts" description:"Lists hosts."`
}

func (ListHosts) response() interface{} {
	return new(ListHostsResponse)
}

// ListHostsResponse represents a list of hosts
type ListHostsResponse struct {
	Count int    `json:"count"`
	Host  []Host `json:"host"`
}

// UpdateHost changes the resources state of a host
type UpdateHost struct {
	Allocationstate string   `json:"allocationstate,omitempty" doc:"Change resource state of host, valid values are [Enable, Disable]. Operation may failed if host in states not allowing Enable/Disable"`
	HostTags        []string `json:"hosttags,omitempty" doc:"list of tags to be added to the host"`
	ID              *UUID    `json:"id" doc:"the ID of the host to update"`
	OSCategoryID    *UUID    `json:"oscategoryid,omitempty" doc:"the id of Os category to update the host with"`
	URL             string   `json:"url,omitempty" doc:"the new uri for the secondary storage: nfs://host/path"`
	_               bool     `name:"updateHost" description:"Updates a host."`
}

func (UpdateHost) response() interface{} {
	return new(Host)
}
