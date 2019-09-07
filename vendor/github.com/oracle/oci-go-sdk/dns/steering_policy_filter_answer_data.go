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

// SteeringPolicyFilterAnswerData The representation of SteeringPolicyFilterAnswerData
type SteeringPolicyFilterAnswerData struct {

	// An expression that is used to select a set of answers that match a condition. For example, answers with matching pool properties.
	AnswerCondition *string `mandatory:"false" json:"answerCondition"`

	// Keeps the answer only if the value is `true`.
	ShouldKeep *bool `mandatory:"false" json:"shouldKeep"`
}

func (m SteeringPolicyFilterAnswerData) String() string {
	return common.PointerString(m)
}
