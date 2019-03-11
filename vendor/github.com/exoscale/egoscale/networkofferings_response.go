// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListNetworkOfferings) Response() interface{} {
	return new(ListNetworkOfferingsResponse)
}

// ListRequest returns itself
func (ls *ListNetworkOfferings) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListNetworkOfferings) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListNetworkOfferings) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListNetworkOfferings) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListNetworkOfferingsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListNetworkOfferingsResponse was expected, got %T", resp))
		return
	}

	for i := range items.NetworkOffering {
		if !callback(&items.NetworkOffering[i], nil) {
			break
		}
	}
}
