// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListAffinityGroups) Response() interface{} {
	return new(ListAffinityGroupsResponse)
}

// ListRequest returns itself
func (ls *ListAffinityGroups) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListAffinityGroups) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListAffinityGroups) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListAffinityGroups) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListAffinityGroupsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListAffinityGroupsResponse was expected, got %T", resp))
		return
	}

	for i := range items.AffinityGroup {
		if !callback(&items.AffinityGroup[i], nil) {
			break
		}
	}
}
