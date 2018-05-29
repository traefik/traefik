package egoscale

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// DNSDomain represents a domain
type DNSDomain struct {
	ID             int64  `json:"id"`
	AccountID      int64  `json:"account_id,omitempty"`
	UserID         int64  `json:"user_id,omitempty"`
	RegistrantID   int64  `json:"registrant_id,omitempty"`
	Name           string `json:"name"`
	UnicodeName    string `json:"unicode_name"`
	Token          string `json:"token"`
	State          string `json:"state"`
	Language       string `json:"language,omitempty"`
	Lockable       bool   `json:"lockable"`
	AutoRenew      bool   `json:"auto_renew"`
	WhoisProtected bool   `json:"whois_protected"`
	RecordCount    int64  `json:"record_count"`
	ServiceCount   int64  `json:"service_count"`
	ExpiresOn      string `json:"expires_on,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// DNSDomainResponse represents a domain creation response
type DNSDomainResponse struct {
	Domain *DNSDomain `json:"domain"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	ID         int64  `json:"id,omitempty"`
	DomainID   int64  `json:"domain_id,omitempty"`
	Name       string `json:"name"`
	TTL        int    `json:"ttl,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Content    string `json:"content"`
	RecordType string `json:"record_type"`
	Prio       int    `json:"prio,omitempty"`
}

// DNSRecordResponse represents the creation of a DNS record
type DNSRecordResponse struct {
	Record DNSRecord `json:"record"`
}

// DNSErrorResponse represents an error in the API
type DNSErrorResponse struct {
	Message string    `json:"message,omitempty"`
	Errors  *DNSError `json:"errors"`
}

// DNSError represents an error
type DNSError struct {
	Name []string `json:"name"`
}

// Error formats the DNSerror into a string
func (req *DNSErrorResponse) Error() error {
	if req.Errors != nil {
		return fmt.Errorf("DNS error: %s", strings.Join(req.Errors.Name, ", "))
	}
	return fmt.Errorf("DNS error: %s", req.Message)
}

// CreateDomain creates a DNS domain
func (exo *Client) CreateDomain(name string) (*DNSDomain, error) {
	m, err := json.Marshal(DNSDomainResponse{
		Domain: &DNSDomain{
			Name: name,
		},
	})
	if err != nil {
		return nil, err
	}

	resp, err := exo.dnsRequest("/v1/domains", string(m), "POST")
	if err != nil {
		return nil, err
	}

	var d *DNSDomainResponse
	if err := json.Unmarshal(resp, &d); err != nil {
		return nil, err
	}

	return d.Domain, nil
}

// GetDomain gets a DNS domain
func (exo *Client) GetDomain(name string) (*DNSDomain, error) {
	resp, err := exo.dnsRequest("/v1/domains/"+name, "", "GET")
	if err != nil {
		return nil, err
	}

	var d *DNSDomainResponse
	if err := json.Unmarshal(resp, &d); err != nil {
		return nil, err
	}

	return d.Domain, nil
}

// DeleteDomain delets a DNS domain
func (exo *Client) DeleteDomain(name string) error {
	_, err := exo.dnsRequest("/v1/domains/"+name, "", "DELETE")
	if err != nil {
		return err
	}

	return nil
}

// GetRecord returns a DNS record
func (exo *Client) GetRecord(domain string, recordID int64) (*DNSRecord, error) {
	id := strconv.FormatInt(recordID, 10)
	resp, err := exo.dnsRequest("/v1/domains/"+domain+"/records/"+id, "", "GET")
	if err != nil {
		return nil, err
	}

	var r DNSRecordResponse
	if err = json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &(r.Record), nil
}

// GetRecords returns the DNS records
func (exo *Client) GetRecords(name string) ([]DNSRecord, error) {
	resp, err := exo.dnsRequest("/v1/domains/"+name+"/records", "", "GET")
	if err != nil {
		return nil, err
	}

	var r []DNSRecordResponse
	if err = json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	records := make([]DNSRecord, 0, len(r))
	for _, rec := range r {
		records = append(records, rec.Record)
	}

	return records, nil
}

// CreateRecord creates a DNS record
func (exo *Client) CreateRecord(name string, rec DNSRecord) (*DNSRecord, error) {
	body, err := json.Marshal(DNSRecordResponse{
		Record: rec,
	})
	if err != nil {
		return nil, err
	}

	resp, err := exo.dnsRequest("/v1/domains/"+name+"/records", string(body), "POST")
	if err != nil {
		return nil, err
	}

	var r DNSRecordResponse
	if err = json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &(r.Record), nil
}

// UpdateRecord updates a DNS record
func (exo *Client) UpdateRecord(name string, rec DNSRecord) (*DNSRecord, error) {
	body, err := json.Marshal(DNSRecordResponse{
		Record: rec,
	})
	if err != nil {
		return nil, err
	}

	id := strconv.FormatInt(rec.ID, 10)
	resp, err := exo.dnsRequest("/v1/domains/"+name+"/records/"+id, string(body), "PUT")
	if err != nil {
		return nil, err
	}

	var r DNSRecordResponse
	if err = json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &(r.Record), nil
}

// DeleteRecord deletes a record
func (exo *Client) DeleteRecord(name string, recordID int64) error {
	id := strconv.FormatInt(recordID, 10)
	_, err := exo.dnsRequest("/v1/domains/"+name+"/records/"+id, "", "DELETE")

	return err
}

func (exo *Client) dnsRequest(uri string, params string, method string) (json.RawMessage, error) {
	url := exo.endpoint + uri
	req, err := http.NewRequest(method, url, strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	var hdr = make(http.Header)
	hdr.Add("X-DNS-TOKEN", exo.apiKey+":"+exo.apiSecret)
	hdr.Add("Accept", "application/json")
	if params != "" {
		hdr.Add("Content-Type", "application/json")
	}
	req.Header = hdr

	response, err := exo.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 400 {
		var e DNSErrorResponse
		if err := json.Unmarshal(b, &e); err != nil {
			return nil, err
		}
		return nil, e.Error()
	}

	return b, nil
}
