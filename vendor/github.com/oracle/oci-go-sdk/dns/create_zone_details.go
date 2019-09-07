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

// CreateZoneDetails The body for defining a new zone.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type CreateZoneDetails struct {

	// The name of the zone.
	Name *string `mandatory:"true" json:"name"`

	// The type of the zone. Must be either `PRIMARY` or `SECONDARY`.
	ZoneType CreateZoneDetailsZoneTypeEnum `mandatory:"true" json:"zoneType"`

	// The OCID of the compartment containing the zone.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	//
	// **Example:** `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	//
	// **Example:** `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// External master servers for the zone. `externalMasters` becomes a
	// required parameter when the `zoneType` value is `SECONDARY`.
	ExternalMasters []ExternalMaster `mandatory:"false" json:"externalMasters"`
}

func (m CreateZoneDetails) String() string {
	return common.PointerString(m)
}

// CreateZoneDetailsZoneTypeEnum Enum with underlying type: string
type CreateZoneDetailsZoneTypeEnum string

// Set of constants representing the allowable values for CreateZoneDetailsZoneTypeEnum
const (
	CreateZoneDetailsZoneTypePrimary   CreateZoneDetailsZoneTypeEnum = "PRIMARY"
	CreateZoneDetailsZoneTypeSecondary CreateZoneDetailsZoneTypeEnum = "SECONDARY"
)

var mappingCreateZoneDetailsZoneType = map[string]CreateZoneDetailsZoneTypeEnum{
	"PRIMARY":   CreateZoneDetailsZoneTypePrimary,
	"SECONDARY": CreateZoneDetailsZoneTypeSecondary,
}

// GetCreateZoneDetailsZoneTypeEnumValues Enumerates the set of values for CreateZoneDetailsZoneTypeEnum
func GetCreateZoneDetailsZoneTypeEnumValues() []CreateZoneDetailsZoneTypeEnum {
	values := make([]CreateZoneDetailsZoneTypeEnum, 0)
	for _, v := range mappingCreateZoneDetailsZoneType {
		values = append(values, v)
	}
	return values
}
