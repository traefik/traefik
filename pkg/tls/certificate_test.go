package tls

import (
	"crypto/x509"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_verifyServerCertMatchesURI(t *testing.T) {
	tests := []struct {
		desc   string
		uri    string
		cert   *x509.Certificate
		expErr require.ErrorAssertionFunc
	}{
		{
			desc:   "returns error when certificate is nil",
			uri:    "spiffe://foo.com",
			expErr: require.Error,
		},
		{
			desc:   "returns error when certificate has no URIs",
			uri:    "spiffe://foo.com",
			cert:   &x509.Certificate{URIs: nil},
			expErr: require.Error,
		},
		{
			desc: "returns error when no URI matches",
			uri:  "spiffe://foo.com",
			cert: &x509.Certificate{URIs: []*url.URL{
				{Scheme: "spiffe", Host: "other.org"},
			}},
			expErr: require.Error,
		},
		{
			desc: "returns nil when URI matches",
			uri:  "spiffe://foo.com",
			cert: &x509.Certificate{URIs: []*url.URL{
				{Scheme: "spiffe", Host: "foo.com"},
			}},
			expErr: require.NoError,
		},
		{
			desc: "returns nil when one of the URI matches",
			uri:  "spiffe://foo.com",
			cert: &x509.Certificate{URIs: []*url.URL{
				{Scheme: "spiffe", Host: "example.org"},
				{Scheme: "spiffe", Host: "foo.com"},
			}},
			expErr: require.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := verifyServerCertMatchesURI(test.uri, test.cert)
			test.expErr(t, err)
		})
	}
}
