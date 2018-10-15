package linodego

/**
 * Pagination and Filtering types and helpers
 */

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/go-resty/resty"
)

// PageOptions are the pagination parameters for List endpoints
type PageOptions struct {
	Page    int `url:"page,omitempty" json:"page"`
	Pages   int `url:"pages,omitempty" json:"pages"`
	Results int `url:"results,omitempty" json:"results"`
}

// ListOptions are the pagination and filtering (TODO) parameters for endpoints
type ListOptions struct {
	*PageOptions
	Filter string
}

// NewListOptions simplified construction of ListOptions using only
// the two writable properties, Page and Filter
func NewListOptions(Page int, Filter string) *ListOptions {
	return &ListOptions{PageOptions: &PageOptions{Page: Page}, Filter: Filter}

}

// listHelper abstracts fetching and pagination for GET endpoints that
// do not require any Ids (top level endpoints).
// When opts (or opts.Page) is nil, all pages will be fetched and
// returned in a single (endpoint-specific)PagedResponse
// opts.results and opts.pages will be updated from the API response
func (c *Client) listHelper(ctx context.Context, i interface{}, opts *ListOptions) error {
	req := c.R(ctx)
	if opts != nil && opts.PageOptions != nil && opts.Page > 0 {
		req.SetQueryParam("page", strconv.Itoa(opts.Page))
	}

	var (
		err     error
		pages   int
		results int
		r       *resty.Response
	)

	if opts != nil && len(opts.Filter) > 0 {
		req.SetHeader("X-Filter", opts.Filter)
	}

	switch v := i.(type) {
	case *LinodeKernelsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(LinodeKernelsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*LinodeKernelsPagedResponse).Pages
			results = r.Result().(*LinodeKernelsPagedResponse).Results
			v.appendData(r.Result().(*LinodeKernelsPagedResponse))
		}
	case *LinodeTypesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(LinodeTypesPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*LinodeTypesPagedResponse).Pages
			results = r.Result().(*LinodeTypesPagedResponse).Results
			v.appendData(r.Result().(*LinodeTypesPagedResponse))
		}
	case *ImagesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(ImagesPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*ImagesPagedResponse).Pages
			results = r.Result().(*ImagesPagedResponse).Results
			v.appendData(r.Result().(*ImagesPagedResponse))
		}
	case *StackscriptsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(StackscriptsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*StackscriptsPagedResponse).Pages
			results = r.Result().(*StackscriptsPagedResponse).Results
			v.appendData(r.Result().(*StackscriptsPagedResponse))
		}
	case *InstancesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(InstancesPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*InstancesPagedResponse).Pages
			results = r.Result().(*InstancesPagedResponse).Results
			v.appendData(r.Result().(*InstancesPagedResponse))
		}
	case *RegionsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(RegionsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*RegionsPagedResponse).Pages
			results = r.Result().(*RegionsPagedResponse).Results
			v.appendData(r.Result().(*RegionsPagedResponse))
		}
	case *VolumesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(VolumesPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*VolumesPagedResponse).Pages
			results = r.Result().(*VolumesPagedResponse).Results
			v.appendData(r.Result().(*VolumesPagedResponse))
		}
	case *DomainsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(DomainsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			response, ok := r.Result().(*DomainsPagedResponse)
			if !ok {
				return fmt.Errorf("Response is not a *DomainsPagedResponse")
			}
			pages = response.Pages
			results = response.Results
			v.appendData(response)
		}
	case *EventsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(EventsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*EventsPagedResponse).Pages
			results = r.Result().(*EventsPagedResponse).Results
			v.appendData(r.Result().(*EventsPagedResponse))
		}
	case *LongviewSubscriptionsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(LongviewSubscriptionsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*LongviewSubscriptionsPagedResponse).Pages
			results = r.Result().(*LongviewSubscriptionsPagedResponse).Results
			v.appendData(r.Result().(*LongviewSubscriptionsPagedResponse))
		}
	case *LongviewClientsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(LongviewClientsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*LongviewClientsPagedResponse).Pages
			results = r.Result().(*LongviewClientsPagedResponse).Results
			v.appendData(r.Result().(*LongviewClientsPagedResponse))
		}
	case *IPAddressesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(IPAddressesPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*IPAddressesPagedResponse).Pages
			results = r.Result().(*IPAddressesPagedResponse).Results
			v.appendData(r.Result().(*IPAddressesPagedResponse))
		}
	case *IPv6PoolsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(IPv6PoolsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*IPv6PoolsPagedResponse).Pages
			results = r.Result().(*IPv6PoolsPagedResponse).Results
			v.appendData(r.Result().(*IPv6PoolsPagedResponse))
		}
	case *IPv6RangesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(IPv6RangesPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*IPv6RangesPagedResponse).Pages
			results = r.Result().(*IPv6RangesPagedResponse).Results
			v.appendData(r.Result().(*IPv6RangesPagedResponse))
			// @TODO consolidate this type with IPv6PoolsPagedResponse?
		}
	case *SSHKeysPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(SSHKeysPagedResponse{}).Get(v.endpoint(c))); err == nil {
			response, ok := r.Result().(*SSHKeysPagedResponse)
			if !ok {
				return fmt.Errorf("Response is not a *SSHKeysPagedResponse")
			}
			pages = response.Pages
			results = response.Results
			v.appendData(response)
		}
	case *TicketsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(TicketsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*TicketsPagedResponse).Pages
			results = r.Result().(*TicketsPagedResponse).Results
			v.appendData(r.Result().(*TicketsPagedResponse))
		}
	case *InvoicesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(InvoicesPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*InvoicesPagedResponse).Pages
			results = r.Result().(*InvoicesPagedResponse).Results
			v.appendData(r.Result().(*InvoicesPagedResponse))
		}
	case *NotificationsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(NotificationsPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*NotificationsPagedResponse).Pages
			results = r.Result().(*NotificationsPagedResponse).Results
			v.appendData(r.Result().(*NotificationsPagedResponse))
		}
	case *NodeBalancersPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(NodeBalancersPagedResponse{}).Get(v.endpoint(c))); err == nil {
			pages = r.Result().(*NodeBalancersPagedResponse).Pages
			results = r.Result().(*NodeBalancersPagedResponse).Results
			v.appendData(r.Result().(*NodeBalancersPagedResponse))
		}
	/**
	case AccountOauthClientsPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*AccountOauthClientsPagedResponse).Pages
			results = r.Result().(*AccountOauthClientsPagedResponse).Results
			v.appendData(r.Result().(*AccountOauthClientsPagedResponse))
		}
	case AccountPaymentsPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*AccountPaymentsPagedResponse).Pages
			results = r.Result().(*AccountPaymentsPagedResponse).Results
			v.appendData(r.Result().(*AccountPaymentsPagedResponse))
		}
	case AccountUsersPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*AccountUsersPagedResponse).Pages
			results = r.Result().(*AccountUsersPagedResponse).Results
			v.appendData(r.Result().(*AccountUsersPagedResponse))
		}
	case ProfileAppsPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ProfileAppsPagedResponse).Pages
			results = r.Result().(*ProfileAppsPagedResponse).Results
			v.appendData(r.Result().(*ProfileAppsPagedResponse))
		}
	case ProfileTokensPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ProfileTokensPagedResponse).Pages
			results = r.Result().(*ProfileTokensPagedResponse).Results
			v.appendData(r.Result().(*ProfileTokensPagedResponse))
		}
	case ProfileWhitelistPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ProfileWhitelistPagedResponse).Pages
			results = r.Result().(*ProfileWhitelistPagedResponse).Results
			v.appendData(r.Result().(*ProfileWhitelistPagedResponse))
		}
	case ManagedContactsPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ManagedContactsPagedResponse).Pages
			results = r.Result().(*ManagedContactsPagedResponse).Results
			v.appendData(r.Result().(*ManagedContactsPagedResponse))
		}
	case ManagedCredentialsPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ManagedCredentialsPagedResponse).Pages
			results = r.Result().(*ManagedCredentialsPagedResponse).Results
			v.appendData(r.Result().(*ManagedCredentialsPagedResponse))
		}
	case ManagedIssuesPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ManagedIssuesPagedResponse).Pages
			results = r.Result().(*ManagedIssuesPagedResponse).Results
			v.appendData(r.Result().(*ManagedIssuesPagedResponse))
		}
	case ManagedLinodeSettingsPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ManagedLinodeSettingsPagedResponse).Pages
			results = r.Result().(*ManagedLinodeSettingsPagedResponse).Results
			v.appendData(r.Result().(*ManagedLinodeSettingsPagedResponse))
		}
	case ManagedServicesPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*ManagedServicesPagedResponse).Pages
			results = r.Result().(*ManagedServicesPagedResponse).Results
			v.appendData(r.Result().(*ManagedServicesPagedResponse))
		}
	**/
	default:
		log.Fatalf("listHelper interface{} %+v used", i)
	}

	if err != nil {
		return err
	}

	if opts == nil {
		for page := 2; page <= pages; page = page + 1 {
			if err := c.listHelper(ctx, i, &ListOptions{PageOptions: &PageOptions{Page: page}}); err != nil {
				return err
			}
		}
	} else {
		if opts.PageOptions == nil {
			opts.PageOptions = &PageOptions{}
		}

		if opts.Page == 0 {
			for page := 2; page <= pages; page = page + 1 {
				opts.Page = page
				if err := c.listHelper(ctx, i, opts); err != nil {
					return err
				}
			}
		}
		opts.Results = results
		opts.Pages = pages
	}

	return nil
}

// listHelperWithID abstracts fetching and pagination for GET endpoints that
// require an Id (second level endpoints).
// When opts (or opts.Page) is nil, all pages will be fetched and
// returned in a single (endpoint-specific)PagedResponse
// opts.results and opts.pages will be updated from the API response
func (c *Client) listHelperWithID(ctx context.Context, i interface{}, id int, opts *ListOptions) error {
	req := c.R(ctx)
	if opts != nil && opts.Page > 0 {
		req.SetQueryParam("page", strconv.Itoa(opts.Page))
	}

	var (
		err     error
		pages   int
		results int
		r       *resty.Response
	)

	if opts != nil && len(opts.Filter) > 0 {
		req.SetHeader("X-Filter", opts.Filter)
	}

	switch v := i.(type) {
	case *InvoiceItemsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(InvoiceItemsPagedResponse{}).Get(v.endpointWithID(c, id))); err == nil {
			pages = r.Result().(*InvoiceItemsPagedResponse).Pages
			results = r.Result().(*InvoiceItemsPagedResponse).Results
			v.appendData(r.Result().(*InvoiceItemsPagedResponse))
		}
	case *DomainRecordsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(DomainRecordsPagedResponse{}).Get(v.endpointWithID(c, id))); err == nil {
			response, ok := r.Result().(*DomainRecordsPagedResponse)
			if !ok {
				return fmt.Errorf("Response is not a *DomainRecordsPagedResponse")
			}
			pages = response.Pages
			results = response.Results
			v.appendData(response)
		}
	case *InstanceConfigsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(InstanceConfigsPagedResponse{}).Get(v.endpointWithID(c, id))); err == nil {
			pages = r.Result().(*InstanceConfigsPagedResponse).Pages
			results = r.Result().(*InstanceConfigsPagedResponse).Results
			v.appendData(r.Result().(*InstanceConfigsPagedResponse))
		}
	case *InstanceDisksPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(InstanceDisksPagedResponse{}).Get(v.endpointWithID(c, id))); err == nil {
			pages = r.Result().(*InstanceDisksPagedResponse).Pages
			results = r.Result().(*InstanceDisksPagedResponse).Results
			v.appendData(r.Result().(*InstanceDisksPagedResponse))
		}
	case *NodeBalancerConfigsPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(NodeBalancerConfigsPagedResponse{}).Get(v.endpointWithID(c, id))); err == nil {
			pages = r.Result().(*NodeBalancerConfigsPagedResponse).Pages
			results = r.Result().(*NodeBalancerConfigsPagedResponse).Results
			v.appendData(r.Result().(*NodeBalancerConfigsPagedResponse))
		}
	case *InstanceVolumesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(InstanceVolumesPagedResponse{}).Get(v.endpointWithID(c, id))); err == nil {
			pages = r.Result().(*InstanceVolumesPagedResponse).Pages
			results = r.Result().(*InstanceVolumesPagedResponse).Results
			v.appendData(r.Result().(*InstanceVolumesPagedResponse))
		}
	/**
	case TicketAttachmentsPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*TicketAttachmentsPagedResponse).Pages
			results = r.Result().(*TicketAttachmentsPagedResponse).Results
			v.appendData(r.Result().(*TicketAttachmentsPagedResponse))
		}
	case TicketRepliesPagedResponse:
		if r, err = req.SetResult(v).Get(v.endpoint(c)); r.Error() != nil {
			return NewError(r)
		} else if err == nil {
			pages = r.Result().(*TicketRepliesPagedResponse).Pages
			results = r.Result().(*TicketRepliesPagedResponse).Results
			v.appendData(r.Result().(*TicketRepliesPagedResponse))
		}
	**/
	default:
		log.Fatalf("Unknown listHelperWithID interface{} %T used", i)
	}

	if err != nil {
		return err
	}

	if opts == nil {
		for page := 2; page <= pages; page = page + 1 {
			if err := c.listHelperWithID(ctx, i, id, &ListOptions{PageOptions: &PageOptions{Page: page}}); err != nil {
				return err
			}
		}
	} else {
		if opts.PageOptions == nil {
			opts.PageOptions = &PageOptions{}
		}
		if opts.Page == 0 {
			for page := 2; page <= pages; page = page + 1 {
				opts.Page = page
				if err := c.listHelperWithID(ctx, i, id, opts); err != nil {
					return err
				}
			}
		}
		opts.Results = results
		opts.Pages = pages
	}

	return nil
}

// listHelperWithTwoIDs abstracts fetching and pagination for GET endpoints that
// require twos IDs (third level endpoints).
// When opts (or opts.Page) is nil, all pages will be fetched and
// returned in a single (endpoint-specific)PagedResponse
// opts.results and opts.pages will be updated from the API response
func (c *Client) listHelperWithTwoIDs(ctx context.Context, i interface{}, firstID, secondID int, opts *ListOptions) error {
	req := c.R(ctx)
	if opts != nil && opts.Page > 0 {
		req.SetQueryParam("page", strconv.Itoa(opts.Page))
	}

	var (
		err     error
		pages   int
		results int
		r       *resty.Response
	)

	if opts != nil && len(opts.Filter) > 0 {
		req.SetHeader("X-Filter", opts.Filter)
	}

	switch v := i.(type) {
	case *NodeBalancerNodesPagedResponse:
		if r, err = coupleAPIErrors(req.SetResult(NodeBalancerNodesPagedResponse{}).Get(v.endpointWithTwoIDs(c, firstID, secondID))); err == nil {
			pages = r.Result().(*NodeBalancerNodesPagedResponse).Pages
			results = r.Result().(*NodeBalancerNodesPagedResponse).Results
			v.appendData(r.Result().(*NodeBalancerNodesPagedResponse))
		}

	default:
		log.Fatalf("Unknown listHelperWithTwoIDs interface{} %T used", i)
	}

	if err != nil {
		return err
	}

	if opts == nil {
		for page := 2; page <= pages; page = page + 1 {
			if err := c.listHelper(ctx, i, &ListOptions{PageOptions: &PageOptions{Page: page}}); err != nil {
				return err
			}
		}
	} else {
		if opts.PageOptions == nil {
			opts.PageOptions = &PageOptions{}
		}
		if opts.Page == 0 {
			for page := 2; page <= pages; page = page + 1 {
				opts.Page = page
				if err := c.listHelperWithTwoIDs(ctx, i, firstID, secondID, opts); err != nil {
					return err
				}
			}
		}
		opts.Results = results
		opts.Pages = pages
	}

	return nil
}
