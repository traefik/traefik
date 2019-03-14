// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListVirtualMachines) Response() interface{} {
	return new(ListVirtualMachinesResponse)
}

// ListRequest returns itself
func (ls *ListVirtualMachines) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListVirtualMachines) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVirtualMachines) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListVirtualMachines) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListVirtualMachinesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListVirtualMachinesResponse was expected, got %T", resp))
		return
	}

	for i := range items.VirtualMachine {
		if !callback(&items.VirtualMachine[i], nil) {
			break
		}
	}
}
