package govultr

import (
	"context"
	"net/http"
	"net/url"
)

// DNSDomainService is the interface to interact with the DNS endpoints on the Vultr API
// Link: https://www.vultr.com/api/#dns
type DNSDomainService interface {
	Create(ctx context.Context, domain, InstanceIP string) error
	Delete(ctx context.Context, domain string) error
	ToggleDNSSec(ctx context.Context, domain string, enabled bool) error
	DNSSecInfo(ctx context.Context, domain string) ([]string, error)
	List(ctx context.Context) ([]DNSDomain, error)
	GetSoa(ctx context.Context, domain string) (*Soa, error)
	UpdateSoa(ctx context.Context, domain, nsPrimary, email string) error
}

// DNSDomainServiceHandler handles interaction with the DNS methods for the Vultr API
type DNSDomainServiceHandler struct {
	client *Client
}

// DNSDomain represents a DNS Domain entry on Vultr
type DNSDomain struct {
	Domain      string `json:"domain"`
	DateCreated string `json:"date_created"`
}

// Soa represents record information for a domain on Vultr
type Soa struct {
	NsPrimary string `json:"nsprimary"`
	Email     string `json:"email"`
}

// Create will create a DNS Domain entry on Vultr
func (d *DNSDomainServiceHandler) Create(ctx context.Context, domain, InstanceIP string) error {

	uri := "/v1/dns/create_domain"

	values := url.Values{
		"domain":   {domain},
		"serverip": {InstanceIP},
	}

	req, err := d.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = d.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

//Delete will delete a domain name and all associated records
func (d *DNSDomainServiceHandler) Delete(ctx context.Context, domain string) error {
	uri := "/v1/dns/delete_domain"

	values := url.Values{
		"domain": {domain},
	}

	req, err := d.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = d.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// ToggleDNSSec will enable or disable DNSSEC for a domain on Vultr
func (d *DNSDomainServiceHandler) ToggleDNSSec(ctx context.Context, domain string, enabled bool) error {

	uri := "/v1/dns/dnssec_enable"

	enable := "no"
	if enabled == true {
		enable = "yes"
	}

	values := url.Values{
		"domain": {domain},
		"enable": {enable},
	}

	req, err := d.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = d.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// DNSSecInfo gets the DNSSec keys for a domain (if enabled)
func (d *DNSDomainServiceHandler) DNSSecInfo(ctx context.Context, domain string) ([]string, error) {

	uri := "/v1/dns/dnssec_info"

	req, err := d.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("domain", domain)
	req.URL.RawQuery = q.Encode()

	var DNSSec []string
	err = d.client.DoWithContext(ctx, req, &DNSSec)

	if err != nil {
		return nil, err
	}

	return DNSSec, nil
}

// List gets all domains associated with the current Vultr account.
func (d *DNSDomainServiceHandler) List(ctx context.Context) ([]DNSDomain, error) {
	uri := "/v1/dns/list"

	req, err := d.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var dnsDomains []DNSDomain
	err = d.client.DoWithContext(ctx, req, &dnsDomains)

	if err != nil {
		return nil, err
	}

	return dnsDomains, nil
}

// GetSoa gets the SOA record information for a domain
func (d *DNSDomainServiceHandler) GetSoa(ctx context.Context, domain string) (*Soa, error) {
	uri := "/v1/dns/soa_info"

	req, err := d.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("domain", domain)
	req.URL.RawQuery = q.Encode()

	soa := new(Soa)
	err = d.client.DoWithContext(ctx, req, soa)

	if err != nil {
		return nil, err
	}

	return soa, nil
}

// UpdateSoa will update the SOA record information for a domain.
func (d *DNSDomainServiceHandler) UpdateSoa(ctx context.Context, domain, nsPrimary, email string) error {

	uri := "/v1/dns/soa_update"

	values := url.Values{
		"domain":    {domain},
		"nsprimary": {nsPrimary},
		"email":     {email},
	}

	req, err := d.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = d.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}
