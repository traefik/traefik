// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListISOs) Response() interface{} {
	return new(ListISOsResponse)
}

// ListRequest returns itself
func (ls *ListISOs) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListISOs) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListISOs) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListISOs) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListISOsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListISOsResponse was expected, got %T", resp))
		return
	}

	for i := range items.ISO {
		if !callback(&items.ISO[i], nil) {
			break
		}
	}
}
