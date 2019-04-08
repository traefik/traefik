// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListInstanceGroups) Response() interface{} {
	return new(ListInstanceGroupsResponse)
}

// ListRequest returns itself
func (ls *ListInstanceGroups) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListInstanceGroups) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListInstanceGroups) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListInstanceGroups) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListInstanceGroupsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListInstanceGroupsResponse was expected, got %T", resp))
		return
	}

	for i := range items.InstanceGroup {
		if !callback(&items.InstanceGroup[i], nil) {
			break
		}
	}
}
