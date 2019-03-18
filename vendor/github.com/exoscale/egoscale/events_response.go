// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListEvents) Response() interface{} {
	return new(ListEventsResponse)
}

// ListRequest returns itself
func (ls *ListEvents) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListEvents) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListEvents) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListEvents) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListEventsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListEventsResponse was expected, got %T", resp))
		return
	}

	for i := range items.Event {
		if !callback(&items.Event[i], nil) {
			break
		}
	}
}
