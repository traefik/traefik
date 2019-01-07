package dnsimple

import "fmt"

// ZoneDistribution is the result of the zone distribution check.
type ZoneDistribution struct {
	Distributed bool `json:"distributed"`
}

// zoneDistributionResponse represents a response from an API method that returns a ZoneDistribution struct.
type zoneDistributionResponse struct {
	Response
	Data *ZoneDistribution `json:"data"`
}

// CheckZoneDistribution checks if a zone is fully distributed across DNSimple nodes.
//
// See https://developer.dnsimple.com/v2/zones/#checkZoneDistribution
func (s *ZonesService) CheckZoneDistribution(accountID string, zoneName string) (*zoneDistributionResponse, error) {
	path := versioned(fmt.Sprintf("/%v/zones/%v/distribution", accountID, zoneName))
	zoneDistributionResponse := &zoneDistributionResponse{}

	resp, err := s.client.get(path, zoneDistributionResponse)
	if err != nil {
		return nil, err
	}

	zoneDistributionResponse.HttpResponse = resp
	return zoneDistributionResponse, nil
}

// CheckZoneRecordDistribution checks if a zone is fully distributed across DNSimple nodes.
//
// See https://developer.dnsimple.com/v2/zones/#checkZoneRecordDistribution
func (s *ZonesService) CheckZoneRecordDistribution(accountID string, zoneName string, recordID int64) (*zoneDistributionResponse, error) {
	path := versioned(fmt.Sprintf("/%v/zones/%v/records/%v/distribution", accountID, zoneName, recordID))
	zoneDistributionResponse := &zoneDistributionResponse{}

	resp, err := s.client.get(path, zoneDistributionResponse)
	if err != nil {
		return nil, err
	}

	zoneDistributionResponse.HttpResponse = resp
	return zoneDistributionResponse, nil
}
