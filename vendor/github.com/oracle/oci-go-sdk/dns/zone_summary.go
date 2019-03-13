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

// ZoneSummary A DNS zone.
// *Warning:* Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type ZoneSummary struct {

	// The name of the zone.
	Name *string `mandatory:"false" json:"name"`

	// The type of the zone. Must be either `PRIMARY` or `SECONDARY`.
	ZoneType ZoneSummaryZoneTypeEnum `mandatory:"false" json:"zoneType,omitempty"`

	// The OCID of the compartment containing the zone.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// Simple key-value pair that is applied without any predefined name, type, or scope.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of predefined tag keys. These predefined keys are scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The canonical absolute URL of the resource.
	Self *string `mandatory:"false" json:"self"`

	// The OCID of the zone.
	Id *string `mandatory:"false" json:"id"`

	// The date and time the resource was created in "YYYY-MM-ddThh:mmZ" format
	// with a Z offset, as defined by RFC 3339.
	// **Example:** `2016-07-22T17:23:59:60Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Version is the never-repeating, totally-orderable, version of the
	// zone, from which the serial field of the zone's SOA record is
	// derived.
	Version *string `mandatory:"false" json:"version"`

	// The current serial of the zone. As seen in the zone's SOA record.
	Serial *int64 `mandatory:"false" json:"serial"`

	// The current state of the zone resource.
	LifecycleState ZoneSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m ZoneSummary) String() string {
	return common.PointerString(m)
}

// ZoneSummaryZoneTypeEnum Enum with underlying type: string
type ZoneSummaryZoneTypeEnum string

// Set of constants representing the allowable values for ZoneSummaryZoneTypeEnum
const (
	ZoneSummaryZoneTypePrimary   ZoneSummaryZoneTypeEnum = "PRIMARY"
	ZoneSummaryZoneTypeSecondary ZoneSummaryZoneTypeEnum = "SECONDARY"
)

var mappingZoneSummaryZoneType = map[string]ZoneSummaryZoneTypeEnum{
	"PRIMARY":   ZoneSummaryZoneTypePrimary,
	"SECONDARY": ZoneSummaryZoneTypeSecondary,
}

// GetZoneSummaryZoneTypeEnumValues Enumerates the set of values for ZoneSummaryZoneTypeEnum
func GetZoneSummaryZoneTypeEnumValues() []ZoneSummaryZoneTypeEnum {
	values := make([]ZoneSummaryZoneTypeEnum, 0)
	for _, v := range mappingZoneSummaryZoneType {
		values = append(values, v)
	}
	return values
}

// ZoneSummaryLifecycleStateEnum Enum with underlying type: string
type ZoneSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for ZoneSummaryLifecycleStateEnum
const (
	ZoneSummaryLifecycleStateActive   ZoneSummaryLifecycleStateEnum = "ACTIVE"
	ZoneSummaryLifecycleStateCreating ZoneSummaryLifecycleStateEnum = "CREATING"
	ZoneSummaryLifecycleStateDeleted  ZoneSummaryLifecycleStateEnum = "DELETED"
	ZoneSummaryLifecycleStateDeleting ZoneSummaryLifecycleStateEnum = "DELETING"
	ZoneSummaryLifecycleStateFailed   ZoneSummaryLifecycleStateEnum = "FAILED"
)

var mappingZoneSummaryLifecycleState = map[string]ZoneSummaryLifecycleStateEnum{
	"ACTIVE":   ZoneSummaryLifecycleStateActive,
	"CREATING": ZoneSummaryLifecycleStateCreating,
	"DELETED":  ZoneSummaryLifecycleStateDeleted,
	"DELETING": ZoneSummaryLifecycleStateDeleting,
	"FAILED":   ZoneSummaryLifecycleStateFailed,
}

// GetZoneSummaryLifecycleStateEnumValues Enumerates the set of values for ZoneSummaryLifecycleStateEnum
func GetZoneSummaryLifecycleStateEnumValues() []ZoneSummaryLifecycleStateEnum {
	values := make([]ZoneSummaryLifecycleStateEnum, 0)
	for _, v := range mappingZoneSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
