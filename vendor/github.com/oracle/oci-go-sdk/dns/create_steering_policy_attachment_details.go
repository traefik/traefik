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

// CreateSteeringPolicyAttachmentDetails The body for defining an attachment between a steering policy and a domain.
// An attachment occludes all records at its domain that are of a covered rtype, constructing
// DNS responses from its steering policy rather than from those domain records.
// The attachment will cover every rtype that matches the rtype of an answer in its policy, and
// will cover all address rtypes (e.g., A and AAAA) if the policy includes at least one CNAME
// answer.
// A domain can have at most one attachment covering any given rtype.
type CreateSteeringPolicyAttachmentDetails struct {

	// The OCID of the attached steering policy.
	SteeringPolicyId *string `mandatory:"true" json:"steeringPolicyId"`

	// The OCID of the attached zone.
	ZoneId *string `mandatory:"true" json:"zoneId"`

	// The attached domain within the attached zone.
	DomainName *string `mandatory:"true" json:"domainName"`

	// A user-friendly name for the steering policy attachment.
	// Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m CreateSteeringPolicyAttachmentDetails) String() string {
	return common.PointerString(m)
}
