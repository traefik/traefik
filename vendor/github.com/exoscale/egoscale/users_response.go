// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListUsers) Response() interface{} {
	return new(ListUsersResponse)
}

// ListRequest returns itself
func (ls *ListUsers) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListUsers) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListUsers) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListUsers) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListUsersResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListUsersResponse was expected, got %T", resp))
		return
	}

	for i := range items.User {
		if !callback(&items.User[i], nil) {
			break
		}
	}
}
