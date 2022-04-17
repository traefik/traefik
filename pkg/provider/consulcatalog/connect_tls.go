package consulcatalog

import (
	"fmt"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
)

// connectCert holds our certificates as a client of the Consul Connect protocol.
type connectCert struct {
	root []string
	leaf keyPair
}

func (c *connectCert) getRoot() []traefiktls.FileOrContent {
	var result []traefiktls.FileOrContent
	for _, r := range c.root {
		result = append(result, traefiktls.FileOrContent(r))
	}
	return result
}

func (c *connectCert) getLeaf() traefiktls.Certificate {
	return traefiktls.Certificate{
		CertFile: traefiktls.FileOrContent(c.leaf.cert),
		KeyFile:  traefiktls.FileOrContent(c.leaf.key),
	}
}

func (c *connectCert) isReady() bool {
	return c != nil && len(c.root) > 0 && c.leaf.cert != "" && c.leaf.key != ""
}

func (c *connectCert) equals(other *connectCert) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	if len(c.root) != len(other.root) {
		return false
	}
	for i, v := range c.root {
		if v != other.root[i] {
			return false
		}
	}
	return c.leaf == other.leaf
}

func (c *connectCert) serversTransport(item itemData) *dynamic.ServersTransport {
	spiffeIDService := connect.SpiffeIDService{
		Namespace:  item.Namespace,
		Datacenter: item.Datacenter,
		Service:    item.Name,
	}

	return &dynamic.ServersTransport{
		// This ensures that the config changes whenever the verifier function changes
		ServerName: fmt.Sprintf("%s-%s-%s", item.Namespace, item.Datacenter, item.Name),
		// InsecureSkipVerify is needed because Go wants to verify a hostname otherwise
		InsecureSkipVerify: true,
		RootCAs:            c.getRoot(),
		Certificates: traefiktls.Certificates{
			c.getLeaf(),
		},
		PeerCertURI: spiffeIDService.URI().String(),
	}
}
