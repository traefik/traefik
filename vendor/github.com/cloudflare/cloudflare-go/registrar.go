package cloudflare

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// RegistrarDomain is the structure of the API response for a new
// Cloudflare Registrar domain.
type RegistrarDomain struct {
	ID                string              `json:"id"`
	Available         bool                `json:"available"`
	SupportedTLD      bool                `json:"supported_tld"`
	CanRegister       bool                `json:"can_register"`
	TransferIn        RegistrarTransferIn `json:"transfer_in"`
	CurrentRegistrar  string              `json:"current_registrar"`
	ExpiresAt         time.Time           `json:"expires_at"`
	RegistryStatuses  string              `json:"registry_statuses"`
	Locked            bool                `json:"locked"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	RegistrantContact RegistrantContact   `json:"registrant_contact"`
}

// RegistrarTransferIn contains the structure for a domain transfer in
// request.
type RegistrarTransferIn struct {
	UnlockDomain      string `json:"unlock_domain"`
	DisablePrivacy    string `json:"disable_privacy"`
	EnterAuthCode     string `json:"enter_auth_code"`
	ApproveTransfer   string `json:"approve_transfer"`
	AcceptFoa         string `json:"accept_foa"`
	CanCancelTransfer bool   `json:"can_cancel_transfer"`
}

// RegistrantContact is the contact details for the domain registration.
type RegistrantContact struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Organization string `json:"organization"`
	Address      string `json:"address"`
	Address2     string `json:"address2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Zip          string `json:"zip"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Fax          string `json:"fax"`
}

// RegistrarDomainConfiguration is the structure for making updates to
// and existing domain.
type RegistrarDomainConfiguration struct {
	NameServers []string `json:"name_servers"`
	Privacy     bool     `json:"privacy"`
	Locked      bool     `json:"locked"`
	AutoRenew   bool     `json:"auto_renew"`
}

// RegistrarDomainDetailResponse is the structure of the detailed
// response from the API for a single domain.
type RegistrarDomainDetailResponse struct {
	Response
	Result RegistrarDomain `json:"result"`
}

// RegistrarDomainsDetailResponse is the structure of the detailed
// response from the API.
type RegistrarDomainsDetailResponse struct {
	Response
	Result []RegistrarDomain `json:"result"`
}

// RegistrarDomain returns a single domain based on the account ID and
// domain name.
//
// API reference: https://api.cloudflare.com/#registrar-domains-get-domain
func (api *API) RegistrarDomain(accountID, domainName string) (RegistrarDomain, error) {
	uri := fmt.Sprintf("/accounts/%s/registrar/domains/%s", accountID, domainName)

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return RegistrarDomain{}, errors.Wrap(err, errMakeRequestError)
	}

	var r RegistrarDomainDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return RegistrarDomain{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// RegistrarDomains returns all registrar domains based on the account
// ID.
//
// API reference: https://api.cloudflare.com/#registrar-domains-list-domains
func (api *API) RegistrarDomains(accountID string) ([]RegistrarDomain, error) {
	uri := "/accounts/" + accountID + "/registrar/domains"

	res, err := api.makeRequest("POST", uri, nil)
	if err != nil {
		return []RegistrarDomain{}, errors.Wrap(err, errMakeRequestError)
	}

	var r RegistrarDomainsDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []RegistrarDomain{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// TransferRegistrarDomain initiates the transfer from another registrar
// to Cloudflare Registrar.
//
// API reference: https://api.cloudflare.com/#registrar-domains-transfer-domain
func (api *API) TransferRegistrarDomain(accountID, domainName string) ([]RegistrarDomain, error) {
	uri := fmt.Sprintf("/accounts/%s/registrar/domains/%s/transfer", accountID, domainName)

	res, err := api.makeRequest("POST", uri, nil)
	if err != nil {
		return []RegistrarDomain{}, errors.Wrap(err, errMakeRequestError)
	}

	var r RegistrarDomainsDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []RegistrarDomain{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// CancelRegistrarDomainTransfer cancels a pending domain transfer.
//
// API reference: https://api.cloudflare.com/#registrar-domains-cancel-transfer
func (api *API) CancelRegistrarDomainTransfer(accountID, domainName string) ([]RegistrarDomain, error) {
	uri := fmt.Sprintf("/accounts/%s/registrar/domains/%s/cancel_transfer", accountID, domainName)

	res, err := api.makeRequest("POST", uri, nil)
	if err != nil {
		return []RegistrarDomain{}, errors.Wrap(err, errMakeRequestError)
	}

	var r RegistrarDomainsDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []RegistrarDomain{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// UpdateRegistrarDomain updates an existing Registrar Domain configuration.
//
// API reference: https://api.cloudflare.com/#registrar-domains-update-domain
func (api *API) UpdateRegistrarDomain(accountID, domainName string, domainConfiguration RegistrarDomainConfiguration) (RegistrarDomain, error) {
	uri := fmt.Sprintf("/accounts/%s/registrar/domains/%s", accountID, domainName)

	res, err := api.makeRequest("PUT", uri, domainConfiguration)
	if err != nil {
		return RegistrarDomain{}, errors.Wrap(err, errMakeRequestError)
	}

	var r RegistrarDomainDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return RegistrarDomain{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}
