package tls

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestUpdateConfigsReportsInvalidInitialTLS(t *testing.T) {
	tests := []struct {
		name    string
		stores  map[string]Store
		options map[string]Options
		certs   []*CertAndStores
	}{
		{name: "certificate", certs: []*CertAndStores{{CertFile: "/missing-cert", KeyFile: "/missing-key"}}},
		{name: "default certificate", stores: map[string]Store{"default": {DefaultCertificate: &Certificate{CertFile: "/missing-cert", KeyFile: "/missing-key"}}}},
		{name: "client CA", options: map[string]Options{"default": {ClientAuth: ClientAuth{ClientAuthType: "RequireAndVerifyClientCert", CAFiles: []types.FileOrContent{"/missing-ca"}}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Error(t, NewManager(nil).UpdateConfigsWithError(t.Context(), test.stores, test.options, test.certs))
		})
	}
}

func TestUpdateConfigsAllowsEmptyInitialStore(t *testing.T) {
	require.NoError(t, NewManager(nil).UpdateConfigsWithError(t.Context(), nil, nil, nil))
}
