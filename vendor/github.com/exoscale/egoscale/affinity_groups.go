package egoscale

import (
	"net/url"
)

// AffinityGroup represents an (anti-)affinity group
type AffinityGroup struct {
	ID                string   `json:"id,omitempty"`
	Account           string   `json:"account,omitempty"`
	Description       string   `json:"description,omitempty"`
	Domain            string   `json:"domain,omitempty"`
	DomainID          string   `json:"domainid,omitempty"`
	Name              string   `json:"name,omitempty"`
	Type              string   `json:"type,omitempty"`
	VirtualMachineIDs []string `json:"virtualmachineIDs,omitempty"` // *I*ds is not a typo
}

// AffinityGroupType represent an affinity group type
type AffinityGroupType struct {
	Type string `json:"type"`
}

// CreateAffinityGroup (Async) represents a new (anti-)affinity group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createAffinityGroup.html
type CreateAffinityGroup struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
}

func (*CreateAffinityGroup) name() string {
	return "createAffinityGroup"
}

func (*CreateAffinityGroup) asyncResponse() interface{} {
	return new(CreateAffinityGroupResponse)
}

// CreateAffinityGroupResponse represents the response of the creation of an (anti-)affinity group
type CreateAffinityGroupResponse struct {
	AffinityGroup AffinityGroup `json:"affinitygroup"`
}

// UpdateVMAffinityGroup (Async) represents a modification of a (anti-)affinity group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateVMAffinityGroup.html
type UpdateVMAffinityGroup struct {
	ID                 string   `json:"id"`
	AffinityGroupIDs   []string `json:"affinitygroupids,omitempty"`   // mutually exclusive with names
	AffinityGroupNames []string `json:"affinitygroupnames,omitempty"` // mutually exclusive with ids
}

func (*UpdateVMAffinityGroup) name() string {
	return "updateVMAffinityGroup"
}

func (*UpdateVMAffinityGroup) asyncResponse() interface{} {
	return new(UpdateVMAffinityGroupResponse)
}

func (req *UpdateVMAffinityGroup) onBeforeSend(params *url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) == 0 && len(req.AffinityGroupNames) == 0 {
		params.Set("affinitygroupids", "")
	}
	return nil
}

// UpdateVMAffinityGroupResponse represents the new VM
type UpdateVMAffinityGroupResponse VirtualMachineResponse

// DeleteAffinityGroup (Async) represents an (anti-)affinity group to be deleted
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteAffinityGroup.html
type DeleteAffinityGroup struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
}

func (*DeleteAffinityGroup) name() string {
	return "deleteAffinityGroup"
}

func (*DeleteAffinityGroup) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ListAffinityGroups represents an (anti-)affinity groups search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAffinityGroups.html
type ListAffinityGroups struct {
	Account          string `json:"account,omitempty"`
	DomainID         string `json:"domainid,omitempty"`
	ID               string `json:"id,omitempty"`
	IsRecursive      bool   `json:"isrecursive,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	ListAll          bool   `json:"listall,omitempty"`
	Name             string `json:"name,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	Type             string `json:"type,omitempty"`
	VirtualMachineID string `json:"virtualmachineid,omitempty"`
}

func (*ListAffinityGroups) name() string {
	return "listAffinityGroups"
}

func (*ListAffinityGroups) response() interface{} {
	return new(ListAffinityGroupsResponse)
}

// ListAffinityGroupTypes represents an (anti-)affinity groups search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAffinityGroupTypes.html
type ListAffinityGroupTypes struct {
	Keyword  string `json:"keyword,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pagesize,omitempty"`
}

func (*ListAffinityGroupTypes) name() string {
	return "listAffinityGroupTypes"
}

func (*ListAffinityGroupTypes) response() interface{} {
	return new(ListAffinityGroupTypesResponse)
}

// ListAffinityGroupsResponse represents a list of (anti-)affinity groups
type ListAffinityGroupsResponse struct {
	Count         int             `json:"count"`
	AffinityGroup []AffinityGroup `json:"affinitygroup"`
}

// ListAffinityGroupTypesResponse represents a list of (anti-)affinity group types
type ListAffinityGroupTypesResponse struct {
	Count             int                 `json:"count"`
	AffinityGroupType []AffinityGroupType `json:"affinitygrouptype"`
}

// Legacy methods

// CreateAffinityGroup creates a group
//
// Deprecated: Use the API directly
func (exo *Client) CreateAffinityGroup(name string, async AsyncInfo) (*AffinityGroup, error) {
	req := &CreateAffinityGroup{
		Name: name,
	}
	resp, err := exo.AsyncRequest(req, async)
	if err != nil {
		return nil, err
	}

	ag := resp.(*CreateAffinityGroupResponse).AffinityGroup
	return &ag, nil
}

// DeleteAffinityGroup deletes a group
//
// Deprecated: Use the API directly
func (exo *Client) DeleteAffinityGroup(name string, async AsyncInfo) error {
	req := &DeleteAffinityGroup{
		Name: name,
	}
	return exo.BooleanAsyncRequest(req, async)
}
