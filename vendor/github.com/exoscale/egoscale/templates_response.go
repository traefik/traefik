// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListTemplates) Response() interface{} {
	return new(ListTemplatesResponse)
}

// ListRequest returns itself
func (ls *ListTemplates) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListTemplates) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListTemplates) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListTemplates) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListTemplatesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListTemplatesResponse was expected, got %T", resp))
		return
	}

	for i := range items.Template {
		if !callback(&items.Template[i], nil) {
			break
		}
	}
}
