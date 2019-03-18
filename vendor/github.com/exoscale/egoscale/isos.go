package egoscale

// ISO represents an attachable ISO disc
type ISO Template

// ResourceType returns the type of the resource
func (ISO) ResourceType() string {
	return "ISO"
}

// ListRequest produces the ListIsos command.
func (iso ISO) ListRequest() (ListCommand, error) {
	req := &ListISOs{
		ID:     iso.ID,
		Name:   iso.Name,
		ZoneID: iso.ZoneID,
	}
	if iso.Bootable {
		*req.Bootable = true
	}
	if iso.IsFeatured {
		req.IsoFilter = "featured"
	}
	if iso.IsPublic {
		*req.IsPublic = true
	}
	if iso.IsReady {
		*req.IsReady = true
	}

	for i := range iso.Tags {
		req.Tags = append(req.Tags, iso.Tags[i])
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListISOs

// ListISOs represents the list all available ISO files request
type ListISOs struct {
	_           bool          `name:"listIsos" description:"Lists all available ISO files."`
	Bootable    *bool         `json:"bootable,omitempty" doc:"True if the ISO is bootable, false otherwise"`
	ID          *UUID         `json:"id,omitempty" doc:"List ISO by id"`
	IsoFilter   string        `json:"isofilter,omitempty" doc:"Possible values are \"featured\", \"self\", \"selfexecutable\",\"sharedexecutable\",\"executable\", and \"community\". * featured : templates that have been marked as featured and public. * self : templates that have been registered or created by the calling user. * selfexecutable : same as self, but only returns templates that can be used to deploy a new VM. * sharedexecutable : templates ready to be deployed that have been granted to the calling user by another user. * executable : templates that are owned by the calling user, or public templates, that can be used to deploy a VM. * community : templates that have been marked as public but not featured. * all : all templates (only usable by admins)."`
	IsPublic    *bool         `json:"ispublic,omitempty" doc:"True if the ISO is publicly available to all users, false otherwise."`
	IsReady     *bool         `json:"isready,omitempty" doc:"True if this ISO is ready to be deployed"`
	Keyword     string        `json:"keyword,omitempty" doc:"List by keyword"`
	Name        string        `json:"name,omitempty" doc:"List all isos by name"`
	Page        int           `json:"page,omitempty"`
	PageSize    int           `json:"pagesize,omitempty"`
	ShowRemoved *bool         `json:"showremoved,omitempty" doc:"Show removed ISOs as well"`
	Tags        []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	ZoneID      *UUID         `json:"zoneid,omitempty" doc:"The ID of the zone"`
}

// ListISOsResponse represents a list of ISO files
type ListISOsResponse struct {
	Count int   `json:"count"`
	ISO   []ISO `json:"iso"`
}

// AttachISO represents the request to attach an ISO to a virtual machine.
type AttachISO struct {
	_                bool  `name:"attachIso" description:"Attaches an ISO to a virtual machine."`
	ID               *UUID `json:"id" doc:"the ID of the ISO file"`
	VirtualMachineID *UUID `json:"virtualmachineid" doc:"the ID of the virtual machine"`
}

// Response returns the struct to unmarshal
func (AttachISO) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (AttachISO) AsyncResponse() interface{} {
	return new(VirtualMachine)
}

// DetachISO represents the request to detach an ISO to a virtual machine.
type DetachISO struct {
	_                bool  `name:"detachIso" description:"Detaches any ISO file (if any) currently attached to a virtual machine."`
	VirtualMachineID *UUID `json:"virtualmachineid" doc:"The ID of the virtual machine"`
}

// Response returns the struct to unmarshal
func (DetachISO) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (DetachISO) AsyncResponse() interface{} {
	return new(VirtualMachine)
}
