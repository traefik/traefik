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

// SteeringPolicyRule The configuration of the sorting and filtering behaviors in a steering policy. Rules can
// filter and sort answers based on weight, priority, endpoint health, and other data.
//
// A rule may optionally include a sequence of cases, each with an optional `caseCondition`
// expression. Cases allow a sequence of conditions to be defined that will apply different
// parameters to the rule when the conditions are met. For more information about cases,
// see Traffic Management API Guide (https://docs.cloud.oracle.com/iaas/Content/TrafficManagement/Concepts/trafficmanagementapi.htm).
//
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type SteeringPolicyRule interface {

	// A user-defined description of the rule's purpose or behavior.
	GetDescription() *string
}

type steeringpolicyrule struct {
	JsonData    []byte
	Description *string `mandatory:"false" json:"description"`
	RuleType    string  `json:"ruleType"`
}

// UnmarshalJSON unmarshals json
func (m *steeringpolicyrule) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalersteeringpolicyrule steeringpolicyrule
	s := struct {
		Model Unmarshalersteeringpolicyrule
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Description = s.Model.Description
	m.RuleType = s.Model.RuleType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *steeringpolicyrule) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.RuleType {
	case "FILTER":
		mm := SteeringPolicyFilterRule{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "WEIGHTED":
		mm := SteeringPolicyWeightedRule{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "LIMIT":
		mm := SteeringPolicyLimitRule{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "HEALTH":
		mm := SteeringPolicyHealthRule{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "PRIORITY":
		mm := SteeringPolicyPriorityRule{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

//GetDescription returns Description
func (m steeringpolicyrule) GetDescription() *string {
	return m.Description
}

func (m steeringpolicyrule) String() string {
	return common.PointerString(m)
}
