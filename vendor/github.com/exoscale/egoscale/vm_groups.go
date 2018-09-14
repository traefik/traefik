package egoscale

// InstanceGroup represents a group of VM
type InstanceGroup struct {
	Account  string `json:"account,omitempty" doc:"the account owning the instance group"`
	Created  string `json:"created,omitempty" doc:"time and date the instance group was created"`
	Domain   string `json:"domain,omitempty" doc:"the domain name of the instance group"`
	DomainID *UUID  `json:"domainid,omitempty" doc:"the domain ID of the instance group"`
	ID       *UUID  `json:"id,omitempty" doc:"the id of the instance group"`
	Name     string `json:"name,omitempty" doc:"the name of the instance group"`
}

// CreateInstanceGroup creates a VM group
type CreateInstanceGroup struct {
	Name     string `json:"name" doc:"the name of the instance group"`
	Account  string `json:"account,omitempty" doc:"the account of the instance group. The account parameter must be used with the domainId parameter."`
	DomainID *UUID  `json:"domainid,omitempty" doc:"the domain ID of account owning the instance group"`
	_        bool   `name:"createInstanceGroup" description:"Creates a vm group"`
}

func (CreateInstanceGroup) response() interface{} {
	return new(InstanceGroup)
}

// UpdateInstanceGroup updates a VM group
type UpdateInstanceGroup struct {
	ID   *UUID  `json:"id" doc:"Instance group ID"`
	Name string `json:"name,omitempty" doc:"new instance group name"`
	_    bool   `name:"updateInstanceGroup" description:"Updates a vm group"`
}

func (UpdateInstanceGroup) response() interface{} {
	return new(InstanceGroup)
}

// DeleteInstanceGroup deletes a VM group
type DeleteInstanceGroup struct {
	ID *UUID `json:"id" doc:"the ID of the instance group"`
	_  bool  `name:"deleteInstanceGroup" description:"Deletes a vm group"`
}

func (DeleteInstanceGroup) response() interface{} {
	return new(booleanResponse)
}

// ListInstanceGroups lists VM groups
type ListInstanceGroups struct {
	Account     string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID    *UUID  `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID          *UUID  `json:"id,omitempty" doc:"list instance groups by ID"`
	IsRecursive *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll     *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name        string `json:"name,omitempty" doc:"list instance groups by name"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	_           bool   `name:"listInstanceGroups" description:"Lists vm groups"`
}

// ListInstanceGroupsResponse represents a list of instance groups
type ListInstanceGroupsResponse struct {
	Count         int             `json:"count"`
	InstanceGroup []InstanceGroup `json:"instancegroup"`
}

func (ListInstanceGroups) response() interface{} {
	return new(ListInstanceGroupsResponse)
}
