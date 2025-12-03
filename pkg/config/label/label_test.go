package label

import (
	"fmt"
	"testing"
	"time"

	"github.com/baqupio/baqup/v3/pkg/config/dynamic"
	"github.com/baqupio/baqup/v3/pkg/tls"
	"github.com/baqupio/baqup/v3/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
)

func pointer[T any](v T) *T { return &v }

func TestDecodeConfiguration(t *testing.T) {
	labels := map[string]string{
		"baqup.http.middlewares.Middleware0.addprefix.prefix":                                    "foobar",
		"baqup.http.middlewares.Middleware1.basicauth.headerfield":                               "foobar",
		"baqup.http.middlewares.Middleware1.basicauth.realm":                                     "foobar",
		"baqup.http.middlewares.Middleware1.basicauth.removeheader":                              "true",
		"baqup.http.middlewares.Middleware1.basicauth.users":                                     "foobar, fiibar",
		"baqup.http.middlewares.Middleware1.basicauth.usersfile":                                 "foobar",
		"baqup.http.middlewares.Middleware2.buffering.maxrequestbodybytes":                       "42",
		"baqup.http.middlewares.Middleware2.buffering.maxresponsebodybytes":                      "42",
		"baqup.http.middlewares.Middleware2.buffering.memrequestbodybytes":                       "42",
		"baqup.http.middlewares.Middleware2.buffering.memresponsebodybytes":                      "42",
		"baqup.http.middlewares.Middleware2.buffering.retryexpression":                           "foobar",
		"baqup.http.middlewares.Middleware3.chain.middlewares":                                   "foobar, fiibar",
		"baqup.http.middlewares.Middleware4.circuitbreaker.expression":                           "foobar",
		"baqup.HTTP.Middlewares.Middleware4.circuitbreaker.checkperiod":                          "1s",
		"baqup.HTTP.Middlewares.Middleware4.circuitbreaker.fallbackduration":                     "1s",
		"baqup.HTTP.Middlewares.Middleware4.circuitbreaker.recoveryduration":                     "1s",
		"baqup.HTTP.Middlewares.Middleware4.circuitbreaker.responsecode":                         "403",
		"baqup.http.middlewares.Middleware5.digestauth.headerfield":                              "foobar",
		"baqup.http.middlewares.Middleware5.digestauth.realm":                                    "foobar",
		"baqup.http.middlewares.Middleware5.digestauth.removeheader":                             "true",
		"baqup.http.middlewares.Middleware5.digestauth.users":                                    "foobar, fiibar",
		"baqup.http.middlewares.Middleware5.digestauth.usersfile":                                "foobar",
		"baqup.http.middlewares.Middleware6.errors.query":                                        "foobar",
		"baqup.http.middlewares.Middleware6.errors.service":                                      "foobar",
		"baqup.http.middlewares.Middleware6.errors.status":                                       "foobar, fiibar",
		"baqup.http.middlewares.Middleware7.forwardauth.address":                                 "foobar",
		"baqup.http.middlewares.Middleware7.forwardauth.authresponseheaders":                     "foobar, fiibar",
		"baqup.http.middlewares.Middleware7.forwardauth.authrequestheaders":                      "foobar, fiibar",
		"baqup.http.middlewares.Middleware7.forwardauth.tls.ca":                                  "foobar",
		"baqup.http.middlewares.Middleware7.forwardauth.tls.caoptional":                          "true",
		"baqup.http.middlewares.Middleware7.forwardauth.tls.cert":                                "foobar",
		"baqup.http.middlewares.Middleware7.forwardauth.tls.insecureskipverify":                  "true",
		"baqup.http.middlewares.Middleware7.forwardauth.tls.key":                                 "foobar",
		"baqup.http.middlewares.Middleware7.forwardauth.trustforwardheader":                      "true",
		"baqup.http.middlewares.Middleware7.forwardauth.forwardbody":                             "true",
		"baqup.http.middlewares.Middleware7.forwardauth.maxbodysize":                             "42",
		"baqup.http.middlewares.Middleware7.forwardauth.preserveRequestMethod":                   "true",
		"baqup.http.middlewares.Middleware8.headers.accesscontrolallowcredentials":               "true",
		"baqup.http.middlewares.Middleware8.headers.allowedhosts":                                "foobar, fiibar",
		"baqup.http.middlewares.Middleware8.headers.accesscontrolallowheaders":                   "X-foobar, X-fiibar",
		"baqup.http.middlewares.Middleware8.headers.accesscontrolallowmethods":                   "GET, PUT",
		"baqup.http.middlewares.Middleware8.headers.accesscontrolalloworiginList":                "foobar, fiibar",
		"baqup.http.middlewares.Middleware8.headers.accesscontrolalloworiginListRegex":           "foobar, fiibar",
		"baqup.http.middlewares.Middleware8.headers.accesscontrolexposeheaders":                  "X-foobar, X-fiibar",
		"baqup.http.middlewares.Middleware8.headers.accesscontrolmaxage":                         "200",
		"baqup.http.middlewares.Middleware8.headers.addvaryheader":                               "true",
		"baqup.http.middlewares.Middleware8.headers.browserxssfilter":                            "true",
		"baqup.http.middlewares.Middleware8.headers.contentsecuritypolicy":                       "foobar",
		"baqup.http.middlewares.Middleware8.headers.contentsecuritypolicyreportonly":             "foobar",
		"baqup.http.middlewares.Middleware8.headers.contenttypenosniff":                          "true",
		"baqup.http.middlewares.Middleware8.headers.custombrowserxssvalue":                       "foobar",
		"baqup.http.middlewares.Middleware8.headers.customframeoptionsvalue":                     "foobar",
		"baqup.http.middlewares.Middleware8.headers.customrequestheaders.name0":                  "foobar",
		"baqup.http.middlewares.Middleware8.headers.customrequestheaders.name1":                  "foobar",
		"baqup.http.middlewares.Middleware8.headers.customresponseheaders.name0":                 "foobar",
		"baqup.http.middlewares.Middleware8.headers.customresponseheaders.name1":                 "foobar",
		"baqup.http.middlewares.Middleware8.headers.forcestsheader":                              "true",
		"baqup.http.middlewares.Middleware8.headers.framedeny":                                   "true",
		"baqup.http.middlewares.Middleware8.headers.hostsproxyheaders":                           "foobar, fiibar",
		"baqup.http.middlewares.Middleware8.headers.isdevelopment":                               "true",
		"baqup.http.middlewares.Middleware8.headers.publickey":                                   "foobar",
		"baqup.http.middlewares.Middleware8.headers.referrerpolicy":                              "foobar",
		"baqup.http.middlewares.Middleware8.headers.featurepolicy":                               "foobar",
		"baqup.http.middlewares.Middleware8.headers.permissionspolicy":                           "foobar",
		"baqup.http.middlewares.Middleware8.headers.sslforcehost":                                "true",
		"baqup.http.middlewares.Middleware8.headers.sslhost":                                     "foobar",
		"baqup.http.middlewares.Middleware8.headers.sslproxyheaders.name0":                       "foobar",
		"baqup.http.middlewares.Middleware8.headers.sslproxyheaders.name1":                       "foobar",
		"baqup.http.middlewares.Middleware8.headers.sslredirect":                                 "true",
		"baqup.http.middlewares.Middleware8.headers.ssltemporaryredirect":                        "true",
		"baqup.http.middlewares.Middleware8.headers.stsincludesubdomains":                        "true",
		"baqup.http.middlewares.Middleware8.headers.stspreload":                                  "true",
		"baqup.http.middlewares.Middleware8.headers.stsseconds":                                  "42",
		"baqup.http.middlewares.Middleware9.ipallowlist.ipstrategy.depth":                        "42",
		"baqup.http.middlewares.Middleware9.ipallowlist.ipstrategy.excludedips":                  "foobar, fiibar",
		"baqup.http.middlewares.Middleware9.ipallowlist.ipstrategy.ipv6subnet":                   "42",
		"baqup.http.middlewares.Middleware9.ipallowlist.sourcerange":                             "foobar, fiibar",
		"baqup.http.middlewares.Middleware10.inflightreq.amount":                                 "42",
		"baqup.http.middlewares.Middleware10.inflightreq.sourcecriterion.ipstrategy.depth":       "42",
		"baqup.http.middlewares.Middleware10.inflightreq.sourcecriterion.ipstrategy.excludedips": "foobar, fiibar",
		"baqup.http.middlewares.Middleware10.inflightreq.sourcecriterion.ipstrategy.ipv6subnet":  "42",
		"baqup.http.middlewares.Middleware10.inflightreq.sourcecriterion.requestheadername":      "foobar",
		"baqup.http.middlewares.Middleware10.inflightreq.sourcecriterion.requesthost":            "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.notafter":                    "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.notbefore":                   "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.sans":                        "true",
		"baqup.http.middlewares.Middleware11.passTLSClientCert.info.serialNumber":                "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.commonname":          "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.country":             "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.domaincomponent":     "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.locality":            "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.organization":        "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.organizationalunit":  "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.province":            "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.subject.serialnumber":        "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.issuer.commonname":           "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.issuer.country":              "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.issuer.domaincomponent":      "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.issuer.locality":             "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.issuer.organization":         "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.issuer.province":             "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.info.issuer.serialnumber":         "true",
		"baqup.http.middlewares.Middleware11.passtlsclientcert.pem":                              "true",
		"baqup.http.middlewares.Middleware12.ratelimit.average":                                  "42",
		"baqup.http.middlewares.Middleware12.ratelimit.period":                                   "1s",
		"baqup.http.middlewares.Middleware12.ratelimit.burst":                                    "42",
		"baqup.http.middlewares.Middleware12.ratelimit.sourcecriterion.requestheadername":        "foobar",
		"baqup.http.middlewares.Middleware12.ratelimit.sourcecriterion.requesthost":              "true",
		"baqup.http.middlewares.Middleware12.ratelimit.sourcecriterion.ipstrategy.depth":         "42",
		"baqup.http.middlewares.Middleware12.ratelimit.sourcecriterion.ipstrategy.excludedips":   "foobar, foobar",
		"baqup.http.middlewares.Middleware12.ratelimit.sourcecriterion.ipstrategy.ipv6subnet":    "42",
		"baqup.http.middlewares.Middleware13.redirectregex.permanent":                            "true",
		"baqup.http.middlewares.Middleware13.redirectregex.regex":                                "foobar",
		"baqup.http.middlewares.Middleware13.redirectregex.replacement":                          "foobar",
		"baqup.http.middlewares.Middleware13b.redirectscheme.scheme":                             "https",
		"baqup.http.middlewares.Middleware13b.redirectscheme.port":                               "80",
		"baqup.http.middlewares.Middleware13b.redirectscheme.permanent":                          "true",
		"baqup.http.middlewares.Middleware14.replacepath.path":                                   "foobar",
		"baqup.http.middlewares.Middleware15.replacepathregex.regex":                             "foobar",
		"baqup.http.middlewares.Middleware15.replacepathregex.replacement":                       "foobar",
		"baqup.http.middlewares.Middleware16.retry.attempts":                                     "42",
		"baqup.http.middlewares.Middleware16.retry.initialinterval":                              "1s",
		"baqup.http.middlewares.Middleware17.stripprefix.prefixes":                               "foobar, fiibar",
		"baqup.http.middlewares.Middleware17.stripprefix.forceslash":                             "true",
		"baqup.http.middlewares.Middleware18.stripprefixregex.regex":                             "foobar, fiibar",
		"baqup.http.middlewares.Middleware19.compress.encodings":                                 "foobar, fiibar",
		"baqup.http.middlewares.Middleware19.compress.minresponsebodybytes":                      "42",
		"baqup.http.middlewares.Middleware20.plugin.tomato.aaa":                                  "foo1",
		"baqup.http.middlewares.Middleware20.plugin.tomato.bbb":                                  "foo2",
		"baqup.http.routers.Router0.entrypoints":                                                 "foobar, fiibar",
		"baqup.http.routers.Router0.middlewares":                                                 "foobar, fiibar",
		"baqup.http.routers.Router0.priority":                                                    "42",
		"baqup.http.routers.Router0.rule":                                                        "foobar",
		"baqup.http.routers.Router0.tls":                                                         "true",
		"baqup.http.routers.Router0.service":                                                     "foobar",
		"baqup.http.routers.Router1.entrypoints":                                                 "foobar, fiibar",
		"baqup.http.routers.Router1.middlewares":                                                 "foobar, fiibar",
		"baqup.http.routers.Router1.priority":                                                    "42",
		"baqup.http.routers.Router1.rule":                                                        "foobar",
		"baqup.http.routers.Router1.service":                                                     "foobar",

		"baqup.http.services.Service0.loadbalancer.healthcheck.headers.name0":        "foobar",
		"baqup.http.services.Service0.loadbalancer.healthcheck.headers.name1":        "foobar",
		"baqup.http.services.Service0.loadbalancer.healthcheck.hostname":             "foobar",
		"baqup.http.services.Service0.loadbalancer.healthcheck.interval":             "1s",
		"baqup.http.services.Service0.loadbalancer.healthcheck.unhealthyinterval":    "1s",
		"baqup.http.services.Service0.loadbalancer.healthcheck.path":                 "foobar",
		"baqup.http.services.Service0.loadbalancer.healthcheck.method":               "foobar",
		"baqup.http.services.Service0.loadbalancer.healthcheck.status":               "401",
		"baqup.http.services.Service0.loadbalancer.healthcheck.port":                 "42",
		"baqup.http.services.Service0.loadbalancer.healthcheck.scheme":               "foobar",
		"baqup.http.services.Service0.loadbalancer.healthcheck.mode":                 "foobar",
		"baqup.http.services.Service0.loadbalancer.healthcheck.timeout":              "1s",
		"baqup.http.services.Service0.loadbalancer.healthcheck.followredirects":      "true",
		"baqup.http.services.Service0.loadbalancer.passhostheader":                   "true",
		"baqup.http.services.Service0.loadbalancer.responseforwarding.flushinterval": "1s",
		"baqup.http.services.Service0.loadbalancer.strategy":                         "foobar",
		"baqup.http.services.Service0.loadbalancer.server.url":                       "foobar",
		"baqup.http.services.Service0.loadbalancer.server.preservepath":              "true",
		"baqup.http.services.Service0.loadbalancer.server.scheme":                    "foobar",
		"baqup.http.services.Service0.loadbalancer.server.port":                      "8080",
		"baqup.http.services.Service0.loadbalancer.sticky.cookie.name":               "foobar",
		"baqup.http.services.Service0.loadbalancer.sticky.cookie.secure":             "true",
		"baqup.http.services.Service0.loadbalancer.sticky.cookie.path":               "/foobar",
		"baqup.http.services.Service0.loadbalancer.sticky.cookie.domain":             "foo.com",
		"baqup.http.services.Service0.loadbalancer.serversTransport":                 "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.headers.name0":        "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.headers.name1":        "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.hostname":             "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.interval":             "1s",
		"baqup.http.services.Service1.loadbalancer.healthcheck.unhealthyinterval":    "1s",
		"baqup.http.services.Service1.loadbalancer.healthcheck.path":                 "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.method":               "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.status":               "401",
		"baqup.http.services.Service1.loadbalancer.healthcheck.port":                 "42",
		"baqup.http.services.Service1.loadbalancer.healthcheck.scheme":               "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.mode":                 "foobar",
		"baqup.http.services.Service1.loadbalancer.healthcheck.timeout":              "1s",
		"baqup.http.services.Service1.loadbalancer.healthcheck.followredirects":      "true",
		"baqup.http.services.Service1.loadbalancer.passhostheader":                   "true",
		"baqup.http.services.Service1.loadbalancer.responseforwarding.flushinterval": "1s",
		"baqup.http.services.Service1.loadbalancer.strategy":                         "foobar",
		"baqup.http.services.Service1.loadbalancer.server.url":                       "foobar",
		"baqup.http.services.Service1.loadbalancer.server.preservepath":              "true",
		"baqup.http.services.Service1.loadbalancer.server.scheme":                    "foobar",
		"baqup.http.services.Service1.loadbalancer.server.port":                      "8080",
		"baqup.http.services.Service1.loadbalancer.sticky":                           "false",
		"baqup.http.services.Service1.loadbalancer.sticky.cookie.name":               "fui",
		"baqup.http.services.Service1.loadbalancer.serversTransport":                 "foobar",

		"baqup.tcp.middlewares.Middleware0.ipallowlist.sourcerange":      "foobar, fiibar",
		"baqup.tcp.middlewares.Middleware2.inflightconn.amount":          "42",
		"baqup.tcp.routers.Router0.rule":                                 "foobar",
		"baqup.tcp.routers.Router0.priority":                             "42",
		"baqup.tcp.routers.Router0.entrypoints":                          "foobar, fiibar",
		"baqup.tcp.routers.Router0.service":                              "foobar",
		"baqup.tcp.routers.Router0.tls.passthrough":                      "false",
		"baqup.tcp.routers.Router0.tls.options":                          "foo",
		"baqup.tcp.routers.Router1.rule":                                 "foobar",
		"baqup.tcp.routers.Router1.priority":                             "42",
		"baqup.tcp.routers.Router1.entrypoints":                          "foobar, fiibar",
		"baqup.tcp.routers.Router1.service":                              "foobar",
		"baqup.tcp.routers.Router1.tls.options":                          "foo",
		"baqup.tcp.routers.Router1.tls.passthrough":                      "false",
		"baqup.tcp.services.Service0.loadbalancer.server.Port":           "42",
		"baqup.tcp.services.Service0.loadbalancer.TerminationDelay":      "42",
		"baqup.tcp.services.Service0.loadbalancer.proxyProtocol.version": "42",
		"baqup.tcp.services.Service0.loadbalancer.serversTransport":      "foo",
		"baqup.tcp.services.Service1.loadbalancer.server.Port":           "42",
		"baqup.tcp.services.Service1.loadbalancer.TerminationDelay":      "42",
		"baqup.tcp.services.Service1.loadbalancer.proxyProtocol":         "true",
		"baqup.tcp.services.Service1.loadbalancer.serversTransport":      "foo",

		"baqup.udp.routers.Router0.entrypoints":                "foobar, fiibar",
		"baqup.udp.routers.Router0.service":                    "foobar",
		"baqup.udp.routers.Router1.entrypoints":                "foobar, fiibar",
		"baqup.udp.routers.Router1.service":                    "foobar",
		"baqup.udp.services.Service0.loadbalancer.server.Port": "42",
		"baqup.udp.services.Service1.loadbalancer.server.Port": "42",

		"baqup.tls.stores.default.defaultgeneratedcert.resolver":    "foobar",
		"baqup.tls.stores.default.defaultgeneratedcert.domain.main": "foobar",
		"baqup.tls.stores.default.defaultgeneratedcert.domain.sans": "foobar, fiibar",
	}

	configuration, err := DecodeConfiguration(labels)
	require.NoError(t, err)

	expected := &dynamic.Configuration{
		TCP: &dynamic.TCPConfiguration{
			Routers: map[string]*dynamic.TCPRouter{
				"Router0": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service:  "foobar",
					Rule:     "foobar",
					Priority: 42,
					TLS: &dynamic.RouterTCPTLSConfig{
						Passthrough: false,
						Options:     "foo",
					},
				},
				"Router1": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service:  "foobar",
					Rule:     "foobar",
					Priority: 42,
					TLS: &dynamic.RouterTCPTLSConfig{
						Passthrough: false,
						Options:     "foo",
					},
				},
			},
			Middlewares: map[string]*dynamic.TCPMiddleware{
				"Middleware0": {
					IPAllowList: &dynamic.TCPIPAllowList{
						SourceRange: []string{"foobar", "fiibar"},
					},
				},
				"Middleware2": {
					InFlightConn: &dynamic.TCPInFlightConn{
						Amount: 42,
					},
				},
			},
			Services: map[string]*dynamic.TCPService{
				"Service0": {
					LoadBalancer: &dynamic.TCPServersLoadBalancer{
						Servers: []dynamic.TCPServer{
							{
								Port: "42",
							},
						},
						TerminationDelay: pointer(42),
						ProxyProtocol:    &dynamic.ProxyProtocol{Version: 42},
						ServersTransport: "foo",
					},
				},
				"Service1": {
					LoadBalancer: &dynamic.TCPServersLoadBalancer{
						Servers: []dynamic.TCPServer{
							{
								Port: "42",
							},
						},
						TerminationDelay: pointer(42),
						ProxyProtocol:    &dynamic.ProxyProtocol{Version: 2},
						ServersTransport: "foo",
					},
				},
			},
		},
		UDP: &dynamic.UDPConfiguration{
			Routers: map[string]*dynamic.UDPRouter{
				"Router0": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service: "foobar",
				},
				"Router1": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service: "foobar",
				},
			},
			Services: map[string]*dynamic.UDPService{
				"Service0": {
					LoadBalancer: &dynamic.UDPServersLoadBalancer{
						Servers: []dynamic.UDPServer{
							{
								Port: "42",
							},
						},
					},
				},
				"Service1": {
					LoadBalancer: &dynamic.UDPServersLoadBalancer{
						Servers: []dynamic.UDPServer{
							{
								Port: "42",
							},
						},
					},
				},
			},
		},
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
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
					TLS:      &dynamic.RouterTLSConfig{},
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
			Middlewares: map[string]*dynamic.Middleware{
				"Middleware0": {
					AddPrefix: &dynamic.AddPrefix{
						Prefix: "foobar",
					},
				},
				"Middleware1": {
					BasicAuth: &dynamic.BasicAuth{
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
					InFlightReq: &dynamic.InFlightReq{
						Amount: 42,
						SourceCriterion: &dynamic.SourceCriterion{
							IPStrategy: &dynamic.IPStrategy{
								Depth:       42,
								ExcludedIPs: []string{"foobar", "fiibar"},
								IPv6Subnet:  intPtr(42),
							},
							RequestHeaderName: "foobar",
							RequestHost:       true,
						},
					},
				},
				"Middleware11": {
					PassTLSClientCert: &dynamic.PassTLSClientCert{
						PEM: true,
						Info: &dynamic.TLSClientCertificateInfo{
							NotAfter:     true,
							NotBefore:    true,
							SerialNumber: true,
							Subject: &dynamic.TLSClientCertificateSubjectDNInfo{
								Country:            true,
								Province:           true,
								Locality:           true,
								Organization:       true,
								OrganizationalUnit: true,
								CommonName:         true,
								SerialNumber:       true,
								DomainComponent:    true,
							},
							Issuer: &dynamic.TLSClientCertificateIssuerDNInfo{
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
					RateLimit: &dynamic.RateLimit{
						Average: 42,
						Burst:   42,
						Period:  ptypes.Duration(time.Second),
						SourceCriterion: &dynamic.SourceCriterion{
							IPStrategy: &dynamic.IPStrategy{
								Depth:       42,
								ExcludedIPs: []string{"foobar", "foobar"},
								IPv6Subnet:  intPtr(42),
							},
							RequestHeaderName: "foobar",
							RequestHost:       true,
						},
					},
				},
				"Middleware13": {
					RedirectRegex: &dynamic.RedirectRegex{
						Regex:       "foobar",
						Replacement: "foobar",
						Permanent:   true,
					},
				},
				"Middleware13b": {
					RedirectScheme: &dynamic.RedirectScheme{
						Scheme:    "https",
						Port:      "80",
						Permanent: true,
					},
				},
				"Middleware14": {
					ReplacePath: &dynamic.ReplacePath{
						Path: "foobar",
					},
				},
				"Middleware15": {
					ReplacePathRegex: &dynamic.ReplacePathRegex{
						Regex:       "foobar",
						Replacement: "foobar",
					},
				},
				"Middleware16": {
					Retry: &dynamic.Retry{
						Attempts:        42,
						InitialInterval: ptypes.Duration(time.Second),
					},
				},
				"Middleware17": {
					StripPrefix: &dynamic.StripPrefix{
						Prefixes: []string{
							"foobar",
							"fiibar",
						},
						ForceSlash: pointer(true),
					},
				},
				"Middleware18": {
					StripPrefixRegex: &dynamic.StripPrefixRegex{
						Regex: []string{
							"foobar",
							"fiibar",
						},
					},
				},
				"Middleware19": {
					Compress: &dynamic.Compress{
						MinResponseBodyBytes: 42,
						Encodings: []string{
							"foobar",
							"fiibar",
						},
					},
				},
				"Middleware2": {
					Buffering: &dynamic.Buffering{
						MaxRequestBodyBytes:  42,
						MemRequestBodyBytes:  42,
						MaxResponseBodyBytes: 42,
						MemResponseBodyBytes: 42,
						RetryExpression:      "foobar",
					},
				},
				"Middleware3": {
					Chain: &dynamic.Chain{
						Middlewares: []string{
							"foobar",
							"fiibar",
						},
					},
				},
				"Middleware4": {
					CircuitBreaker: &dynamic.CircuitBreaker{
						Expression:       "foobar",
						CheckPeriod:      ptypes.Duration(time.Second),
						FallbackDuration: ptypes.Duration(time.Second),
						RecoveryDuration: ptypes.Duration(time.Second),
						ResponseCode:     403,
					},
				},
				"Middleware5": {
					DigestAuth: &dynamic.DigestAuth{
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
					Errors: &dynamic.ErrorPage{
						Status: []string{
							"foobar",
							"fiibar",
						},
						Service: "foobar",
						Query:   "foobar",
					},
				},
				"Middleware7": {
					ForwardAuth: &dynamic.ForwardAuth{
						Address: "foobar",
						TLS: &dynamic.ClientTLS{
							CA:                 "foobar",
							Cert:               "foobar",
							Key:                "foobar",
							InsecureSkipVerify: true,
							CAOptional:         pointer(true),
						},
						TrustForwardHeader: true,
						AuthResponseHeaders: []string{
							"foobar",
							"fiibar",
						},
						AuthRequestHeaders: []string{
							"foobar",
							"fiibar",
						},
						ForwardBody:           true,
						MaxBodySize:           pointer(int64(42)),
						PreserveRequestMethod: true,
					},
				},
				"Middleware8": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
						CustomResponseHeaders: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
						AccessControlAllowCredentials: true,
						AccessControlAllowHeaders: []string{
							"X-foobar",
							"X-fiibar",
						},
						AccessControlAllowMethods: []string{
							"GET",
							"PUT",
						},
						AccessControlAllowOriginList: []string{
							"foobar",
							"fiibar",
						},
						AccessControlAllowOriginListRegex: []string{
							"foobar",
							"fiibar",
						},
						AccessControlExposeHeaders: []string{
							"X-foobar",
							"X-fiibar",
						},
						AccessControlMaxAge: 200,
						AddVaryHeader:       true,
						AllowedHosts: []string{
							"foobar",
							"fiibar",
						},
						HostsProxyHeaders: []string{
							"foobar",
							"fiibar",
						},
						SSLRedirect:          pointer(true),
						SSLTemporaryRedirect: pointer(true),
						SSLHost:              pointer("foobar"),
						SSLProxyHeaders: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
						SSLForceHost:                    pointer(true),
						STSSeconds:                      42,
						STSIncludeSubdomains:            true,
						STSPreload:                      true,
						ForceSTSHeader:                  true,
						FrameDeny:                       true,
						CustomFrameOptionsValue:         "foobar",
						ContentTypeNosniff:              true,
						BrowserXSSFilter:                true,
						CustomBrowserXSSValue:           "foobar",
						ContentSecurityPolicy:           "foobar",
						ContentSecurityPolicyReportOnly: "foobar",
						PublicKey:                       "foobar",
						ReferrerPolicy:                  "foobar",
						FeaturePolicy:                   pointer("foobar"),
						PermissionsPolicy:               "foobar",
						IsDevelopment:                   true,
					},
				},
				"Middleware9": {
					IPAllowList: &dynamic.IPAllowList{
						SourceRange: []string{
							"foobar",
							"fiibar",
						},
						IPStrategy: &dynamic.IPStrategy{
							Depth: 42,
							ExcludedIPs: []string{
								"foobar",
								"fiibar",
							},
							IPv6Subnet: intPtr(42),
						},
					},
				},
				"Middleware20": {
					Plugin: map[string]dynamic.PluginConf{
						"tomato": {
							"aaa": "foo1",
							"bbb": "foo2",
						},
					},
				},
			},
			Services: map[string]*dynamic.Service{
				"Service0": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: "foobar",
						Sticky: &dynamic.Sticky{
							Cookie: &dynamic.Cookie{
								Name:     "foobar",
								Secure:   true,
								HTTPOnly: false,
								Domain:   "foo.com",
								Path:     func(v string) *string { return &v }("/foobar"),
							},
						},
						Servers: []dynamic.Server{
							{
								URL:          "foobar",
								PreservePath: true,
								Scheme:       "foobar",
								Port:         "8080",
							},
						},
						HealthCheck: &dynamic.ServerHealthCheck{
							Scheme:            "foobar",
							Mode:              "foobar",
							Path:              "foobar",
							Method:            "foobar",
							Status:            401,
							Port:              42,
							Interval:          ptypes.Duration(time.Second),
							UnhealthyInterval: pointer(ptypes.Duration(time.Second)),
							Timeout:           ptypes.Duration(time.Second),
							Hostname:          "foobar",
							Headers: map[string]string{
								"name0": "foobar",
								"name1": "foobar",
							},
							FollowRedirects: pointer(true),
						},
						PassHostHeader: pointer(true),
						ResponseForwarding: &dynamic.ResponseForwarding{
							FlushInterval: ptypes.Duration(time.Second),
						},
						ServersTransport: "foobar",
					},
				},
				"Service1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: "foobar",
						Servers: []dynamic.Server{
							{
								URL:          "foobar",
								PreservePath: true,
								Scheme:       "foobar",
								Port:         "8080",
							},
						},
						HealthCheck: &dynamic.ServerHealthCheck{
							Scheme:            "foobar",
							Mode:              "foobar",
							Path:              "foobar",
							Method:            "foobar",
							Status:            401,
							Port:              42,
							Interval:          ptypes.Duration(time.Second),
							UnhealthyInterval: pointer(ptypes.Duration(time.Second)),
							Timeout:           ptypes.Duration(time.Second),
							Hostname:          "foobar",
							Headers: map[string]string{
								"name0": "foobar",
								"name1": "foobar",
							},
							FollowRedirects: pointer(true),
						},
						PassHostHeader: pointer(true),
						ResponseForwarding: &dynamic.ResponseForwarding{
							FlushInterval: ptypes.Duration(time.Second),
						},
						ServersTransport: "foobar",
					},
				},
			},
		},
		TLS: &dynamic.TLSConfiguration{
			Stores: map[string]tls.Store{
				"default": {
					DefaultGeneratedCert: &tls.GeneratedCert{
						Resolver: "foobar",
						Domain: &types.Domain{
							Main: "foobar",
							SANs: []string{"foobar", "fiibar"},
						},
					},
				},
			},
		},
	}

	assert.Nil(t, configuration.HTTP.ServersTransports)
	assert.Nil(t, configuration.TCP.ServersTransports)
	assert.Equal(t, expected, configuration)
}

func TestEncodeConfiguration(t *testing.T) {
	configuration := &dynamic.Configuration{
		TCP: &dynamic.TCPConfiguration{
			Routers: map[string]*dynamic.TCPRouter{
				"Router0": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service:  "foobar",
					Rule:     "foobar",
					Priority: 42,
					TLS: &dynamic.RouterTCPTLSConfig{
						Passthrough: false,
						Options:     "foo",
					},
				},
				"Router1": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service:  "foobar",
					Rule:     "foobar",
					Priority: 42,
					TLS: &dynamic.RouterTCPTLSConfig{
						Passthrough: false,
						Options:     "foo",
					},
				},
			},
			Middlewares: map[string]*dynamic.TCPMiddleware{
				"Middleware0": {
					IPAllowList: &dynamic.TCPIPAllowList{
						SourceRange: []string{"foobar", "fiibar"},
					},
				},
				"Middleware2": {
					InFlightConn: &dynamic.TCPInFlightConn{
						Amount: 42,
					},
				},
			},
			Services: map[string]*dynamic.TCPService{
				"Service0": {
					LoadBalancer: &dynamic.TCPServersLoadBalancer{
						Servers: []dynamic.TCPServer{
							{
								Port: "42",
							},
						},
						ServersTransport: "foo",
						TerminationDelay: pointer(42),
					},
				},
				"Service1": {
					LoadBalancer: &dynamic.TCPServersLoadBalancer{
						Servers: []dynamic.TCPServer{
							{
								Port: "42",
							},
						},
						ServersTransport: "foo",
						TerminationDelay: pointer(42),
					},
				},
			},
		},
		UDP: &dynamic.UDPConfiguration{
			Routers: map[string]*dynamic.UDPRouter{
				"Router0": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service: "foobar",
				},
				"Router1": {
					EntryPoints: []string{
						"foobar",
						"fiibar",
					},
					Service: "foobar",
				},
			},
			Services: map[string]*dynamic.UDPService{
				"Service0": {
					LoadBalancer: &dynamic.UDPServersLoadBalancer{
						Servers: []dynamic.UDPServer{
							{
								Port: "42",
							},
						},
					},
				},
				"Service1": {
					LoadBalancer: &dynamic.UDPServersLoadBalancer{
						Servers: []dynamic.UDPServer{
							{
								Port: "42",
							},
						},
					},
				},
			},
		},
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
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
					TLS:      &dynamic.RouterTLSConfig{},
					Observability: &dynamic.RouterObservabilityConfig{
						AccessLogs: pointer(true),
						Tracing:    pointer(true),
						Metrics:    pointer(true),
					},
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
					Observability: &dynamic.RouterObservabilityConfig{
						AccessLogs: pointer(true),
						Tracing:    pointer(true),
						Metrics:    pointer(true),
					},
				},
			},
			Middlewares: map[string]*dynamic.Middleware{
				"Middleware0": {
					AddPrefix: &dynamic.AddPrefix{
						Prefix: "foobar",
					},
				},
				"Middleware1": {
					BasicAuth: &dynamic.BasicAuth{
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
					InFlightReq: &dynamic.InFlightReq{
						Amount: 42,
						SourceCriterion: &dynamic.SourceCriterion{
							IPStrategy: &dynamic.IPStrategy{
								Depth:       42,
								ExcludedIPs: []string{"foobar", "fiibar"},
								IPv6Subnet:  intPtr(42),
							},
							RequestHeaderName: "foobar",
							RequestHost:       true,
						},
					},
				},
				"Middleware11": {
					PassTLSClientCert: &dynamic.PassTLSClientCert{
						PEM: true,
						Info: &dynamic.TLSClientCertificateInfo{
							NotAfter:     true,
							NotBefore:    true,
							SerialNumber: true,
							Subject: &dynamic.TLSClientCertificateSubjectDNInfo{
								Country:            true,
								Province:           true,
								Locality:           true,
								Organization:       true,
								OrganizationalUnit: true,
								CommonName:         true,
								SerialNumber:       true,
								DomainComponent:    true,
							},
							Issuer: &dynamic.TLSClientCertificateIssuerDNInfo{
								Country:         true,
								Province:        true,
								Locality:        true,
								Organization:    true,
								CommonName:      true,
								SerialNumber:    true,
								DomainComponent: true,
							}, Sans: true,
						},
					},
				},
				"Middleware12": {
					RateLimit: &dynamic.RateLimit{
						Average: 42,
						Burst:   42,
						Period:  ptypes.Duration(time.Second),
						SourceCriterion: &dynamic.SourceCriterion{
							IPStrategy: &dynamic.IPStrategy{
								Depth:       42,
								ExcludedIPs: []string{"foobar", "foobar"},
								IPv6Subnet:  intPtr(42),
							},
							RequestHeaderName: "foobar",
							RequestHost:       true,
						},
					},
				},
				"Middleware13": {
					RedirectRegex: &dynamic.RedirectRegex{
						Regex:       "foobar",
						Replacement: "foobar",
						Permanent:   true,
					},
				},
				"Middleware13b": {
					RedirectScheme: &dynamic.RedirectScheme{
						Scheme:    "https",
						Port:      "80",
						Permanent: true,
					},
				},
				"Middleware14": {
					ReplacePath: &dynamic.ReplacePath{
						Path: "foobar",
					},
				},
				"Middleware15": {
					ReplacePathRegex: &dynamic.ReplacePathRegex{
						Regex:       "foobar",
						Replacement: "foobar",
					},
				},
				"Middleware16": {
					Retry: &dynamic.Retry{
						Attempts:        42,
						InitialInterval: ptypes.Duration(time.Second),
					},
				},
				"Middleware17": {
					StripPrefix: &dynamic.StripPrefix{
						Prefixes: []string{
							"foobar",
							"fiibar",
						},
						ForceSlash: pointer(true),
					},
				},
				"Middleware18": {
					StripPrefixRegex: &dynamic.StripPrefixRegex{
						Regex: []string{
							"foobar",
							"fiibar",
						},
					},
				},
				"Middleware19": {
					Compress: &dynamic.Compress{
						MinResponseBodyBytes: 42,
						Encodings: []string{
							"foobar",
							"fiibar",
						},
					},
				},
				"Middleware2": {
					Buffering: &dynamic.Buffering{
						MaxRequestBodyBytes:  42,
						MemRequestBodyBytes:  42,
						MaxResponseBodyBytes: 42,
						MemResponseBodyBytes: 42,
						RetryExpression:      "foobar",
					},
				},
				"Middleware20": {
					Plugin: map[string]dynamic.PluginConf{
						"tomato": {
							"aaa": "foo1",
							"bbb": "foo2",
						},
					},
				},
				"Middleware3": {
					Chain: &dynamic.Chain{
						Middlewares: []string{
							"foobar",
							"fiibar",
						},
					},
				},
				"Middleware4": {
					CircuitBreaker: &dynamic.CircuitBreaker{
						Expression:       "foobar",
						CheckPeriod:      ptypes.Duration(time.Second),
						FallbackDuration: ptypes.Duration(time.Second),
						RecoveryDuration: ptypes.Duration(time.Second),
						ResponseCode:     404,
					},
				},
				"Middleware5": {
					DigestAuth: &dynamic.DigestAuth{
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
					Errors: &dynamic.ErrorPage{
						Status: []string{
							"foobar",
							"fiibar",
						},
						Service: "foobar",
						Query:   "foobar",
					},
				},
				"Middleware7": {
					ForwardAuth: &dynamic.ForwardAuth{
						Address: "foobar",
						TLS: &dynamic.ClientTLS{
							CA:                 "foobar",
							Cert:               "foobar",
							Key:                "foobar",
							InsecureSkipVerify: true,
							CAOptional:         pointer(true),
						},
						TrustForwardHeader: true,
						AuthResponseHeaders: []string{
							"foobar",
							"fiibar",
						},
						AuthRequestHeaders: []string{
							"foobar",
							"fiibar",
						},
						ForwardBody:           true,
						MaxBodySize:           pointer(int64(42)),
						PreserveRequestMethod: true,
					},
				},
				"Middleware8": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
						CustomResponseHeaders: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
						AccessControlAllowCredentials: true,
						AccessControlAllowHeaders: []string{
							"X-foobar",
							"X-fiibar",
						},
						AccessControlAllowMethods: []string{
							"GET",
							"PUT",
						},
						AccessControlAllowOriginList: []string{
							"foobar",
							"fiibar",
						},
						AccessControlAllowOriginListRegex: []string{
							"foobar",
							"fiibar",
						},
						AccessControlExposeHeaders: []string{
							"X-foobar",
							"X-fiibar",
						},
						AccessControlMaxAge: 200,
						AddVaryHeader:       true,
						AllowedHosts: []string{
							"foobar",
							"fiibar",
						},
						HostsProxyHeaders: []string{
							"foobar",
							"fiibar",
						},
						SSLRedirect:          pointer(true),
						SSLTemporaryRedirect: pointer(true),
						SSLHost:              pointer("foobar"),
						SSLProxyHeaders: map[string]string{
							"name0": "foobar",
							"name1": "foobar",
						},
						SSLForceHost:                    pointer(true),
						STSSeconds:                      42,
						STSIncludeSubdomains:            true,
						STSPreload:                      true,
						ForceSTSHeader:                  true,
						FrameDeny:                       true,
						CustomFrameOptionsValue:         "foobar",
						ContentTypeNosniff:              true,
						BrowserXSSFilter:                true,
						CustomBrowserXSSValue:           "foobar",
						ContentSecurityPolicy:           "foobar",
						ContentSecurityPolicyReportOnly: "foobar",
						PublicKey:                       "foobar",
						ReferrerPolicy:                  "foobar",
						FeaturePolicy:                   pointer("foobar"),
						PermissionsPolicy:               "foobar",
						IsDevelopment:                   true,
					},
				},
				"Middleware9": {
					IPAllowList: &dynamic.IPAllowList{
						SourceRange: []string{
							"foobar",
							"fiibar",
						},
						IPStrategy: &dynamic.IPStrategy{
							Depth: 42,
							ExcludedIPs: []string{
								"foobar",
								"fiibar",
							},
							IPv6Subnet: intPtr(42),
						},
					},
				},
			},
			Services: map[string]*dynamic.Service{
				"Service0": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: "foobar",
						Sticky: &dynamic.Sticky{
							Cookie: &dynamic.Cookie{
								Name:     "foobar",
								HTTPOnly: true,
								Domain:   "foo.com",
								Path:     func(v string) *string { return &v }("/foobar"),
							},
						},
						Servers: []dynamic.Server{
							{
								URL:          "foobar",
								PreservePath: true,
								Scheme:       "foobar",
								Port:         "8080",
							},
						},
						HealthCheck: &dynamic.ServerHealthCheck{
							Scheme:            "foobar",
							Path:              "foobar",
							Method:            "foobar",
							Status:            401,
							Port:              42,
							Interval:          ptypes.Duration(time.Second),
							UnhealthyInterval: pointer(ptypes.Duration(time.Second)),
							Timeout:           ptypes.Duration(time.Second),
							Hostname:          "foobar",
							Headers: map[string]string{
								"name0": "foobar",
								"name1": "foobar",
							},
						},
						PassHostHeader: pointer(true),
						ResponseForwarding: &dynamic.ResponseForwarding{
							FlushInterval: ptypes.Duration(time.Second),
						},
						ServersTransport: "foobar",
					},
				},
				"Service1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: "foobar",
						Servers: []dynamic.Server{
							{
								URL:          "foobar",
								PreservePath: true,
								Scheme:       "foobar",
								Port:         "8080",
							},
						},
						HealthCheck: &dynamic.ServerHealthCheck{
							Scheme:            "foobar",
							Path:              "foobar",
							Method:            "foobar",
							Status:            401,
							Port:              42,
							Interval:          ptypes.Duration(time.Second),
							UnhealthyInterval: pointer(ptypes.Duration(time.Second)),
							Timeout:           ptypes.Duration(time.Second),
							Hostname:          "foobar",
							Headers: map[string]string{
								"name0": "foobar",
								"name1": "foobar",
							},
						},
						PassHostHeader: pointer(true),
						ResponseForwarding: &dynamic.ResponseForwarding{
							FlushInterval: ptypes.Duration(time.Second),
						},
						ServersTransport: "foobar",
					},
				},
			},
		},
		TLS: &dynamic.TLSConfiguration{
			Stores: map[string]tls.Store{
				"default": {
					DefaultGeneratedCert: &tls.GeneratedCert{
						Resolver: "foobar",
						Domain: &types.Domain{
							Main: "foobar",
							SANs: []string{"foobar", "fiibar"},
						},
					},
				},
			},
		},
	}

	labels, err := EncodeConfiguration(configuration)
	require.NoError(t, err)

	expected := map[string]string{
		"baqup.HTTP.Middlewares.Middleware0.AddPrefix.Prefix":                                    "foobar",
		"baqup.HTTP.Middlewares.Middleware1.BasicAuth.HeaderField":                               "foobar",
		"baqup.HTTP.Middlewares.Middleware1.BasicAuth.Realm":                                     "foobar",
		"baqup.HTTP.Middlewares.Middleware1.BasicAuth.RemoveHeader":                              "true",
		"baqup.HTTP.Middlewares.Middleware1.BasicAuth.Users":                                     "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware1.BasicAuth.UsersFile":                                 "foobar",
		"baqup.HTTP.Middlewares.Middleware2.Buffering.MaxRequestBodyBytes":                       "42",
		"baqup.HTTP.Middlewares.Middleware2.Buffering.MaxResponseBodyBytes":                      "42",
		"baqup.HTTP.Middlewares.Middleware2.Buffering.MemRequestBodyBytes":                       "42",
		"baqup.HTTP.Middlewares.Middleware2.Buffering.MemResponseBodyBytes":                      "42",
		"baqup.HTTP.Middlewares.Middleware2.Buffering.RetryExpression":                           "foobar",
		"baqup.HTTP.Middlewares.Middleware3.Chain.Middlewares":                                   "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware4.CircuitBreaker.Expression":                           "foobar",
		"baqup.HTTP.Middlewares.Middleware4.CircuitBreaker.CheckPeriod":                          "1000000000",
		"baqup.HTTP.Middlewares.Middleware4.CircuitBreaker.FallbackDuration":                     "1000000000",
		"baqup.HTTP.Middlewares.Middleware4.CircuitBreaker.RecoveryDuration":                     "1000000000",
		"baqup.HTTP.Middlewares.Middleware4.CircuitBreaker.ResponseCode":                         "404",
		"baqup.HTTP.Middlewares.Middleware5.DigestAuth.HeaderField":                              "foobar",
		"baqup.HTTP.Middlewares.Middleware5.DigestAuth.Realm":                                    "foobar",
		"baqup.HTTP.Middlewares.Middleware5.DigestAuth.RemoveHeader":                             "true",
		"baqup.HTTP.Middlewares.Middleware5.DigestAuth.Users":                                    "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware5.DigestAuth.UsersFile":                                "foobar",
		"baqup.HTTP.Middlewares.Middleware6.Errors.Query":                                        "foobar",
		"baqup.HTTP.Middlewares.Middleware6.Errors.Service":                                      "foobar",
		"baqup.HTTP.Middlewares.Middleware6.Errors.Status":                                       "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.Address":                                 "foobar",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.AuthResponseHeaders":                     "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.AuthRequestHeaders":                      "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.ForwardBody":                             "true",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.MaxBodySize":                             "42",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.TLS.CA":                                  "foobar",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.TLS.CAOptional":                          "true",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.TLS.Cert":                                "foobar",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.TLS.InsecureSkipVerify":                  "true",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.TLS.Key":                                 "foobar",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.TrustForwardHeader":                      "true",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.PreserveLocationHeader":                  "false",
		"baqup.HTTP.Middlewares.Middleware7.ForwardAuth.PreserveRequestMethod":                   "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AccessControlAllowCredentials":               "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AccessControlAllowHeaders":                   "X-foobar, X-fiibar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AccessControlAllowMethods":                   "GET, PUT",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AccessControlAllowOriginList":                "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AccessControlAllowOriginListRegex":           "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AccessControlExposeHeaders":                  "X-foobar, X-fiibar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AccessControlMaxAge":                         "200",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AddVaryHeader":                               "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.AllowedHosts":                                "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.BrowserXSSFilter":                            "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.ContentSecurityPolicy":                       "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.ContentSecurityPolicyReportOnly":             "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.ContentTypeNosniff":                          "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.CustomBrowserXSSValue":                       "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.CustomFrameOptionsValue":                     "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.CustomRequestHeaders.name0":                  "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.CustomRequestHeaders.name1":                  "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.CustomResponseHeaders.name0":                 "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.CustomResponseHeaders.name1":                 "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.ForceSTSHeader":                              "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.FrameDeny":                                   "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.HostsProxyHeaders":                           "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.IsDevelopment":                               "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.PublicKey":                                   "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.ReferrerPolicy":                              "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.FeaturePolicy":                               "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.PermissionsPolicy":                           "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.SSLForceHost":                                "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.SSLHost":                                     "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.SSLProxyHeaders.name0":                       "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.SSLProxyHeaders.name1":                       "foobar",
		"baqup.HTTP.Middlewares.Middleware8.Headers.SSLRedirect":                                 "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.SSLTemporaryRedirect":                        "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.STSIncludeSubdomains":                        "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.STSPreload":                                  "true",
		"baqup.HTTP.Middlewares.Middleware8.Headers.STSSeconds":                                  "42",
		"baqup.HTTP.Middlewares.Middleware9.IPAllowList.IPStrategy.Depth":                        "42",
		"baqup.HTTP.Middlewares.Middleware9.IPAllowList.IPStrategy.ExcludedIPs":                  "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware9.IPAllowList.IPStrategy.IPv6Subnet":                   "42",
		"baqup.HTTP.Middlewares.Middleware9.IPAllowList.RejectStatusCode":                        "0",
		"baqup.HTTP.Middlewares.Middleware9.IPAllowList.SourceRange":                             "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware10.InFlightReq.Amount":                                 "42",
		"baqup.HTTP.Middlewares.Middleware10.InFlightReq.SourceCriterion.IPStrategy.Depth":       "42",
		"baqup.HTTP.Middlewares.Middleware10.InFlightReq.SourceCriterion.IPStrategy.ExcludedIPs": "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware10.InFlightReq.SourceCriterion.IPStrategy.IPv6Subnet":  "42",
		"baqup.HTTP.Middlewares.Middleware10.InFlightReq.SourceCriterion.RequestHeaderName":      "foobar",
		"baqup.HTTP.Middlewares.Middleware10.InFlightReq.SourceCriterion.RequestHost":            "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.NotAfter":                    "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.NotBefore":                   "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Sans":                        "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.SerialNumber":                "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Country":             "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Province":            "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Locality":            "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.Organization":        "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.OrganizationalUnit":  "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.CommonName":          "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.SerialNumber":        "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Subject.DomainComponent":     "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Issuer.Country":              "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Issuer.Province":             "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Issuer.Locality":             "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Issuer.Organization":         "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Issuer.CommonName":           "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Issuer.SerialNumber":         "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.Info.Issuer.DomainComponent":      "true",
		"baqup.HTTP.Middlewares.Middleware11.PassTLSClientCert.PEM":                              "true",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.Average":                                  "42",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.Period":                                   "1000000000",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.Burst":                                    "42",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.SourceCriterion.RequestHeaderName":        "foobar",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.SourceCriterion.RequestHost":              "true",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.SourceCriterion.IPStrategy.Depth":         "42",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.SourceCriterion.IPStrategy.ExcludedIPs":   "foobar, foobar",
		"baqup.HTTP.Middlewares.Middleware12.RateLimit.SourceCriterion.IPStrategy.IPv6Subnet":    "42",
		"baqup.HTTP.Middlewares.Middleware13.RedirectRegex.Regex":                                "foobar",
		"baqup.HTTP.Middlewares.Middleware13.RedirectRegex.Replacement":                          "foobar",
		"baqup.HTTP.Middlewares.Middleware13.RedirectRegex.Permanent":                            "true",
		"baqup.HTTP.Middlewares.Middleware13b.RedirectScheme.Scheme":                             "https",
		"baqup.HTTP.Middlewares.Middleware13b.RedirectScheme.Port":                               "80",
		"baqup.HTTP.Middlewares.Middleware13b.RedirectScheme.Permanent":                          "true",
		"baqup.HTTP.Middlewares.Middleware14.ReplacePath.Path":                                   "foobar",
		"baqup.HTTP.Middlewares.Middleware15.ReplacePathRegex.Regex":                             "foobar",
		"baqup.HTTP.Middlewares.Middleware15.ReplacePathRegex.Replacement":                       "foobar",
		"baqup.HTTP.Middlewares.Middleware16.Retry.Attempts":                                     "42",
		"baqup.HTTP.Middlewares.Middleware16.Retry.InitialInterval":                              "1000000000",
		"baqup.HTTP.Middlewares.Middleware17.StripPrefix.Prefixes":                               "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware17.StripPrefix.ForceSlash":                             "true",
		"baqup.HTTP.Middlewares.Middleware18.StripPrefixRegex.Regex":                             "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware19.Compress.Encodings":                                 "foobar, fiibar",
		"baqup.HTTP.Middlewares.Middleware19.Compress.MinResponseBodyBytes":                      "42",
		"baqup.HTTP.Middlewares.Middleware20.Plugin.tomato.aaa":                                  "foo1",
		"baqup.HTTP.Middlewares.Middleware20.Plugin.tomato.bbb":                                  "foo2",

		"baqup.HTTP.Routers.Router0.EntryPoints":              "foobar, fiibar",
		"baqup.HTTP.Routers.Router0.Middlewares":              "foobar, fiibar",
		"baqup.HTTP.Routers.Router0.Priority":                 "42",
		"baqup.HTTP.Routers.Router0.Rule":                     "foobar",
		"baqup.HTTP.Routers.Router0.Service":                  "foobar",
		"baqup.HTTP.Routers.Router0.TLS":                      "true",
		"baqup.HTTP.Routers.Router0.Observability.AccessLogs": "true",
		"baqup.HTTP.Routers.Router0.Observability.Tracing":    "true",
		"baqup.HTTP.Routers.Router0.Observability.Metrics":    "true",
		"baqup.HTTP.Routers.Router1.EntryPoints":              "foobar, fiibar",
		"baqup.HTTP.Routers.Router1.Middlewares":              "foobar, fiibar",
		"baqup.HTTP.Routers.Router1.Priority":                 "42",
		"baqup.HTTP.Routers.Router1.Rule":                     "foobar",
		"baqup.HTTP.Routers.Router1.Service":                  "foobar",
		"baqup.HTTP.Routers.Router1.Observability.AccessLogs": "true",
		"baqup.HTTP.Routers.Router1.Observability.Tracing":    "true",
		"baqup.HTTP.Routers.Router1.Observability.Metrics":    "true",

		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Headers.name0":        "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Headers.name1":        "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Hostname":             "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Interval":             "1000000000",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.UnhealthyInterval":    "1000000000",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Path":                 "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Method":               "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Status":               "401",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Port":                 "42",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Scheme":               "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.HealthCheck.Timeout":              "1000000000",
		"baqup.HTTP.Services.Service0.LoadBalancer.PassHostHeader":                   "true",
		"baqup.HTTP.Services.Service0.LoadBalancer.ResponseForwarding.FlushInterval": "1000000000",
		"baqup.HTTP.Services.Service0.LoadBalancer.Strategy":                         "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.server.URL":                       "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.server.PreservePath":              "true",
		"baqup.HTTP.Services.Service0.LoadBalancer.server.Port":                      "8080",
		"baqup.HTTP.Services.Service0.LoadBalancer.server.Scheme":                    "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.Sticky.Cookie.Name":               "foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.Sticky.Cookie.HTTPOnly":           "true",
		"baqup.HTTP.Services.Service0.LoadBalancer.Sticky.Cookie.Secure":             "false",
		"baqup.HTTP.Services.Service0.LoadBalancer.Sticky.Cookie.MaxAge":             "0",
		"baqup.HTTP.Services.Service0.LoadBalancer.Sticky.Cookie.Path":               "/foobar",
		"baqup.HTTP.Services.Service0.LoadBalancer.Sticky.Cookie.Domain":             "foo.com",
		"baqup.HTTP.Services.Service0.LoadBalancer.ServersTransport":                 "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Headers.name0":        "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Headers.name1":        "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Hostname":             "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Interval":             "1000000000",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.UnhealthyInterval":    "1000000000",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Path":                 "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Method":               "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Status":               "401",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Port":                 "42",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Scheme":               "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.HealthCheck.Timeout":              "1000000000",
		"baqup.HTTP.Services.Service1.LoadBalancer.PassHostHeader":                   "true",
		"baqup.HTTP.Services.Service1.LoadBalancer.ResponseForwarding.FlushInterval": "1000000000",
		"baqup.HTTP.Services.Service1.LoadBalancer.Strategy":                         "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.server.URL":                       "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.server.PreservePath":              "true",
		"baqup.HTTP.Services.Service1.LoadBalancer.server.Port":                      "8080",
		"baqup.HTTP.Services.Service1.LoadBalancer.server.Scheme":                    "foobar",
		"baqup.HTTP.Services.Service1.LoadBalancer.ServersTransport":                 "foobar",

		"baqup.TCP.Middlewares.Middleware0.IPAllowList.SourceRange": "foobar, fiibar",
		"baqup.TCP.Middlewares.Middleware2.InFlightConn.Amount":     "42",
		"baqup.TCP.Routers.Router0.Rule":                            "foobar",
		"baqup.TCP.Routers.Router0.Priority":                        "42",
		"baqup.TCP.Routers.Router0.EntryPoints":                     "foobar, fiibar",
		"baqup.TCP.Routers.Router0.Service":                         "foobar",
		"baqup.TCP.Routers.Router0.TLS.Passthrough":                 "false",
		"baqup.TCP.Routers.Router0.TLS.Options":                     "foo",
		"baqup.TCP.Routers.Router1.Rule":                            "foobar",
		"baqup.TCP.Routers.Router1.Priority":                        "42",
		"baqup.TCP.Routers.Router1.EntryPoints":                     "foobar, fiibar",
		"baqup.TCP.Routers.Router1.Service":                         "foobar",
		"baqup.TCP.Routers.Router1.TLS.Passthrough":                 "false",
		"baqup.TCP.Routers.Router1.TLS.Options":                     "foo",
		"baqup.TCP.Services.Service0.LoadBalancer.server.Port":      "42",
		"baqup.TCP.Services.Service0.LoadBalancer.server.TLS":       "false",
		"baqup.TCP.Services.Service0.LoadBalancer.ServersTransport": "foo",
		"baqup.TCP.Services.Service0.LoadBalancer.TerminationDelay": "42",
		"baqup.TCP.Services.Service1.LoadBalancer.server.Port":      "42",
		"baqup.TCP.Services.Service1.LoadBalancer.server.TLS":       "false",
		"baqup.TCP.Services.Service1.LoadBalancer.ServersTransport": "foo",
		"baqup.TCP.Services.Service1.LoadBalancer.TerminationDelay": "42",

		"baqup.TLS.Stores.default.DefaultGeneratedCert.Resolver":    "foobar",
		"baqup.TLS.Stores.default.DefaultGeneratedCert.Domain.Main": "foobar",
		"baqup.TLS.Stores.default.DefaultGeneratedCert.Domain.SANs": "foobar, fiibar",

		"baqup.UDP.Routers.Router0.EntryPoints":                "foobar, fiibar",
		"baqup.UDP.Routers.Router0.Service":                    "foobar",
		"baqup.UDP.Routers.Router1.EntryPoints":                "foobar, fiibar",
		"baqup.UDP.Routers.Router1.Service":                    "foobar",
		"baqup.UDP.Services.Service0.LoadBalancer.server.Port": "42",
		"baqup.UDP.Services.Service1.LoadBalancer.server.Port": "42",
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

func intPtr(value int) *int {
	return &value
}
