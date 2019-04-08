// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListTags) Response() interface{} {
	return new(ListTagsResponse)
}

// ListRequest returns itself
func (ls *ListTags) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListTags) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListTags) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListTags) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListTagsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListTagsResponse was expected, got %T", resp))
		return
	}

	for i := range items.Tag {
		if !callback(&items.Tag[i], nil) {
			break
		}
	}
}
