package dnsimple

import (
	"fmt"
)

// VanityNameServersService handles communication with Vanity Name Servers
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/vanity/
type VanityNameServersService struct {
	client *Client
}

// VanityNameServer represents data for a single vanity name server
type VanityNameServer struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	IPv4      string `json:"ipv4,omitempty"`
	IPv6      string `json:"ipv6,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func vanityNameServerPath(accountID string, domainIdentifier string) string {
	return fmt.Sprintf("/%v/vanity/%v", accountID, domainIdentifier)
}

// vanityNameServerResponse represents a response for vanity name server enable and disable operations.
type vanityNameServerResponse struct {
	Response
	Data []VanityNameServer `json:"data"`
}

// EnableVanityNameServers Vanity Name Servers for the given domain
//
// See https://developer.dnsimple.com/v2/vanity/#enable
func (s *VanityNameServersService) EnableVanityNameServers(accountID string, domainIdentifier string) (*vanityNameServerResponse, error) {
	path := versioned(vanityNameServerPath(accountID, domainIdentifier))
	vanityNameServerResponse := &vanityNameServerResponse{}

	resp, err := s.client.put(path, nil, vanityNameServerResponse)
	if err != nil {
		return nil, err
	}

	vanityNameServerResponse.HttpResponse = resp
	return vanityNameServerResponse, nil
}

// DisableVanityNameServers Vanity Name Servers for the given domain
//
// See https://developer.dnsimple.com/v2/vanity/#disable
func (s *VanityNameServersService) DisableVanityNameServers(accountID string, domainIdentifier string) (*vanityNameServerResponse, error) {
	path := versioned(vanityNameServerPath(accountID, domainIdentifier))
	vanityNameServerResponse := &vanityNameServerResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	vanityNameServerResponse.HttpResponse = resp
	return vanityNameServerResponse, nil
}
