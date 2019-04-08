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

// SteeringPolicyHealthRule The representation of SteeringPolicyHealthRule
type SteeringPolicyHealthRule struct {

	// Your description of the rule's purpose and/or behavior.
	Description *string `mandatory:"false" json:"description"`

	Cases []SteeringPolicyHealthRuleCase `mandatory:"false" json:"cases"`
}

//GetDescription returns Description
func (m SteeringPolicyHealthRule) GetDescription() *string {
	return m.Description
}

func (m SteeringPolicyHealthRule) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m SteeringPolicyHealthRule) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeSteeringPolicyHealthRule SteeringPolicyHealthRule
	s := struct {
		DiscriminatorParam string `json:"ruleType"`
		MarshalTypeSteeringPolicyHealthRule
	}{
		"HEALTH",
		(MarshalTypeSteeringPolicyHealthRule)(m),
	}

	return json.Marshal(&s)
}
