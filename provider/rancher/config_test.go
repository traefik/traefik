package rancher

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderBuildConfiguration(t *testing.T) {
	provider := &Provider{
		Domain:           "rancher.localhost",
		ExposedByDefault: true,
	}

	testCases := []struct {
		desc              string
		services          []rancherData
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "without services",
			services:          []rancherData{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when all labels are set",
			services: []rancherData{
				{
					Labels: map[string]string{
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",

						label.TraefikBackend: "foobar",

						label.TraefikBackendCircuitBreakerExpression:         "NetworkErrorRatio() > 0.5",
						label.TraefikBackendResponseForwardingFlushInterval:  "10ms",
						label.TraefikBackendHealthCheckScheme:                "http",
						label.TraefikBackendHealthCheckPath:                  "/health",
						label.TraefikBackendHealthCheckPort:                  "880",
						label.TraefikBackendHealthCheckInterval:              "6",
						label.TraefikBackendHealthCheckHostname:              "foo.com",
						label.TraefikBackendHealthCheckHeaders:               "Foo:bar || Bar:foo",
						label.TraefikBackendLoadBalancerMethod:               "drr",
						label.TraefikBackendLoadBalancerSticky:               "true",
						label.TraefikBackendLoadBalancerStickiness:           "true",
						label.TraefikBackendLoadBalancerStickinessCookieName: "chocolate",
						label.TraefikBackendMaxConnAmount:                    "666",
						label.TraefikBackendMaxConnExtractorFunc:             "client.ip",
						label.TraefikBackendBufferingMaxResponseBodyBytes:    "10485760",
						label.TraefikBackendBufferingMemResponseBodyBytes:    "2097152",
						label.TraefikBackendBufferingMaxRequestBodyBytes:     "10485760",
						label.TraefikBackendBufferingMemRequestBodyBytes:     "2097152",
						label.TraefikBackendBufferingRetryExpression:         "IsNetworkError() && Attempts() <= 2",

						label.TraefikFrontendPassTLSClientCertPem:                         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCommonName:       "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCountry:          "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent:  "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerLocality:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerOrganization:     "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerProvince:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber:     "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName:      "true",
						label.TraefikFrontendPassTLSClientCertInfosNotBefore:              "true",
						label.TraefikFrontendPassTLSClientCertInfosNotAfter:               "true",
						label.TraefikFrontendPassTLSClientCertInfosSans:                   "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCountry:         "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent: "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectLocality:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization:    "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectProvince:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber:    "true",

						label.TraefikFrontendAuthBasic:                        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicRemoveHeader:            "true",
						label.TraefikFrontendAuthBasicUsers:                   "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicUsersFile:               ".htpasswd",
						label.TraefikFrontendAuthDigestRemoveHeader:           "true",
						label.TraefikFrontendAuthDigestUsers:                  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthDigestUsersFile:              ".htpasswd",
						label.TraefikFrontendAuthForwardAddress:               "auth.server",
						label.TraefikFrontendAuthForwardTrustForwardHeader:    "true",
						label.TraefikFrontendAuthForwardTLSCa:                 "ca.crt",
						label.TraefikFrontendAuthForwardTLSCaOptional:         "true",
						label.TraefikFrontendAuthForwardTLSCert:               "server.crt",
						label.TraefikFrontendAuthForwardTLSKey:                "server.key",
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.TraefikFrontendAuthHeaderField:                  "X-WebAuth-User",

						label.TraefikFrontendEntryPoints:               "http,https",
						label.TraefikFrontendPassHostHeader:            "true",
						label.TraefikFrontendPassTLSCert:               "true",
						label.TraefikFrontendPriority:                  "666",
						label.TraefikFrontendRedirectEntryPoint:        "https",
						label.TraefikFrontendRedirectRegex:             "nope",
						label.TraefikFrontendRedirectReplacement:       "nope",
						label.TraefikFrontendRedirectPermanent:         "true",
						label.TraefikFrontendRule:                      "Host:traefik.io",
						label.TraefikFrontendWhiteListSourceRange:      "10.10.10.10",
						label.TraefikFrontendWhiteListUseXForwardedFor: "true",

						label.TraefikFrontendRequestHeaders:          "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendResponseHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendAllowedHosts:            "foo,bar,bor",
						label.TraefikFrontendHostsProxyHeaders:       "foo,bar,bor",
						label.TraefikFrontendSSLHost:                 "foo",
						label.TraefikFrontendCustomFrameOptionsValue: "foo",
						label.TraefikFrontendContentSecurityPolicy:   "foo",
						label.TraefikFrontendPublicKey:               "foo",
						label.TraefikFrontendReferrerPolicy:          "foo",
						label.TraefikFrontendCustomBrowserXSSValue:   "foo",
						label.TraefikFrontendSTSSeconds:              "666",
						label.TraefikFrontendSSLForceHost:            "true",
						label.TraefikFrontendSSLRedirect:             "true",
						label.TraefikFrontendSSLTemporaryRedirect:    "true",
						label.TraefikFrontendSTSIncludeSubdomains:    "true",
						label.TraefikFrontendSTSPreload:              "true",
						label.TraefikFrontendForceSTSHeader:          "true",
						label.TraefikFrontendFrameDeny:               "true",
						label.TraefikFrontendContentTypeNosniff:      "true",
						label.TraefikFrontendBrowserXSSFilter:        "true",
						label.TraefikFrontendIsDevelopment:           "true",

						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",

						label.TraefikFrontendRateLimitExtractorFunc:                                        "client.ip",
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
					},
					Health:     "healthy",
					Containers: []string{"10.0.0.1", "10.0.0.2"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-traefik-io": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-foobar",
					Routes: map[string]types.Route{
						"route-frontend-Host-traefik-io": {
							Rule: "Host:traefik.io",
						},
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					PassTLSClientCert: &types.TLSClientHeaders{
						PEM: true,
						Infos: &types.TLSClientCertificateInfos{
							NotBefore: true,
							Sans:      true,
							NotAfter:  true,
							Subject: &types.TLSCLientCertificateDNInfos{
								CommonName:      true,
								Country:         true,
								DomainComponent: true,
								Locality:        true,
								Organization:    true,
								Province:        true,
								SerialNumber:    true,
							},
							Issuer: &types.TLSCLientCertificateDNInfos{
								CommonName:      true,
								Country:         true,
								DomainComponent: true,
								Locality:        true,
								Organization:    true,
								Province:        true,
								SerialNumber:    true,
							},
						},
					},
					Auth: &types.Auth{
						HeaderField: "X-WebAuth-User",
						Basic: &types.Basic{
							RemoveHeader: true,
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					WhiteList: &types.WhiteList{
						SourceRange: []string{
							"10.10.10.10",
						},
						UseXForwardedFor: true,
					},
					Headers: &types.Headers{
						CustomRequestHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						CustomResponseHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						AllowedHosts: []string{
							"foo",
							"bar",
							"bor",
						},
						HostsProxyHeaders: []string{
							"foo",
							"bar",
							"bor",
						},
						SSLRedirect:          true,
						SSLTemporaryRedirect: true,
						SSLForceHost:         true,
						SSLHost:              "foo",
						SSLProxyHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						STSSeconds:              666,
						STSIncludeSubdomains:    true,
						STSPreload:              true,
						ForceSTSHeader:          true,
						FrameDeny:               true,
						CustomFrameOptionsValue: "foo",
						ContentTypeNosniff:      true,
						BrowserXSSFilter:        true,
						CustomBrowserXSSValue:   "foo",
						ContentSecurityPolicy:   "foo",
						PublicKey:               "foo",
						ReferrerPolicy:          "foo",
						IsDevelopment:           true,
					},
					Errors: map[string]*types.ErrorPage{
						"foo": {
							Status:  []string{"404"},
							Query:   "foo_query",
							Backend: "backend-foobar",
						},
						"bar": {
							Status:  []string{"500", "600"},
							Query:   "bar_query",
							Backend: "backend-foobar",
						},
					},
					RateLimit: &types.RateLimit{
						ExtractorFunc: "client.ip",
						RateSet: map[string]*types.Rate{
							"foo": {
								Period:  flaeg.Duration(6 * time.Second),
								Average: 12,
								Burst:   18,
							},
							"bar": {
								Period:  flaeg.Duration(3 * time.Second),
								Average: 6,
								Burst:   9,
							},
						},
					},
					Redirect: &types.Redirect{
						EntryPoint:  "https",
						Regex:       "",
						Replacement: "",
						Permanent:   true,
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "https://10.0.0.1:666",
							Weight: 12,
						},
						"server-1": {
							URL:    "https://10.0.0.2:666",
							Weight: 12,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					ResponseForwarding: &types.ResponseForwarding{
						FlushInterval: "10ms",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
						Sticky: true,
						Stickiness: &types.Stickiness{
							CookieName: "chocolate",
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        666,
						ExtractorFunc: "client.ip",
					},
					HealthCheck: &types.HealthCheck{
						Scheme:   "http",
						Path:     "/health",
						Port:     880,
						Interval: "6",
						Hostname: "foo.com",
						Headers: map[string]string{
							"Foo": "bar",
							"Bar": "foo",
						},
					},
					Buffering: &types.Buffering{
						MaxResponseBodyBytes: 10485760,
						MemResponseBodyBytes: 2097152,
						MaxRequestBodyBytes:  10485760,
						MemRequestBodyBytes:  2097152,
						RetryExpression:      "IsNetworkError() && Attempts() <= 2",
					},
				},
			},
		},
		{
			desc: "when all segment labels are set",
			services: []rancherData{
				{
					Labels: map[string]string{
						label.Prefix + "sauternes." + label.SuffixPort:     "666",
						label.Prefix + "sauternes." + label.SuffixProtocol: "https",
						label.Prefix + "sauternes." + label.SuffixWeight:   "12",

						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerCommonName:       "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerCountry:          "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerDomainComponent:  "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerLocality:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerOrganization:     "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerProvince:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerSerialNumber:     "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertPem:                         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosNotAfter:               "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosNotBefore:              "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSans:                   "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectCommonName:      "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectCountry:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectDomainComponent: "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectLocality:        "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectOrganization:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectProvince:        "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectSerialNumber:    "true",

						label.Prefix + "sauternes." + label.SuffixFrontendRule:                             "Host:traefik.wtf",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicRemoveHeader:            "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicUsers:                   "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicUsersFile:               ".htpasswd",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestRemoveHeader:           "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestUsers:                  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestUsersFile:              ".htpasswd",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardAddress:               "auth.server",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTrustForwardHeader:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCa:                 "ca.crt",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCaOptional:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCert:               "server.crt",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSKey:                "server.key",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthHeaderField:                  "X-WebAuth-User",

						label.Prefix + "sauternes." + label.SuffixFrontendEntryPoints:               "http,https",
						label.Prefix + "sauternes." + label.SuffixFrontendPassHostHeader:            "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSCert:               "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPriority:                  "666",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectEntryPoint:        "https",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectRegex:             "nope",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectReplacement:       "nope",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectPermanent:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendWhiteListSourceRange:      "10.10.10.10",
						label.Prefix + "sauternes." + label.SuffixFrontendWhiteListUseXForwardedFor: "true",

						label.Prefix + "sauternes." + label.SuffixFrontendRequestHeaders:                 "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "sauternes." + label.SuffixFrontendResponseHeaders:                "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersAllowedHosts:            "foo,bar,bor",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersHostsProxyHeaders:       "foo,bar,bor",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLHost:                 "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersCustomFrameOptionsValue: "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersContentSecurityPolicy:   "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersPublicKey:               "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersReferrerPolicy:          "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersCustomBrowserXSSValue:   "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSTSSeconds:              "666",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLForceHost:            "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLRedirect:             "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLTemporaryRedirect:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSTSIncludeSubdomains:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSTSPreload:              "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersForceSTSHeader:          "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersFrameDeny:               "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersContentTypeNosniff:      "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersBrowserXSSFilter:        "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersIsDevelopment:           "true",

						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",

						label.Prefix + "sauternes." + label.SuffixFrontendRateLimitExtractorFunc:                          "client.ip",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
					},
					Health:     "healthy",
					Containers: []string{"10.0.0.1", "10.0.0.2"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes": {
					EntryPoints: []string{"http", "https"},
					Backend:     "backend-sauternes",
					Routes: map[string]types.Route{
						"route-frontend-sauternes": {
							Rule: "Host:traefik.wtf",
						},
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					PassTLSClientCert: &types.TLSClientHeaders{
						PEM: true,
						Infos: &types.TLSClientCertificateInfos{
							NotBefore: true,
							Sans:      true,
							NotAfter:  true,
							Subject: &types.TLSCLientCertificateDNInfos{
								CommonName:      true,
								Country:         true,
								DomainComponent: true,
								Locality:        true,
								Organization:    true,
								Province:        true,
								SerialNumber:    true,
							},
							Issuer: &types.TLSCLientCertificateDNInfos{
								CommonName:      true,
								Country:         true,
								DomainComponent: true,
								Locality:        true,
								Organization:    true,
								Province:        true,
								SerialNumber:    true,
							},
						},
					},
					Auth: &types.Auth{
						HeaderField: "X-WebAuth-User",
						Basic: &types.Basic{
							RemoveHeader: true,
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					WhiteList: &types.WhiteList{
						SourceRange: []string{
							"10.10.10.10",
						},
						UseXForwardedFor: true,
					},
					Headers: &types.Headers{
						CustomRequestHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						CustomResponseHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						AllowedHosts:         []string{"foo", "bar", "bor"},
						HostsProxyHeaders:    []string{"foo", "bar", "bor"},
						SSLRedirect:          true,
						SSLTemporaryRedirect: true,
						SSLForceHost:         true,
						SSLHost:              "foo",
						SSLProxyHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						STSSeconds:              666,
						STSIncludeSubdomains:    true,
						STSPreload:              true,
						ForceSTSHeader:          true,
						FrameDeny:               true,
						CustomFrameOptionsValue: "foo",
						ContentTypeNosniff:      true,
						BrowserXSSFilter:        true,
						CustomBrowserXSSValue:   "foo",
						ContentSecurityPolicy:   "foo",
						PublicKey:               "foo",
						ReferrerPolicy:          "foo",
						IsDevelopment:           true,
					},
					Errors: map[string]*types.ErrorPage{
						"bar": {
							Status:  []string{"500", "600"},
							Backend: "backend-foobar",
							Query:   "bar_query",
						},
						"foo": {
							Status:  []string{"404"},
							Backend: "backend-foobar",
							Query:   "foo_query",
						},
					},
					RateLimit: &types.RateLimit{
						ExtractorFunc: "client.ip",
						RateSet: map[string]*types.Rate{
							"foo": {
								Period:  flaeg.Duration(6 * time.Second),
								Average: 12,
								Burst:   18,
							},
							"bar": {
								Period:  flaeg.Duration(3 * time.Second),
								Average: 6,
								Burst:   9,
							},
						},
					},
					Redirect: &types.Redirect{
						EntryPoint:  "https",
						Regex:       "",
						Replacement: "",
						Permanent:   true,
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-sauternes": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "https://10.0.0.1:666",
							Weight: 12,
						},
						"server-1": {
							URL:    "https://10.0.0.2:666",
							Weight: 12,
						},
					},
				},
			},
		},
		{
			desc: "with services",
			services: []rancherData{
				{
					Name: "test/service",
					Labels: map[string]string{
						label.TraefikPort:                       "80",
						label.TraefikFrontendAuthBasicUsers:     "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicUsersFile: ".htpasswd",
						label.TraefikFrontendRedirectEntryPoint: "https",
					},
					Health:     "healthy",
					Containers: []string{"127.0.0.1"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-service-rancher-localhost": {
					Backend:        "backend-test-service",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Basic: &types.Basic{
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					Priority: 0,
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-service-rancher-localhost": {
							Rule: "Host:test.service.rancher.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test-service": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "with basic auth backward compatibility",
			services: []rancherData{
				{
					Name: "test/service",
					Labels: map[string]string{
						label.TraefikPort:              "80",
						label.TraefikFrontendAuthBasic: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					},
					Health:     "healthy",
					Containers: []string{"127.0.0.1"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-service-rancher-localhost": {
					Backend:        "backend-test-service",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Basic: &types.Basic{
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
						},
					},
					Priority: 0,
					Routes: map[string]types.Route{
						"route-frontend-Host-test-service-rancher-localhost": {
							Rule: "Host:test.service.rancher.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test-service": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "with digest auth",
			services: []rancherData{
				{
					Name: "test/service",
					Labels: map[string]string{
						label.TraefikPort:                           "80",
						label.TraefikFrontendAuthDigestUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthDigestUsersFile:    ".htpasswd",
						label.TraefikFrontendAuthDigestRemoveHeader: "true",
					},
					Health:     "healthy",
					Containers: []string{"127.0.0.1"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-service-rancher-localhost": {
					Backend:        "backend-test-service",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Digest: &types.Digest{
							RemoveHeader: true,
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					Priority: 0,
					Routes: map[string]types.Route{
						"route-frontend-Host-test-service-rancher-localhost": {
							Rule: "Host:test.service.rancher.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test-service": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "with forward auth",
			services: []rancherData{
				{
					Name: "test/service",
					Labels: map[string]string{
						label.TraefikPort:                                     "80",
						label.TraefikFrontendAuthForwardAddress:               "auth.server",
						label.TraefikFrontendAuthForwardTrustForwardHeader:    "true",
						label.TraefikFrontendAuthForwardTLSCa:                 "ca.crt",
						label.TraefikFrontendAuthForwardTLSCaOptional:         "true",
						label.TraefikFrontendAuthForwardTLSCert:               "server.crt",
						label.TraefikFrontendAuthForwardTLSKey:                "server.key",
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.TraefikFrontendAuthHeaderField:                  "X-WebAuth-User",
						label.TraefikFrontendAuthForwardAuthResponseHeaders:   "X-Auth-User,X-Auth-Token",
					},
					Health:     "healthy",
					Containers: []string{"127.0.0.1"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-service-rancher-localhost": {
					Backend:        "backend-test-service",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						HeaderField: "X-WebAuth-User",
						Forward: &types.Forward{
							Address: "auth.server",
							TLS: &types.ClientTLS{
								CA:                 "ca.crt",
								CAOptional:         true,
								InsecureSkipVerify: true,
								Cert:               "server.crt",
								Key:                "server.key",
							},
							TrustForwardHeader:  true,
							AuthResponseHeaders: []string{"X-Auth-User", "X-Auth-Token"},
						},
					},
					Priority: 0,
					Routes: map[string]types.Route{
						"route-frontend-Host-test-service-rancher-localhost": {
							Rule: "Host:test.service.rancher.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test-service": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actualConfig := provider.buildConfiguration(test.services)
			require.NotNil(t, actualConfig)

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestProviderServiceFilter(t *testing.T) {
	provider := &Provider{
		Domain:                    "rancher.localhost",
		EnableServiceHealthFilter: true,
	}

	constraint, _ := types.NewConstraint("tag==ch*se")
	provider.Constraints = types.Constraints{constraint}

	testCases := []struct {
		desc     string
		service  rancherData
		expected bool
	}{
		{
			desc: "missing Port labels, don't respect constraint",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "don't respect constraint",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "false",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "unhealthy",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "unhealthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "inactive",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "not-cheesy",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "inactive",
			},
			expected: false,
		},
		{
			desc: "healthy & active, tag: cheese",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: true,
		},
		{
			desc: "healthy & active, tag: chose",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "chose",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: true,
		},
		{
			desc: "healthy & upgraded",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheeeeese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "upgraded",
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.serviceFilter(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestContainerFilter(t *testing.T) {
	testCases := []struct {
		name        string
		healthState string
		state       string
		expected    bool
	}{
		{
			healthState: "unhealthy",
			state:       "running",
			expected:    false,
		},
		{
			healthState: "healthy",
			state:       "stopped",
			expected:    false,
		},
		{
			state:    "stopped",
			expected: false,
		},
		{
			healthState: "healthy",
			state:       "running",
			expected:    true,
		},
		{
			healthState: "updating-healthy",
			state:       "updating-running",
			expected:    true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.healthState+" "+test.state, func(t *testing.T) {
			t.Parallel()

			actual := containerFilter(test.name, test.healthState, test.state)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetFrontendName(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "default",
			service: rancherData{
				Name: "foo",
			},
			expected: "Host-foo-rancher-localhost",
		},
		{
			desc: "with Headers label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
				},
			},
			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			desc: "with Host label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Host:foo.bar",
				},
			},
			expected: "Host-foo-bar",
		},
		{
			desc: "with Path label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Path:/test",
				},
			},
			expected: "Path-test",
		},
		{
			desc: "with PathPrefix label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "PathPrefix:/test2",
				},
			},
			expected: "PathPrefix-test2",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			segmentProperties := label.ExtractTraefikLabels(test.service.Labels)
			test.service.SegmentLabels = segmentProperties[""]

			actual := provider.getFrontendName(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetFrontendRule(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "host",
			service: rancherData{
				Name: "foo",
			},
			expected: "Host:foo.rancher.localhost",
		},
		{
			desc: "with domain label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikDomain: "traefik.localhost",
				},
			},
			expected: "Host:test-service.traefik.localhost",
		},
		{
			desc: "host with /",
			service: rancherData{
				Name: "foo/bar",
			},
			expected: "Host:foo.bar.rancher.localhost",
		},
		{
			desc: "with Host label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Host:foo.bar.com",
				},
			},
			expected: "Host:foo.bar.com",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			segmentProperties := label.ExtractTraefikLabels(test.service.Labels)
			test.service.SegmentLabels = segmentProperties[""]

			actual := provider.getFrontendRule(test.service.Name, test.service.SegmentLabels)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendName(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "without label",
			service: rancherData{
				Name: "test-service",
			},
			expected: "test-service",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikBackend: "foobar",
				},
			},

			expected: "foobar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			segmentProperties := label.ExtractTraefikLabels(test.service.Labels)
			test.service.SegmentLabels = segmentProperties[""]

			actual := getBackendName(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetServers(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected map[string]types.Server
	}{
		{
			desc: "should return nil when no server labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return nil when no server IPs",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikWeight: "7",
				},
				Containers: []string{},
				Health:     "healthy",
				State:      "active",
			},
			expected: nil,
		},
		{
			desc: "should return nil when no server IPs",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikWeight: "7",
				},
				Containers: []string{""},
				Health:     "healthy",
				State:      "active",
			},
			expected: nil,
		},
		{
			desc: "should use default weight when invalid weight value",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikWeight: "kls",
				},
				Containers: []string{"10.10.10.0"},
				Health:     "healthy",
				State:      "active",
			},
			expected: map[string]types.Server{
				"server-0": {
					URL:    "http://10.10.10.0:",
					Weight: label.DefaultWeight,
				},
			},
		},
		{
			desc: "should return a map when configuration keys are defined",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikWeight: "6",
				},
				Containers: []string{"10.10.10.0", "10.10.10.1"},
				Health:     "healthy",
				State:      "active",
			},
			expected: map[string]types.Server{
				"server-0": {
					URL:    "http://10.10.10.0:",
					Weight: 6,
				},
				"server-1": {
					URL:    "http://10.10.10.1:",
					Weight: 6,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			segmentProperties := label.ExtractTraefikLabels(test.service.Labels)
			test.service.SegmentLabels = segmentProperties[""]

			actual := getServers(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}
