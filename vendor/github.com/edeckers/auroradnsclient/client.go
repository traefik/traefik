package auroradnsclient

import (
	"github.com/edeckers/auroradnsclient/requests"
)

// AuroraDNSClient is a client for accessing the Aurora DNS API
type AuroraDNSClient struct {
	requestor *requests.AuroraRequestor
}

// NewAuroraDNSClient instantiates a new client
func NewAuroraDNSClient(endpoint string, userID string, key string) (*AuroraDNSClient, error) {
	requestor, err := requests.NewAuroraRequestor(endpoint, userID, key)
	if err != nil {
		return nil, err
	}

	return &AuroraDNSClient{
		requestor: requestor,
	}, nil
}
