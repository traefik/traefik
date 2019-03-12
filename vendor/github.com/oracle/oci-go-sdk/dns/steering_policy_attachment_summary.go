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

// SteeringPolicyAttachmentSummary An attachment between a steering policy and a domain.
type SteeringPolicyAttachmentSummary struct {

	// The OCID of the attached steering policy.
	SteeringPolicyId *string `mandatory:"false" json:"steeringPolicyId"`

	// The OCID of the attached zone.
	ZoneId *string `mandatory:"false" json:"zoneId"`

	// The attached domain within the attached zone.
	DomainName *string `mandatory:"false" json:"domainName"`

	// A user-friendly name for the steering policy attachment.
	// Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The record types covered by the attachment at the domain. The set of record types is
	// determined by aggregating the record types from the answers defined in the steering
	// policy.
	Rtypes []string `mandatory:"false" json:"rtypes"`

	// The OCID of the compartment containing the steering policy attachment.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The canonical absolute URL of the resource.
	Self *string `mandatory:"false" json:"self"`

	// The OCID of the resource.
	Id *string `mandatory:"false" json:"id"`

	// The date and time the resource was created in "YYYY-MM-ddThh:mmZ" format
	// with a Z offset, as defined by RFC 3339.
	// **Example:** `2016-07-22T17:23:59:60Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The current state of the resource.
	LifecycleState SteeringPolicyAttachmentSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m SteeringPolicyAttachmentSummary) String() string {
	return common.PointerString(m)
}

// SteeringPolicyAttachmentSummaryLifecycleStateEnum Enum with underlying type: string
type SteeringPolicyAttachmentSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for SteeringPolicyAttachmentSummaryLifecycleStateEnum
const (
	SteeringPolicyAttachmentSummaryLifecycleStateCreating SteeringPolicyAttachmentSummaryLifecycleStateEnum = "CREATING"
	SteeringPolicyAttachmentSummaryLifecycleStateActive   SteeringPolicyAttachmentSummaryLifecycleStateEnum = "ACTIVE"
	SteeringPolicyAttachmentSummaryLifecycleStateDeleting SteeringPolicyAttachmentSummaryLifecycleStateEnum = "DELETING"
)

var mappingSteeringPolicyAttachmentSummaryLifecycleState = map[string]SteeringPolicyAttachmentSummaryLifecycleStateEnum{
	"CREATING": SteeringPolicyAttachmentSummaryLifecycleStateCreating,
	"ACTIVE":   SteeringPolicyAttachmentSummaryLifecycleStateActive,
	"DELETING": SteeringPolicyAttachmentSummaryLifecycleStateDeleting,
}

// GetSteeringPolicyAttachmentSummaryLifecycleStateEnumValues Enumerates the set of values for SteeringPolicyAttachmentSummaryLifecycleStateEnum
func GetSteeringPolicyAttachmentSummaryLifecycleStateEnumValues() []SteeringPolicyAttachmentSummaryLifecycleStateEnum {
	values := make([]SteeringPolicyAttachmentSummaryLifecycleStateEnum, 0)
	for _, v := range mappingSteeringPolicyAttachmentSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
