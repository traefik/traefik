// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//DnsClient a client for Dns
type DnsClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewDnsClientWithConfigurationProvider Creates a new default Dns client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewDnsClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client DnsClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = DnsClient{BaseClient: baseClient}
	client.BasePath = "20180115"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *DnsClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).Endpoint("dns")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *DnsClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
	if ok, err := common.IsConfigurationProviderValid(configProvider); !ok {
		return err
	}

	// Error has been checked already
	region, _ := configProvider.Region()
	client.SetRegion(region)
	client.config = &configProvider
	return nil
}

// ConfigurationProvider the ConfigurationProvider used in this client, or null if none set
func (client *DnsClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// CreateSteeringPolicy Creates a new steering policy in the specified compartment.
func (client DnsClient) CreateSteeringPolicy(ctx context.Context, request CreateSteeringPolicyRequest) (response CreateSteeringPolicyResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createSteeringPolicy, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateSteeringPolicyResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateSteeringPolicyResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateSteeringPolicyResponse")
	}
	return
}

// createSteeringPolicy implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) createSteeringPolicy(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/steeringPolicies")
	if err != nil {
		return nil, err
	}

	var response CreateSteeringPolicyResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateSteeringPolicyAttachment Creates a new attachment between a steering policy and a domain.
// For the purposes of access control, the attachment is automatically placed
// into the same compartment as the containing zone of the domain.
func (client DnsClient) CreateSteeringPolicyAttachment(ctx context.Context, request CreateSteeringPolicyAttachmentRequest) (response CreateSteeringPolicyAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createSteeringPolicyAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateSteeringPolicyAttachmentResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateSteeringPolicyAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateSteeringPolicyAttachmentResponse")
	}
	return
}

// createSteeringPolicyAttachment implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) createSteeringPolicyAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/steeringPolicyAttachments")
	if err != nil {
		return nil, err
	}

	var response CreateSteeringPolicyAttachmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateZone Creates a new zone in the specified compartment. The `compartmentId`
// query parameter is required if the `Content-Type` header for the
// request is `text/dns`.
func (client DnsClient) CreateZone(ctx context.Context, request CreateZoneRequest) (response CreateZoneResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.createZone, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateZoneResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateZoneResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateZoneResponse")
	}
	return
}

// createZone implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) createZone(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/zones")
	if err != nil {
		return nil, err
	}

	var response CreateZoneResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteDomainRecords Deletes all records at the specified zone and domain.
func (client DnsClient) DeleteDomainRecords(ctx context.Context, request DeleteDomainRecordsRequest) (response DeleteDomainRecordsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteDomainRecords, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteDomainRecordsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteDomainRecordsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteDomainRecordsResponse")
	}
	return
}

// deleteDomainRecords implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) deleteDomainRecords(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/zones/{zoneNameOrId}/records/{domain}")
	if err != nil {
		return nil, err
	}

	var response DeleteDomainRecordsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteRRSet Deletes all records in the specified RRSet.
func (client DnsClient) DeleteRRSet(ctx context.Context, request DeleteRRSetRequest) (response DeleteRRSetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteRRSet, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteRRSetResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteRRSetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteRRSetResponse")
	}
	return
}

// deleteRRSet implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) deleteRRSet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/zones/{zoneNameOrId}/records/{domain}/{rtype}")
	if err != nil {
		return nil, err
	}

	var response DeleteRRSetResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteSteeringPolicy Deletes the specified steering policy.
// A `204` response indicates that the delete has been successful.
// Deletion will fail if the policy is attached to any zones.
func (client DnsClient) DeleteSteeringPolicy(ctx context.Context, request DeleteSteeringPolicyRequest) (response DeleteSteeringPolicyResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteSteeringPolicy, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteSteeringPolicyResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteSteeringPolicyResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteSteeringPolicyResponse")
	}
	return
}

// deleteSteeringPolicy implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) deleteSteeringPolicy(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/steeringPolicies/{steeringPolicyId}")
	if err != nil {
		return nil, err
	}

	var response DeleteSteeringPolicyResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteSteeringPolicyAttachment Deletes the specified steering policy attachment.
// A `204` response indicates that the delete has been successful.
func (client DnsClient) DeleteSteeringPolicyAttachment(ctx context.Context, request DeleteSteeringPolicyAttachmentRequest) (response DeleteSteeringPolicyAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteSteeringPolicyAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteSteeringPolicyAttachmentResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteSteeringPolicyAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteSteeringPolicyAttachmentResponse")
	}
	return
}

// deleteSteeringPolicyAttachment implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) deleteSteeringPolicyAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/steeringPolicyAttachments/{steeringPolicyAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response DeleteSteeringPolicyAttachmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteZone Deletes the specified zone and all its steering policy attachments.
// A `204` response indicates that zone has been successfully deleted.
func (client DnsClient) DeleteZone(ctx context.Context, request DeleteZoneRequest) (response DeleteZoneResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteZone, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteZoneResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteZoneResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteZoneResponse")
	}
	return
}

// deleteZone implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) deleteZone(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/zones/{zoneNameOrId}")
	if err != nil {
		return nil, err
	}

	var response DeleteZoneResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetDomainRecords Gets a list of all records at the specified zone and domain.
// The results are sorted by `rtype` in alphabetical order by default. You
// can optionally filter and/or sort the results using the listed parameters.
func (client DnsClient) GetDomainRecords(ctx context.Context, request GetDomainRecordsRequest) (response GetDomainRecordsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getDomainRecords, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetDomainRecordsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetDomainRecordsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetDomainRecordsResponse")
	}
	return
}

// getDomainRecords implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) getDomainRecords(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/zones/{zoneNameOrId}/records/{domain}")
	if err != nil {
		return nil, err
	}

	var response GetDomainRecordsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetRRSet Gets a list of all records in the specified RRSet. The results are
// sorted by `recordHash` by default.
func (client DnsClient) GetRRSet(ctx context.Context, request GetRRSetRequest) (response GetRRSetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getRRSet, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetRRSetResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetRRSetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetRRSetResponse")
	}
	return
}

// getRRSet implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) getRRSet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/zones/{zoneNameOrId}/records/{domain}/{rtype}")
	if err != nil {
		return nil, err
	}

	var response GetRRSetResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetSteeringPolicy Gets information about the specified steering policy.
func (client DnsClient) GetSteeringPolicy(ctx context.Context, request GetSteeringPolicyRequest) (response GetSteeringPolicyResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getSteeringPolicy, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetSteeringPolicyResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetSteeringPolicyResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetSteeringPolicyResponse")
	}
	return
}

// getSteeringPolicy implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) getSteeringPolicy(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/steeringPolicies/{steeringPolicyId}")
	if err != nil {
		return nil, err
	}

	var response GetSteeringPolicyResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetSteeringPolicyAttachment Gets information about the specified steering policy attachment.
func (client DnsClient) GetSteeringPolicyAttachment(ctx context.Context, request GetSteeringPolicyAttachmentRequest) (response GetSteeringPolicyAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getSteeringPolicyAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetSteeringPolicyAttachmentResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetSteeringPolicyAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetSteeringPolicyAttachmentResponse")
	}
	return
}

// getSteeringPolicyAttachment implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) getSteeringPolicyAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/steeringPolicyAttachments/{steeringPolicyAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response GetSteeringPolicyAttachmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetZone Gets information about the specified zone, including its creation date,
// zone type, and serial.
func (client DnsClient) GetZone(ctx context.Context, request GetZoneRequest) (response GetZoneResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getZone, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetZoneResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetZoneResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetZoneResponse")
	}
	return
}

// getZone implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) getZone(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/zones/{zoneNameOrId}")
	if err != nil {
		return nil, err
	}

	var response GetZoneResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetZoneRecords Gets all records in the specified zone. The results are
// sorted by `domain` in alphabetical order by default. For more
// information about records, see Resource Record (RR) TYPEs (https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4).
func (client DnsClient) GetZoneRecords(ctx context.Context, request GetZoneRecordsRequest) (response GetZoneRecordsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getZoneRecords, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetZoneRecordsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetZoneRecordsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetZoneRecordsResponse")
	}
	return
}

// getZoneRecords implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) getZoneRecords(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/zones/{zoneNameOrId}/records")
	if err != nil {
		return nil, err
	}

	var response GetZoneRecordsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListSteeringPolicies Gets a list of all steering policies in the specified compartment.
func (client DnsClient) ListSteeringPolicies(ctx context.Context, request ListSteeringPoliciesRequest) (response ListSteeringPoliciesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listSteeringPolicies, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListSteeringPoliciesResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListSteeringPoliciesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListSteeringPoliciesResponse")
	}
	return
}

// listSteeringPolicies implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) listSteeringPolicies(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/steeringPolicies")
	if err != nil {
		return nil, err
	}

	var response ListSteeringPoliciesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListSteeringPolicyAttachments Lists the steering policy attachments in the specified compartment.
func (client DnsClient) ListSteeringPolicyAttachments(ctx context.Context, request ListSteeringPolicyAttachmentsRequest) (response ListSteeringPolicyAttachmentsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listSteeringPolicyAttachments, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListSteeringPolicyAttachmentsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListSteeringPolicyAttachmentsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListSteeringPolicyAttachmentsResponse")
	}
	return
}

// listSteeringPolicyAttachments implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) listSteeringPolicyAttachments(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/steeringPolicyAttachments")
	if err != nil {
		return nil, err
	}

	var response ListSteeringPolicyAttachmentsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListZones Gets a list of all zones in the specified compartment. The collection
// can be filtered by name, time created, and zone type.
func (client DnsClient) ListZones(ctx context.Context, request ListZonesRequest) (response ListZonesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listZones, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListZonesResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListZonesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListZonesResponse")
	}
	return
}

// listZones implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) listZones(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/zones")
	if err != nil {
		return nil, err
	}

	var response ListZonesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// PatchDomainRecords Updates records in the specified zone at a domain. You can update one record or all records for the specified zone depending on the changes provided in the request body. You can also add or remove records using this function.
func (client DnsClient) PatchDomainRecords(ctx context.Context, request PatchDomainRecordsRequest) (response PatchDomainRecordsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.patchDomainRecords, policy)
	if err != nil {
		if ociResponse != nil {
			response = PatchDomainRecordsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(PatchDomainRecordsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into PatchDomainRecordsResponse")
	}
	return
}

// patchDomainRecords implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) patchDomainRecords(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPatch, "/zones/{zoneNameOrId}/records/{domain}")
	if err != nil {
		return nil, err
	}

	var response PatchDomainRecordsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// PatchRRSet Updates records in the specified RRSet.
func (client DnsClient) PatchRRSet(ctx context.Context, request PatchRRSetRequest) (response PatchRRSetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.patchRRSet, policy)
	if err != nil {
		if ociResponse != nil {
			response = PatchRRSetResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(PatchRRSetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into PatchRRSetResponse")
	}
	return
}

// patchRRSet implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) patchRRSet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPatch, "/zones/{zoneNameOrId}/records/{domain}/{rtype}")
	if err != nil {
		return nil, err
	}

	var response PatchRRSetResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// PatchZoneRecords Updates a collection of records in the specified zone. You can update
// one record or all records for the specified zone depending on the
// changes provided in the request body. You can also add or remove records
// using this function.
func (client DnsClient) PatchZoneRecords(ctx context.Context, request PatchZoneRecordsRequest) (response PatchZoneRecordsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.patchZoneRecords, policy)
	if err != nil {
		if ociResponse != nil {
			response = PatchZoneRecordsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(PatchZoneRecordsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into PatchZoneRecordsResponse")
	}
	return
}

// patchZoneRecords implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) patchZoneRecords(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPatch, "/zones/{zoneNameOrId}/records")
	if err != nil {
		return nil, err
	}

	var response PatchZoneRecordsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateDomainRecords Replaces records in the specified zone at a domain with the records
// specified in the request body. If a specified record does not exist,
// it will be created. If the record exists, then it will be updated to
// represent the record in the body of the request. If a record in the zone
// does not exist in the request body, the record will be removed from the
// zone.
func (client DnsClient) UpdateDomainRecords(ctx context.Context, request UpdateDomainRecordsRequest) (response UpdateDomainRecordsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateDomainRecords, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateDomainRecordsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateDomainRecordsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateDomainRecordsResponse")
	}
	return
}

// updateDomainRecords implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) updateDomainRecords(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/zones/{zoneNameOrId}/records/{domain}")
	if err != nil {
		return nil, err
	}

	var response UpdateDomainRecordsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateRRSet Replaces records in the specified RRSet.
func (client DnsClient) UpdateRRSet(ctx context.Context, request UpdateRRSetRequest) (response UpdateRRSetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateRRSet, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateRRSetResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateRRSetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateRRSetResponse")
	}
	return
}

// updateRRSet implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) updateRRSet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/zones/{zoneNameOrId}/records/{domain}/{rtype}")
	if err != nil {
		return nil, err
	}

	var response UpdateRRSetResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateSteeringPolicy Updates the specified steering policy with your new information.
func (client DnsClient) UpdateSteeringPolicy(ctx context.Context, request UpdateSteeringPolicyRequest) (response UpdateSteeringPolicyResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateSteeringPolicy, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateSteeringPolicyResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateSteeringPolicyResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateSteeringPolicyResponse")
	}
	return
}

// updateSteeringPolicy implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) updateSteeringPolicy(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/steeringPolicies/{steeringPolicyId}")
	if err != nil {
		return nil, err
	}

	var response UpdateSteeringPolicyResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateSteeringPolicyAttachment Updates the specified steering policy attachment with your new information.
func (client DnsClient) UpdateSteeringPolicyAttachment(ctx context.Context, request UpdateSteeringPolicyAttachmentRequest) (response UpdateSteeringPolicyAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateSteeringPolicyAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateSteeringPolicyAttachmentResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateSteeringPolicyAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateSteeringPolicyAttachmentResponse")
	}
	return
}

// updateSteeringPolicyAttachment implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) updateSteeringPolicyAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/steeringPolicyAttachments/{steeringPolicyAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response UpdateSteeringPolicyAttachmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateZone Updates the specified secondary zone with your new external master
// server information. For more information about secondary zone, see
// Manage DNS Service Zone (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Tasks/managingdnszones.htm).
func (client DnsClient) UpdateZone(ctx context.Context, request UpdateZoneRequest) (response UpdateZoneResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateZone, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateZoneResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateZoneResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateZoneResponse")
	}
	return
}

// updateZone implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) updateZone(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/zones/{zoneNameOrId}")
	if err != nil {
		return nil, err
	}

	var response UpdateZoneResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateZoneRecords Replaces records in the specified zone with the records specified in the
// request body. If a specified record does not exist, it will be created.
// If the record exists, then it will be updated to represent the record in
// the body of the request. If a record in the zone does not exist in the
// request body, the record will be removed from the zone.
func (client DnsClient) UpdateZoneRecords(ctx context.Context, request UpdateZoneRecordsRequest) (response UpdateZoneRecordsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateZoneRecords, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateZoneRecordsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateZoneRecordsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateZoneRecordsResponse")
	}
	return
}

// updateZoneRecords implements the OCIOperation interface (enables retrying operations)
func (client DnsClient) updateZoneRecords(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/zones/{zoneNameOrId}/records")
	if err != nil {
		return nil, err
	}

	var response UpdateZoneRecordsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}
