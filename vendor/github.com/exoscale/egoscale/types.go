package egoscale

import (
	"net/http"
)

// Client represents the CloudStack API client
type Client struct {
	client    *http.Client
	endpoint  string
	apiKey    string
	apiSecret string
}

// Topology represents a view of the servers
type Topology struct {
	Zones          map[string]string
	Images         map[string]map[int64]string
	Profiles       map[string]string
	Keypairs       []string
	SecurityGroups map[string]string
	AffinityGroups map[string]string
}
