package namecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

var _ = bytes.MinRead

// ListRecords returns all records for a zone.
func (n *NameCom) ListRecords(request *ListRecordsRequest) (*ListRecordsResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/records", request.DomainName)

	values := url.Values{}
	if request.PerPage != 0 {
		values.Set("perPage", fmt.Sprintf("%d", request.PerPage))
	}
	if request.Page != 0 {
		values.Set("page", fmt.Sprintf("%d", request.Page))
	}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &ListRecordsResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetRecord returns details about an individual record.
func (n *NameCom) GetRecord(request *GetRecordRequest) (*Record, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/records/%d", request.DomainName, request.ID)

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &Record{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateRecord creates a new record in the zone.
func (n *NameCom) CreateRecord(request *Record) (*Record, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/records", request.DomainName)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Record{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UpdateRecord replaces the record with the new record that is passed.
func (n *NameCom) UpdateRecord(request *Record) (*Record, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/records/%d", request.DomainName, request.ID)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.put(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Record{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteRecord deletes a record from the zone.
func (n *NameCom) DeleteRecord(request *DeleteRecordRequest) (*EmptyResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/records/%d", request.DomainName, request.ID)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.delete(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &EmptyResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
