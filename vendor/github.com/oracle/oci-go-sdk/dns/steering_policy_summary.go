// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
)

// SteeringPolicySummary A DNS steering policy.
// *Warning:* Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type SteeringPolicySummary struct {

	// The OCID of the compartment containing the steering policy.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// A user-friendly name for the steering policy.
	// Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The Time To Live for responses from the steering policy, in seconds.
	// If not specified during creation, a value of 30 seconds will be used.
	Ttl *int `mandatory:"false" json:"ttl"`

	// The OCID of the health check monitor providing health data about the answers of the
	// steering policy.
	// A steering policy answer with `rdata` matching a monitored endpoint will use the health
	// data of that endpoint.
	// A steering policy answer with `rdata` not matching any monitored endpoint will be assumed
	// healthy.
	HealthCheckMonitorId *string `mandatory:"false" json:"healthCheckMonitorId"`

	// The common pattern (or lack thereof) to which the steering policy adheres. This
	// value restricts the possible configurations of rules, but thereby supports
	// specifically tailored interfaces. Values other than "CUSTOM" require the rules to
	// begin with an unconditional FILTER that keeps answers contingent upon
	// `answer.isDisabled != true`, followed
	// _if and only if the policy references a health check monitor_ by an unconditional
	// HEALTH rule, and require the last rule to be an unconditional LIMIT.
	// What must precede the LIMIT rule is determined by the template value:
	// - FAILOVER requires exactly an unconditional PRIORITY rule that ranks answers by pool.
	//   Each answer pool must have a unique priority value assigned to it. Answer data must
	//   be defined in the `defaultAnswerData` property for the rule and the `cases` property
	//   must not be defined.
	// - LOAD_BALANCE requires exactly an unconditional WEIGHTED rule that shuffles answers
	//   by name. Answer data must be defined in the `defaultAnswerData` property for the
	//   rule and the `cases` property must not be defined.
	// - ROUTE_BY_GEO requires exactly one PRIORITY rule that ranks answers by pool using the
	//   geographical location of the client as a condition. Within that rule you may only
	//   use `query.client.geoKey` in the `caseCondition` expressions for defining the cases.
	//   For each case in the PRIORITY rule each answer pool must have a unique priority
	//   value assigned to it. Answer data can only be defined within cases and
	//   `defaultAnswerData` cannot be used in the PRIORITY rule.
	// - ROUTE_BY_ASN requires exactly one PRIORITY rule that ranks answers by pool using the
	//   ASN of the client as a condition. Within that rule you may only use
	//   `query.client.asn` in the `caseCondition` expressions for defining the cases.
	//   For each case in the PRIORITY rule each answer pool must have a unique priority
	//   value assigned to it. Answer data can only be defined within cases and
	//   `defaultAnswerData` cannot be used in the PRIORITY rule.
	// - ROUTE_BY_IP requires exactly one PRIORITY rule that ranks answers by pool using the
	//   IP subnet of the client as a condition. Within that rule you may only use
	//   `query.client.address` in the `caseCondition` expressions for defining the cases.
	//   For each case in the PRIORITY rule each answer pool must have a unique priority
	//   value assigned to it. Answer data can only be defined within cases and
	//   `defaultAnswerData` cannot be used in the PRIORITY rule.
	// - CUSTOM allows an arbitrary configuration of rules.
	// For an existing steering policy, the template value may be changed to any of the
	// supported options but the resulting policy must conform to the requirements for the
	// new template type or else a Bad Request error will be returned.
	Template SteeringPolicySummaryTemplateEnum `mandatory:"false" json:"template,omitempty"`

	// Simple key-value pair that is applied without any predefined name, type, or scope.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of predefined tag keys. These predefined keys are scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The canonical absolute URL of the resource.
	Self *string `mandatory:"false" json:"self"`

	// The OCID of the resource.
	Id *string `mandatory:"false" json:"id"`

	// The date and time the resource was created in "YYYY-MM-ddThh:mmZ" format
	// with a Z offset, as defined by RFC 3339.
	// **Example:** `2016-07-22T17:23:59:60Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The current state of the resource.
	LifecycleState SteeringPolicySummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m SteeringPolicySummary) String() string {
	return common.PointerString(m)
}

// SteeringPolicySummaryTemplateEnum Enum with underlying type: string
type SteeringPolicySummaryTemplateEnum string

// Set of constants representing the allowable values for SteeringPolicySummaryTemplateEnum
const (
	SteeringPolicySummaryTemplateFailover    SteeringPolicySummaryTemplateEnum = "FAILOVER"
	SteeringPolicySummaryTemplateLoadBalance SteeringPolicySummaryTemplateEnum = "LOAD_BALANCE"
	SteeringPolicySummaryTemplateRouteByGeo  SteeringPolicySummaryTemplateEnum = "ROUTE_BY_GEO"
	SteeringPolicySummaryTemplateRouteByAsn  SteeringPolicySummaryTemplateEnum = "ROUTE_BY_ASN"
	SteeringPolicySummaryTemplateRouteByIp   SteeringPolicySummaryTemplateEnum = "ROUTE_BY_IP"
	SteeringPolicySummaryTemplateCustom      SteeringPolicySummaryTemplateEnum = "CUSTOM"
)

var mappingSteeringPolicySummaryTemplate = map[string]SteeringPolicySummaryTemplateEnum{
	"FAILOVER":     SteeringPolicySummaryTemplateFailover,
	"LOAD_BALANCE": SteeringPolicySummaryTemplateLoadBalance,
	"ROUTE_BY_GEO": SteeringPolicySummaryTemplateRouteByGeo,
	"ROUTE_BY_ASN": SteeringPolicySummaryTemplateRouteByAsn,
	"ROUTE_BY_IP":  SteeringPolicySummaryTemplateRouteByIp,
	"CUSTOM":       SteeringPolicySummaryTemplateCustom,
}

// GetSteeringPolicySummaryTemplateEnumValues Enumerates the set of values for SteeringPolicySummaryTemplateEnum
func GetSteeringPolicySummaryTemplateEnumValues() []SteeringPolicySummaryTemplateEnum {
	values := make([]SteeringPolicySummaryTemplateEnum, 0)
	for _, v := range mappingSteeringPolicySummaryTemplate {
		values = append(values, v)
	}
	return values
}

// SteeringPolicySummaryLifecycleStateEnum Enum with underlying type: string
type SteeringPolicySummaryLifecycleStateEnum string

// Set of constants representing the allowable values for SteeringPolicySummaryLifecycleStateEnum
const (
	SteeringPolicySummaryLifecycleStateActive   SteeringPolicySummaryLifecycleStateEnum = "ACTIVE"
	SteeringPolicySummaryLifecycleStateCreating SteeringPolicySummaryLifecycleStateEnum = "CREATING"
	SteeringPolicySummaryLifecycleStateDeleted  SteeringPolicySummaryLifecycleStateEnum = "DELETED"
	SteeringPolicySummaryLifecycleStateDeleting SteeringPolicySummaryLifecycleStateEnum = "DELETING"
)

var mappingSteeringPolicySummaryLifecycleState = map[string]SteeringPolicySummaryLifecycleStateEnum{
	"ACTIVE":   SteeringPolicySummaryLifecycleStateActive,
	"CREATING": SteeringPolicySummaryLifecycleStateCreating,
	"DELETED":  SteeringPolicySummaryLifecycleStateDeleted,
	"DELETING": SteeringPolicySummaryLifecycleStateDeleting,
}

// GetSteeringPolicySummaryLifecycleStateEnumValues Enumerates the set of values for SteeringPolicySummaryLifecycleStateEnum
func GetSteeringPolicySummaryLifecycleStateEnumValues() []SteeringPolicySummaryLifecycleStateEnum {
	values := make([]SteeringPolicySummaryLifecycleStateEnum, 0)
	for _, v := range mappingSteeringPolicySummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
