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

// ChangeZoneCompartmentDetails The representation of ChangeZoneCompartmentDetails
type ChangeZoneCompartmentDetails struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment
	// into which the zone should be moved.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`
}

func (m ChangeZoneCompartmentDetails) String() string {
	return common.PointerString(m)
}
