package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Token represents a Token object
type Token struct {
	// This token's unique ID, which can be used to revoke it.
	ID int `json:"id"`

	// The scopes this token was created with. These define what parts of the Account the token can be used to access. Many command-line tools, such as the Linode CLI, require tokens with access to *. Tokens with more restrictive scopes are generally more secure.
	Scopes string `json:"scopes"`

	// This token's label. This is for display purposes only, but can be used to more easily track what you're using each token for. (1-100 Characters)
	Label string `json:"label"`

	// The token used to access the API. When the token is created, the full token is returned here. Otherwise, only the first 16 characters are returned.
	Token string `json:"token"`

	// The date and time this token was created.
	Created    *time.Time `json:"-"`
	CreatedStr string     `json:"created"`

	// When this token will expire. Personal Access Tokens cannot be renewed, so after this time the token will be completely unusable and a new token will need to be generated. Tokens may be created with "null" as their expiry and will never expire unless revoked.
	Expiry    *time.Time `json:"-"`
	ExpiryStr string     `json:"expiry"`
}

// TokenCreateOptions fields are those accepted by CreateToken
type TokenCreateOptions struct {
	// The scopes this token was created with. These define what parts of the Account the token can be used to access. Many command-line tools, such as the Linode CLI, require tokens with access to *. Tokens with more restrictive scopes are generally more secure.
	Scopes string `json:"scopes"`

	// This token's label. This is for display purposes only, but can be used to more easily track what you're using each token for. (1-100 Characters)
	Label string `json:"label"`

	// When this token will expire. Personal Access Tokens cannot be renewed, so after this time the token will be completely unusable and a new token will need to be generated. Tokens may be created with "null" as their expiry and will never expire unless revoked.
	Expiry *time.Time `json:"expiry"`
}

// TokenUpdateOptions fields are those accepted by UpdateToken
type TokenUpdateOptions struct {
	// This token's label. This is for display purposes only, but can be used to more easily track what you're using each token for. (1-100 Characters)
	Label string `json:"label"`
}

// GetCreateOptions converts a Token to TokenCreateOptions for use in CreateToken
func (i Token) GetCreateOptions() (o TokenCreateOptions) {
	o.Label = i.Label
	o.Expiry = copyTime(i.Expiry)
	o.Scopes = i.Scopes
	return
}

// GetUpdateOptions converts a Token to TokenUpdateOptions for use in UpdateToken
func (i Token) GetUpdateOptions() (o TokenUpdateOptions) {
	o.Label = i.Label
	return
}

// TokensPagedResponse represents a paginated Token API response
type TokensPagedResponse struct {
	*PageOptions
	Data []Token `json:"data"`
}

// endpoint gets the endpoint URL for Token
func (TokensPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Tokens.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Tokens when processing paginated Token responses
func (resp *TokensPagedResponse) appendData(r *TokensPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListTokens lists Tokens
func (c *Client) ListTokens(ctx context.Context, opts *ListOptions) ([]Token, error) {
	response := TokensPagedResponse{}
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
func (i *Token) fixDates() *Token {
	i.Created, _ = parseDates(i.CreatedStr)
	i.Expiry, _ = parseDates(i.ExpiryStr)
	return i
}

// GetToken gets the token with the provided ID
func (c *Client) GetToken(ctx context.Context, id int) (*Token, error) {
	e, err := c.Tokens.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&Token{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Token).fixDates(), nil
}

// CreateToken creates a Token
func (c *Client) CreateToken(ctx context.Context, createOpts TokenCreateOptions) (*Token, error) {
	var body string
	e, err := c.Tokens.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Token{})

	// Format the Time as a string to meet the ISO8601 requirement
	createOptsFixed := struct {
		Label  string  `json:"label"`
		Scopes string  `json:"scopes"`
		Expiry *string `json:"expiry"`
	}{}
	createOptsFixed.Label = createOpts.Label
	createOptsFixed.Scopes = createOpts.Scopes
	if createOpts.Expiry != nil {
		iso8601Expiry := createOpts.Expiry.UTC().Format("2006-01-02T15:04:05")
		createOptsFixed.Expiry = &iso8601Expiry
	}

	if bodyData, err := json.Marshal(createOptsFixed); err == nil {
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
	return r.Result().(*Token).fixDates(), nil
}

// UpdateToken updates the Token with the specified id
func (c *Client) UpdateToken(ctx context.Context, id int, updateOpts TokenUpdateOptions) (*Token, error) {
	var body string
	e, err := c.Tokens.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&Token{})

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
	return r.Result().(*Token).fixDates(), nil
}

// DeleteToken deletes the Token with the specified id
func (c *Client) DeleteToken(ctx context.Context, id int) error {
	e, err := c.Tokens.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
