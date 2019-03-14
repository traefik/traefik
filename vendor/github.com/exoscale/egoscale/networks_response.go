// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListNetworks) Response() interface{} {
	return new(ListNetworksResponse)
}

// ListRequest returns itself
func (ls *ListNetworks) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListNetworks) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListNetworks) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListNetworks) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListNetworksResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListNetworksResponse was expected, got %T", resp))
		return
	}

	for i := range items.Network {
		if !callback(&items.Network[i], nil) {
			break
		}
	}
}
