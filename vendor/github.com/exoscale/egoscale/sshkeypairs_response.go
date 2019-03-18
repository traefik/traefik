// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListSSHKeyPairs) Response() interface{} {
	return new(ListSSHKeyPairsResponse)
}

// ListRequest returns itself
func (ls *ListSSHKeyPairs) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListSSHKeyPairs) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListSSHKeyPairs) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListSSHKeyPairs) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListSSHKeyPairsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListSSHKeyPairsResponse was expected, got %T", resp))
		return
	}

	for i := range items.SSHKeyPair {
		if !callback(&items.SSHKeyPair[i], nil) {
			break
		}
	}
}
