// +build ignore

package linodego

/*
 - replace "Template" with "NameOfResource"
 - replace "template" with "nameOfResource"
 - copy template_test.go and do the same
 - When updating Template structs,
   - use pointers where ever null'able would have a different meaning if the wrapper
	 supplied "" or 0 instead
 - Add "NameOfResource" to client.go, resources.go, pagination.go
*/

import (
	"context"
	"encoding/json"
	"fmt"
)

// Template represents a Template object
type Template struct {
	ID int `json:"id"`
	// UpdatedStr string `json:"updated"`
	// Updated *time.Time `json:"-"`
}

// TemplateCreateOptions fields are those accepted by CreateTemplate
type TemplateCreateOptions struct {
}

// TemplateUpdateOptions fields are those accepted by UpdateTemplate
type TemplateUpdateOptions struct {
}

// GetCreateOptions converts a Template to TemplateCreateOptions for use in CreateTemplate
func (i Template) GetCreateOptions() (o TemplateCreateOptions) {
	// o.Label = i.Label
	// o.Description = copyString(o.Description)
	return
}

// GetUpdateOptions converts a Template to TemplateUpdateOptions for use in UpdateTemplate
func (i Template) GetUpdateOptions() (o TemplateUpdateOptions) {
	// o.Label = i.Label
	// o.Description = copyString(o.Description)
	return
}

// TemplatesPagedResponse represents a paginated Template API response
type TemplatesPagedResponse struct {
	*PageOptions
	Data []Template `json:"data"`
}

// endpoint gets the endpoint URL for Template
func (TemplatesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Templates.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Templates when processing paginated Template responses
func (resp *TemplatesPagedResponse) appendData(r *TemplatesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListTemplates lists Templates
func (c *Client) ListTemplates(ctx context.Context, opts *ListOptions) ([]Template, error) {
	response := TemplatesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	for i := range response.Data {
		response.Data[i].fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (i *Template) fixDates() *Template {
	// i.Created, _ = parseDates(i.CreatedStr)
	// i.Updated, _ = parseDates(i.UpdatedStr)
	return i
}

// GetTemplate gets the template with the provided ID
func (c *Client) GetTemplate(ctx context.Context, id int) (*Template, error) {
	e, err := c.Templates.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&Template{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Template).fixDates(), nil
}

// CreateTemplate creates a Template
func (c *Client) CreateTemplate(ctx context.Context, createOpts TemplateCreateOptions) (*Template, error) {
	var body string
	e, err := c.Templates.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Template{})

	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Template).fixDates(), nil
}

// UpdateTemplate updates the Template with the specified id
func (c *Client) UpdateTemplate(ctx context.Context, id int, updateOpts TemplateUpdateOptions) (*Template, error) {
	var body string
	e, err := c.Templates.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&Template{})

	if bodyData, err := json.Marshal(updateOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Template).fixDates(), nil
}

// DeleteTemplate deletes the Template with the specified id
func (c *Client) DeleteTemplate(ctx context.Context, id int) error {
	e, err := c.Templates.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
