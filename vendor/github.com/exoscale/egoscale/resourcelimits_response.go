// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListResourceLimits) Response() interface{} {
	return new(ListResourceLimitsResponse)
}

// ListRequest returns itself
func (ls *ListResourceLimits) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListResourceLimits) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListResourceLimits) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListResourceLimits) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListResourceLimitsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListResourceLimitsResponse was expected, got %T", resp))
		return
	}

	for i := range items.ResourceLimit {
		if !callback(&items.ResourceLimit[i], nil) {
			break
		}
	}
}
