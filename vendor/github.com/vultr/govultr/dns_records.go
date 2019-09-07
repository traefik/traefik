package govultr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// DNSRecordService is the interface to interact with the DNS Records endpoints on the Vultr API
// Link: https://www.vultr.com/api/#dns
type DNSRecordService interface {
	Create(ctx context.Context, domain, recordType, name, data string, ttl, priority int) error
	Delete(ctx context.Context, domain, recordID string) error
	List(ctx context.Context, domain string) ([]DNSRecord, error)
	Update(ctx context.Context, domain string, dnsRecord *DNSRecord) error
}

// DNSRecordsServiceHandler handles interaction with the DNS Records methods for the Vultr API
type DNSRecordsServiceHandler struct {
	client *Client
}

// DNSRecord represents a DNS record on Vultr
type DNSRecord struct {
	RecordID int    `json:"RECORDID"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	Priority int    `json:"priority"`
	TTL      int    `json:"ttl"`
}

// Create will add a DNS record.
func (d *DNSRecordsServiceHandler) Create(ctx context.Context, domain, recordType, name, data string, ttl, priority int) error {

	uri := "/v1/dns/create_record"

	values := url.Values{
		"domain":   {domain},
		"name":     {name},
		"type":     {recordType},
		"data":     {data},
		"ttl":      {strconv.Itoa(ttl)},
		"priority": {strconv.Itoa(priority)},
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

// Delete will delete a domain name and all associated records.
func (d *DNSRecordsServiceHandler) Delete(ctx context.Context, domain, recordID string) error {

	uri := "/v1/dns/delete_record"

	values := url.Values{
		"domain":   {domain},
		"RECORDID": {recordID},
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

// List will list all the records associated with a particular domain on Vultr
func (d *DNSRecordsServiceHandler) List(ctx context.Context, domain string) ([]DNSRecord, error) {

	uri := "/v1/dns/records"

	req, err := d.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("domain", domain)
	req.URL.RawQuery = q.Encode()

	var dnsRecord []DNSRecord
	err = d.client.DoWithContext(ctx, req, &dnsRecord)

	if err != nil {
		return nil, err
	}

	return dnsRecord, nil
}

// Update will update a DNS record
func (d *DNSRecordsServiceHandler) Update(ctx context.Context, domain string, dnsRecord *DNSRecord) error {

	uri := "/v1/dns/update_record"

	values := url.Values{
		"domain":   {domain},
		"RECORDID": {strconv.Itoa(dnsRecord.RecordID)},
	}

	// Optional
	if dnsRecord.Name != "" {
		values.Add("name", dnsRecord.Name)
	}
	if dnsRecord.Data != "" {
		values.Add("data", dnsRecord.Data)
	}
	if dnsRecord.TTL != 0 {
		values.Add("ttl", strconv.Itoa(dnsRecord.TTL))
	}
	if dnsRecord.Priority != 0 {
		values.Add("priority", strconv.Itoa(dnsRecord.Priority))
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
