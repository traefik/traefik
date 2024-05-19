package consulcatalog

import (
	"fmt"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// connectCert holds our certificates as a client of the Consul Connect protocol.
type connectCert struct {
	root []string
	leaf keyPair
}

func (c *connectCert) getRoot() []types.FileOrContent {
	var result []types.FileOrContent
	for _, r := range c.root {
		result = append(result, types.FileOrContent(r))
	}
	return result
}

func (c *connectCert) getLeaf() traefiktls.Certificate {
	return traefiktls.Certificate{
		CertFile: types.FileOrContent(c.leaf.cert),
		KeyFile:  types.FileOrContent(c.leaf.key),
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
	spiffeID := fmt.Sprintf("spiffe:///ns/%s/dc/%s/svc/%s",
		item.Namespace,
		item.Datacenter,
		item.Name,
	)

	return &dynamic.ServersTransport{
		// This ensures that the config changes whenever the verifier function changes
		ServerName: fmt.Sprintf("%s-%s-%s", item.Namespace, item.Datacenter, item.Name),
		// InsecureSkipVerify is needed because Go wants to verify a hostname otherwise
		InsecureSkipVerify: true,
		RootCAs:            c.getRoot(),
		Certificates: traefiktls.Certificates{
			c.getLeaf(),
		},
		PeerCertURI: spiffeID,
	}
}

func (c *connectCert) tcpServersTransport(item itemData) *dynamic.TCPServersTransport {
	spiffeID := fmt.Sprintf("spiffe:///ns/%s/dc/%s/svc/%s",
		item.Namespace,
		item.Datacenter,
		item.Name,
	)

	return &dynamic.TCPServersTransport{
		TLS: &dynamic.TLSClientConfig{
			// This ensures that the config changes whenever the verifier function changes
			ServerName: fmt.Sprintf("%s-%s-%s", item.Namespace, item.Datacenter, item.Name),
			// InsecureSkipVerify is needed because Go wants to verify a hostname otherwise
			InsecureSkipVerify: true,
			RootCAs:            c.getRoot(),
			Certificates: traefiktls.Certificates{
				c.getLeaf(),
			},
			PeerCertURI: spiffeID,
		},
	}
}
