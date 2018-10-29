package linodego

import "context"

// Account associated with the token in use
type Account struct {
	FirstName  string      `json:"first_name"`
	LastName   string      `json:"last_name"`
	Email      string      `json:"email"`
	Company    string      `json:"company"`
	Address1   string      `json:"address1"`
	Address2   string      `json:"address2"`
	Balance    float32     `json:"balance"`
	City       string      `json:"city"`
	State      string      `json:"state"`
	Zip        string      `json:"zip"`
	Country    string      `json:"country"`
	TaxID      string      `json:"tax_id"`
	CreditCard *CreditCard `json:"credit_card"`
}

// CreditCard information associated with the Account.
type CreditCard struct {
	LastFour string `json:"last_four"`
	Expiry   string `json:"expiry"`
}

// fixDates converts JSON timestamps to Go time.Time values
func (v *Account) fixDates() *Account {
	return v
}

// GetAccount gets the contact and billing information related to the Account
func (c *Client) GetAccount(ctx context.Context) (*Account, error) {
	e, err := c.Account.Endpoint()
	if err != nil {
		return nil, err
	}
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&Account{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Account).fixDates(), nil
}
