// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListServiceOfferings) Response() interface{} {
	return new(ListServiceOfferingsResponse)
}

// ListRequest returns itself
func (ls *ListServiceOfferings) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListServiceOfferings) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListServiceOfferings) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListServiceOfferings) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListServiceOfferingsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListServiceOfferingsResponse was expected, got %T", resp))
		return
	}

	for i := range items.ServiceOffering {
		if !callback(&items.ServiceOffering[i], nil) {
			break
		}
	}
}
