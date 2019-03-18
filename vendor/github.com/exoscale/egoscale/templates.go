package egoscale

// Template represents a machine to be deployed.
type Template struct {
	Account               string            `json:"account,omitempty" doc:"the account name to which the template belongs"`
	AccountID             *UUID             `json:"accountid,omitempty" doc:"the account id to which the template belongs"`
	Bootable              bool              `json:"bootable,omitempty" doc:"true if the ISO is bootable, false otherwise"`
	Checksum              string            `json:"checksum,omitempty" doc:"checksum of the template"`
	Created               string            `json:"created,omitempty" doc:"the date this template was created"`
	CrossZones            bool              `json:"crossZones,omitempty" doc:"true if the template is managed across all Zones, false otherwise"`
	Details               map[string]string `json:"details,omitempty" doc:"additional key/value details tied with template"`
	DisplayText           string            `json:"displaytext,omitempty" doc:"the template display text"`
	Format                string            `json:"format,omitempty" doc:"the format of the template."`
	HostID                *UUID             `json:"hostid,omitempty" doc:"the ID of the secondary storage host for the template"`
	HostName              string            `json:"hostname,omitempty" doc:"the name of the secondary storage host for the template"`
	ID                    *UUID             `json:"id,omitempty" doc:"the template ID"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty" doc:"true if template contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	IsExtractable         bool              `json:"isextractable,omitempty" doc:"true if the template is extractable, false otherwise"`
	IsFeatured            bool              `json:"isfeatured,omitempty" doc:"true if this template is a featured template, false otherwise"`
	IsPublic              bool              `json:"ispublic,omitempty" doc:"true if this template is a public template, false otherwise"`
	IsReady               bool              `json:"isready,omitempty" doc:"true if the template is ready to be deployed from, false otherwise."`
	Name                  string            `json:"name,omitempty" doc:"the template name"`
	OsTypeID              *UUID             `json:"ostypeid,omitempty" doc:"the ID of the OS type for this template."`
	OsTypeName            string            `json:"ostypename,omitempty" doc:"the name of the OS type for this template."`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty" doc:"true if the reset password feature is enabled, false otherwise"`
	Removed               string            `json:"removed,omitempty" doc:"the date this template was removed"`
	Size                  int64             `json:"size,omitempty" doc:"the size of the template"`
	SourceTemplateID      *UUID             `json:"sourcetemplateid,omitempty" doc:"the template ID of the parent template if present"`
	SSHKeyEnabled         bool              `json:"sshkeyenabled,omitempty" doc:"true if template is sshkey enabled, false otherwise"`
	Status                string            `json:"status,omitempty" doc:"the status of the template"`
	Tags                  []ResourceTag     `json:"tags,omitempty" doc:"the list of resource tags associated with tempate"`
	TemplateDirectory     string            `json:"templatedirectory,omitempty" doc:"Template directory"`
	TemplateTag           string            `json:"templatetag,omitempty" doc:"the tag of this template"`
	TemplateType          string            `json:"templatetype,omitempty" doc:"the type of the template"`
	URL                   string            `json:"url,omitempty" doc:"Original URL of the template where it was downloaded"`
	ZoneID                *UUID             `json:"zoneid,omitempty" doc:"the ID of the zone for this template"`
	ZoneName              string            `json:"zonename,omitempty" doc:"the name of the zone for this template"`
}

// ResourceType returns the type of the resource
func (Template) ResourceType() string {
	return "Template"
}

// ListRequest builds the ListTemplates request
func (template Template) ListRequest() (ListCommand, error) {
	req := &ListTemplates{
		ID:     template.ID,
		Name:   template.Name,
		ZoneID: template.ZoneID,
	}
	if template.IsFeatured {
		req.TemplateFilter = "featured"
	}
	if template.Removed != "" {
		*req.ShowRemoved = true
	}

	for i := range template.Tags {
		req.Tags = append(req.Tags, template.Tags[i])
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListTemplates

// ListTemplates represents a template query filter
type ListTemplates struct {
	TemplateFilter string        `json:"templatefilter" doc:"Possible values are \"featured\", \"self\", \"selfexecutable\",\"sharedexecutable\",\"executable\", and \"community\". * featured : templates that have been marked as featured and public. * self : templates that have been registered or created by the calling user. * selfexecutable : same as self, but only returns templates that can be used to deploy a new VM. * sharedexecutable : templates ready to be deployed that have been granted to the calling user by another user. * executable : templates that are owned by the calling user, or public templates, that can be used to deploy a VM. * community : templates that have been marked as public but not featured."`
	ID             *UUID         `json:"id,omitempty" doc:"the template ID"`
	Keyword        string        `json:"keyword,omitempty" doc:"List by keyword"`
	Name           string        `json:"name,omitempty" doc:"the template name"`
	Page           int           `json:"page,omitempty"`
	PageSize       int           `json:"pagesize,omitempty"`
	ShowRemoved    *bool         `json:"showremoved,omitempty" doc:"Show removed templates as well"`
	Tags           []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	ZoneID         *UUID         `json:"zoneid,omitempty" doc:"list templates by zoneid"`
	_              bool          `name:"listTemplates" description:"List all public, private, and privileged templates."`
}

// ListTemplatesResponse represents a list of templates
type ListTemplatesResponse struct {
	Count    int        `json:"count"`
	Template []Template `json:"template"`
}

// OSCategory represents an OS category
type OSCategory struct {
	ID   *UUID  `json:"id,omitempty" doc:"the ID of the OS category"`
	Name string `json:"name,omitempty" doc:"the name of the OS category"`
}

// ListRequest builds the ListOSCategories request
func (osCat OSCategory) ListRequest() (ListCommand, error) {
	req := &ListOSCategories{
		Name: osCat.Name,
		ID:   osCat.ID,
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListOSCategories

// ListOSCategories lists the OS categories
type ListOSCategories struct {
	ID       *UUID  `json:"id,omitempty" doc:"list Os category by id"`
	Keyword  string `json:"keyword,omitempty" doc:"List by keyword"`
	Name     string `json:"name,omitempty" doc:"list os category by name"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pagesize,omitempty"`
	_        bool   `name:"listOsCategories" description:"Lists all supported OS categories for this cloud."`
}

// ListOSCategoriesResponse represents a list of OS categories
type ListOSCategoriesResponse struct {
	Count      int          `json:"count"`
	OSCategory []OSCategory `json:"oscategory"`
}
