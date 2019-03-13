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

// SteeringPolicyLimitRule The representation of SteeringPolicyLimitRule
type SteeringPolicyLimitRule struct {

	// Your description of the rule's purpose and/or behavior.
	Description *string `mandatory:"false" json:"description"`

	Cases []SteeringPolicyLimitRuleCase `mandatory:"false" json:"cases"`

	// Defines a default count if `cases` is not defined for the rule or a matching case does
	// not define `count`. `defaultCount` is **not** applied if `cases` is defined and there
	// are no matching cases.
	DefaultCount *int `mandatory:"false" json:"defaultCount"`
}

//GetDescription returns Description
func (m SteeringPolicyLimitRule) GetDescription() *string {
	return m.Description
}

func (m SteeringPolicyLimitRule) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m SteeringPolicyLimitRule) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeSteeringPolicyLimitRule SteeringPolicyLimitRule
	s := struct {
		DiscriminatorParam string `json:"ruleType"`
		MarshalTypeSteeringPolicyLimitRule
	}{
		"LIMIT",
		(MarshalTypeSteeringPolicyLimitRule)(m),
	}

	return json.Marshal(&s)
}
