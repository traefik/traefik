// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal
func (ListAsyncJobs) Response() interface{} {
	return new(ListAsyncJobsResponse)
}

// ListRequest returns itself
func (ls *ListAsyncJobs) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current apge
func (ls *ListAsyncJobs) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListAsyncJobs) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue
func (ListAsyncJobs) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListAsyncJobsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListAsyncJobsResponse was expected, got %T", resp))
		return
	}

	for i := range items.AsyncJob {
		if !callback(&items.AsyncJob[i], nil) {
			break
		}
	}
}
