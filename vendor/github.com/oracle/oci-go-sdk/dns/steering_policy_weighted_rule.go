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

// SteeringPolicyWeightedRule The representation of SteeringPolicyWeightedRule
type SteeringPolicyWeightedRule struct {

	// A user-defined description of the rule's purpose or behavior.
	Description *string `mandatory:"false" json:"description"`

	// An array of `caseConditions`. A rule may optionally include a sequence of cases defining alternate
	// configurations for how it should behave during processing for any given DNS query. When a rule has
	// no sequence of `cases`, it is always evaluated with the same configuration during processing. When
	// a rule has an empty sequence of `cases`, it is always ignored during processing. When a rule has a
	// non-empty sequence of `cases`, its behavior during processing is configured by the first matching
	// `case` in the sequence. When a rule has no matching cases the rule is ignored. A rule case with no
	// `caseCondition` always matches. A rule case with a `caseCondition` matches only when that expression
	// evaluates to true for the given query.
	Cases []SteeringPolicyWeightedRuleCase `mandatory:"false" json:"cases"`

	// Defines a default set of answer conditions and values that are applied to an answer when
	// `cases` is not defined for the rule or a matching case does not have any matching
	// `answerCondition`s in its `answerData`. `defaultAnswerData` is not applied if `cases` is
	// defined and there are no matching cases. In this scenario, the next rule will be processed.
	DefaultAnswerData []SteeringPolicyWeightedAnswerData `mandatory:"false" json:"defaultAnswerData"`
}

//GetDescription returns Description
func (m SteeringPolicyWeightedRule) GetDescription() *string {
	return m.Description
}

func (m SteeringPolicyWeightedRule) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m SteeringPolicyWeightedRule) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeSteeringPolicyWeightedRule SteeringPolicyWeightedRule
	s := struct {
		DiscriminatorParam string `json:"ruleType"`
		MarshalTypeSteeringPolicyWeightedRule
	}{
		"WEIGHTED",
		(MarshalTypeSteeringPolicyWeightedRule)(m),
	}

	return json.Marshal(&s)
}
