package label

import (
	"fmt"
	"testing"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	labels := map[string]string{
		"traefik.middlewares.Middleware0.addprefix.prefix":                                "foobar",
		"traefik.middlewares.Middleware1.basicauth.headerfield":                           "foobar",
		"traefik.middlewares.Middleware1.basicauth.realm":                                 "foobar",
		"traefik.middlewares.Middleware1.basicauth.removeheader":                          "true",
		"traefik.middlewares.Middleware1.basicauth.users":                                 "foobar, fiibar",
		"traefik.middlewares.Middleware1.basicauth.usersfile":                             "foobar",
		"traefik.middlewares.Middleware2.buffering.maxrequestbodybytes":                   "42",
		"traefik.middlewares.Middleware2.buffering.maxresponsebodybytes":                  "42",
		"traefik.middlewares.Middleware2.buffering.memrequestbodybytes":                   "42",
		"traefik.middlewares.Middleware2.buffering.memresponsebodybytes":                  "42",
		"traefik.middlewares.Middleware2.buffering.retryexpression":                       "foobar",
		"traefik.middlewares.Middleware3.chain.middlewares":                               "foobar, fiibar",
		"traefik.middlewares.Middleware4.circuitbreaker.expression":                       "foobar",
		"traefik.middlewares.Middleware5.digestauth.headerfield":                          "foobar",
		"traefik.middlewares.Middleware5.digestauth.realm":                                "foobar",
		"traefik.middlewares.Middleware5.digestauth.removeheader":                         "true",
		"traefik.middlewares.Middleware5.digestauth.users":                                "foobar, fiibar",
		"traefik.middlewares.Middleware5.digestauth.usersfile":                            "foobar",
		"traefik.middlewares.Middleware6.errors.query":                                    "foobar",
		"traefik.middlewares.Middleware6.errors.service":                                  "foobar",
		"traefik.middlewares.Middleware6.errors.status":                                   "foobar, fiibar",
		"traefik.middlewares.Middleware7.forwardauth.address":                             "foobar",
		"traefik.middlewares.Middleware7.forwardauth.authresponseheaders":                 "foobar, fiibar",
		"traefik.middlewares.Middleware7.forwardauth.tls.ca":                              "foobar",
		"traefik.middlewares.Middleware7.forwardauth.tls.caoptional":                      "true",
		"traefik.middlewares.Middleware7.forwardauth.tls.cert":                            "foobar",
		"traefik.middlewares.Middleware7.forwardauth.tls.insecureskipverify":              "true",
		"traefik.middlewares.Middleware7.forwardauth.tls.key":                             "foobar",
		"traefik.middlewares.Middleware7.forwardauth.trustforwardheader":                  "true",
		"traefik.middlewares.Middleware8.headers.allowedhosts":                            "foobar, fiibar",
		"traefik.middlewares.Middleware8.headers.browserxssfilter":                        "true",
		"traefik.middlewares.Middleware8.headers.contentsecuritypolicy":                   "foobar",
		"traefik.middlewares.Middleware8.headers.contenttypenosniff":                      "true",
		"traefik.middlewares.Middleware8.headers.custombrowserxssvalue":                   "foobar",
		"traefik.middlewares.Middleware8.headers.customframeoptionsvalue":                 "foobar",
		"traefik.middlewares.Middleware8.headers.customrequestheaders.name0":              "foobar",
		"traefik.middlewares.Middleware8.headers.customrequestheaders.name1":              "foobar",
		"traefik.middlewares.Middleware8.headers.customresponseheaders.name0":             "foobar",
		"traefik.middlewares.Middleware8.headers.customresponseheaders.name1":             "foobar",
		"traefik.middlewares.Middleware8.headers.forcestsheader":                          "true",
		"traefik.middlewares.Middleware8.headers.framedeny":                               "true",
		"traefik.middlewares.Middleware8.headers.hostsproxyheaders":                       "foobar, fiibar",
		"traefik.middlewares.Middleware8.headers.isdevelopment":                           "true",
		"traefik.middlewares.Middleware8.headers.publickey":                               "foobar",
		"traefik.middlewares.Middleware8.headers.referrerpolicy":                          "foobar",
		"traefik.middlewares.Middleware8.headers.sslforcehost":                            "true",
		"traefik.middlewares.Middleware8.headers.sslhost":                                 "foobar",
		"traefik.middlewares.Middleware8.headers.sslproxyheaders.name0":                   "foobar",
		"traefik.middlewares.Middleware8.headers.sslproxyheaders.name1":                   "foobar",
		"traefik.middlewares.Middleware8.headers.sslredirect":                             "true",
		"traefik.middlewares.Middleware8.headers.ssltemporaryredirect":                    "true",
		"traefik.middlewares.Middleware8.headers.stsincludesubdomains":                    "true",
		"traefik.middlewares.Middleware8.headers.stspreload":                              "true",
		"traefik.middlewares.Middleware8.headers.stsseconds":                              "42",
		"traefik.middlewares.Middleware9.ipwhitelist.ipstrategy.depth":                    "42",
		"traefik.middlewares.Middleware9.ipwhitelist.ipstrategy.excludedips":              "foobar, fiibar",
		"traefik.middlewares.Middleware9.ipwhitelist.sourcerange":                         "foobar, fiibar",
		"traefik.middlewares.Middleware10.maxconn.amount":                                 "42",
		"traefik.middlewares.Middleware10.maxconn.extractorfunc":                          "foobar",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.notafter":                "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.notbefore":               "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.sans":                    "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.subject.commonname":      "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.subject.country":         "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.subject.domaincomponent": "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.subject.locality":        "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.subject.organization":    "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.subject.province":        "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.info.subject.serialnumber":    "true",
		"traefik.middlewares.Middleware11.passtlsclientcert.pem":                          "true",
		"traefik.middlewares.Middleware12.ratelimit.extractorfunc":                        "foobar",
		"traefik.middlewares.Middleware12.ratelimit.rateset.Rate0.average":                "42",
		"traefik.middlewares.Middleware12.ratelimit.rateset.Rate0.burst":                  "42",
		"traefik.middlewares.Middleware12.ratelimit.rateset.Rate0.period":                 "42",
		"traefik.middlewares.Middleware12.ratelimit.rateset.Rate1.average":                "42",
		"traefik.middlewares.Middleware12.ratelimit.rateset.Rate1.burst":                  "42",
		"traefik.middlewares.Middleware12.ratelimit.rateset.Rate1.period":                 "42",
		"traefik.middlewares.Middleware13.redirect.permanent":                             "true",
		"traefik.middlewares.Middleware13.redirect.regex":                                 "foobar",
		"traefik.middlewares.Middleware13.redirect.replacement":                           "foobar",
		"traefik.middlewares.Middleware14.replacepath.path":                               "foobar",
		"traefik.middlewares.Middleware15.replacepathregex.regex":                         "foobar",
		"traefik.middlewares.Middleware15.replacepathregex.replacement":                   "foobar",
		"traefik.middlewares.Middleware16.retry.attempts":                                 "42",
		"traefik.middlewares.Middleware17.stripprefix.prefixes":                           "foobar, fiibar",
		"traefik.middlewares.Middleware18.stripprefixregex.regex":                         "foobar, fiibar",
		"traefik.middlewares.Middleware19.compress":                                       "true",

		"traefik.routers.Router0.entrypoints": "foobar, fiibar",
		"traefik.routers.Router0.middlewares": "foobar, fiibar",
		"traefik.routers.Router0.priority":    "42",
		"traefik.routers.Router0.rule":        "foobar",
		"traefik.routers.Router0.service":     "foobar",
		"traefik.routers.Router1.entrypoints": "foobar, fiibar",
		"traefik.routers.Router1.middlewares": "foobar, fiibar",
		"traefik.routers.Router1.priority":    "42",
		"traefik.routers.Router1.rule":        "foobar",
		"traefik.routers.Router1.service":     "foobar",

		"traefik.services.Service0.loadbalancer.healthcheck.headers.name0":        "foobar",
		"traefik.services.Service0.loadbalancer.healthcheck.headers.name1":        "foobar",
		"traefik.services.Service0.loadbalancer.healthcheck.hostname":             "foobar",
		"traefik.services.Service0.loadbalancer.healthcheck.interval":             "foobar",
		"traefik.services.Service0.loadbalancer.healthcheck.path":                 "foobar",
		"traefik.services.Service0.loadbalancer.healthcheck.port":                 "42",
		"traefik.services.Service0.loadbalancer.healthcheck.scheme":               "foobar",
		"traefik.services.Service0.loadbalancer.healthcheck.timeout":              "foobar",
		"traefik.services.Service0.loadbalancer.method":                           "foobar",
		"traefik.services.Service0.loadbalancer.passhostheader":                   "true",
		"traefik.services.Service0.loadbalancer.responseforwarding.flushinterval": "foobar",
		"traefik.services.Service0.loadbalancer.server.url":                       "foobar",
		"traefik.services.Service0.loadbalancer.server.weight":                    "42",
		"traefik.services.Service0.loadbalancer.stickiness.cookiename":            "foobar",
		"traefik.services.Service1.loadbalancer.healthcheck.headers.name0":        "foobar",
		"traefik.services.Service1.loadbalancer.healthcheck.headers.name1":        "foobar",
		"traefik.services.Service1.loadbalancer.healthcheck.hostname":             "foobar",
		"traefik.services.Service1.loadbalancer.healthcheck.interval":             "foobar",
		"traefik.services.Service1.loadbalancer.healthcheck.path":                 "foobar",
		"traefik.services.Service1.loadbalancer.healthcheck.port":                 "42",
		"traefik.services.Service1.loadbalancer.healthcheck.scheme":               "foobar",
		"traefik.services.Service1.loadbalancer.healthcheck.timeout":              "foobar",
		"traefik.services.Service1.loadbalancer.method":                           "foobar",
		"traefik.services.Service1.loadbalancer.passhostheader":                   "true",
		"traefik.services.Service1.loadbalancer.responseforwarding.flushinterval": "foobar",
		"traefik.services.Service1.loadbalancer.server.url":                       "foobar",
		"traefik.services.Service1.loadbalancer.server.weight":                    "42",
		"traefik.services.Service1.loadbalancer.stickiness":                       "false",
		"traefik.services.Service1.loadbalancer.stickiness.cookiename":            "fui",
	}

	configuration, err := Decode(labels)
	require.NoError(t, err)

	expected := &config.Configuration{
		Routers: map[string]*config.Router{
			"Router0": {
				EntryPoints: []string{
					"foobar",
					"fiibar",
				},
				Middlewares: []string{
					"foobar",
					"fiibar",
				},
				Service:  "foobar",
				Rule:     "foobar",
				Priority: 42,
			},
			"Router1": {
				EntryPoints: []string{
					"foobar",
					"fiibar",
				},
				Middlewares: []string{
					"foobar",
					"fiibar",
				},
				Service:  "foobar",
				Rule:     "foobar",
				Priority: 42,
			},
		},
		Middlewares: map[string]*config.Middleware{
			"Middleware0": {
				AddPrefix: &config.AddPrefix{
					Prefix: "foobar",
				},
			},
			"Middleware1": {
				BasicAuth: &config.BasicAuth{
					Users: []string{
						"foobar",
						"fiibar",
					},
					UsersFile:    "foobar",
					Realm:        "foobar",
					RemoveHeader: true,
					HeaderField:  "foobar",
				},
			},
			"Middleware10": {
				MaxConn: &config.MaxConn{
					Amount:        42,
					ExtractorFunc: "foobar",
				},
			},
			"Middleware11": {
				PassTLSClientCert: &config.PassTLSClientCert{
					PEM: true,
					Info: &config.TLSClientCertificateInfo{
						NotAfter:  true,
						NotBefore: true,
						Subject: &config.TLSCLientCertificateDNInfo{
							Country:         true,
							Province:        true,
							Locality:        true,
							Organization:    true,
							CommonName:      true,
							SerialNumber:    true,
							DomainComponent: true,
						},
						Sans: true,
					},
				},
			},
			"Middleware12": {
				RateLimit: &config.RateLimit{
					RateSet: map[string]*config.Rate{
						"Rate0": {
							Period:  parse.Duration(42 * time.Nanosecond),
							Average: 42,
							Burst:   42,
						},
						"Rate1": {
							Period:  parse.Duration(42 * time.Nanosecond),
							Average: 42,
							Burst:   42,
						},
					},
					ExtractorFunc: "foobar",
				},
			},
			"Middleware13": {
				Redirect: &config.Redirect{
					Regex:       "foobar",
					Replacement: "foobar",
					Permanent:   true,
				},
			},
			"Middleware14": {
				ReplacePath: &config.ReplacePath{
					Path: "foobar",
				},
			},
			"Middleware15": {
				ReplacePathRegex: &config.ReplacePathRegex{
					Regex:       "foobar",
					Replacement: "foobar",
				},
			},
			"Middleware16": {
				Retry: &config.Retry{
					Attempts: 42,
				},
			},
			"Middleware17": {
				StripPrefix: &config.StripPrefix{
					Prefixes: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware18": {
				StripPrefixRegex: &config.StripPrefixRegex{
					Regex: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware19": {
				Compress: &config.Compress{},
			},
			"Middleware2": {
				Buffering: &config.Buffering{
					MaxRequestBodyBytes:  42,
					MemRequestBodyBytes:  42,
					MaxResponseBodyBytes: 42,
					MemResponseBodyBytes: 42,
					RetryExpression:      "foobar",
				},
			},
			"Middleware3": {
				Chain: &config.Chain{
					Middlewares: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware4": {
				CircuitBreaker: &config.CircuitBreaker{
					Expression: "foobar",
				},
			},
			"Middleware5": {
				DigestAuth: &config.DigestAuth{
					Users: []string{
						"foobar",
						"fiibar",
					},
					UsersFile:    "foobar",
					RemoveHeader: true,
					Realm:        "foobar",
					HeaderField:  "foobar",
				},
			},
			"Middleware6": {
				Errors: &config.ErrorPage{
					Status: []string{
						"foobar",
						"fiibar",
					},
					Service: "foobar",
					Query:   "foobar",
				},
			},
			"Middleware7": {
				ForwardAuth: &config.ForwardAuth{
					Address: "foobar",
					TLS: &config.ClientTLS{
						CA:                 "foobar",
						CAOptional:         true,
						Cert:               "foobar",
						Key:                "foobar",
						InsecureSkipVerify: true,
					},
					TrustForwardHeader: true,
					AuthResponseHeaders: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware8": {
				Headers: &config.Headers{
					CustomRequestHeaders: map[string]string{
						"name0": "foobar",
						"name1": "foobar",
					},
					CustomResponseHeaders: map[string]string{
						"name0": "foobar",
						"name1": "foobar",
					},
					AllowedHosts: []string{
						"foobar",
						"fiibar",
					},
					HostsProxyHeaders: []string{
						"foobar",
						"fiibar",
					},
					SSLRedirect:          true,
					SSLTemporaryRedirect: true,
					SSLHost:              "foobar",
					SSLProxyHeaders: map[string]string{
						"name0": "foobar",
						"name1": "foobar",
					},
					SSLForceHost:            true,
					STSSeconds:              42,
					STSIncludeSubdomains:    true,
					STSPreload:              true,
					ForceSTSHeader:          true,
					FrameDeny:               true,
					CustomFrameOptionsValue: "foobar",
					ContentTypeNosniff:      true,
					BrowserXSSFilter:        true,
					CustomBrowserXSSValue:   "foobar",
					ContentSecurityPolicy:   "foobar",
					PublicKey:               "foobar",
					ReferrerPolicy:          "foobar",
					IsDevelopment:           true,
				},
			},
			"Middleware9": {
				IPWhiteList: &config.IPWhiteList{
					SourceRange: []string{
						"foobar",
						"fiibar",
					},
					IPStrategy: &config.IPStrategy{
						Depth: 42,
						ExcludedIPs: []string{
							"foobar",
							"fiibar",
						},
					},
				},
			},
		},
		Services: map[string]*config.Service{
			"Service0": {
				LoadBalancer: &config.LoadBalancerService{
					Stickiness: &config.Stickiness{
						CookieName: "foobar",
					},
					Servers: []config.Server{
						{
							URL:    "foobar",
							Weight: 42,
						},
					},
					Method: "foobar",
					HealthCheck: &config.HealthCheck{
						Scheme:   "foobar",
						Path:     "foobar",
						Port:     42,
						Interval: "foobar",
						Timeout:  "foobar",
						Hostname: "foobar",
						Headers: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
					},
					PassHostHeader: true,
					ResponseForwarding: &config.ResponseForwarding{
						FlushInterval: "foobar",
					},
				},
			},
			"Service1": {
				LoadBalancer: &config.LoadBalancerService{
					Servers: []config.Server{
						{
							URL:    "foobar",
							Weight: 42,
						},
					},
					Method: "foobar",
					HealthCheck: &config.HealthCheck{
						Scheme:   "foobar",
						Path:     "foobar",
						Port:     42,
						Interval: "foobar",
						Timeout:  "foobar",
						Hostname: "foobar",
						Headers: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
					},
					PassHostHeader: true,
					ResponseForwarding: &config.ResponseForwarding{
						FlushInterval: "foobar",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, configuration)
}

func TestEncode(t *testing.T) {
	configuration := &config.Configuration{
		Routers: map[string]*config.Router{
			"Router0": {
				EntryPoints: []string{
					"foobar",
					"fiibar",
				},
				Middlewares: []string{
					"foobar",
					"fiibar",
				},
				Service:  "foobar",
				Rule:     "foobar",
				Priority: 42,
			},
			"Router1": {
				EntryPoints: []string{
					"foobar",
					"fiibar",
				},
				Middlewares: []string{
					"foobar",
					"fiibar",
				},
				Service:  "foobar",
				Rule:     "foobar",
				Priority: 42,
			},
		},
		Middlewares: map[string]*config.Middleware{
			"Middleware0": {
				AddPrefix: &config.AddPrefix{
					Prefix: "foobar",
				},
			},
			"Middleware1": {
				BasicAuth: &config.BasicAuth{
					Users: []string{
						"foobar",
						"fiibar",
					},
					UsersFile:    "foobar",
					Realm:        "foobar",
					RemoveHeader: true,
					HeaderField:  "foobar",
				},
			},
			"Middleware10": {
				MaxConn: &config.MaxConn{
					Amount:        42,
					ExtractorFunc: "foobar",
				},
			},
			"Middleware11": {
				PassTLSClientCert: &config.PassTLSClientCert{
					PEM: true,
					Info: &config.TLSClientCertificateInfo{
						NotAfter:  true,
						NotBefore: true,
						Subject: &config.TLSCLientCertificateDNInfo{
							Country:         true,
							Province:        true,
							Locality:        true,
							Organization:    true,
							CommonName:      true,
							SerialNumber:    true,
							DomainComponent: true,
						},
						Sans: true,
					},
				},
			},
			"Middleware12": {
				RateLimit: &config.RateLimit{
					RateSet: map[string]*config.Rate{
						"Rate0": {
							Period:  parse.Duration(42 * time.Nanosecond),
							Average: 42,
							Burst:   42,
						},
						"Rate1": {
							Period:  parse.Duration(42 * time.Nanosecond),
							Average: 42,
							Burst:   42,
						},
					},
					ExtractorFunc: "foobar",
				},
			},
			"Middleware13": {
				Redirect: &config.Redirect{
					Regex:       "foobar",
					Replacement: "foobar",
					Permanent:   true,
				},
			},
			"Middleware14": {
				ReplacePath: &config.ReplacePath{
					Path: "foobar",
				},
			},
			"Middleware15": {
				ReplacePathRegex: &config.ReplacePathRegex{
					Regex:       "foobar",
					Replacement: "foobar",
				},
			},
			"Middleware16": {
				Retry: &config.Retry{
					Attempts: 42,
				},
			},
			"Middleware17": {
				StripPrefix: &config.StripPrefix{
					Prefixes: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware18": {
				StripPrefixRegex: &config.StripPrefixRegex{
					Regex: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware19": {
				Compress: &config.Compress{},
			},
			"Middleware2": {
				Buffering: &config.Buffering{
					MaxRequestBodyBytes:  42,
					MemRequestBodyBytes:  42,
					MaxResponseBodyBytes: 42,
					MemResponseBodyBytes: 42,
					RetryExpression:      "foobar",
				},
			},
			"Middleware3": {
				Chain: &config.Chain{
					Middlewares: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware4": {
				CircuitBreaker: &config.CircuitBreaker{
					Expression: "foobar",
				},
			},
			"Middleware5": {
				DigestAuth: &config.DigestAuth{
					Users: []string{
						"foobar",
						"fiibar",
					},
					UsersFile:    "foobar",
					RemoveHeader: true,
					Realm:        "foobar",
					HeaderField:  "foobar",
				},
			},
			"Middleware6": {
				Errors: &config.ErrorPage{
					Status: []string{
						"foobar",
						"fiibar",
					},
					Service: "foobar",
					Query:   "foobar",
				},
			},
			"Middleware7": {
				ForwardAuth: &config.ForwardAuth{
					Address: "foobar",
					TLS: &config.ClientTLS{
						CA:                 "foobar",
						CAOptional:         true,
						Cert:               "foobar",
						Key:                "foobar",
						InsecureSkipVerify: true,
					},
					TrustForwardHeader: true,
					AuthResponseHeaders: []string{
						"foobar",
						"fiibar",
					},
				},
			},
			"Middleware8": {
				Headers: &config.Headers{
					CustomRequestHeaders: map[string]string{
						"name0": "foobar",
						"name1": "foobar",
					},
					CustomResponseHeaders: map[string]string{
						"name0": "foobar",
						"name1": "foobar",
					},
					AllowedHosts: []string{
						"foobar",
						"fiibar",
					},
					HostsProxyHeaders: []string{
						"foobar",
						"fiibar",
					},
					SSLRedirect:          true,
					SSLTemporaryRedirect: true,
					SSLHost:              "foobar",
					SSLProxyHeaders: map[string]string{
						"name0": "foobar",
						"name1": "foobar",
					},
					SSLForceHost:            true,
					STSSeconds:              42,
					STSIncludeSubdomains:    true,
					STSPreload:              true,
					ForceSTSHeader:          true,
					FrameDeny:               true,
					CustomFrameOptionsValue: "foobar",
					ContentTypeNosniff:      true,
					BrowserXSSFilter:        true,
					CustomBrowserXSSValue:   "foobar",
					ContentSecurityPolicy:   "foobar",
					PublicKey:               "foobar",
					ReferrerPolicy:          "foobar",
					IsDevelopment:           true,
				},
			},
			"Middleware9": {
				IPWhiteList: &config.IPWhiteList{
					SourceRange: []string{
						"foobar",
						"fiibar",
					},
					IPStrategy: &config.IPStrategy{
						Depth: 42,
						ExcludedIPs: []string{
							"foobar",
							"fiibar",
						},
					},
				},
			},
		},
		Services: map[string]*config.Service{
			"Service0": {
				LoadBalancer: &config.LoadBalancerService{
					Stickiness: &config.Stickiness{
						CookieName: "foobar",
					},
					Servers: []config.Server{
						{
							URL:    "foobar",
							Weight: 42,
						},
					},
					Method: "foobar",
					HealthCheck: &config.HealthCheck{
						Scheme:   "foobar",
						Path:     "foobar",
						Port:     42,
						Interval: "foobar",
						Timeout:  "foobar",
						Hostname: "foobar",
						Headers: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
					},
					PassHostHeader: true,
					ResponseForwarding: &config.ResponseForwarding{
						FlushInterval: "foobar",
					},
				},
			},
			"Service1": {
				LoadBalancer: &config.LoadBalancerService{
					Servers: []config.Server{
						{
							URL:    "foobar",
							Weight: 42,
						},
					},
					Method: "foobar",
					HealthCheck: &config.HealthCheck{
						Scheme:   "foobar",
						Path:     "foobar",
						Port:     42,
						Interval: "foobar",
						Timeout:  "foobar",
						Hostname: "foobar",
						Headers: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
					},
					PassHostHeader: true,
					ResponseForwarding: &config.ResponseForwarding{
						FlushInterval: "foobar",
					},
				},
			},
		},
	}

	labels, err := Encode(configuration)
	require.NoError(t, err)

	expected := map[string]string{
		"traefik.Middlewares.Middleware0.AddPrefix.Prefix":                                "foobar",
		"traefik.Middlewares.Middleware1.BasicAuth.HeaderField":                           "foobar",
		"traefik.Middlewares.Middleware1.BasicAuth.Realm":                                 "foobar",
		"traefik.Middlewares.Middleware1.BasicAuth.RemoveHeader":                          "true",
		"traefik.Middlewares.Middleware1.BasicAuth.Users":                                 "foobar, fiibar",
		"traefik.Middlewares.Middleware1.BasicAuth.UsersFile":                             "foobar",
		"traefik.Middlewares.Middleware2.Buffering.MaxRequestBodyBytes":                   "42",
		"traefik.Middlewares.Middleware2.Buffering.MaxResponseBodyBytes":                  "42",
		"traefik.Middlewares.Middleware2.Buffering.MemRequestBodyBytes":                   "42",
		"traefik.Middlewares.Middleware2.Buffering.MemResponseBodyBytes":                  "42",
		"traefik.Middlewares.Middleware2.Buffering.RetryExpression":                       "foobar",
		"traefik.Middlewares.Middleware3.Chain.Middlewares":                               "foobar, fiibar",
		"traefik.Middlewares.Middleware4.CircuitBreaker.Expression":                       "foobar",
		"traefik.Middlewares.Middleware5.DigestAuth.HeaderField":                          "foobar",
		"traefik.Middlewares.Middleware5.DigestAuth.Realm":                                "foobar",
		"traefik.Middlewares.Middleware5.DigestAuth.RemoveHeader":                         "true",
		"traefik.Middlewares.Middleware5.DigestAuth.Users":                                "foobar, fiibar",
		"traefik.Middlewares.Middleware5.DigestAuth.UsersFile":                            "foobar",
		"traefik.Middlewares.Middleware6.Errors.Query":                                    "foobar",
		"traefik.Middlewares.Middleware6.Errors.Service":                                  "foobar",
		"traefik.Middlewares.Middleware6.Errors.Status":                                   "foobar, fiibar",
		"traefik.Middlewares.Middleware7.ForwardAuth.Address":                             "foobar",
		"traefik.Middlewares.Middleware7.ForwardAuth.AuthResponseHeaders":                 "foobar, fiibar",
		"traefik.Middlewares.Middleware7.ForwardAuth.TLS.CA":                              "foobar",
		"traefik.Middlewares.Middleware7.ForwardAuth.TLS.CAOptional":                      "true",
		"traefik.Middlewares.Middleware7.ForwardAuth.TLS.Cert":                            "foobar",
		"traefik.Middlewares.Middleware7.ForwardAuth.TLS.InsecureSkipVerify":              "true",
		"traefik.Middlewares.Middleware7.ForwardAuth.TLS.Key":                             "foobar",
		"traefik.Middlewares.Middleware7.ForwardAuth.TrustForwardHeader":                  "true",
		"traefik.Middlewares.Middleware8.Headers.AllowedHosts":                            "foobar, fiibar",
		"traefik.Middlewares.Middleware8.Headers.BrowserXSSFilter":                        "true",
		"traefik.Middlewares.Middleware8.Headers.ContentSecurityPolicy":                   "foobar",
		"traefik.Middlewares.Middleware8.Headers.ContentTypeNosniff":                      "true",
		"traefik.Middlewares.Middleware8.Headers.CustomBrowserXSSValue":                   "foobar",
		"traefik.Middlewares.Middleware8.Headers.CustomFrameOptionsValue":                 "foobar",
		"traefik.Middlewares.Middleware8.Headers.CustomRequestHeaders.name0":              "foobar",
		"traefik.Middlewares.Middleware8.Headers.CustomRequestHeaders.name1":              "foobar",
		"traefik.Middlewares.Middleware8.Headers.CustomResponseHeaders.name0":             "foobar",
		"traefik.Middlewares.Middleware8.Headers.CustomResponseHeaders.name1":             "foobar",
		"traefik.Middlewares.Middleware8.Headers.ForceSTSHeader":                          "true",
		"traefik.Middlewares.Middleware8.Headers.FrameDeny":                               "true",
		"traefik.Middlewares.Middleware8.Headers.HostsProxyHeaders":                       "foobar, fiibar",
		"traefik.Middlewares.Middleware8.Headers.IsDevelopment":                           "true",
		"traefik.Middlewares.Middleware8.Headers.PublicKey":                               "foobar",
		"traefik.Middlewares.Middleware8.Headers.ReferrerPolicy":                          "foobar",
		"traefik.Middlewares.Middleware8.Headers.SSLForceHost":                            "true",
		"traefik.Middlewares.Middleware8.Headers.SSLHost":                                 "foobar",
		"traefik.Middlewares.Middleware8.Headers.SSLProxyHeaders.name0":                   "foobar",
		"traefik.Middlewares.Middleware8.Headers.SSLProxyHeaders.name1":                   "foobar",
		"traefik.Middlewares.Middleware8.Headers.SSLRedirect":                             "true",
		"traefik.Middlewares.Middleware8.Headers.SSLTemporaryRedirect":                    "true",
		"traefik.Middlewares.Middleware8.Headers.STSIncludeSubdomains":                    "true",
		"traefik.Middlewares.Middleware8.Headers.STSPreload":                              "true",
		"traefik.Middlewares.Middleware8.Headers.STSSeconds":                              "42",
		"traefik.Middlewares.Middleware9.IPWhiteList.IPStrategy.Depth":                    "42",
		"traefik.Middlewares.Middleware9.IPWhiteList.IPStrategy.ExcludedIPs":              "foobar, fiibar",
		"traefik.Middlewares.Middleware9.IPWhiteList.SourceRange":                         "foobar, fiibar",
		"traefik.Middlewares.Middleware10.MaxConn.Amount":                                 "42",
		"traefik.Middlewares.Middleware10.MaxConn.ExtractorFunc":                          "foobar",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.NotAfter":                "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.NotBefore":               "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Sans":                    "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.CommonName":      "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Country":         "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.DomainComponent": "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Locality":        "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Organization":    "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Province":        "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.SerialNumber":    "true",
		"traefik.Middlewares.Middleware11.PassTLSClientCert.PEM":                          "true",
		"traefik.Middlewares.Middleware12.RateLimit.ExtractorFunc":                        "foobar",
		"traefik.Middlewares.Middleware12.RateLimit.RateSet.Rate0.Average":                "42",
		"traefik.Middlewares.Middleware12.RateLimit.RateSet.Rate0.Burst":                  "42",
		"traefik.Middlewares.Middleware12.RateLimit.RateSet.Rate0.Period":                 "42",
		"traefik.Middlewares.Middleware12.RateLimit.RateSet.Rate1.Average":                "42",
		"traefik.Middlewares.Middleware12.RateLimit.RateSet.Rate1.Burst":                  "42",
		"traefik.Middlewares.Middleware12.RateLimit.RateSet.Rate1.Period":                 "42",
		"traefik.Middlewares.Middleware13.Redirect.Permanent":                             "true",
		"traefik.Middlewares.Middleware13.Redirect.Regex":                                 "foobar",
		"traefik.Middlewares.Middleware13.Redirect.Replacement":                           "foobar",
		"traefik.Middlewares.Middleware14.ReplacePath.Path":                               "foobar",
		"traefik.Middlewares.Middleware15.ReplacePathRegex.Regex":                         "foobar",
		"traefik.Middlewares.Middleware15.ReplacePathRegex.Replacement":                   "foobar",
		"traefik.Middlewares.Middleware16.Retry.Attempts":                                 "42",
		"traefik.Middlewares.Middleware17.StripPrefix.Prefixes":                           "foobar, fiibar",
		"traefik.Middlewares.Middleware18.StripPrefixRegex.Regex":                         "foobar, fiibar",
		"traefik.Middlewares.Middleware19.Compress":                                       "true",

		"traefik.Routers.Router0.EntryPoints": "foobar, fiibar",
		"traefik.Routers.Router0.Middlewares": "foobar, fiibar",
		"traefik.Routers.Router0.Priority":    "42",
		"traefik.Routers.Router0.Rule":        "foobar",
		"traefik.Routers.Router0.Service":     "foobar",
		"traefik.Routers.Router1.EntryPoints": "foobar, fiibar",
		"traefik.Routers.Router1.Middlewares": "foobar, fiibar",
		"traefik.Routers.Router1.Priority":    "42",
		"traefik.Routers.Router1.Rule":        "foobar",
		"traefik.Routers.Router1.Service":     "foobar",

		"traefik.Services.Service0.LoadBalancer.HealthCheck.Headers.name0":        "foobar",
		"traefik.Services.Service0.LoadBalancer.HealthCheck.Headers.name1":        "foobar",
		"traefik.Services.Service0.LoadBalancer.HealthCheck.Hostname":             "foobar",
		"traefik.Services.Service0.LoadBalancer.HealthCheck.Interval":             "foobar",
		"traefik.Services.Service0.LoadBalancer.HealthCheck.Path":                 "foobar",
		"traefik.Services.Service0.LoadBalancer.HealthCheck.Port":                 "42",
		"traefik.Services.Service0.LoadBalancer.HealthCheck.Scheme":               "foobar",
		"traefik.Services.Service0.LoadBalancer.HealthCheck.Timeout":              "foobar",
		"traefik.Services.Service0.LoadBalancer.Method":                           "foobar",
		"traefik.Services.Service0.LoadBalancer.PassHostHeader":                   "true",
		"traefik.Services.Service0.LoadBalancer.ResponseForwarding.FlushInterval": "foobar",
		"traefik.Services.Service0.LoadBalancer.server.URL":                       "foobar",
		"traefik.Services.Service0.LoadBalancer.server.Weight":                    "42",
		"traefik.Services.Service0.LoadBalancer.Stickiness.CookieName":            "foobar",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Headers.name0":        "foobar",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Headers.name1":        "foobar",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Hostname":             "foobar",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Interval":             "foobar",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Path":                 "foobar",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Port":                 "42",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Scheme":               "foobar",
		"traefik.Services.Service1.LoadBalancer.HealthCheck.Timeout":              "foobar",
		"traefik.Services.Service1.LoadBalancer.Method":                           "foobar",
		"traefik.Services.Service1.LoadBalancer.PassHostHeader":                   "true",
		"traefik.Services.Service1.LoadBalancer.ResponseForwarding.FlushInterval": "foobar",
		"traefik.Services.Service1.LoadBalancer.server.URL":                       "foobar",
		"traefik.Services.Service1.LoadBalancer.server.Weight":                    "42",
	}

	for key, val := range expected {
		if _, ok := labels[key]; !ok {
			fmt.Println("missing in labels:", key, val)
		}
	}

	for key, val := range labels {
		if _, ok := expected[key]; !ok {
			fmt.Println("missing in expected:", key, val)
		}
	}
	assert.Equal(t, expected, labels)
}
