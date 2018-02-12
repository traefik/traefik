package egoscale

// InstanceGroup represents a group of VM
type InstanceGroup struct {
	ID        string `json:"id"`
	Account   string `json:"account,omitempty"`
	Created   string `json:"created,omitempty"`
	Domain    string `json:"domain,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	Name      string `json:"name,omitempty"`
	Project   string `json:"project,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

// InstanceGroupResponse represents a VM group
type InstanceGroupResponse struct {
	InstanceGroup InstanceGroup `json:"instancegroup"`
}

// CreateInstanceGroup creates a VM group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createInstanceGroup.html
type CreateInstanceGroup struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

func (*CreateInstanceGroup) name() string {
	return "createInstanceGroup"
}

func (*CreateInstanceGroup) response() interface{} {
	return new(CreateInstanceGroupResponse)
}

// CreateInstanceGroupResponse represents a freshly created VM group
type CreateInstanceGroupResponse InstanceGroupResponse

// UpdateInstanceGroup creates a VM group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateInstanceGroup.html
type UpdateInstanceGroup struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

func (*UpdateInstanceGroup) name() string {
	return "updateInstanceGroup"
}

func (*UpdateInstanceGroup) response() interface{} {
	return new(UpdateInstanceGroupResponse)
}

// UpdateInstanceGroupResponse represents an updated VM group
type UpdateInstanceGroupResponse InstanceGroupResponse

// DeleteInstanceGroup creates a VM group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteInstanceGroup.html
type DeleteInstanceGroup struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

func (*DeleteInstanceGroup) name() string {
	return "deleteInstanceGroup"
}

func (*DeleteInstanceGroup) response() interface{} {
	return new(booleanSyncResponse)
}

// ListInstanceGroups lists VM groups
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listInstanceGroups.html
type ListInstanceGroups struct {
	Account     string `json:"account,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
	ID          string `json:"id,omitempty"`
	IsRecursive bool   `json:"isrecursive,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	ListAll     bool   `json:"listall,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	State       string `json:"state,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
}

func (*ListInstanceGroups) name() string {
	return "listInstanceGroups"
}

func (*ListInstanceGroups) response() interface{} {
	return new(ListInstanceGroupsResponse)
}

// ListInstanceGroupsResponse represents a list of instance groups
type ListInstanceGroupsResponse struct {
	Count         int             `json:"count"`
	InstanceGroup []InstanceGroup `json:"instancegroup"`
}
