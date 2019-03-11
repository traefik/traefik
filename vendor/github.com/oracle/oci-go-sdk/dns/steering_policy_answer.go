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

// SteeringPolicyAnswer DNS record data with metadata for processing in a steering policy.
// *Warning:* Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type SteeringPolicyAnswer struct {

	// A user-friendly name for the answer, unique within the steering policy.
	Name *string `mandatory:"true" json:"name"`

	// The canonical name for the record's type. Only A, AAAA, and CNAME are supported. For more
	// information, see Resource Record (RR) TYPEs (https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4).
	Rtype *string `mandatory:"true" json:"rtype"`

	// The record's data, as whitespace-delimited tokens in
	// type-specific presentation format.
	Rdata *string `mandatory:"true" json:"rdata"`

	// The freeform name of a group of one or more records (e.g., a data center or a geographic
	// region) in which this one is included.
	Pool *string `mandatory:"false" json:"pool"`

	// Whether or not an answer should be excluded from responses, e.g. because the corresponding
	// server is down for maintenance. Note, however, that such filtering is not automatic and
	// will only take place if a rule implements it.
	IsDisabled *bool `mandatory:"false" json:"isDisabled"`
}

func (m SteeringPolicyAnswer) String() string {
	return common.PointerString(m)
}
