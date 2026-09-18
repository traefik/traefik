package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_buildHostRule(t *testing.T) {
	testCases := []struct {
		desc             string
		hostnames        []gatev1.Hostname
		expectedRule     string
		expectedPriority int
		expectErr        bool
	}{
		{
			desc:         "Empty (should not happen)",
			expectedRule: "",
		},
		{
			desc: "One Host",
			hostnames: []gatev1.Hostname{
				"Foo",
			},
			expectedRule:     `Host("Foo")`,
			expectedPriority: 3,
		},
		{
			desc: "Multiple Hosts",
			hostnames: []gatev1.Hostname{
				"Foo",
				"Bar",
				"Bir",
			},
			expectedRule:     `(Host("Foo") || Host("Bar") || Host("Bir"))`,
			expectedPriority: 3,
		},
		{
			desc: "Several Host and wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
				"bar.foo",
				"foo.foo",
			},
			expectedRule:     `(HostRegexp("^[a-z0-9-\\.]+\\.bar\\.foo$") || Host("bar.foo") || Host("foo.foo"))`,
			expectedPriority: 9,
		},
		{
			desc: "Host with wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
			},
			expectedRule:     `HostRegexp("^[a-z0-9-\\.]+\\.bar\\.foo$")`,
			expectedPriority: 9,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, priority := buildHostRule(test.hostnames)
			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPriority, priority)
		})
	}
}

func Test_buildMatchRule(t *testing.T) {
	testCases := []struct {
		desc             string
		match            gatev1.HTTPRouteMatch
		hostnames        []gatev1.Hostname
		expectedRule     string
		expectedPriority int
		expectedError    bool
	}{
		{
			desc:             "Empty rule and matches",
			expectedRule:     `PathPrefix("/")`,
			expectedPriority: 1,
		},
		{
			desc:             "One Host rule without match",
			hostnames:        []gatev1.Hostname{"foo.com"},
			expectedRule:     `Host("foo.com") && PathPrefix("/")`,
			expectedPriority: 8,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch",
			match: gatev1.HTTPRouteMatch{
				Path: new(gatev1.HTTPPathMatch{
					Type:  new(gatev1.PathMatchPathPrefix),
					Value: new("/"),
				}),
				Headers: nil,
			},
			expectedRule:     `PathPrefix("/")`,
			expectedPriority: 1,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch Type",
			match: gatev1.HTTPRouteMatch{
				Path: new(gatev1.HTTPPathMatch{
					Type:  new(gatev1.PathMatchPathPrefix),
					Value: new("/"),
				}),
				Headers: []gatev1.HTTPHeaderMatch{
					{Name: "foo", Value: "bar"},
				},
			},
			expectedRule:     `PathPrefix("/") && Header("foo","bar")`,
			expectedPriority: 101,
		},
		{
			desc:             "One HTTPRouteMatch with nil HTTPPathMatch",
			match:            gatev1.HTTPRouteMatch{Path: nil},
			expectedRule:     `PathPrefix("/")`,
			expectedPriority: 1,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Type",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  nil,
					Value: new("/foo/"),
				},
			},
			expectedRule:     `(Path("/foo") || PathPrefix("/foo/"))`,
			expectedPriority: 10500,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Values",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  new(gatev1.PathMatchExact),
					Value: nil,
				},
			},
			expectedRule:     `Path("/")`,
			expectedPriority: 100000,
		},
		{
			desc: "One Path",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  new(gatev1.PathMatchExact),
					Value: new("/foo/"),
				},
			},
			expectedRule:     `Path("/foo/")`,
			expectedPriority: 100000,
		},
		{
			desc: "Path && Header",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  new(gatev1.PathMatchExact),
					Value: new("/foo/"),
				},
				Headers: []gatev1.HTTPHeaderMatch{
					{
						Type:  new(gatev1.HeaderMatchExact),
						Name:  "my-header",
						Value: "foo",
					},
				},
			},
			expectedRule:     `Path("/foo/") && Header("my-header","foo")`,
			expectedPriority: 100100,
		},
		{
			desc:      "Host && Path && Header",
			hostnames: []gatev1.Hostname{"foo.com"},
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  new(gatev1.PathMatchExact),
					Value: new("/foo/"),
				},
				Headers: []gatev1.HTTPHeaderMatch{
					{
						Type:  new(gatev1.HeaderMatchExact),
						Name:  "my-header",
						Value: "foo",
					},
				},
			},
			expectedRule:     `Host("foo.com") && Path("/foo/") && Header("my-header","foo")`,
			expectedPriority: 100107,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, priority := buildMatchRule(test.hostnames, test.match)
			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPriority, priority)
		})
	}
}

func TestGetHTTPServiceProtocol(t *testing.T) {
	testCases := []struct {
		desc        string
		port        corev1.ServicePort
		expected    string
		expectedErr bool
	}{
		{
			desc:     "TCP port without appProtocol",
			port:     corev1.ServicePort{Name: "web", Protocol: corev1.ProtocolTCP, Port: 80},
			expected: "http",
		},
		{
			desc:     "Port 443 without appProtocol serves plain HTTP",
			port:     corev1.ServicePort{Name: "web", Protocol: corev1.ProtocolTCP, Port: 443},
			expected: "http",
		},
		{
			desc:     "HTTPS named port without appProtocol serves plain HTTP",
			port:     corev1.ServicePort{Name: "https", Protocol: corev1.ProtocolTCP, Port: 8443},
			expected: "http",
		},
		{
			desc:     "HTTP appProtocol",
			port:     corev1.ServicePort{Name: "web", Protocol: corev1.ProtocolTCP, Port: 443, AppProtocol: new("http")},
			expected: "http",
		},
		{
			desc:     "HTTPS appProtocol",
			port:     corev1.ServicePort{Name: "web", Protocol: corev1.ProtocolTCP, Port: 443, AppProtocol: new("https")},
			expected: "https",
		},
		{
			desc:     "H2C appProtocol",
			port:     corev1.ServicePort{Name: "web", Protocol: corev1.ProtocolTCP, Port: 80, AppProtocol: new("kubernetes.io/h2c")},
			expected: "h2c",
		},
		{
			desc:        "Non TCP protocol",
			port:        corev1.ServicePort{Name: "web", Protocol: corev1.ProtocolUDP, Port: 80},
			expectedErr: true,
		},
		{
			desc:        "Unsupported appProtocol",
			port:        corev1.ServicePort{Name: "web", Protocol: corev1.ProtocolTCP, Port: 80, AppProtocol: new("gopher")},
			expectedErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			protocol, err := getHTTPServiceProtocol(test.port)
			if test.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, protocol)
		})
	}
}

func TestLoadHTTPRoutes_plainHTTPServiceOnPort443(t *testing.T) {
	k8sObjects, gwObjects := readResources(t, []string{"httproute/with_plain_http_service_on_port_443.yml"})

	kubeClient := kubefake.NewClientset(k8sObjects...)
	gwClient := newGatewaySimpleClientSet(t, gwObjects...)

	client := newClientImpl(kubeClient, gwClient)

	eventCh, err := client.WatchAll(nil, make(chan struct{}))
	require.NoError(t, err)

	if len(k8sObjects) > 0 || len(gwObjects) > 0 {
		// just wait for the first event
		<-eventCh
	}

	p := Provider{
		EntryPoints: map[string]Entrypoint{"web": {Address: ":80"}},
		client:      client,
	}

	conf := p.loadConfigurationFromGateways(t.Context())

	var urls []string
	for _, svc := range conf.HTTP.Services {
		if svc.LoadBalancer == nil {
			continue
		}
		for _, server := range svc.LoadBalancer.Servers {
			urls = append(urls, server.URL)
		}
	}

	assert.ElementsMatch(t, []string{"http://10.10.0.20:8080", "http://10.10.0.21:8080"}, urls)
}
