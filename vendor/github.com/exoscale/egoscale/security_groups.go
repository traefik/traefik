package egoscale

import (
	"net/url"
	"strconv"
)

// SecurityGroup represent a firewalling set of rules
type SecurityGroup struct {
	ID                  string        `json:"id"`
	Account             string        `json:"account,omitempty"`
	Description         string        `json:"description,omitempty"`
	Domain              string        `json:"domain,omitempty"`
	DomainID            string        `json:"domainid,omitempty"`
	Name                string        `json:"name"`
	Project             string        `json:"project,omitempty"`
	ProjectID           string        `json:"projectid,omitempty"`
	VirtualMachineCount int           `json:"virtualmachinecount,omitempty"` // CloudStack 4.6+
	VirtualMachineIDs   []string      `json:"virtualmachineids,omitempty"`   // CloudStack 4.6+
	IngressRule         []IngressRule `json:"ingressrule"`
	EgressRule          []EgressRule  `json:"egressrule"`
	Tags                []ResourceTag `json:"tags,omitempty"`
	JobID               string        `json:"jobid,omitempty"`
	JobStatus           JobStatusType `json:"jobstatus,omitempty"`
}

// ResourceType returns the type of the resource
func (*SecurityGroup) ResourceType() string {
	return "SecurityGroup"
}

// IngressRule represents the ingress rule
type IngressRule struct {
	RuleID                string              `json:"ruleid"`
	Account               string              `json:"account,omitempty"`
	Cidr                  string              `json:"cidr,omitempty"`
	Description           string              `json:"description,omitempty"`
	IcmpType              int                 `json:"icmptype,omitempty"`
	IcmpCode              int                 `json:"icmpcode,omitempty"`
	StartPort             int                 `json:"startport,omitempty"`
	EndPort               int                 `json:"endport,omitempty"`
	Protocol              string              `json:"protocol,omitempty"`
	Tags                  []ResourceTag       `json:"tags,omitempty"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
	JobID                 string              `json:"jobid,omitempty"`
	JobStatus             JobStatusType       `json:"jobstatus,omitempty"`
}

// EgressRule represents the ingress rule
type EgressRule IngressRule

// UserSecurityGroup represents the traffic of another security group
type UserSecurityGroup struct {
	Group   string `json:"group,omitempty"`
	Account string `json:"account,omitempty"`
}

// SecurityGroupResponse represents a generic security group response
type SecurityGroupResponse struct {
	SecurityGroup SecurityGroup `json:"securitygroup"`
}

// CreateSecurityGroup represents a security group creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/createSecurityGroup.html
type CreateSecurityGroup struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (*CreateSecurityGroup) name() string {
	return "createSecurityGroup"
}

func (*CreateSecurityGroup) response() interface{} {
	return new(CreateSecurityGroupResponse)
}

// CreateSecurityGroupResponse represents a new security group
type CreateSecurityGroupResponse SecurityGroupResponse

// DeleteSecurityGroup represents a security group deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSecurityGroup.html
type DeleteSecurityGroup struct {
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ID        string `json:"id,omitempty"`   // Mutually exclusive with name
	Name      string `json:"name,omitempty"` // Mutually exclusive with id
	ProjectID string `json:"project,omitempty"`
}

func (*DeleteSecurityGroup) name() string {
	return "deleteSecurityGroup"
}

func (*DeleteSecurityGroup) response() interface{} {
	return new(booleanSyncResponse)
}

// AuthorizeSecurityGroupIngress (Async) represents the ingress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupIngress.html
type AuthorizeSecurityGroupIngress struct {
	Account               string              `json:"account,omitempty"`
	CidrList              []string            `json:"cidrlist,omitempty"`
	Description           string              `json:"description,omitempty"`
	DomainID              string              `json:"domainid,omitempty"`
	IcmpType              int                 `json:"icmptype,omitempty"`
	IcmpCode              int                 `json:"icmpcode,omitempty"`
	StartPort             int                 `json:"startport,omitempty"`
	EndPort               int                 `json:"endport,omitempty"`
	ProjectID             string              `json:"projectid,omitempty"`
	Protocol              string              `json:"protocol,omitempty"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
}

func (*AuthorizeSecurityGroupIngress) name() string {
	return "authorizeSecurityGroupIngress"
}

func (*AuthorizeSecurityGroupIngress) asyncResponse() interface{} {
	return new(AuthorizeSecurityGroupIngressResponse)
}

func (req *AuthorizeSecurityGroupIngress) onBeforeSend(params *url.Values) error {
	// ICMP code and type may be zero but can also be omitted...
	if req.Protocol == "ICMP" {
		params.Set("icmpcode", strconv.FormatInt(int64(req.IcmpCode), 10))
		params.Set("icmptype", strconv.FormatInt(int64(req.IcmpType), 10))
	}
	return nil
}

// AuthorizeSecurityGroupIngressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupIngressResponse SecurityGroupResponse

// AuthorizeSecurityGroupEgress (Async) represents the egress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupEgress.html
type AuthorizeSecurityGroupEgress AuthorizeSecurityGroupIngress

func (*AuthorizeSecurityGroupEgress) name() string {
	return "authorizeSecurityGroupEgress"
}

func (*AuthorizeSecurityGroupEgress) asyncResponse() interface{} {
	return new(AuthorizeSecurityGroupEgressResponse)
}

func (req *AuthorizeSecurityGroupEgress) onBeforeSend(params *url.Values) error {
	return (*AuthorizeSecurityGroupIngress)(req).onBeforeSend(params)
}

// AuthorizeSecurityGroupEgressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupEgressResponse CreateSecurityGroupResponse

// RevokeSecurityGroupIngress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupIngress.html
type RevokeSecurityGroupIngress struct {
	ID string `json:"id"`
}

func (*RevokeSecurityGroupIngress) name() string {
	return "revokeSecurityGroupIngress"
}

func (*RevokeSecurityGroupIngress) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// RevokeSecurityGroupEgress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupEgress.html
type RevokeSecurityGroupEgress struct {
	ID string `json:"id"`
}

func (*RevokeSecurityGroupEgress) name() string {
	return "revokeSecurityGroupEgress"
}

func (*RevokeSecurityGroupEgress) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ListSecurityGroups represents a search for security groups
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listSecurityGroups.html
type ListSecurityGroups struct {
	Account           string        `json:"account,omitempty"`
	DomainID          string        `json:"domainid,omitempty"`
	ID                string        `json:"id,omitempty"`
	IsRecursive       bool          `json:"isrecursive,omitempty"`
	Keyword           string        `json:"keyword,omitempty"`
	ListAll           bool          `json:"listall,omitempty"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	ProjectID         string        `json:"projectid,omitempty"`
	Type              string        `json:"type,omitempty"`
	SecurityGroupName string        `json:"securitygroupname,omitempty"`
	Tags              []ResourceTag `json:"tags,omitempty"`
	VirtualMachineID  string        `json:"virtualmachineid,omitempty"`
}

func (*ListSecurityGroups) name() string {
	return "listSecurityGroups"
}

func (*ListSecurityGroups) response() interface{} {
	return new(ListSecurityGroupsResponse)
}

// ListSecurityGroupsResponse represents a list of security groups
type ListSecurityGroupsResponse struct {
	Count         int             `json:"count"`
	SecurityGroup []SecurityGroup `json:"securitygroup"`
}

// CreateIngressRule creates a set of ingress rules
//
// Deprecated: use the API directly
func (exo *Client) CreateIngressRule(req *AuthorizeSecurityGroupIngress, async AsyncInfo) ([]IngressRule, error) {
	resp, err := exo.AsyncRequest(req, async)
	if err != nil {
		return nil, err
	}
	return resp.(*AuthorizeSecurityGroupIngressResponse).SecurityGroup.IngressRule, nil
}

// CreateEgressRule creates a set of egress rules
//
// Deprecated: use the API directly
func (exo *Client) CreateEgressRule(req *AuthorizeSecurityGroupEgress, async AsyncInfo) ([]EgressRule, error) {
	resp, err := exo.AsyncRequest(req, async)
	if err != nil {
		return nil, err
	}
	return resp.(*AuthorizeSecurityGroupEgressResponse).SecurityGroup.EgressRule, nil
}

// CreateSecurityGroupWithRules create a security group with its rules
// Warning: it doesn't rollback in case of a failure!
//
// Deprecated: use the API directly
func (exo *Client) CreateSecurityGroupWithRules(name string, ingress []AuthorizeSecurityGroupIngress, egress []AuthorizeSecurityGroupEgress, async AsyncInfo) (*SecurityGroup, error) {
	req := &CreateSecurityGroup{
		Name: name,
	}
	resp, err := exo.Request(req)
	if err != nil {
		return nil, err
	}

	sg := resp.(*SecurityGroupResponse).SecurityGroup
	reqs := make([]AsyncCommand, 0, len(ingress)+len(egress))
	// Egress rules
	for _, ereq := range egress {
		ereq.SecurityGroupID = sg.ID
		reqs = append(reqs, &ereq)

	}
	// Ingress rules
	for _, ireq := range ingress {
		ireq.SecurityGroupID = sg.ID
		reqs = append(reqs, &ireq)
	}

	for _, r := range reqs {
		_, err := exo.AsyncRequest(r, async)
		if err != nil {
			return nil, err
		}
	}

	r, err := exo.Request(&ListSecurityGroups{
		ID: sg.ID,
	})
	if err != nil {
		return nil, err
	}

	sg = r.(*ListSecurityGroupsResponse).SecurityGroup[0]
	return &sg, nil
}

// DeleteSecurityGroup deletes a security group
//
// Deprecated: use the API directly
func (exo *Client) DeleteSecurityGroup(name string) error {
	req := &DeleteSecurityGroup{
		Name: name,
	}
	return exo.BooleanRequest(req)
}
