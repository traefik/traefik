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

// UpdateSteeringPolicyAttachmentDetails The body for updating a steering policy attachment.
//
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type UpdateSteeringPolicyAttachmentDetails struct {

	// A user-friendly name for the steering policy attachment.
	// Does not have to be unique and can be changed.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m UpdateSteeringPolicyAttachmentDetails) String() string {
	return common.PointerString(m)
}
