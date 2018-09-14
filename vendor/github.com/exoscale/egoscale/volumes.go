package egoscale

import (
	"fmt"
)

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
	Domain                     string        `json:"domain,omitempty" doc:"the domain associated with the disk volume"`
	DomainID                   *UUID         `json:"domainid,omitempty" doc:"the ID of the domain associated with the disk volume"`
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
	TemplateID                 string        `json:"templateid,omitempty" doc:"the ID of the template for the virtual machine. A -1 is returned if the virtual machine was created from an ISO file."` // no *UUID because of the -1 thingy...
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
		Account:          vol.Account,
		DomainID:         vol.DomainID,
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
	ShrinkOk       *bool `json:"shrinkok,omitempty" doc:"Verify OK to Shrink"`
	Size           int64 `json:"size,omitempty" doc:"New volume size in G"`
	_              bool  `name:"resizeVolume" description:"Resizes a volume"`
}

func (ResizeVolume) response() interface{} {
	return new(AsyncJobResult)
}

func (ResizeVolume) asyncResponse() interface{} {
	return new(Volume)
}

// ListVolumes represents a query listing volumes
type ListVolumes struct {
	Account          string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DiskOfferingID   *UUID         `json:"diskofferingid,omitempty" doc:"list volumes by disk offering"`
	DisplayVolume    *bool         `json:"displayvolume,omitempty" doc:"list resources by display flag; only ROOT admin is eligible to pass this parameter"`
	DomainID         *UUID         `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	HostID           *UUID         `json:"hostid,omitempty" doc:"list volumes on specified host"`
	ID               *UUID         `json:"id,omitempty" doc:"the ID of the disk volume"`
	IsRecursive      *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword          string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll          *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name             string        `json:"name,omitempty" doc:"the name of the disk volume"`
	Page             int           `json:"page,omitempty"`
	PageSize         int           `json:"pagesize,omitempty"`
	PodID            *UUID         `json:"podid,omitempty" doc:"the pod id the disk volume belongs to"`
	StorageID        *UUID         `json:"storageid,omitempty" doc:"the ID of the storage pool, available to ROOT admin only"`
	Tags             []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	Type             string        `json:"type,omitempty" doc:"the type of disk volume"`
	VirtualMachineID *UUID         `json:"virtualmachineid,omitempty" doc:"the ID of the virtual machine"`
	ZoneID           *UUID         `json:"zoneid,omitempty" doc:"the ID of the availability zone"`
	_                bool          `name:"listVolumes" description:"Lists all volumes."`
}

// ListVolumesResponse represents a list of volumes
type ListVolumesResponse struct {
	Count  int      `json:"count"`
	Volume []Volume `json:"volume"`
}

func (ListVolumes) response() interface{} {
	return new(ListVolumesResponse)
}

// SetPage sets the current page
func (ls *ListVolumes) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVolumes) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (ListVolumes) each(resp interface{}, callback IterateItemFunc) {
	volumes, ok := resp.(*ListVolumesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListVolumesResponse expected, got %T", resp))
		return
	}

	for i := range volumes.Volume {
		if !callback(&volumes.Volume[i], nil) {
			break
		}
	}
}
