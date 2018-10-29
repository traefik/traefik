package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Stackscript represents a Linode StackScript
type Stackscript struct {
	CreatedStr string `json:"created"`
	UpdatedStr string `json:"updated"`

	ID                int               `json:"id"`
	Username          string            `json:"username"`
	Label             string            `json:"label"`
	Description       string            `json:"description"`
	Images            []string          `json:"images"`
	DeploymentsTotal  int               `json:"deployments_total"`
	DeploymentsActive int               `json:"deployments_active"`
	IsPublic          bool              `json:"is_public"`
	Created           *time.Time        `json:"-"`
	Updated           *time.Time        `json:"-"`
	RevNote           string            `json:"rev_note"`
	Script            string            `json:"script"`
	UserDefinedFields *[]StackscriptUDF `json:"user_defined_fields"`
	UserGravatarID    string            `json:"user_gravatar_id"`
}

// StackscriptUDF define a single variable that is accepted by a Stackscript
type StackscriptUDF struct {
	// A human-readable label for the field that will serve as the input prompt for entering the value during deployment.
	Label string `json:"label"`

	// The name of the field.
	Name string `json:"name"`

	// An example value for the field.
	Example string `json:"example"`

	// A list of acceptable single values for the field.
	OneOf string `json:"oneOf,omitempty"`

	// A list of acceptable values for the field in any quantity, combination or order.
	ManyOf string `json:"manyOf,omitempty"`

	// The default value. If not specified, this value will be used.
	Default string `json:"default,omitempty"`
}

// StackscriptCreateOptions fields are those accepted by CreateStackscript
type StackscriptCreateOptions struct {
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	IsPublic    bool     `json:"is_public"`
	RevNote     string   `json:"rev_note"`
	Script      string   `json:"script"`
}

// StackscriptUpdateOptions fields are those accepted by UpdateStackscript
type StackscriptUpdateOptions StackscriptCreateOptions

// GetCreateOptions converts a Stackscript to StackscriptCreateOptions for use in CreateStackscript
func (i Stackscript) GetCreateOptions() StackscriptCreateOptions {
	return StackscriptCreateOptions{
		Label:       i.Label,
		Description: i.Description,
		Images:      i.Images,
		IsPublic:    i.IsPublic,
		RevNote:     i.RevNote,
		Script:      i.Script,
	}
}

// GetUpdateOptions converts a Stackscript to StackscriptUpdateOptions for use in UpdateStackscript
func (i Stackscript) GetUpdateOptions() StackscriptUpdateOptions {
	return StackscriptUpdateOptions{
		Label:       i.Label,
		Description: i.Description,
		Images:      i.Images,
		IsPublic:    i.IsPublic,
		RevNote:     i.RevNote,
		Script:      i.Script,
	}
}

// StackscriptsPagedResponse represents a paginated Stackscript API response
type StackscriptsPagedResponse struct {
	*PageOptions
	Data []Stackscript `json:"data"`
}

// endpoint gets the endpoint URL for Stackscript
func (StackscriptsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.StackScripts.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Stackscripts when processing paginated Stackscript responses
func (resp *StackscriptsPagedResponse) appendData(r *StackscriptsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListStackscripts lists Stackscripts
func (c *Client) ListStackscripts(ctx context.Context, opts *ListOptions) ([]Stackscript, error) {
	response := StackscriptsPagedResponse{}
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
func (i *Stackscript) fixDates() *Stackscript {
	i.Created, _ = parseDates(i.CreatedStr)
	i.Updated, _ = parseDates(i.UpdatedStr)
	return i
}

// GetStackscript gets the Stackscript with the provided ID
func (c *Client) GetStackscript(ctx context.Context, id int) (*Stackscript, error) {
	e, err := c.StackScripts.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := coupleAPIErrors(c.R(ctx).
		SetResult(&Stackscript{}).
		Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Stackscript).fixDates(), nil
}

// CreateStackscript creates a StackScript
func (c *Client) CreateStackscript(ctx context.Context, createOpts StackscriptCreateOptions) (*Stackscript, error) {
	var body string
	e, err := c.StackScripts.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Stackscript{})

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
	return r.Result().(*Stackscript).fixDates(), nil
}

// UpdateStackscript updates the StackScript with the specified id
func (c *Client) UpdateStackscript(ctx context.Context, id int, updateOpts StackscriptUpdateOptions) (*Stackscript, error) {
	var body string
	e, err := c.StackScripts.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&Stackscript{})

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
	return r.Result().(*Stackscript).fixDates(), nil
}

// DeleteStackscript deletes the StackScript with the specified id
func (c *Client) DeleteStackscript(ctx context.Context, id int) error {
	e, err := c.StackScripts.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
