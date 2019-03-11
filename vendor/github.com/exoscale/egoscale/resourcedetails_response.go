// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListResourceDetails) Response() interface{} {
	return new(ListResourceDetailsResponse)
}

// ListRequest returns itself
func (ls *ListResourceDetails) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListResourceDetails) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListResourceDetails) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListResourceDetails) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListResourceDetailsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListResourceDetailsResponse was expected, got %T", resp))
		return
	}

	for i := range items.ResourceDetail {
		if !callback(&items.ResourceDetail[i], nil) {
			break
		}
	}
}
