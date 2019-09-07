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

// SteeringPolicyAnswer DNS record data with metadata for processing in a steering policy.
//
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type SteeringPolicyAnswer struct {

	// A user-friendly name for the answer, unique within the steering policy.
	// An answer's `name` property can be referenced in `answerCondition` properties
	// of rules using `answer.name`.
	// **Example:**
	//   "rules": [
	//     {
	//       "ruleType": "FILTER",
	//       "defaultAnswerData":  [
	//         {
	//           "answerCondition": "answer.name == 'server 1'",
	//           "shouldKeep": true
	//         }
	//       ]
	//     }
	//   ]
	Name *string `mandatory:"true" json:"name"`

	// The canonical name for the record's type. Only A, AAAA, and CNAME are supported. For more
	// information, see Supported DNS Resource Record Types (https://docs.cloud.oracle.com/iaas/Content/DNS/Reference/supporteddnsresource.htm).
	Rtype *string `mandatory:"true" json:"rtype"`

	// The record's data, as whitespace-delimited tokens in
	// type-specific presentation format. All RDATA is normalized and the
	// returned presentation of your RDATA may differ from its initial input.
	// For more information about RDATA, see Supported DNS Resource Record Types (https://docs.cloud.oracle.com/iaas/Content/DNS/Reference/supporteddnsresource.htm).
	Rdata *string `mandatory:"true" json:"rdata"`

	// The freeform name of a group of one or more records in which this record is included,
	// such as "LAX data center". An answer's `pool` property can be referenced in `answerCondition`
	// properties of rules using `answer.pool`.
	// **Example:**
	//   "rules": [
	//     {
	//       "ruleType": "FILTER",
	//       "defaultAnswerData":  [
	//         {
	//           "answerCondition": "answer.pool == 'US East Servers'",
	//           "shouldKeep": true
	//         }
	//       ]
	//     }
	//   ]
	Pool *string `mandatory:"false" json:"pool"`

	// Set this property to `true` to indicate that the answer is administratively disabled,
	// such as when the corresponding server is down for maintenance. An answer's `isDisabled`
	// property can be referenced in `answerCondition` properties in rules using `answer.isDisabled`.
	// **Example:**
	//   "rules": [
	//     {
	//       "ruleType": "FILTER",
	//       "defaultAnswerData": [
	//         {
	//           "answerCondition": "answer.isDisabled != true",
	//           "shouldKeep": true
	//         }
	//       ]
	//     },
	IsDisabled *bool `mandatory:"false" json:"isDisabled"`
}

func (m SteeringPolicyAnswer) String() string {
	return common.PointerString(m)
}
