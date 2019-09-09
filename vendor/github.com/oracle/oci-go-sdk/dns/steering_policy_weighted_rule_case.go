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

// SteeringPolicyWeightedRuleCase The representation of SteeringPolicyWeightedRuleCase
type SteeringPolicyWeightedRuleCase struct {

	// An expression that uses conditions at the time of a DNS query to indicate
	// whether a case matches. Conditions may include the geographical location, IP
	// subnet, or ASN the DNS query originated. **Example:** If you have an
	// office that uses the subnet `192.0.2.0/24` you could use a `caseCondition`
	// expression `query.client.subnet in ('192.0.2.0/24')` to define a case that
	// matches queries from that office.
	CaseCondition *string `mandatory:"false" json:"caseCondition"`

	// An array of `SteeringPolicyWeightedAnswerData` objects.
	AnswerData []SteeringPolicyWeightedAnswerData `mandatory:"false" json:"answerData"`
}

func (m SteeringPolicyWeightedRuleCase) String() string {
	return common.PointerString(m)
}
