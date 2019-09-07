package dnsimple

import (
	"fmt"
)

// WhoisPrivacy represents a whois privacy in DNSimple.
type WhoisPrivacy struct {
	ID        int64  `json:"id,omitempty"`
	DomainID  int64  `json:"domain_id,omitempty"`
	Enabled   bool   `json:"enabled,omitempty"`
	ExpiresOn string `json:"expires_on,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// WhoisPrivacyRenewal represents a whois privacy renewal in DNSimple.
type WhoisPrivacyRenewal struct {
	ID             int64  `json:"id,omitempty"`
	DomainID       int64  `json:"domain_id,omitempty"`
	WhoisPrivacyID int64  `json:"whois_privacy_id,omitempty"`
	State          string `json:"string,omitempty"`
	Enabled        bool   `json:"enabled,omitempty"`
	ExpiresOn      string `json:"expires_on,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// whoisPrivacyResponse represents a response from an API method that returns a WhoisPrivacy struct.
type whoisPrivacyResponse struct {
	Response
	Data *WhoisPrivacy `json:"data"`
}

// whoisPrivacyRenewalResponse represents a response from an API method that returns a WhoisPrivacyRenewal struct.
type whoisPrivacyRenewalResponse struct {
	Response
	Data *WhoisPrivacyRenewal `json:"data"`
}

// GetWhoisPrivacy gets the whois privacy for the domain.
//
// See https://developer.dnsimple.com/v2/registrar/whois-privacy/#get
func (s *RegistrarService) GetWhoisPrivacy(accountID string, domainName string) (*whoisPrivacyResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/whois_privacy", accountID, domainName))
	privacyResponse := &whoisPrivacyResponse{}

	resp, err := s.client.get(path, privacyResponse)
	if err != nil {
		return nil, err
	}

	privacyResponse.HttpResponse = resp
	return privacyResponse, nil
}

// EnableWhoisPrivacy enables the whois privacy for the domain.
//
// See https://developer.dnsimple.com/v2/registrar/whois-privacy/#enable
func (s *RegistrarService) EnableWhoisPrivacy(accountID string, domainName string) (*whoisPrivacyResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/whois_privacy", accountID, domainName))
	privacyResponse := &whoisPrivacyResponse{}

	resp, err := s.client.put(path, nil, privacyResponse)
	if err != nil {
		return nil, err
	}

	privacyResponse.HttpResponse = resp
	return privacyResponse, nil
}

// DisableWhoisPrivacy disables the whois privacy for the domain.
//
// See https://developer.dnsimple.com/v2/registrar/whois-privacy/#enable
func (s *RegistrarService) DisableWhoisPrivacy(accountID string, domainName string) (*whoisPrivacyResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/whois_privacy", accountID, domainName))
	privacyResponse := &whoisPrivacyResponse{}

	resp, err := s.client.delete(path, nil, privacyResponse)
	if err != nil {
		return nil, err
	}

	privacyResponse.HttpResponse = resp
	return privacyResponse, nil
}

// RenewWhoisPrivacy renews the whois privacy for the domain.
//
// See https://developer.dnsimple.com/v2/registrar/whois-privacy/#renew
func (s *RegistrarService) RenewWhoisPrivacy(accountID string, domainName string) (*whoisPrivacyRenewalResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/whois_privacy/renewals", accountID, domainName))
	privacyRenewalResponse := &whoisPrivacyRenewalResponse{}

	resp, err := s.client.post(path, nil, privacyRenewalResponse)
	if err != nil {
		return nil, err
	}

	privacyRenewalResponse.HttpResponse = resp
	return privacyRenewalResponse, nil
}
