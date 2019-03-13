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

// SteeringPolicyFilterRule The representation of SteeringPolicyFilterRule
type SteeringPolicyFilterRule struct {

	// Your description of the rule's purpose and/or behavior.
	Description *string `mandatory:"false" json:"description"`

	Cases []SteeringPolicyFilterRuleCase `mandatory:"false" json:"cases"`

	// Defines a default set of answer conditions and values that are applied to an answer when
	// `cases` is not defined for the rule or a matching case does not have any matching
	// `answerCondition`s in its `answerData`. `defaultAnswerData` is **not** applied if `cases` is
	// defined and there are no matching cases.
	DefaultAnswerData []SteeringPolicyFilterAnswerData `mandatory:"false" json:"defaultAnswerData"`
}

//GetDescription returns Description
func (m SteeringPolicyFilterRule) GetDescription() *string {
	return m.Description
}

func (m SteeringPolicyFilterRule) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m SteeringPolicyFilterRule) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeSteeringPolicyFilterRule SteeringPolicyFilterRule
	s := struct {
		DiscriminatorParam string `json:"ruleType"`
		MarshalTypeSteeringPolicyFilterRule
	}{
		"FILTER",
		(MarshalTypeSteeringPolicyFilterRule)(m),
	}

	return json.Marshal(&s)
}
