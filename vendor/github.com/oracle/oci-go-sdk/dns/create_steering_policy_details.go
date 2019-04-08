// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// CreateSteeringPolicyDetails The body for defining a new steering policy.
// *Warning:* Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type CreateSteeringPolicyDetails struct {

	// The OCID of the compartment containing the steering policy.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name for the steering policy.
	// Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

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
	Template CreateSteeringPolicyDetailsTemplateEnum `mandatory:"true" json:"template"`

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

	// Simple key-value pair that is applied without any predefined name, type, or scope.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of predefined tag keys. These predefined keys are scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The set of all answers that can potentially issue from the steering policy.
	Answers []SteeringPolicyAnswer `mandatory:"false" json:"answers"`

	// The pipeline of rules that will be processed in sequence to reduce the pool of answers
	// to a response for any given request.
	// The first rule receives a shuffled list of all answers, and every other rule receives
	// the list of answers emitted by the one preceding it. The last rule populates the
	// response.
	Rules []SteeringPolicyRule `mandatory:"false" json:"rules"`
}

func (m CreateSteeringPolicyDetails) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *CreateSteeringPolicyDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Ttl                  *int                                    `json:"ttl"`
		HealthCheckMonitorId *string                                 `json:"healthCheckMonitorId"`
		FreeformTags         map[string]string                       `json:"freeformTags"`
		DefinedTags          map[string]map[string]interface{}       `json:"definedTags"`
		Answers              []SteeringPolicyAnswer                  `json:"answers"`
		Rules                []steeringpolicyrule                    `json:"rules"`
		CompartmentId        *string                                 `json:"compartmentId"`
		DisplayName          *string                                 `json:"displayName"`
		Template             CreateSteeringPolicyDetailsTemplateEnum `json:"template"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	m.Ttl = model.Ttl
	m.HealthCheckMonitorId = model.HealthCheckMonitorId
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
	m.CompartmentId = model.CompartmentId
	m.DisplayName = model.DisplayName
	m.Template = model.Template
	return
}

// CreateSteeringPolicyDetailsTemplateEnum Enum with underlying type: string
type CreateSteeringPolicyDetailsTemplateEnum string

// Set of constants representing the allowable values for CreateSteeringPolicyDetailsTemplateEnum
const (
	CreateSteeringPolicyDetailsTemplateFailover    CreateSteeringPolicyDetailsTemplateEnum = "FAILOVER"
	CreateSteeringPolicyDetailsTemplateLoadBalance CreateSteeringPolicyDetailsTemplateEnum = "LOAD_BALANCE"
	CreateSteeringPolicyDetailsTemplateRouteByGeo  CreateSteeringPolicyDetailsTemplateEnum = "ROUTE_BY_GEO"
	CreateSteeringPolicyDetailsTemplateRouteByAsn  CreateSteeringPolicyDetailsTemplateEnum = "ROUTE_BY_ASN"
	CreateSteeringPolicyDetailsTemplateRouteByIp   CreateSteeringPolicyDetailsTemplateEnum = "ROUTE_BY_IP"
	CreateSteeringPolicyDetailsTemplateCustom      CreateSteeringPolicyDetailsTemplateEnum = "CUSTOM"
)

var mappingCreateSteeringPolicyDetailsTemplate = map[string]CreateSteeringPolicyDetailsTemplateEnum{
	"FAILOVER":     CreateSteeringPolicyDetailsTemplateFailover,
	"LOAD_BALANCE": CreateSteeringPolicyDetailsTemplateLoadBalance,
	"ROUTE_BY_GEO": CreateSteeringPolicyDetailsTemplateRouteByGeo,
	"ROUTE_BY_ASN": CreateSteeringPolicyDetailsTemplateRouteByAsn,
	"ROUTE_BY_IP":  CreateSteeringPolicyDetailsTemplateRouteByIp,
	"CUSTOM":       CreateSteeringPolicyDetailsTemplateCustom,
}

// GetCreateSteeringPolicyDetailsTemplateEnumValues Enumerates the set of values for CreateSteeringPolicyDetailsTemplateEnum
func GetCreateSteeringPolicyDetailsTemplateEnumValues() []CreateSteeringPolicyDetailsTemplateEnum {
	values := make([]CreateSteeringPolicyDetailsTemplateEnum, 0)
	for _, v := range mappingCreateSteeringPolicyDetailsTemplate {
		values = append(values, v)
	}
	return values
}
