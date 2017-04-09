package dnsimple

import (
	"fmt"
)

// TldsService handles communication with the Tld related
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/tlds/
type TldsService struct {
	client *Client
}

// Tld represents a TLD in DNSimple.
type Tld struct {
	Tld                 string `json:"tld"`
	TldType             int    `json:"tld_type"`
	WhoisPrivacy        bool   `json:"whois_privacy"`
	AutoRenewOnly       bool   `json:"auto_renew_only"`
	MinimumRegistration int    `json:"minimum_registration"`
	RegistrationEnabled bool   `json:"registration_enabled"`
	RenewalEnabled      bool   `json:"renewal_enabled"`
	TransferEnabled     bool   `json:"transfer_enabled"`
}

// TldExtendedAttribute represents an extended attributes supported or required
// by a specific TLD.
//
// See https://developer.dnsimple.com/v2/tlds/
type TldExtendedAttribute struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Required    bool                         `json:"required"`
	Options     []TldExtendedAttributeOption `json:"options"`
}

// TldExtendedAttributeOption represents a single option you can assign to an extended attributes.
//
// See https://developer.dnsimple.com/v2/tlds/
type TldExtendedAttributeOption struct {
	Title       string `json:"title"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// tldResponse represents a response from an API method that returns a Tld struct.
type tldResponse struct {
	Response
	Data *Tld `json:"data"`
}

// tldsResponse represents a response from an API method that returns a collection of Tld struct.
type tldsResponse struct {
	Response
	Data []Tld `json:"data"`
}

// tldExtendedAttributesResponse represents a response from an API method that returns
// a collection of Tld extended attributes.
type tldExtendedAttributesResponse struct {
	Response
	Data []TldExtendedAttribute `json:"data"`
}

// ListTlds lists the supported TLDs.
//
// See https://developer.dnsimple.com/v2/tlds/#list
func (s *TldsService) ListTlds(options *ListOptions) (*tldsResponse, error) {
	path := versioned("/tlds")
	tldsResponse := &tldsResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, tldsResponse)
	if err != nil {
		return tldsResponse, err
	}

	tldsResponse.HttpResponse = resp
	return tldsResponse, nil
}

// GetTld fetches a TLD.
//
// See https://developer.dnsimple.com/v2/tlds/#get
func (s *TldsService) GetTld(tld string) (*tldResponse, error) {
	path := versioned(fmt.Sprintf("/tlds/%s", tld))
	tldResponse := &tldResponse{}

	resp, err := s.client.get(path, tldResponse)
	if err != nil {
		return nil, err
	}

	tldResponse.HttpResponse = resp
	return tldResponse, nil
}

// GetTld fetches the extended attributes of a TLD.
//
// See https://developer.dnsimple.com/v2/tlds/#get
func (s *TldsService) GetTldExtendedAttributes(tld string) (*tldExtendedAttributesResponse, error) {
	path := versioned(fmt.Sprintf("/tlds/%s/extended_attributes", tld))
	tldResponse := &tldExtendedAttributesResponse{}

	resp, err := s.client.get(path, tldResponse)
	if err != nil {
		return nil, err
	}

	tldResponse.HttpResponse = resp
	return tldResponse, nil
}
