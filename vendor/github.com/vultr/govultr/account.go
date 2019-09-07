package govultr

import (
	"context"
	"net/http"
)

// AccountService is the interface to interact with Accounts endpoint on the Vultr API
// Link: https://www.vultr.com/api/#account
type AccountService interface {
	GetInfo(ctx context.Context) (*Account, error)
}

// AccountServiceHandler handles interaction with the account methods for the Vultr API
type AccountServiceHandler struct {
	client *Client
}

// Account represents a Vultr account
type Account struct {
	Balance           string `json:"balance"`
	PendingCharges    string `json:"pending_charges"`
	LastPaymentDate   string `json:"last_payment_date"`
	LastPaymentAmount string `json:"last_payment_amount"`
}

// GetInfo Vultr account info
func (a *AccountServiceHandler) GetInfo(ctx context.Context) (*Account, error) {

	uri := "/v1/account/info"
	req, err := a.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	account := new(Account)
	err = a.client.DoWithContext(ctx, req, account)

	if err != nil {
		return nil, err
	}

	return account, nil
}
