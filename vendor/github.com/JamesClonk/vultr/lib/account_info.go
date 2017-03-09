package lib

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// AccountInfo of Vultr account
type AccountInfo struct {
	Balance           float64 `json:"balance"`
	PendingCharges    float64 `json:"pending_charges"`
	LastPaymentDate   string  `json:"last_payment_date"`
	LastPaymentAmount float64 `json:"last_payment_amount"`
}

// GetAccountInfo retrieves the Vultr account information about current balance, pending charges, etc..
func (c *Client) GetAccountInfo() (info AccountInfo, err error) {
	if err := c.get(`account/info`, &info); err != nil {
		return AccountInfo{}, err
	}
	return
}

// UnmarshalJSON implements json.Unmarshaller on AccountInfo.
// This is needed because the Vultr API is inconsistent in it's JSON responses for account info.
// Some fields can change type, from JSON number to JSON string and vice-versa.
func (a *AccountInfo) UnmarshalJSON(data []byte) (err error) {
	if a == nil {
		*a = AccountInfo{}
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	b, err := strconv.ParseFloat(fmt.Sprintf("%v", fields["balance"]), 64)
	if err != nil {
		return err
	}
	a.Balance = b

	pc, err := strconv.ParseFloat(fmt.Sprintf("%v", fields["pending_charges"]), 64)
	if err != nil {
		return err
	}
	a.PendingCharges = pc

	lpa, err := strconv.ParseFloat(fmt.Sprintf("%v", fields["last_payment_amount"]), 64)
	if err != nil {
		return err
	}
	a.LastPaymentAmount = lpa

	a.LastPaymentDate = fmt.Sprintf("%v", fields["last_payment_date"])

	return
}
