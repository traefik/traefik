package dnsimple

import (
	"fmt"
)

// ZoneRecord represents a DNS record in DNSimple.
type ZoneRecord struct {
	ID           int64    `json:"id,omitempty"`
	ZoneID       string   `json:"zone_id,omitempty"`
	ParentID     int64    `json:"parent_id,omitempty"`
	Type         string   `json:"type,omitempty"`
	Name         string   `json:"name"`
	Content      string   `json:"content,omitempty"`
	TTL          int      `json:"ttl,omitempty"`
	Priority     int      `json:"priority,omitempty"`
	SystemRecord bool     `json:"system_record,omitempty"`
	Regions      []string `json:"regions,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
}

func zoneRecordPath(accountID string, zoneName string, recordID int64) (path string) {
	path = fmt.Sprintf("/%v/zones/%v/records", accountID, zoneName)
	if recordID != 0 {
		path += fmt.Sprintf("/%v", recordID)
	}
	return
}

// zoneRecordResponse represents a response from an API method that returns a ZoneRecord struct.
type zoneRecordResponse struct {
	Response
	Data *ZoneRecord `json:"data"`
}

// zoneRecordsResponse represents a response from an API method that returns a collection of ZoneRecord struct.
type zoneRecordsResponse struct {
	Response
	Data []ZoneRecord `json:"data"`
}

// ZoneRecordListOptions specifies the optional parameters you can provide
// to customize the ZonesService.ListZoneRecords method.
type ZoneRecordListOptions struct {
	// Select records where the name matches given string.
	Name string `url:"name,omitempty"`

	// Select records where the name contains given string.
	NameLike string `url:"name_like,omitempty"`

	// Select records of given type.
	// Eg. TXT, A, NS.
	Type string `url:"type,omitempty"`

	ListOptions
}

// ListRecords lists the zone records for a zone.
//
// See https://developer.dnsimple.com/v2/zones/records/#listZoneRecords
func (s *ZonesService) ListRecords(accountID string, zoneName string, options *ZoneRecordListOptions) (*zoneRecordsResponse, error) {
	path := versioned(zoneRecordPath(accountID, zoneName, 0))
	recordsResponse := &zoneRecordsResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, recordsResponse)
	if err != nil {
		return nil, err
	}

	recordsResponse.HttpResponse = resp
	return recordsResponse, nil
}

// CreateRecord creates a zone record.
//
// See https://developer.dnsimple.com/v2/zones/records/#createZoneRecord
func (s *ZonesService) CreateRecord(accountID string, zoneName string, recordAttributes ZoneRecord) (*zoneRecordResponse, error) {
	path := versioned(zoneRecordPath(accountID, zoneName, 0))
	recordResponse := &zoneRecordResponse{}

	resp, err := s.client.post(path, recordAttributes, recordResponse)
	if err != nil {
		return nil, err
	}

	recordResponse.HttpResponse = resp
	return recordResponse, nil
}

// GetRecord fetches a zone record.
//
// See https://developer.dnsimple.com/v2/zones/records/#getZoneRecord
func (s *ZonesService) GetRecord(accountID string, zoneName string, recordID int64) (*zoneRecordResponse, error) {
	path := versioned(zoneRecordPath(accountID, zoneName, recordID))
	recordResponse := &zoneRecordResponse{}

	resp, err := s.client.get(path, recordResponse)
	if err != nil {
		return nil, err
	}

	recordResponse.HttpResponse = resp
	return recordResponse, nil
}

// UpdateRecord updates a zone record.
//
// See https://developer.dnsimple.com/v2/zones/records/#updateZoneRecord
func (s *ZonesService) UpdateRecord(accountID string, zoneName string, recordID int64, recordAttributes ZoneRecord) (*zoneRecordResponse, error) {
	path := versioned(zoneRecordPath(accountID, zoneName, recordID))
	recordResponse := &zoneRecordResponse{}
	resp, err := s.client.patch(path, recordAttributes, recordResponse)

	if err != nil {
		return nil, err
	}

	recordResponse.HttpResponse = resp
	return recordResponse, nil
}

// DeleteRecord PERMANENTLY deletes a zone record from the zone.
//
// See https://developer.dnsimple.com/v2/zones/records/#deleteZoneRecord
func (s *ZonesService) DeleteRecord(accountID string, zoneName string, recordID int64) (*zoneRecordResponse, error) {
	path := versioned(zoneRecordPath(accountID, zoneName, recordID))
	recordResponse := &zoneRecordResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	recordResponse.HttpResponse = resp
	return recordResponse, nil
}
