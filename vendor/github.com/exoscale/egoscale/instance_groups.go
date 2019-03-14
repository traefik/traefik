package egoscale

// InstanceGroup represents a group of VM
type InstanceGroup struct {
	Account string `json:"account,omitempty" doc:"the account owning the instance group"`
	Created string `json:"created,omitempty" doc:"time and date the instance group was created"`
	ID      *UUID  `json:"id,omitempty" doc:"the id of the instance group"`
	Name    string `json:"name,omitempty" doc:"the name of the instance group"`
}

// ListRequest builds the ListInstanceGroups request
func (ig InstanceGroup) ListRequest() (ListCommand, error) {
	req := &ListInstanceGroups{
		ID:   ig.ID,
		Name: ig.Name,
	}

	return req, nil
}

// CreateInstanceGroup creates a VM group
type CreateInstanceGroup struct {
	Name string `json:"name" doc:"the name of the instance group"`
	_    bool   `name:"createInstanceGroup" description:"Creates a vm group"`
}

// Response returns the struct to unmarshal
func (CreateInstanceGroup) Response() interface{} {
	return new(InstanceGroup)
}

// UpdateInstanceGroup updates a VM group
type UpdateInstanceGroup struct {
	ID   *UUID  `json:"id" doc:"Instance group ID"`
	Name string `json:"name,omitempty" doc:"new instance group name"`
	_    bool   `name:"updateInstanceGroup" description:"Updates a vm group"`
}

// Response returns the struct to unmarshal
func (UpdateInstanceGroup) Response() interface{} {
	return new(InstanceGroup)
}

// DeleteInstanceGroup deletes a VM group
type DeleteInstanceGroup struct {
	ID *UUID `json:"id" doc:"the ID of the instance group"`
	_  bool  `name:"deleteInstanceGroup" description:"Deletes a vm group"`
}

// Response returns the struct to unmarshal
func (DeleteInstanceGroup) Response() interface{} {
	return new(BooleanResponse)
}

//go:generate go run generate/main.go -interface=Listable ListInstanceGroups

// ListInstanceGroups lists VM groups
type ListInstanceGroups struct {
	ID       *UUID  `json:"id,omitempty" doc:"List instance groups by ID"`
	Keyword  string `json:"keyword,omitempty" doc:"List by keyword"`
	Name     string `json:"name,omitempty" doc:"List instance groups by name"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pagesize,omitempty"`
	_        bool   `name:"listInstanceGroups" description:"Lists vm groups"`
}

// ListInstanceGroupsResponse represents a list of instance groups
type ListInstanceGroupsResponse struct {
	Count         int             `json:"count"`
	InstanceGroup []InstanceGroup `json:"instancegroup"`
}
