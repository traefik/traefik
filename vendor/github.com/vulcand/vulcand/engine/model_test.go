package engine

import (
	"crypto/tls"
	"encoding/json"
	"testing"
	"time"

	"github.com/vulcand/route"
	"github.com/vulcand/vulcand/plugin"
	"github.com/vulcand/vulcand/plugin/connlimit"
	. "gopkg.in/check.v1"
)

func TestBackend(t *testing.T) { TestingT(t) }

type BackendSuite struct {
}

var _ = Suite(&BackendSuite{})

func (s *BackendSuite) TestHostNew(c *C) {
	h, err := NewHost("localhost", HostSettings{})
	c.Assert(err, IsNil)
	c.Assert(h.Name, Equals, "localhost")
	c.Assert(h.Name, Equals, h.GetId())
	c.Assert(h.String(), Not(Equals), "")
}

func (s *BackendSuite) TestHostBad(c *C) {
	h, err := NewHost("", HostSettings{})
	c.Assert(err, NotNil)
	c.Assert(h, IsNil)
}

func (s *BackendSuite) TestFrontendDefaults(c *C) {
	f, err := NewHTTPFrontend(route.NewMux(), "f1", "b1", `Path("/home")`, HTTPFrontendSettings{})
	c.Assert(err, IsNil)
	c.Assert(f.GetId(), Equals, "f1")
	c.Assert(f.String(), Not(Equals), "")
	c.Assert(f.Route, Equals, `Path("/home")`)
}

func (s *BackendSuite) TestNewFrontendWithOptions(c *C) {
	settings := HTTPFrontendSettings{
		Limits: HTTPFrontendLimits{
			MaxMemBodyBytes: 12,
			MaxBodyBytes:    400,
		},
		FailoverPredicate:  "IsNetworkError() && Attempts() <= 1",
		Hostname:           "host1",
		TrustForwardHeader: true,
	}
	f, err := NewHTTPFrontend(route.NewMux(), "f1", "b1", `Path("/home")`, settings)
	c.Assert(err, IsNil)
	c.Assert(f.Id, Equals, "f1")

	o := f.HTTPSettings()

	c.Assert(o.Limits.MaxMemBodyBytes, Equals, int64(12))
	c.Assert(o.Limits.MaxBodyBytes, Equals, int64(400))

	c.Assert(o.FailoverPredicate, NotNil)
	c.Assert(o.TrustForwardHeader, Equals, true)
	c.Assert(o.Hostname, Equals, "host1")
}

func (s *BackendSuite) TestFrontendBadParams(c *C) {
	// Bad route
	_, err := NewHTTPFrontend(route.NewMux(), "f1", "b1", "/home  -- afawf \\~", HTTPFrontendSettings{})
	c.Assert(err, NotNil)

	// Empty params
	_, err = NewHTTPFrontend(route.NewMux(), "", "", "", HTTPFrontendSettings{})
	c.Assert(err, NotNil)
}

func (s *BackendSuite) TestFrontendBadOptions(c *C) {
	settings := []HTTPFrontendSettings{
		HTTPFrontendSettings{
			FailoverPredicate: "bad predicate",
		},
	}
	for _, s := range settings {
		f, err := NewHTTPFrontend(route.NewMux(), "f1", "b", `Path("/home")`, s)
		c.Assert(err, NotNil)
		c.Assert(f, IsNil)
	}
}

func (s *BackendSuite) TestBackendNew(c *C) {
	b, err := NewHTTPBackend("b1", HTTPBackendSettings{})
	c.Assert(err, IsNil)
	c.Assert(b.Type, Equals, HTTP)
	c.Assert(b.GetId(), Equals, "b1")
	c.Assert(b.String(), Not(Equals), "")
}

func (s *BackendSuite) TestNewBackendWithOptions(c *C) {
	options := HTTPBackendSettings{
		Timeouts: HTTPBackendTimeouts{
			Read:         "1s",
			Dial:         "2s",
			TLSHandshake: "3s",
		},
		KeepAlive: HTTPBackendKeepAlive{
			Period:              "4s",
			MaxIdleConnsPerHost: 3,
		},
	}
	b, err := NewHTTPBackend("b1", options)
	c.Assert(err, IsNil)
	c.Assert(b.GetId(), Equals, "b1")

	o, err := b.TransportSettings()
	c.Assert(err, IsNil)

	c.Assert(o.Timeouts.Read, Equals, time.Second)
	c.Assert(o.Timeouts.Dial, Equals, 2*time.Second)
	c.Assert(o.Timeouts.TLSHandshake, Equals, 3*time.Second)

	c.Assert(o.KeepAlive.Period, Equals, 4*time.Second)
	c.Assert(o.KeepAlive.MaxIdleConnsPerHost, Equals, 3)
}

func (s *BackendSuite) TestBackendSettingsEq(c *C) {
	options := []struct {
		a HTTPBackendSettings
		b HTTPBackendSettings
		e bool
	}{
		{
			a: HTTPBackendSettings{},
			b: HTTPBackendSettings{},
			e: true,
		},

		{
			a: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{Dial: "1s"}},
			b: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{Dial: "1s"}},
			e: true,
		},
		{
			a: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{Dial: "2s"}},
			b: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{Dial: "1s"}},
			e: false,
		},
		{
			a: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{Read: "2s"}},
			b: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{Read: "1s"}},
			e: false,
		},
		{
			a: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{TLSHandshake: "2s"}},
			b: HTTPBackendSettings{Timeouts: HTTPBackendTimeouts{TLSHandshake: "1s"}},
			e: false,
		},

		{
			a: HTTPBackendSettings{KeepAlive: HTTPBackendKeepAlive{Period: "2s"}},
			b: HTTPBackendSettings{KeepAlive: HTTPBackendKeepAlive{Period: "1s"}},
			e: false,
		},
		{
			a: HTTPBackendSettings{KeepAlive: HTTPBackendKeepAlive{MaxIdleConnsPerHost: 1}},
			b: HTTPBackendSettings{KeepAlive: HTTPBackendKeepAlive{MaxIdleConnsPerHost: 2}},
			e: false,
		},

		{
			a: HTTPBackendSettings{TLS: &TLSSettings{}},
			b: HTTPBackendSettings{TLS: &TLSSettings{}},
			e: true,
		},
		{
			a: HTTPBackendSettings{TLS: &TLSSettings{}},
			b: HTTPBackendSettings{TLS: &TLSSettings{SessionTicketsDisabled: true}},
			e: false,
		},
	}
	for _, o := range options {
		c.Assert(o.a.Equals(o.b), Equals, o.e)
	}
}

func (s *BackendSuite) TestOCSPSettingsEq(c *C) {
	options := []struct {
		a *OCSPSettings
		b *OCSPSettings
		e bool
	}{
		{&OCSPSettings{}, &OCSPSettings{}, true},
		{&OCSPSettings{Period: "2m0s"}, &OCSPSettings{Period: "2m"}, true},
		{&OCSPSettings{Period: "2m0s"}, &OCSPSettings{Period: "3m"}, false},
		{&OCSPSettings{Period: "bla"}, &OCSPSettings{Period: "2m"}, false},
		{&OCSPSettings{Period: "2m"}, &OCSPSettings{Period: "bla"}, false},
		{&OCSPSettings{Enabled: true}, &OCSPSettings{Enabled: false}, false},
		{
			&OCSPSettings{Enabled: true, Responders: []string{"http://a.com", "http://b.com"}},
			&OCSPSettings{Enabled: true, Responders: []string{"http://a.com", "http://b.com"}},
			true,
		},
		{
			&OCSPSettings{Enabled: true, Responders: []string{"http://a.com", "http://b.com"}},
			&OCSPSettings{Enabled: true, Responders: []string{"http://a.com"}},
			false,
		},
		{
			&OCSPSettings{Enabled: true, Responders: []string{"http://a.com", "http://b.com"}},
			&OCSPSettings{Enabled: true, Responders: []string{"http://a.com", "http://c.com"}},
			false,
		},
	}
	for _, o := range options {
		c.Assert(o.a.Equals(o.b), Equals, o.e)
	}
}

func (s *BackendSuite) TestListenerSettingsEq(c *C) {
	options := []struct {
		a Listener
		b Listener
		e bool
		c string
	}{
		{
			a: Listener{},
			b: Listener{},
			e: true,
			c: "empty",
		},
		{
			a: Listener{Settings: &HTTPSListenerSettings{}},
			b: Listener{Settings: &HTTPSListenerSettings{}},
			e: true,
			c: "emtpy https",
		},
		{
			a: Listener{Settings: &HTTPSListenerSettings{TLS: TLSSettings{}}},
			b: Listener{Settings: &HTTPSListenerSettings{TLS: TLSSettings{}}},
			e: true,
			c: "default https",
		},
		{
			a: Listener{Settings: &HTTPSListenerSettings{TLS: TLSSettings{SessionTicketsDisabled: false}}},
			b: Listener{Settings: &HTTPSListenerSettings{TLS: TLSSettings{}}},
			e: true,
			c: "different https",
		},
		{
			a: Listener{Settings: &HTTPSListenerSettings{TLS: TLSSettings{SessionTicketsDisabled: true}}},
			b: Listener{Settings: &HTTPSListenerSettings{TLS: TLSSettings{}}},
			e: false,
			c: "session tickets",
		},
	}
	for _, o := range options {
		c.Assert((&o.a).SettingsEquals(&o.b), Equals, o.e, Commentf("TC: %v", o.c))
	}
}

func (s *BackendSuite) TestNewBackendWithBadOptions(c *C) {
	options := []HTTPBackendSettings{
		HTTPBackendSettings{
			Timeouts: HTTPBackendTimeouts{
				Read: "1what?",
			},
		},
		HTTPBackendSettings{
			Timeouts: HTTPBackendTimeouts{
				Dial: "1what?",
			},
		},
		HTTPBackendSettings{
			Timeouts: HTTPBackendTimeouts{
				TLSHandshake: "1what?",
			},
		},
		HTTPBackendSettings{
			KeepAlive: HTTPBackendKeepAlive{
				Period: "1what?",
			},
		},
	}
	for _, o := range options {
		b, err := NewHTTPBackend("b1", o)
		c.Assert(err, NotNil)
		c.Assert(b, IsNil)
	}
}

func (s *BackendSuite) TestNewServer(c *C) {
	sv, err := NewServer("s1", "http://falhost")
	c.Assert(err, IsNil)
	c.Assert(sv.GetId(), Equals, "s1")
	c.Assert(sv.String(), Not(Equals), "")
}

func (s *BackendSuite) TestNewServerBadParams(c *C) {
	_, err := NewServer("s1", "http---")
	c.Assert(err, NotNil)
}

func (s *BackendSuite) TestNewListener(c *C) {
	_, err := NewListener("id", "http", "tcp", "127.0.0.1:4000", "", nil)
	c.Assert(err, IsNil)
}

func (s *BackendSuite) TestNewListenerBadParams(c *C) {
	_, err := NewListener("id", "http", "tcp", "", "", nil)
	c.Assert(err, NotNil)

	_, err = NewListener("id", "", "tcp", "127.0.0.1:4000", "", nil)
	c.Assert(err, NotNil)

	_, err = NewListener("id", "http", "tcp", "127.0.0.1:4000", "blabla", nil)
	c.Assert(err, NotNil)
}

func (s *BackendSuite) TestFrontendsFromJSON(c *C) {
	f, err := NewHTTPFrontend(route.NewMux(), "f1", "b1", `Path("/path")`, HTTPFrontendSettings{})
	c.Assert(err, IsNil)

	bytes, err := json.Marshal(f)

	fs := []Frontend{*f}

	bytes, err = json.Marshal(map[string]interface{}{"Frontends": fs})

	r := plugin.NewRegistry()
	c.Assert(r.AddSpec(connlimit.GetSpec()), IsNil)

	out, err := FrontendsFromJSON(route.NewMux(), bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)
	c.Assert(out, DeepEquals, fs)
}

func (s *BackendSuite) MiddlewareFromJSON(c *C) {
	cl, err := connlimit.NewConnLimit(10, "client.ip")
	c.Assert(err, IsNil)

	m := &Middleware{Id: "c1", Type: "connlimit", Middleware: cl}

	bytes, err := json.Marshal(m)
	c.Assert(err, IsNil)

	out, err := MiddlewareFromJSON(bytes, plugin.NewRegistry().GetSpec)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)
	c.Assert(out, DeepEquals, m)
}

func (s *BackendSuite) TestBackendFromJSON(c *C) {
	b, err := NewHTTPBackend("b1", HTTPBackendSettings{})
	c.Assert(err, IsNil)

	bytes, err := json.Marshal(b)
	c.Assert(err, IsNil)

	out, err := BackendFromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	c.Assert(out, DeepEquals, b)
}

func (s *BackendSuite) TestServerFromJSON(c *C) {
	e, err := NewServer("sv1", "http://localhost")
	c.Assert(err, IsNil)

	bytes, err := json.Marshal(e)
	c.Assert(err, IsNil)

	out, err := ServerFromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	c.Assert(out, DeepEquals, e)
}

func (s *BackendSuite) TestNewTLSSettings(c *C) {
	tcs := []struct {
		S TLSSettings
		C *tls.Config
	}{
		// Make sure defaults are set as expected
		{
			S: TLSSettings{},
			C: &tls.Config{
				MinVersion: tls.VersionTLS10,
				MaxVersion: tls.VersionTLS12,

				SessionTicketsDisabled:   false,
				PreferServerCipherSuites: false,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,

					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,

					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				},

				InsecureSkipVerify: false,
			},
		},
		{
			S: TLSSettings{
				MinVersion: "VersionTLS11",
				MaxVersion: "VersionTLS12",

				SessionTicketsDisabled:   false,
				PreferServerCipherSuites: true,

				CipherSuites: []string{
					"TLS_RSA_WITH_RC4_128_SHA",
					"TLS_RSA_WITH_3DES_EDE_CBC_SHA",

					"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
					"TLS_ECDHE_RSA_WITH_RC4_128_SHA",
					"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",

					"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
					"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",

					"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
					"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",

					"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
					"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",

					"TLS_RSA_WITH_AES_256_CBC_SHA",
					"TLS_RSA_WITH_AES_128_CBC_SHA",
				},

				InsecureSkipVerify: false,

				SessionCache: TLSSessionCache{
					Type: "LRU",
					Settings: &LRUSessionCacheSettings{
						Capacity: 12,
					},
				},
			},
			C: &tls.Config{
				MinVersion: tls.VersionTLS11,
				MaxVersion: tls.VersionTLS12,

				SessionTicketsDisabled:   false,
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					tls.TLS_RSA_WITH_RC4_128_SHA,
					tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,

					tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
					tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
					tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,

					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,

					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,

					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				},

				InsecureSkipVerify: false,

				ClientSessionCache: tls.NewLRUClientSessionCache(12),
			},
		},
		{
			S: TLSSettings{
				MinVersion: "VersionTLS10",
				MaxVersion: "VersionTLS12",

				SessionTicketsDisabled:   true,
				PreferServerCipherSuites: false,

				CipherSuites: []string{
					"TLS_RSA_WITH_RC4_128_SHA",
				},

				InsecureSkipVerify: true,
			},
			C: &tls.Config{
				MinVersion: tls.VersionTLS10,
				MaxVersion: tls.VersionTLS12,

				SessionTicketsDisabled:   true,
				PreferServerCipherSuites: false,
				CipherSuites: []uint16{
					tls.TLS_RSA_WITH_RC4_128_SHA,
				},

				InsecureSkipVerify: true,
			},
		},
	}
	for _, tc := range tcs {
		cfg, err := NewTLSConfig(&tc.S)
		c.Assert(err, IsNil)

		c.Assert(cfg.MinVersion, Equals, tc.C.MinVersion)
		c.Assert(cfg.MaxVersion, Equals, tc.C.MaxVersion)

		c.Assert(cfg.SessionTicketsDisabled, Equals, tc.C.SessionTicketsDisabled)
		c.Assert(cfg.PreferServerCipherSuites, Equals, tc.C.PreferServerCipherSuites)
		c.Assert(cfg.CipherSuites, DeepEquals, tc.C.CipherSuites)
		c.Assert(cfg.InsecureSkipVerify, Equals, tc.C.InsecureSkipVerify)
		if tc.C.ClientSessionCache != nil {
			c.Assert(cfg.ClientSessionCache, FitsTypeOf, tc.C.ClientSessionCache)
		}
	}
}

func (s *BackendSuite) TestNewTLSSettingsBadParams(c *C) {
	tcs := []TLSSettings{
		TLSSettings{
			MinVersion: "blabla",
		},
		TLSSettings{
			MaxVersion: "blabla",
		},
		TLSSettings{
			CipherSuites: []string{"blabla"},
		},
		TLSSettings{
			SessionCache: TLSSessionCache{
				Type: "what?",
				Settings: &LRUSessionCacheSettings{
					Capacity: 12,
				},
			},
		},
		TLSSettings{
			SessionCache: TLSSessionCache{
				Type: "LRU",
				Settings: &LRUSessionCacheSettings{
					Capacity: -5,
				},
			},
		},
	}
	for _, tc := range tcs {
		cfg, err := NewTLSConfig(&tc)
		c.Assert(err, NotNil)
		c.Assert(cfg, IsNil)
	}
}

func (s *BackendSuite) TestTLSSettingsEq(c *C) {
	tcs := []struct {
		A  TLSSettings
		B  TLSSettings
		R  bool
		TC string
	}{
		{
			A:  TLSSettings{},
			B:  TLSSettings{},
			R:  true,
			TC: "defaults",
		},
		{
			A: TLSSettings{
				SessionCache: TLSSessionCache{
					Type: "LRU",
				},
			},
			B: TLSSettings{
				SessionCache: TLSSessionCache{
					Type: "LRU",
				},
			},
			R:  true,
			TC: "cache defaults",
		},
		{
			A: TLSSettings{
				SessionCache: TLSSessionCache{
					Type: "LRU",
					Settings: &LRUSessionCacheSettings{
						Capacity: 5,
					},
				},
			},
			B: TLSSettings{
				SessionCache: TLSSessionCache{
					Type: "LRU",
					Settings: &LRUSessionCacheSettings{
						Capacity: 5,
					},
				},
			},
			R:  true,
			TC: "cache params",
		},
		{
			A: TLSSettings{
				SessionCache: TLSSessionCache{
					Type: "LRU",
					Settings: &LRUSessionCacheSettings{
						Capacity: 5,
					},
				},
			},
			B: TLSSettings{
				SessionCache: TLSSessionCache{
					Type: "LRU",
					Settings: &LRUSessionCacheSettings{
						Capacity: 4,
					},
				},
			},
			R:  false,
			TC: "cache params neq",
		},
		{
			A: TLSSettings{
				SessionTicketsDisabled: true,
			},
			B:  TLSSettings{},
			R:  false,
			TC: "no session tickets",
		},
		{
			A: TLSSettings{
				InsecureSkipVerify: true,
			},
			B:  TLSSettings{},
			R:  false,
			TC: "insecure",
		},
		{
			A: TLSSettings{
				MinVersion: "VersionTLS12",
			},
			B:  TLSSettings{},
			R:  false,
			TC: "different min",
		},
		{
			A: TLSSettings{
				MaxVersion: "VersionTLS10",
			},
			B:  TLSSettings{},
			R:  false,
			TC: "different max",
		},
		{
			A: TLSSettings{
				PreferServerCipherSuites: true,
			},
			B:  TLSSettings{},
			R:  false,
			TC: "prefer server csuites",
		},
		{
			A: TLSSettings{
				CipherSuites: []string{
					"TLS_RSA_WITH_RC4_128_SHA",
				},
			},
			B: TLSSettings{
				CipherSuites: []string{
					"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
				},
			},
			R:  false,
			TC: "different csuites",
		},
		{
			A: TLSSettings{
				CipherSuites: []string{
					"TLS_RSA_WITH_RC4_128_SHA",
				},
			},
			B: TLSSettings{
				CipherSuites: []string{},
			},
			R:  false,
			TC: "different csuites 1",
		},
	}
	for _, tc := range tcs {
		c.Assert(tc.A.Equals(&tc.B), Equals, tc.R, Commentf("TC: %v", tc.TC))
	}
}
