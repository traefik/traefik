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

// Zone A DNS zone.
// *Warning:* Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type Zone struct {

	// The name of the zone.
	Name *string `mandatory:"false" json:"name"`

	// The type of the zone. Must be either `PRIMARY` or `SECONDARY`.
	ZoneType ZoneZoneTypeEnum `mandatory:"false" json:"zoneType,omitempty"`

	// The OCID of the compartment containing the zone.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// Simple key-value pair that is applied without any predefined name, type, or scope.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of predefined tag keys. These predefined keys are scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// External master servers for the zone. `externalMasters` becomes a
	// required parameter when the `zoneType` value is `SECONDARY`.
	ExternalMasters []ExternalMaster `mandatory:"false" json:"externalMasters"`

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
	LifecycleState ZoneLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The authoritative nameservers for the zone.
	Nameservers []Nameserver `mandatory:"false" json:"nameservers"`
}

func (m Zone) String() string {
	return common.PointerString(m)
}

// ZoneZoneTypeEnum Enum with underlying type: string
type ZoneZoneTypeEnum string

// Set of constants representing the allowable values for ZoneZoneTypeEnum
const (
	ZoneZoneTypePrimary   ZoneZoneTypeEnum = "PRIMARY"
	ZoneZoneTypeSecondary ZoneZoneTypeEnum = "SECONDARY"
)

var mappingZoneZoneType = map[string]ZoneZoneTypeEnum{
	"PRIMARY":   ZoneZoneTypePrimary,
	"SECONDARY": ZoneZoneTypeSecondary,
}

// GetZoneZoneTypeEnumValues Enumerates the set of values for ZoneZoneTypeEnum
func GetZoneZoneTypeEnumValues() []ZoneZoneTypeEnum {
	values := make([]ZoneZoneTypeEnum, 0)
	for _, v := range mappingZoneZoneType {
		values = append(values, v)
	}
	return values
}

// ZoneLifecycleStateEnum Enum with underlying type: string
type ZoneLifecycleStateEnum string

// Set of constants representing the allowable values for ZoneLifecycleStateEnum
const (
	ZoneLifecycleStateActive   ZoneLifecycleStateEnum = "ACTIVE"
	ZoneLifecycleStateCreating ZoneLifecycleStateEnum = "CREATING"
	ZoneLifecycleStateDeleted  ZoneLifecycleStateEnum = "DELETED"
	ZoneLifecycleStateDeleting ZoneLifecycleStateEnum = "DELETING"
	ZoneLifecycleStateFailed   ZoneLifecycleStateEnum = "FAILED"
)

var mappingZoneLifecycleState = map[string]ZoneLifecycleStateEnum{
	"ACTIVE":   ZoneLifecycleStateActive,
	"CREATING": ZoneLifecycleStateCreating,
	"DELETED":  ZoneLifecycleStateDeleted,
	"DELETING": ZoneLifecycleStateDeleting,
	"FAILED":   ZoneLifecycleStateFailed,
}

// GetZoneLifecycleStateEnumValues Enumerates the set of values for ZoneLifecycleStateEnum
func GetZoneLifecycleStateEnumValues() []ZoneLifecycleStateEnum {
	values := make([]ZoneLifecycleStateEnum, 0)
	for _, v := range mappingZoneLifecycleState {
		values = append(values, v)
	}
	return values
}
