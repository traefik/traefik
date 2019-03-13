// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListSteeringPolicyAttachmentsRequest wrapper for the ListSteeringPolicyAttachments operation
type ListSteeringPolicyAttachmentsRequest struct {

	// The OCID of the compartment the resource belongs to.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The maximum number of items to return in a page of the collection.
	Limit *int64 `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The OCID of a resource.
	Id *string `mandatory:"false" contributesTo:"query" name:"id"`

	// The displayName of a resource.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// Search by steering policy OCID.
	// Will match any resource whose steering policy id matches the provided value.
	SteeringPolicyId *string `mandatory:"false" contributesTo:"query" name:"steeringPolicyId"`

	// Search by zone OCID.
	// Will match any resource whose zone id matches the provided value.
	ZoneId *string `mandatory:"false" contributesTo:"query" name:"zoneId"`

	// Search by domain.
	// Will match any record whose domain (case-insensitive) equals the provided value.
	Domain *string `mandatory:"false" contributesTo:"query" name:"domain"`

	// Search by domain.
	// Will match any record whose domain (case-insensitive) contains the provided value.
	DomainContains *string `mandatory:"false" contributesTo:"query" name:"domainContains"`

	// An RFC 3339 (https://www.ietf.org/rfc/rfc3339.txt) timestamp that states
	// all returned resources were created on or after the indicated time.
	TimeCreatedGreaterThanOrEqualTo *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeCreatedGreaterThanOrEqualTo"`

	// An RFC 3339 (https://www.ietf.org/rfc/rfc3339.txt) timestamp that states
	// all returned resources were created before the indicated time.
	TimeCreatedLessThan *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeCreatedLessThan"`

	// The state of a resource.
	LifecycleState SteeringPolicyAttachmentSummaryLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// The field by which to sort steering policy attachments.
	SortBy ListSteeringPolicyAttachmentsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The order to sort the resources.
	SortOrder ListSteeringPolicyAttachmentsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListSteeringPolicyAttachmentsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListSteeringPolicyAttachmentsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListSteeringPolicyAttachmentsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListSteeringPolicyAttachmentsResponse wrapper for the ListSteeringPolicyAttachments operation
type ListSteeringPolicyAttachmentsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []SteeringPolicyAttachmentSummary instances
	Items []SteeringPolicyAttachmentSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works,
	// see List Pagination (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// The total number of items that match the query.
	OpcTotalItems *int `presentIn:"header" name:"opc-total-items"`

	// Unique Oracle-assigned identifier for the request. If you need to
	// contact Oracle about a particular request, please provide the request
	// ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListSteeringPolicyAttachmentsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListSteeringPolicyAttachmentsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListSteeringPolicyAttachmentsSortByEnum Enum with underlying type: string
type ListSteeringPolicyAttachmentsSortByEnum string

// Set of constants representing the allowable values for ListSteeringPolicyAttachmentsSortByEnum
const (
	ListSteeringPolicyAttachmentsSortByDisplayname ListSteeringPolicyAttachmentsSortByEnum = "displayName"
	ListSteeringPolicyAttachmentsSortByTimecreated ListSteeringPolicyAttachmentsSortByEnum = "timeCreated"
	ListSteeringPolicyAttachmentsSortByDomainname  ListSteeringPolicyAttachmentsSortByEnum = "domainName"
)

var mappingListSteeringPolicyAttachmentsSortBy = map[string]ListSteeringPolicyAttachmentsSortByEnum{
	"displayName": ListSteeringPolicyAttachmentsSortByDisplayname,
	"timeCreated": ListSteeringPolicyAttachmentsSortByTimecreated,
	"domainName":  ListSteeringPolicyAttachmentsSortByDomainname,
}

// GetListSteeringPolicyAttachmentsSortByEnumValues Enumerates the set of values for ListSteeringPolicyAttachmentsSortByEnum
func GetListSteeringPolicyAttachmentsSortByEnumValues() []ListSteeringPolicyAttachmentsSortByEnum {
	values := make([]ListSteeringPolicyAttachmentsSortByEnum, 0)
	for _, v := range mappingListSteeringPolicyAttachmentsSortBy {
		values = append(values, v)
	}
	return values
}

// ListSteeringPolicyAttachmentsSortOrderEnum Enum with underlying type: string
type ListSteeringPolicyAttachmentsSortOrderEnum string

// Set of constants representing the allowable values for ListSteeringPolicyAttachmentsSortOrderEnum
const (
	ListSteeringPolicyAttachmentsSortOrderAsc  ListSteeringPolicyAttachmentsSortOrderEnum = "ASC"
	ListSteeringPolicyAttachmentsSortOrderDesc ListSteeringPolicyAttachmentsSortOrderEnum = "DESC"
)

var mappingListSteeringPolicyAttachmentsSortOrder = map[string]ListSteeringPolicyAttachmentsSortOrderEnum{
	"ASC":  ListSteeringPolicyAttachmentsSortOrderAsc,
	"DESC": ListSteeringPolicyAttachmentsSortOrderDesc,
}

// GetListSteeringPolicyAttachmentsSortOrderEnumValues Enumerates the set of values for ListSteeringPolicyAttachmentsSortOrderEnum
func GetListSteeringPolicyAttachmentsSortOrderEnumValues() []ListSteeringPolicyAttachmentsSortOrderEnum {
	values := make([]ListSteeringPolicyAttachmentsSortOrderEnum, 0)
	for _, v := range mappingListSteeringPolicyAttachmentsSortOrder {
		values = append(values, v)
	}
	return values
}
