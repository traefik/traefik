package lib

import (
	"fmt"
	"net/url"
)

// DNSDomain represents a DNS domain on Vultr
type DNSDomain struct {
	Domain  string `json:"domain"`
	Created string `json:"date_created"`
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

// GetDNSDomains returns a list of available domains on Vultr account
func (c *Client) GetDNSDomains() (dnsdomains []DNSDomain, err error) {
	if err := c.get(`dns/list`, &dnsdomains); err != nil {
		return nil, err
	}
	return dnsdomains, nil
}

// GetDNSRecords returns a list of all DNS records of a particular domain
func (c *Client) GetDNSRecords(domain string) (dnsrecords []DNSRecord, err error) {
	if err := c.get(`dns/records?domain=`+domain, &dnsrecords); err != nil {
		return nil, err
	}
	return dnsrecords, nil
}

// CreateDNSDomain creates a new DNS domain name on Vultr
func (c *Client) CreateDNSDomain(domain, serverIP string) error {
	values := url.Values{
		"domain":   {domain},
		"serverIP": {serverIP},
	}

	if err := c.post(`dns/create_domain`, values, nil); err != nil {
		return err
	}
	return nil
}

// DeleteDNSDomain deletes an existing DNS domain name
func (c *Client) DeleteDNSDomain(domain string) error {
	values := url.Values{
		"domain": {domain},
	}

	if err := c.post(`dns/delete_domain`, values, nil); err != nil {
		return err
	}
	return nil
}

// CreateDNSRecord creates a new DNS record
func (c *Client) CreateDNSRecord(domain, name, rtype, data string, priority, ttl int) error {
	values := url.Values{
		"domain":   {domain},
		"name":     {name},
		"type":     {rtype},
		"data":     {data},
		"priority": {fmt.Sprintf("%v", priority)},
		"ttl":      {fmt.Sprintf("%v", ttl)},
	}

	if err := c.post(`dns/create_record`, values, nil); err != nil {
		return err
	}
	return nil
}

// UpdateDNSRecord updates an existing DNS record
func (c *Client) UpdateDNSRecord(domain string, dnsrecord DNSRecord) error {
	values := url.Values{
		"domain":   {domain},
		"RECORDID": {fmt.Sprintf("%v", dnsrecord.RecordID)},
	}

	if dnsrecord.Name != "" {
		values.Add("name", dnsrecord.Name)
	}
	if dnsrecord.Data != "" {
		values.Add("data", dnsrecord.Data)
	}
	if dnsrecord.Priority != 0 {
		values.Add("priority", fmt.Sprintf("%v", dnsrecord.Priority))
	}
	if dnsrecord.TTL != 0 {
		values.Add("ttl", fmt.Sprintf("%v", dnsrecord.TTL))
	}

	if err := c.post(`dns/update_record`, values, nil); err != nil {
		return err
	}
	return nil
}

// DeleteDNSRecord deletes an existing DNS record
func (c *Client) DeleteDNSRecord(domain string, recordID int) error {
	values := url.Values{
		"domain":   {domain},
		"RECORDID": {fmt.Sprintf("%v", recordID)},
	}

	if err := c.post(`dns/delete_record`, values, nil); err != nil {
		return err
	}
	return nil
}
