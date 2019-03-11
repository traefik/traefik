package egoscale

// Volume represents a volume linked to a VM
type Volume struct {
	Account                    string        `json:"account,omitempty" doc:"the account associated with the disk volume"`
	Attached                   string        `json:"attached,omitempty" doc:"the date the volume was attached to a VM instance"`
	ChainInfo                  string        `json:"chaininfo,omitempty" doc:"the chain info of the volume"`
	ClusterID                  *UUID         `json:"clusterid,omitempty" doc:"ID of the cluster"`
	ClusterName                string        `json:"clustername,omitempty" doc:"name of the cluster"`
	Created                    string        `json:"created,omitempty" doc:"the date the disk volume was created"`
	Destroyed                  bool          `json:"destroyed,omitempty" doc:"the boolean state of whether the volume is destroyed or not"`
	DeviceID                   int64         `json:"deviceid,omitempty" doc:"the ID of the device on user vm the volume is attahed to. This tag is not returned when the volume is detached."`
	DiskBytesReadRate          int64         `json:"diskBytesReadRate,omitempty" doc:"bytes read rate of the disk volume"`
	DiskBytesWriteRate         int64         `json:"diskBytesWriteRate,omitempty" doc:"bytes write rate of the disk volume"`
	DiskIopsReadRate           int64         `json:"diskIopsReadRate,omitempty" doc:"io requests read rate of the disk volume"`
	DiskIopsWriteRate          int64         `json:"diskIopsWriteRate,omitempty" doc:"io requests write rate of the disk volume"`
	DiskOfferingDisplayText    string        `json:"diskofferingdisplaytext,omitempty" doc:"the display text of the disk offering"`
	DiskOfferingID             *UUID         `json:"diskofferingid,omitempty" doc:"ID of the disk offering"`
	DiskOfferingName           string        `json:"diskofferingname,omitempty" doc:"name of the disk offering"`
	DisplayVolume              bool          `json:"displayvolume,omitempty" doc:"an optional field whether to the display the volume to the end user or not."`
	Hypervisor                 string        `json:"hypervisor,omitempty" doc:"Hypervisor the volume belongs to"`
	ID                         *UUID         `json:"id,omitempty" doc:"ID of the disk volume"`
	IsExtractable              *bool         `json:"isextractable,omitempty" doc:"true if the volume is extractable, false otherwise"`
	IsoDisplayText             string        `json:"isodisplaytext,omitempty" doc:"an alternate display text of the ISO attached to the virtual machine"`
	IsoID                      *UUID         `json:"isoid,omitempty" doc:"the ID of the ISO attached to the virtual machine"`
	IsoName                    string        `json:"isoname,omitempty" doc:"the name of the ISO attached to the virtual machine"`
	MaxIops                    int64         `json:"maxiops,omitempty" doc:"max iops of the disk volume"`
	MinIops                    int64         `json:"miniops,omitempty" doc:"min iops of the disk volume"`
	Name                       string        `json:"name,omitempty" doc:"name of the disk volume"`
	Path                       string        `json:"path,omitempty" doc:"the path of the volume"`
	PodID                      *UUID         `json:"podid,omitempty" doc:"ID of the pod"`
	PodName                    string        `json:"podname,omitempty" doc:"name of the pod"`
	QuiesceVM                  bool          `json:"quiescevm,omitempty" doc:"need quiesce vm or not when taking snapshot"`
	ServiceOfferingDisplayText string        `json:"serviceofferingdisplaytext,omitempty" doc:"the display text of the service offering for root disk"`
	ServiceOfferingID          *UUID         `json:"serviceofferingid,omitempty" doc:"ID of the service offering for root disk"`
	ServiceOfferingName        string        `json:"serviceofferingname,omitempty" doc:"name of the service offering for root disk"`
	Size                       uint64        `json:"size,omitempty" doc:"size of the disk volume"`
	SnapshotID                 *UUID         `json:"snapshotid,omitempty" doc:"ID of the snapshot from which this volume was created"`
	State                      string        `json:"state,omitempty" doc:"the state of the disk volume"`
	Status                     string        `json:"status,omitempty" doc:"the status of the volume"`
	Storage                    string        `json:"storage,omitempty" doc:"name of the primary storage hosting the disk volume"`
	StorageID                  *UUID         `json:"storageid,omitempty" doc:"id of the primary storage hosting the disk volume; returned to admin user only"`
	StorageType                string        `json:"storagetype,omitempty" doc:"shared or local storage"`
	Tags                       []ResourceTag `json:"tags,omitempty" doc:"the list of resource tags associated with volume"`
	TemplateDisplayText        string        `json:"templatedisplaytext,omitempty" doc:"an alternate display text of the template for the virtual machine"`
	TemplateID                 *UUID         `json:"templateid,omitempty" doc:"the ID of the template for the virtual machine. A -1 is returned if the virtual machine was created from an ISO file."` // no *UUID because of the -1 thingy...
	TemplateName               string        `json:"templatename,omitempty" doc:"the name of the template for the virtual machine"`
	Type                       string        `json:"type,omitempty" doc:"type of the disk volume (ROOT or DATADISK)"`
	VirtualMachineID           *UUID         `json:"virtualmachineid,omitempty" doc:"id of the virtual machine"`
	VMDisplayName              string        `json:"vmdisplayname,omitempty" doc:"display name of the virtual machine"`
	VMName                     string        `json:"vmname,omitempty" doc:"name of the virtual machine"`
	VMState                    string        `json:"vmstate,omitempty" doc:"state of the virtual machine"`
	ZoneID                     *UUID         `json:"zoneid,omitempty" doc:"ID of the availability zone"`
	ZoneName                   string        `json:"zonename,omitempty" doc:"name of the availability zone"`
}

// ResourceType returns the type of the resource
func (Volume) ResourceType() string {
	return "Volume"
}

// ListRequest builds the ListVolumes request
func (vol Volume) ListRequest() (ListCommand, error) {
	req := &ListVolumes{
		Name:             vol.Name,
		Type:             vol.Type,
		VirtualMachineID: vol.VirtualMachineID,
		ZoneID:           vol.ZoneID,
	}

	return req, nil
}

// ResizeVolume (Async) resizes a volume
type ResizeVolume struct {
	ID             *UUID `json:"id" doc:"the ID of the disk volume"`
	DiskOfferingID *UUID `json:"diskofferingid,omitempty" doc:"new disk offering id"`
	Size           int64 `json:"size,omitempty" doc:"New volume size in G (must be larger than current size since shrinking the disk is not supported)"`
	_              bool  `name:"resizeVolume" description:"Resizes a volume"`
}

// Response returns the struct to unmarshal
func (ResizeVolume) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (ResizeVolume) AsyncResponse() interface{} {
	return new(Volume)
}

//go:generate go run generate/main.go -interface=Listable ListVolumes

// ListVolumes represents a query listing volumes
type ListVolumes struct {
	DiskOfferingID   *UUID         `json:"diskofferingid,omitempty" doc:"List volumes by disk offering"`
	ID               *UUID         `json:"id,omitempty" doc:"The ID of the disk volume"`
	Keyword          string        `json:"keyword,omitempty" doc:"List by keyword"`
	Name             string        `json:"name,omitempty" doc:"The name of the disk volume"`
	Page             int           `json:"page,omitempty"`
	PageSize         int           `json:"pagesize,omitempty"`
	Tags             []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	Type             string        `json:"type,omitempty" doc:"The type of disk volume"`
	VirtualMachineID *UUID         `json:"virtualmachineid,omitempty" doc:"The ID of the virtual machine"`
	ZoneID           *UUID         `json:"zoneid,omitempty" doc:"The ID of the availability zone"`
	_                bool          `name:"listVolumes" description:"Lists all volumes."`
}

// ListVolumesResponse represents a list of volumes
type ListVolumesResponse struct {
	Count  int      `json:"count"`
	Volume []Volume `json:"volume"`
}
