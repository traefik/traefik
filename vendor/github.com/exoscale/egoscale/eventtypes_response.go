// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListEventTypes) Response() interface{} {
	return new(ListEventTypesResponse)
}

// ListRequest returns itself
func (ls *ListEventTypes) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListEventTypes) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListEventTypes) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListEventTypes) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListEventTypesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListEventTypesResponse was expected, got %T", resp))
		return
	}

	for i := range items.EventType {
		if !callback(&items.EventType[i], nil) {
			break
		}
	}
}
