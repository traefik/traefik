// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
)

// SteeringPolicyPriorityAnswerData The representation of SteeringPolicyPriorityAnswerData
type SteeringPolicyPriorityAnswerData struct {
	AnswerCondition *string `mandatory:"false" json:"answerCondition"`

	Value *int `mandatory:"false" json:"value"`
}

func (m SteeringPolicyPriorityAnswerData) String() string {
	return common.PointerString(m)
}
