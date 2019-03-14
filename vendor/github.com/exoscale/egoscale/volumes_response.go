// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListVolumes) Response() interface{} {
	return new(ListVolumesResponse)
}

// ListRequest returns itself
func (ls *ListVolumes) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListVolumes) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVolumes) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListVolumes) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListVolumesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListVolumesResponse was expected, got %T", resp))
		return
	}

	for i := range items.Volume {
		if !callback(&items.Volume[i], nil) {
			break
		}
	}
}
