package egoscale

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (exo *Client) CreateDomain(name string) (*DNSDomain, error) {
	var hdr = make(http.Header)
	var domain DNSDomainCreateRequest

	hdr.Add("X-DNS-TOKEN", exo.apiKey + ":" + exo.apiSecret)
	domain.Domain.Name = name
	m, err := json.Marshal(domain); if err != nil {
		return nil, err
	}

	resp, err := exo.DetailedRequest("/v1/domains", string(m), "POST", hdr)
	if err != nil {
		return nil, err
	}

	var d *DNSDomain
	if err := json.Unmarshal(resp, &d); err != nil {
		return nil, err
	}

	return d, nil
}

func (exo *Client) GetDomain(name string) (*DNSDomain, error) {
	var hdr = make(http.Header)

	hdr.Add("X-DNS-TOKEN", exo.apiKey + ":" + exo.apiSecret)
	hdr.Add("Accept", "application/json")

	resp, err := exo.DetailedRequest("/v1/domains/" + name, "", "GET", hdr)
	if err != nil {
		return nil, err
	}

	var d *DNSDomain
	if err := json.Unmarshal(resp, &d); err != nil {
		return nil, err
	}

	return d, nil
}

func (exo *Client) DeleteDomain(name string) (error) {
	var hdr = make(http.Header)
	hdr.Add("X-DNS-TOKEN", exo.apiKey + ":" + exo.apiSecret)
	hdr.Add("Accept", "application/json")

	_, err := exo.DetailedRequest("/v1/domains/" + name, "", "DELETE", hdr)
	if err != nil {
		return err
	}

	return nil
}

func (exo *Client) GetRecords(name string) ([]*DNSRecordResponse, error) {
	var hdr = make(http.Header)
	hdr.Add("X-DNS-TOKEN", exo.apiKey + ":" + exo.apiSecret)
	hdr.Add("Accept", "application/json")

	resp, err := exo.DetailedRequest("/v1/domains/" + name + "/records", "", "GET", hdr)
	if err != nil {
		return nil, err
	}

	var r []*DNSRecordResponse
	if err = json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r, nil
}

func(exo *Client) CreateRecord(name string, rec DNSRecord) (*DNSRecordResponse, error) {
	var hdr = make(http.Header)
	hdr.Add("X-DNS-TOKEN", exo.apiKey + ":" + exo.apiSecret)
	hdr.Add("Accept", "application/json")
	hdr.Add("Content-Type", "application/json")

	var rr DNSRecordResponse
	rr.Record = rec

	body, err := json.Marshal(rr); if err != nil {
		return nil, err
	}

	resp, err := exo.DetailedRequest("/v1/domains/" + name + "/records", string(body), "POST", hdr)
	if err != nil {
		return nil, err
	}

	var r *DNSRecordResponse
	if err = json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r, nil	
}

func(exo *Client) UpdateRecord(name string, rec DNSRecord) (*DNSRecordResponse, error) {
	var hdr = make(http.Header)
	id := strconv.FormatInt(rec.Id, 10)
	hdr.Add("X-DNS-TOKEN", exo.apiKey + ":" + exo.apiSecret)
	hdr.Add("Accept", "application/json")
	hdr.Add("Content-Type", "application/json")

	var rr DNSRecordResponse
	rr.Record = rec

	body, err := json.Marshal(rr); if err != nil {
		return nil, err
	}

	resp, err := exo.DetailedRequest("/v1/domains/" + name + "/records/" + id,
		string(body), "PUT", hdr)
	if err != nil {
		return nil, err
	}

	var r *DNSRecordResponse
	if err = json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r, nil
}

func(exo *Client) DeleteRecord(name string, rec DNSRecord) (error) {
	var hdr = make(http.Header)
	id := strconv.FormatInt(rec.Id, 10)
	hdr.Add("X-DNS-TOKEN", exo.apiKey + ":" + exo.apiSecret)
	hdr.Add("Accept", "application/json")
	hdr.Add("Content-Type", "application/json")

	_, err := exo.DetailedRequest("/v1/domains/" + name + "/records/" + id,
		"", "DELETE", hdr)
	if err != nil {
		return err
	}

	return nil
}