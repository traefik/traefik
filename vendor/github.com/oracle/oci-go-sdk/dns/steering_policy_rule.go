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

// SteeringPolicyRule Configuration for sorting and/or filtering the list of remaining candidate answers, subject to
// rule type and the values of type-specific parameters and/or data associated with answers.
// A rule may optionally include a sequence of cases, each with an optional `caseCondition`
// expression.  If it does, the first case with a matching `caseCondition` or with no
// `caseCondition` at all is used to set rule parameter values and/or answer-associated data,
// and the rule will be ignored during processing of any request that does not match any case.
// Rules without a sequence of cases are processed unconditionally, and rules with an _empty_
// sequence of cases are **ignored** unconditionally.
// Data is associated with answers one-by-one in a similar fashion—for each answer, the first
// answerData item with a matching `answerCondition` or with no `answerCondition` at all is used
// to associate data with the answer, and the absence of any such item associates with the answer
// a default value.  Rule-level default answer data is always processed, but case-level answer
// data will override it on a per-answer basis.
// To prevent empty responses, any attempt to filter away all answers is suppressed at runtime.
type SteeringPolicyRule interface {

	// Your description of the rule's purpose and/or behavior.
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
