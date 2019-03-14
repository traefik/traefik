// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListAccounts) Response() interface{} {
	return new(ListAccountsResponse)
}

// ListRequest returns itself
func (ls *ListAccounts) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListAccounts) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListAccounts) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListAccounts) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListAccountsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListAccountsResponse was expected, got %T", resp))
		return
	}

	for i := range items.Account {
		if !callback(&items.Account[i], nil) {
			break
		}
	}
}
