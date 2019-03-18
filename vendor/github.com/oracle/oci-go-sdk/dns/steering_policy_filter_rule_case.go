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

// SteeringPolicyFilterRuleCase The representation of SteeringPolicyFilterRuleCase
type SteeringPolicyFilterRuleCase struct {
	CaseCondition *string `mandatory:"false" json:"caseCondition"`

	AnswerData []SteeringPolicyFilterAnswerData `mandatory:"false" json:"answerData"`
}

func (m SteeringPolicyFilterRuleCase) String() string {
	return common.PointerString(m)
}
