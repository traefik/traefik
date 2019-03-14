package egoscale

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// SecurityGroup represent a firewalling set of rules
type SecurityGroup struct {
	Account     string        `json:"account,omitempty" doc:"the account owning the security group"`
	Description string        `json:"description,omitempty" doc:"the description of the security group"`
	EgressRule  []EgressRule  `json:"egressrule,omitempty" doc:"the list of egress rules associated with the security group"`
	ID          *UUID         `json:"id" doc:"the ID of the security group"`
	IngressRule []IngressRule `json:"ingressrule,omitempty" doc:"the list of ingress rules associated with the security group"`
	Name        string        `json:"name,omitempty" doc:"the name of the security group"`
}

// UserSecurityGroup converts a SecurityGroup to a UserSecurityGroup
func (sg SecurityGroup) UserSecurityGroup() UserSecurityGroup {
	return UserSecurityGroup{
		Group: sg.Name,
	}
}

// ListRequest builds the ListSecurityGroups request
func (sg SecurityGroup) ListRequest() (ListCommand, error) {
	req := &ListSecurityGroups{
		ID:                sg.ID,
		SecurityGroupName: sg.Name,
	}

	return req, nil
}

// Delete deletes the given Security Group
func (sg SecurityGroup) Delete(ctx context.Context, client *Client) error {
	if sg.ID == nil && sg.Name == "" {
		return fmt.Errorf("a SecurityGroup may only be deleted using ID or Name")
	}

	req := &DeleteSecurityGroup{}

	if sg.ID != nil {
		req.ID = sg.ID
	} else {
		req.Name = sg.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

// RuleByID returns IngressRule or EgressRule by a rule ID
func (sg SecurityGroup) RuleByID(ruleID UUID) (*IngressRule, *EgressRule) {
	for i, in := range sg.IngressRule {
		if in.RuleID.Equal(ruleID) {
			return &sg.IngressRule[i], nil
		}
	}

	for i, out := range sg.EgressRule {
		if out.RuleID.Equal(ruleID) {
			return nil, &sg.EgressRule[i]
		}
	}

	return nil, nil
}

// IngressRule represents the ingress rule
type IngressRule struct {
	CIDR              *CIDR  `json:"cidr,omitempty" doc:"the CIDR notation for the base IP address of the security group rule"`
	Description       string `json:"description,omitempty" doc:"description of the security group rule"`
	EndPort           uint16 `json:"endport,omitempty" doc:"the ending port of the security group rule "`
	IcmpCode          uint8  `json:"icmpcode,omitempty" doc:"the code for the ICMP message response"`
	IcmpType          uint8  `json:"icmptype,omitempty" doc:"the type of the ICMP message response"`
	Protocol          string `json:"protocol,omitempty" doc:"the protocol of the security group rule"`
	RuleID            *UUID  `json:"ruleid" doc:"the id of the security group rule"`
	SecurityGroupName string `json:"securitygroupname,omitempty" doc:"security group name"`
	StartPort         uint16 `json:"startport,omitempty" doc:"the starting port of the security group rule"`
}

// EgressRule represents the ingress rule
type EgressRule IngressRule

// UserSecurityGroup represents the traffic of another security group
type UserSecurityGroup struct {
	Group string `json:"group,omitempty"`
}

// String gives the UserSecurityGroup name
func (usg UserSecurityGroup) String() string {
	return usg.Group
}

// CreateSecurityGroup represents a security group creation
type CreateSecurityGroup struct {
	Name        string `json:"name" doc:"name of the security group"`
	Description string `json:"description,omitempty" doc:"the description of the security group"`
	_           bool   `name:"createSecurityGroup" description:"Creates a security group"`
}

// Response returns the struct to unmarshal
func (CreateSecurityGroup) Response() interface{} {
	return new(SecurityGroup)
}

// DeleteSecurityGroup represents a security group deletion
type DeleteSecurityGroup struct {
	ID   *UUID  `json:"id,omitempty" doc:"The ID of the security group. Mutually exclusive with name parameter"`
	Name string `json:"name,omitempty" doc:"The ID of the security group. Mutually exclusive with id parameter"`
	_    bool   `name:"deleteSecurityGroup" description:"Deletes security group"`
}

// Response returns the struct to unmarshal
func (DeleteSecurityGroup) Response() interface{} {
	return new(BooleanResponse)
}

// AuthorizeSecurityGroupIngress (Async) represents the ingress rule creation
type AuthorizeSecurityGroupIngress struct {
	CIDRList              []CIDR              `json:"cidrlist,omitempty" doc:"the cidr list associated"`
	Description           string              `json:"description,omitempty" doc:"the description of the ingress/egress rule"`
	EndPort               uint16              `json:"endport,omitempty" doc:"end port for this ingress/egress rule"`
	IcmpCode              uint8               `json:"icmpcode,omitempty" doc:"error code for this icmp message"`
	IcmpType              uint8               `json:"icmptype,omitempty" doc:"type of the icmp message being sent"`
	Protocol              string              `json:"protocol,omitempty" doc:"TCP is default. UDP, ICMP, ICMPv6, AH, ESP, GRE, IPIP are the other supported protocols"`
	SecurityGroupID       *UUID               `json:"securitygroupid,omitempty" doc:"The ID of the security group. Mutually exclusive with securitygroupname parameter"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty" doc:"The name of the security group. Mutually exclusive with securitygroupid parameter"`
	StartPort             uint16              `json:"startport,omitempty" doc:"start port for this ingress/egress rule"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty" doc:"user to security group mapping"`
	_                     bool                `name:"authorizeSecurityGroupIngress" description:"Authorize a particular ingress/egress rule for this security group"`
}

// Response returns the struct to unmarshal
func (AuthorizeSecurityGroupIngress) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (AuthorizeSecurityGroupIngress) AsyncResponse() interface{} {
	return new(SecurityGroup)
}

func (req AuthorizeSecurityGroupIngress) onBeforeSend(params url.Values) error {
	// ICMP code and type may be zero but can also be omitted...
	if strings.HasPrefix(strings.ToLower(req.Protocol), "icmp") {
		params.Set("icmpcode", strconv.FormatInt(int64(req.IcmpCode), 10))
		params.Set("icmptype", strconv.FormatInt(int64(req.IcmpType), 10))
	}
	// StartPort may be zero but can also be omitted...
	if req.EndPort != 0 && req.StartPort == 0 {
		params.Set("startport", "0")
	}
	return nil
}

// AuthorizeSecurityGroupEgress (Async) represents the egress rule creation
type AuthorizeSecurityGroupEgress AuthorizeSecurityGroupIngress

// Response returns the struct to unmarshal
func (AuthorizeSecurityGroupEgress) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (AuthorizeSecurityGroupEgress) AsyncResponse() interface{} {
	return new(SecurityGroup)
}

func (req AuthorizeSecurityGroupEgress) onBeforeSend(params url.Values) error {
	return (AuthorizeSecurityGroupIngress)(req).onBeforeSend(params)
}

// RevokeSecurityGroupIngress (Async) represents the ingress/egress rule deletion
type RevokeSecurityGroupIngress struct {
	ID *UUID `json:"id" doc:"The ID of the ingress rule"`
	_  bool  `name:"revokeSecurityGroupIngress" description:"Deletes a particular ingress rule from this security group"`
}

// Response returns the struct to unmarshal
func (RevokeSecurityGroupIngress) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (RevokeSecurityGroupIngress) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

// RevokeSecurityGroupEgress (Async) represents the ingress/egress rule deletion
type RevokeSecurityGroupEgress struct {
	ID *UUID `json:"id" doc:"The ID of the egress rule"`
	_  bool  `name:"revokeSecurityGroupEgress" description:"Deletes a particular egress rule from this security group"`
}

// Response returns the struct to unmarshal
func (RevokeSecurityGroupEgress) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (RevokeSecurityGroupEgress) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

//go:generate go run generate/main.go -interface=Listable ListSecurityGroups

// ListSecurityGroups represents a search for security groups
type ListSecurityGroups struct {
	ID                *UUID  `json:"id,omitempty" doc:"list the security group by the id provided"`
	Keyword           string `json:"keyword,omitempty" doc:"List by keyword"`
	Page              int    `json:"page,omitempty"`
	PageSize          int    `json:"pagesize,omitempty"`
	SecurityGroupName string `json:"securitygroupname,omitempty" doc:"lists security groups by name"`
	VirtualMachineID  *UUID  `json:"virtualmachineid,omitempty" doc:"lists security groups by virtual machine id"`
	_                 bool   `name:"listSecurityGroups" description:"Lists security groups"`
}

// ListSecurityGroupsResponse represents a list of security groups
type ListSecurityGroupsResponse struct {
	Count         int             `json:"count"`
	SecurityGroup []SecurityGroup `json:"securitygroup"`
}
