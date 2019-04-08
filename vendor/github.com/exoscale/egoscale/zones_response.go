// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListZones) Response() interface{} {
	return new(ListZonesResponse)
}

// ListRequest returns itself
func (ls *ListZones) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListZones) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListZones) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListZones) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListZonesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListZonesResponse was expected, got %T", resp))
		return
	}

	for i := range items.Zone {
		if !callback(&items.Zone[i], nil) {
			break
		}
	}
}
