// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.cloud.oracle.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

// SortOrderEnum Enum with underlying type: string
type SortOrderEnum string

// Set of constants representing the allowable values for SortOrderEnum
const (
	SortOrderAsc  SortOrderEnum = "ASC"
	SortOrderDesc SortOrderEnum = "DESC"
)

var mappingSortOrder = map[string]SortOrderEnum{
	"ASC":  SortOrderAsc,
	"DESC": SortOrderDesc,
}

// GetSortOrderEnumValues Enumerates the set of values for SortOrderEnum
func GetSortOrderEnumValues() []SortOrderEnum {
	values := make([]SortOrderEnum, 0)
	for _, v := range mappingSortOrder {
		values = append(values, v)
	}
	return values
}
