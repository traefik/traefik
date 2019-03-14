// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListSecurityGroups) Response() interface{} {
	return new(ListSecurityGroupsResponse)
}

// ListRequest returns itself
func (ls *ListSecurityGroups) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListSecurityGroups) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListSecurityGroups) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListSecurityGroups) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListSecurityGroupsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListSecurityGroupsResponse was expected, got %T", resp))
		return
	}

	for i := range items.SecurityGroup {
		if !callback(&items.SecurityGroup[i], nil) {
			break
		}
	}
}
