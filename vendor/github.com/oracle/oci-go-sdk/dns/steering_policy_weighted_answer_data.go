// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.cloud.oracle.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
)

// SteeringPolicyWeightedAnswerData The representation of SteeringPolicyWeightedAnswerData
type SteeringPolicyWeightedAnswerData struct {

	// The weight assigned to the set of selected answers. Answers with a higher weight will be served
	// more frequently. Answers can be given a value between `0` and `255`.
	Value *int `mandatory:"true" json:"value"`

	// An expression that is used to select a set of answers that match a condition. For example, answers with matching pool properties.
	AnswerCondition *string `mandatory:"false" json:"answerCondition"`
}

func (m SteeringPolicyWeightedAnswerData) String() string {
	return common.PointerString(m)
}
