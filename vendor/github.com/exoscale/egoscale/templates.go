package egoscale

// Template represents a machine to be deployed
type Template struct {
	Account               string            `json:"account,omitempty"`
	AccountID             string            `json:"accountid,omitempty"`
	Bootable              bool              `json:"bootable,omitempty"`
	Checksum              string            `json:"checksum,omitempty"`
	Created               string            `json:"created,omitempty"`
	CrossZones            bool              `json:"crossZones,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	DisplayText           string            `json:"displaytext,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainID              string            `json:"domainid,omitempty"`
	Format                string            `json:"format,omitempty"`
	HostID                string            `json:"hostid,omitempty"`
	HostName              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	ID                    string            `json:"id,omitempty"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty"`
	IsExtractable         bool              `json:"isextractable,omitempty"`
	IsFeatured            bool              `json:"isfeatured,omitempty"`
	IsPublic              bool              `json:"ispublic,omitempty"`
	IsReady               bool              `json:"isready,omitempty"`
	Name                  string            `json:"name,omitempty"`
	OsTypeID              string            `json:"ostypeid,omitempty"`
	OsTypeName            string            `json:"ostypename,omitempty"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty"`
	Project               string            `json:"project,omitempty"`
	ProjectID             string            `json:"projectid,omitempty"`
	Removed               string            `json:"removed,omitempty"`
	Size                  int64             `json:"size,omitempty"`
	SourceTemplateID      string            `json:"sourcetemplateid,omitempty"`
	SSHKeyEnabled         bool              `json:"sshkeyenabled,omitempty"`
	Status                string            `json:"status,omitempty"`
	Zoneid                string            `json:"zoneid,omitempty"`
	Zonename              string            `json:"zonename,omitempty"`
}

// ResourceType returns the type of the resource
func (*Template) ResourceType() string {
	return "Template"
}

// ListTemplates represents a template query filter
type ListTemplates struct {
	TemplateFilter string        `json:"templatefilter"` // featured, etc.
	Account        string        `json:"account,omitempty"`
	DomainID       string        `json:"domainid,omitempty"`
	Hypervisor     string        `json:"hypervisor,omitempty"`
	ID             string        `json:"id,omitempty"`
	IsRecursive    bool          `json:"isrecursive,omitempty"`
	Keyword        string        `json:"keyword,omitempty"`
	ListAll        bool          `json:"listall,omitempty"`
	Name           string        `json:"name,omitempty"`
	Page           int           `json:"page,omitempty"`
	PageSize       int           `json:"pagesize,omitempty"`
	ProjectID      string        `json:"projectid,omitempty"`
	ShowRemoved    bool          `json:"showremoved,omitempty"`
	Tags           []ResourceTag `json:"tags,omitempty"`
	ZoneID         string        `json:"zoneid,omitempty"`
}

func (*ListTemplates) name() string {
	return "listTemplates"
}

func (*ListTemplates) response() interface{} {
	return new(ListTemplatesResponse)
}

// ListTemplatesResponse represents a list of templates
type ListTemplatesResponse struct {
	Count    int        `json:"count"`
	Template []Template `json:"template"`
}
