package dnsimple

import "fmt"

// DelegationSignerRecord represents a delegation signer record for a domain in DNSimple.
type DelegationSignerRecord struct {
	ID         int64  `json:"id,omitempty"`
	DomainID   int64  `json:"domain_id,omitempty"`
	Algorithm  string `json:"algorithm"`
	Digest     string `json:"digest"`
	DigestType string `json:"digest_type"`
	Keytag     string `json:"keytag"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

func delegationSignerRecordPath(accountID string, domainIdentifier string, dsRecordID int64) (path string) {
	path = fmt.Sprintf("%v/ds_records", domainPath(accountID, domainIdentifier))
	if dsRecordID != 0 {
		path += fmt.Sprintf("/%v", dsRecordID)
	}
	return
}

// delegationSignerRecordResponse represents a response from an API method that returns a DelegationSignerRecord struct.
type delegationSignerRecordResponse struct {
	Response
	Data *DelegationSignerRecord `json:"data"`
}

// delegationSignerRecordResponse represents a response from an API method that returns a DelegationSignerRecord struct.
type delegationSignerRecordsResponse struct {
	Response
	Data []DelegationSignerRecord `json:"data"`
}

// ListDelegationSignerRecords lists the delegation signer records for a domain.
//
// See https://developer.dnsimple.com/v2/domains/dnssec/#ds-record-list
func (s *DomainsService) ListDelegationSignerRecords(accountID string, domainIdentifier string, options *ListOptions) (*delegationSignerRecordsResponse, error) {
	path := versioned(delegationSignerRecordPath(accountID, domainIdentifier, 0))
	dsRecordsResponse := &delegationSignerRecordsResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, dsRecordsResponse)
	if err != nil {
		return nil, err
	}

	dsRecordsResponse.HttpResponse = resp
	return dsRecordsResponse, nil
}

// CreateDelegationSignerRecord creates a new delegation signer record.
//
// See https://developer.dnsimple.com/v2/domains/dnssec/#ds-record-create
func (s *DomainsService) CreateDelegationSignerRecord(accountID string, domainIdentifier string, dsRecordAttributes DelegationSignerRecord) (*delegationSignerRecordResponse, error) {
	path := versioned(delegationSignerRecordPath(accountID, domainIdentifier, 0))
	dsRecordResponse := &delegationSignerRecordResponse{}

	resp, err := s.client.post(path, dsRecordAttributes, dsRecordResponse)
	if err != nil {
		return nil, err
	}

	dsRecordResponse.HttpResponse = resp
	return dsRecordResponse, nil
}

// GetDelegationSignerRecord fetches a delegation signer record.
//
// See https://developer.dnsimple.com/v2/domains/dnssec/#ds-record-get
func (s *DomainsService) GetDelegationSignerRecord(accountID string, domainIdentifier string, dsRecordID int64) (*delegationSignerRecordResponse, error) {
	path := versioned(delegationSignerRecordPath(accountID, domainIdentifier, dsRecordID))
	dsRecordResponse := &delegationSignerRecordResponse{}

	resp, err := s.client.get(path, dsRecordResponse)
	if err != nil {
		return nil, err
	}

	dsRecordResponse.HttpResponse = resp
	return dsRecordResponse, nil
}

// DeleteDelegationSignerRecord PERMANENTLY deletes a delegation signer record
// from the domain.
//
// See https://developer.dnsimple.com/v2/domains/dnssec/#ds-record-delete
func (s *DomainsService) DeleteDelegationSignerRecord(accountID string, domainIdentifier string, dsRecordID int64) (*delegationSignerRecordResponse, error) {
	path := versioned(delegationSignerRecordPath(accountID, domainIdentifier, dsRecordID))
	dsRecordResponse := &delegationSignerRecordResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	dsRecordResponse.HttpResponse = resp
	return dsRecordResponse, nil
}
