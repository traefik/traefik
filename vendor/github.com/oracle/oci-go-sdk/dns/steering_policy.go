// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.cloud.oracle.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// SteeringPolicy A DNS steering policy.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type SteeringPolicy struct {

	// The OCID of the compartment containing the steering policy.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// A user-friendly name for the steering policy. Does not have to be unique and can be changed.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The Time To Live (TTL) for responses from the steering policy, in seconds.
	// If not specified during creation, a value of 30 seconds will be used.
	Ttl *int `mandatory:"false" json:"ttl"`

	// The OCID of the health check monitor providing health data about the answers of the
	// steering policy. A steering policy answer with `rdata` matching a monitored endpoint
	// will use the health data of that endpoint. A steering policy answer with `rdata` not
	// matching any monitored endpoint will be assumed healthy.
	//
	// **Note:** To use the Health Check monitoring feature in a steering policy, a monitor
	// must be created using the Health Checks service first. For more information on how to
	// create a monitor, please see Managing Health Checks (https://docs.cloud.oracle.com/iaas/Content/HealthChecks/Tasks/managinghealthchecks.htm).
	HealthCheckMonitorId *string `mandatory:"false" json:"healthCheckMonitorId"`

	// A set of predefined rules based on the desired purpose of the steering policy. Each
	// template utilizes Traffic Management's rules in a different order to produce the desired
	// results when answering DNS queries.
	//
	// **Example:** The `FAILOVER` template determines answers by filtering the policy's answers
	// using the `FILTER` rule first, then the following rules in succession: `HEALTH`, `PRIORITY`,
	// and `LIMIT`. This gives the domain dynamic failover capability.
	//
	// It is **strongly recommended** to use a template other than `CUSTOM` when creating
	// a steering policy.
	//
	// All templates require the rule order to begin with an unconditional `FILTER` rule that keeps
	// answers contingent upon `answer.isDisabled != true`, except for `CUSTOM`. A defined
	// `HEALTH` rule must follow the `FILTER` rule if the policy references a `healthCheckMonitorId`.
	// The last rule of a template must must be a `LIMIT` rule. For more information about templates
	// and code examples, see Traffic Management API Guide (https://docs.cloud.oracle.com/iaas/Content/TrafficManagement/Concepts/trafficmanagementapi.htm).
	// **Template Types**
	// * `FAILOVER` - Uses health check information on your endpoints to determine which DNS answers
	// to serve. If an endpoint fails a health check, the answer for that endpoint will be removed
	// from the list of available answers until the endpoint is detected as healthy.
	//
	// * `LOAD_BALANCE` - Distributes web traffic to specified endpoints based on defined weights.
	//
	// * `ROUTE_BY_GEO` - Answers DNS queries based on the query's geographic location. For a list of geographic
	// locations to route by, see Traffic Management Geographic Locations (https://docs.cloud.oracle.com/iaas/Content/TrafficManagement/Reference/trafficmanagementgeo.htm).
	//
	// * `ROUTE_BY_ASN` - Answers DNS queries based on the query's originating ASN.
	//
	// * `ROUTE_BY_IP` - Answers DNS queries based on the query's IP address.
	//
	// * `CUSTOM` - Allows a customized configuration of rules.
	Template SteeringPolicyTemplateEnum `mandatory:"false" json:"template,omitempty"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	//
	// **Example:** `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	//
	// **Example:** `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The set of all answers that can potentially issue from the steering policy.
	Answers []SteeringPolicyAnswer `mandatory:"false" json:"answers"`

	// The series of rules that will be processed in sequence to reduce the pool of answers
	// to a response for any given request.
	//
	// The first rule receives a shuffled list of all answers, and every other rule receives
	// the list of answers emitted by the one preceding it. The last rule populates the
	// response.
	Rules []SteeringPolicyRule `mandatory:"false" json:"rules"`

	// The canonical absolute URL of the resource.
	Self *string `mandatory:"false" json:"self"`

	// The OCID of the resource.
	Id *string `mandatory:"false" json:"id"`

	// The date and time the resource was created, expressed in RFC 3339 timestamp format.
	// **Example:** `2016-07-22T17:23:59:60Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The current state of the resource.
	LifecycleState SteeringPolicyLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m SteeringPolicy) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *SteeringPolicy) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		CompartmentId        *string                           `json:"compartmentId"`
		DisplayName          *string                           `json:"displayName"`
		Ttl                  *int                              `json:"ttl"`
		HealthCheckMonitorId *string                           `json:"healthCheckMonitorId"`
		Template             SteeringPolicyTemplateEnum        `json:"template"`
		FreeformTags         map[string]string                 `json:"freeformTags"`
		DefinedTags          map[string]map[string]interface{} `json:"definedTags"`
		Answers              []SteeringPolicyAnswer            `json:"answers"`
		Rules                []steeringpolicyrule              `json:"rules"`
		Self                 *string                           `json:"self"`
		Id                   *string                           `json:"id"`
		TimeCreated          *common.SDKTime                   `json:"timeCreated"`
		LifecycleState       SteeringPolicyLifecycleStateEnum  `json:"lifecycleState"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	m.CompartmentId = model.CompartmentId
	m.DisplayName = model.DisplayName
	m.Ttl = model.Ttl
	m.HealthCheckMonitorId = model.HealthCheckMonitorId
	m.Template = model.Template
	m.FreeformTags = model.FreeformTags
	m.DefinedTags = model.DefinedTags
	m.Answers = make([]SteeringPolicyAnswer, len(model.Answers))
	for i, n := range model.Answers {
		m.Answers[i] = n
	}
	m.Rules = make([]SteeringPolicyRule, len(model.Rules))
	for i, n := range model.Rules {
		nn, err := n.UnmarshalPolymorphicJSON(n.JsonData)
		if err != nil {
			return err
		}
		if nn != nil {
			m.Rules[i] = nn.(SteeringPolicyRule)
		} else {
			m.Rules[i] = nil
		}
	}
	m.Self = model.Self
	m.Id = model.Id
	m.TimeCreated = model.TimeCreated
	m.LifecycleState = model.LifecycleState
	return
}

// SteeringPolicyTemplateEnum Enum with underlying type: string
type SteeringPolicyTemplateEnum string

// Set of constants representing the allowable values for SteeringPolicyTemplateEnum
const (
	SteeringPolicyTemplateFailover    SteeringPolicyTemplateEnum = "FAILOVER"
	SteeringPolicyTemplateLoadBalance SteeringPolicyTemplateEnum = "LOAD_BALANCE"
	SteeringPolicyTemplateRouteByGeo  SteeringPolicyTemplateEnum = "ROUTE_BY_GEO"
	SteeringPolicyTemplateRouteByAsn  SteeringPolicyTemplateEnum = "ROUTE_BY_ASN"
	SteeringPolicyTemplateRouteByIp   SteeringPolicyTemplateEnum = "ROUTE_BY_IP"
	SteeringPolicyTemplateCustom      SteeringPolicyTemplateEnum = "CUSTOM"
)

var mappingSteeringPolicyTemplate = map[string]SteeringPolicyTemplateEnum{
	"FAILOVER":     SteeringPolicyTemplateFailover,
	"LOAD_BALANCE": SteeringPolicyTemplateLoadBalance,
	"ROUTE_BY_GEO": SteeringPolicyTemplateRouteByGeo,
	"ROUTE_BY_ASN": SteeringPolicyTemplateRouteByAsn,
	"ROUTE_BY_IP":  SteeringPolicyTemplateRouteByIp,
	"CUSTOM":       SteeringPolicyTemplateCustom,
}

// GetSteeringPolicyTemplateEnumValues Enumerates the set of values for SteeringPolicyTemplateEnum
func GetSteeringPolicyTemplateEnumValues() []SteeringPolicyTemplateEnum {
	values := make([]SteeringPolicyTemplateEnum, 0)
	for _, v := range mappingSteeringPolicyTemplate {
		values = append(values, v)
	}
	return values
}

// SteeringPolicyLifecycleStateEnum Enum with underlying type: string
type SteeringPolicyLifecycleStateEnum string

// Set of constants representing the allowable values for SteeringPolicyLifecycleStateEnum
const (
	SteeringPolicyLifecycleStateActive   SteeringPolicyLifecycleStateEnum = "ACTIVE"
	SteeringPolicyLifecycleStateCreating SteeringPolicyLifecycleStateEnum = "CREATING"
	SteeringPolicyLifecycleStateDeleted  SteeringPolicyLifecycleStateEnum = "DELETED"
	SteeringPolicyLifecycleStateDeleting SteeringPolicyLifecycleStateEnum = "DELETING"
)

var mappingSteeringPolicyLifecycleState = map[string]SteeringPolicyLifecycleStateEnum{
	"ACTIVE":   SteeringPolicyLifecycleStateActive,
	"CREATING": SteeringPolicyLifecycleStateCreating,
	"DELETED":  SteeringPolicyLifecycleStateDeleted,
	"DELETING": SteeringPolicyLifecycleStateDeleting,
}

// GetSteeringPolicyLifecycleStateEnumValues Enumerates the set of values for SteeringPolicyLifecycleStateEnum
func GetSteeringPolicyLifecycleStateEnumValues() []SteeringPolicyLifecycleStateEnum {
	values := make([]SteeringPolicyLifecycleStateEnum, 0)
	for _, v := range mappingSteeringPolicyLifecycleState {
		values = append(values, v)
	}
	return values
}
