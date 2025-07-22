package tailscale

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestProvider_findDomains(t *testing.T) {
	testCases := []struct {
		desc   string
		config dynamic.Configuration
		want   []string
	}{
		{
			desc: "ignore domain with non-matching resolver",
			config: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"foo": {
							Rule: "Host(`machine.http.ts.net`)",
							TLS:  &dynamic.RouterTLSConfig{CertResolver: "bar"},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Rule: "HostSNI(`machine.tcp.ts.net`)",
							TLS:  &dynamic.RouterTCPTLSConfig{CertResolver: "bar"},
						},
					},
				},
			},
		},
		{
			desc: "sanitize domains",
			config: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"dup": {
							Rule: "Host(`machine.http.ts.net`)",
							TLS:  &dynamic.RouterTLSConfig{CertResolver: "foo"},
						},
						"malformed": {
							Rule: "Host(`machine.http.ts.foo`)",
							TLS:  &dynamic.RouterTLSConfig{CertResolver: "foo"},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"dup": {
							Rule: "HostSNI(`machine.http.ts.net`)",
							TLS:  &dynamic.RouterTCPTLSConfig{CertResolver: "foo"},
						},
						"malformed": {
							Rule: "HostSNI(`machine.tcp.ts.foo`)",
							TLS:  &dynamic.RouterTCPTLSConfig{CertResolver: "foo"},
						},
					},
				},
			},
			want: []string{"machine.http.ts.net"},
		},
		{
			desc: "domains from HTTP and TCP router rule",
			config: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"foo": {
							Rule: "Host(`machine.http.ts.net`)",
							TLS:  &dynamic.RouterTLSConfig{CertResolver: "foo"},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Rule: "HostSNI(`machine.tcp.ts.net`)",
							TLS:  &dynamic.RouterTCPTLSConfig{CertResolver: "foo"},
						},
					},
				},
			},
			want: []string{"machine.http.ts.net", "machine.tcp.ts.net"},
		},
		{
			desc: "domains from HTTP and TCP TLS configuration",
			config: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"foo": {
							Rule: "Host(`machine.http.ts.net`)",
							TLS: &dynamic.RouterTLSConfig{
								Domains:      []types.Domain{{Main: "main.http.ts.net"}},
								CertResolver: "foo",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Rule: "HostSNI(`machine.tcp.ts.net`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Domains:      []types.Domain{{Main: "main.tcp.ts.net"}},
								CertResolver: "foo",
							},
						},
					},
				},
			},
			want: []string{"main.http.ts.net", "main.tcp.ts.net"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{ResolverName: "foo"}

			got := p.findDomains(t.Context(), test.config)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestProvider_findNewDomains(t *testing.T) {
	p := Provider{
		ResolverName: "foo",
		certByDomain: map[string]traefiktls.Certificate{
			"foo.com": {},
		},
	}

	got := p.findNewDomains([]string{"foo.com", "bar.com"})
	assert.Equal(t, []string{"bar.com"}, got)
}

func TestProvider_purgeUnusedCerts(t *testing.T) {
	p := Provider{
		ResolverName: "foo",
		certByDomain: map[string]traefiktls.Certificate{
			"foo.com": {},
			"bar.com": {},
		},
	}

	got := p.purgeUnusedCerts([]string{"foo.com"})
	assert.True(t, got)

	assert.Len(t, p.certByDomain, 1)
	assert.Contains(t, p.certByDomain, "foo.com")
}

func TestProvider_sendDynamicConfig(t *testing.T) {
	testCases := []struct {
		desc         string
		certByDomain map[string]traefiktls.Certificate
		want         []*traefiktls.CertAndStores
	}{
		{
			desc: "without certificates",
		},
		{
			desc: "with certificates",
			certByDomain: map[string]traefiktls.Certificate{
				"foo.com": {CertFile: "foo.crt", KeyFile: "foo.key"},
				"bar.com": {CertFile: "bar.crt", KeyFile: "bar.key"},
			},
			want: []*traefiktls.CertAndStores{
				{
					Certificate: traefiktls.Certificate{CertFile: "bar.crt", KeyFile: "bar.key"},
					Stores:      []string{traefiktls.DefaultTLSStoreName},
				},
				{
					Certificate: traefiktls.Certificate{CertFile: "foo.crt", KeyFile: "foo.key"},
					Stores:      []string{traefiktls.DefaultTLSStoreName},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			msgCh := make(chan dynamic.Message, 1)
			p := Provider{
				ResolverName: "foo",
				dynMessages:  msgCh,
				certByDomain: test.certByDomain,
			}

			p.sendDynamicConfig()

			got := <-msgCh

			assert.Equal(t, "foo.tailscale", got.ProviderName)
			assert.NotNil(t, got.Configuration)
			assert.Equal(t, &dynamic.TLSConfiguration{Certificates: test.want}, got.Configuration.TLS)
		})
	}
}

func Test_sanitizeDomains(t *testing.T) {
	testCases := []struct {
		desc    string
		domains []string
		want    []string
	}{
		{
			desc:    "duplicate domains",
			domains: []string{"foo.domain.ts.net", "foo.domain.ts.net"},
			want:    []string{"foo.domain.ts.net"},
		},
		{
			desc:    "not a Tailscale domain",
			domains: []string{"foo.domain.ts.com"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := sanitizeDomains(t.Context(), test.domains)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_isTailscaleDomain(t *testing.T) {
	testCases := []struct {
		desc   string
		domain string
		want   bool
	}{
		{
			desc:   "valid domains",
			domain: "machine.domains.ts.net",
			want:   true,
		},
		{
			desc:   "bad suffix",
			domain: "machine.domains.foo.net",
			want:   false,
		},
		{
			desc:   "too much labels",
			domain: "foo.machine.domains.ts.net",
			want:   false,
		},
		{
			desc:   "not enough labels",
			domain: "domains.ts.net",
			want:   false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := isTailscaleDomain(test.domain)
			assert.Equal(t, test.want, got)
		})
	}
}
