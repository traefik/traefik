package file

import (
	"testing"

	"github.com/containous/traefik/v2/pkg/config/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getRootFieldNames(t *testing.T) {
	testCases := []struct {
		desc     string
		element  interface{}
		expected []string
	}{
		{
			desc:     "simple fields",
			element:  &Yo{},
			expected: []string{"Foo", "Fii", "Fuu", "Yi"},
		},
		{
			desc:     "embedded struct",
			element:  &Yu{},
			expected: []string{"Foo", "Fii", "Fuu"},
		},
		{
			desc:     "embedded struct pointer",
			element:  &Ye{},
			expected: []string{"Foo", "Fii", "Fuu"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			names := getRootFieldNames(test.element)

			assert.Equal(t, test.expected, names)
		})
	}
}

func Test_decodeFileToNode_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		confFile string
	}{
		{
			desc:     "non existing file",
			confFile: "./fixtures/not_existing.toml",
		},
		{
			desc:     "file without content",
			confFile: "./fixtures/empty.toml",
		},
		{
			desc:     "file without any valid configuration",
			confFile: "./fixtures/no_conf.toml",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			node, err := decodeFileToNode(test.confFile,
				"Global", "ServersTransport", "EntryPoints", "Providers", "API", "Metrics", "Ping", "Log", "AccessLog", "Tracing", "HostResolver", "CertificatesResolvers")

			require.Error(t, err)
			assert.Nil(t, node)
		})
	}
}

func Test_decodeFileToNode_compare(t *testing.T) {
	nodeToml, err := decodeFileToNode("./fixtures/sample.toml",
		"Global", "ServersTransport", "EntryPoints", "Providers", "API", "Metrics", "Ping", "Log", "AccessLog", "Tracing", "HostResolver", "CertificatesResolvers")
	require.NoError(t, err)

	nodeYaml, err := decodeFileToNode("./fixtures/sample.yml")
	require.NoError(t, err)

	assert.Equal(t, nodeToml, nodeYaml)
}

func Test_decodeFileToNode_Toml(t *testing.T) {
	node, err := decodeFileToNode("./fixtures/sample.toml",
		"Global", "ServersTransport", "EntryPoints", "Providers", "API", "Metrics", "Ping", "Log", "AccessLog", "Tracing", "HostResolver", "CertificatesResolvers")
	require.NoError(t, err)

	expected := &parser.Node{
		Name: "traefik",
		Children: []*parser.Node{
			{Name: "accessLog", Children: []*parser.Node{
				{Name: "bufferingSize", Value: "42"},
				{Name: "fields", Children: []*parser.Node{
					{Name: "defaultMode", Value: "foobar"},
					{Name: "headers", Children: []*parser.Node{
						{Name: "defaultMode", Value: "foobar"},
						{Name: "names", Children: []*parser.Node{
							{Name: "name0", Value: "foobar"},
							{Name: "name1", Value: "foobar"}}}}},
					{Name: "names", Children: []*parser.Node{
						{Name: "name0", Value: "foobar"},
						{Name: "name1", Value: "foobar"}}}}},
				{Name: "filePath", Value: "foobar"},
				{Name: "filters", Children: []*parser.Node{
					{Name: "minDuration", Value: "42"},
					{Name: "retryAttempts", Value: "true"},
					{Name: "statusCodes", Value: "foobar,foobar"}}},
				{Name: "format", Value: "foobar"}}},
			{Name: "api", Children: []*parser.Node{
				{Name: "dashboard", Value: "true"},
				{Name: "entryPoint", Value: "foobar"},
				{Name: "middlewares", Value: "foobar,foobar"},
				{Name: "statistics", Children: []*parser.Node{
					{Name: "recentErrors", Value: "42"}}}}},
			{Name: "certificatesResolvers", Children: []*parser.Node{
				{Name: "default", Children: []*parser.Node{
					{Name: "acme",
						Children: []*parser.Node{
							{Name: "acmeLogging", Value: "true"},
							{Name: "caServer", Value: "foobar"},
							{Name: "dnsChallenge", Children: []*parser.Node{
								{Name: "delayBeforeCheck", Value: "42"},
								{Name: "disablePropagationCheck", Value: "true"},
								{Name: "provider", Value: "foobar"},
								{Name: "resolvers", Value: "foobar,foobar"},
							}},
							{Name: "email", Value: "foobar"},
							{Name: "entryPoint", Value: "foobar"},
							{Name: "httpChallenge", Children: []*parser.Node{
								{Name: "entryPoint", Value: "foobar"}}},
							{Name: "keyType", Value: "foobar"},
							{Name: "storage", Value: "foobar"},
							{Name: "tlsChallenge"},
						},
					},
				}},
			}},
			{Name: "entryPoints", Children: []*parser.Node{
				{Name: "EntryPoint0", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "forwardedHeaders", Children: []*parser.Node{
						{Name: "insecure", Value: "true"},
						{Name: "trustedIPs", Value: "foobar,foobar"}}},
					{Name: "proxyProtocol", Children: []*parser.Node{
						{Name: "insecure", Value: "true"},
						{Name: "trustedIPs", Value: "foobar,foobar"}}},
					{Name: "transport", Children: []*parser.Node{
						{Name: "lifeCycle", Children: []*parser.Node{
							{Name: "graceTimeOut", Value: "42"},
							{Name: "requestAcceptGraceTimeout", Value: "42"}}},
						{Name: "respondingTimeouts", Children: []*parser.Node{
							{Name: "idleTimeout", Value: "42"},
							{Name: "readTimeout", Value: "42"},
							{Name: "writeTimeout", Value: "42"}}}}}}}}},
			{Name: "global", Children: []*parser.Node{
				{Name: "checkNewVersion", Value: "true"},
				{Name: "sendAnonymousUsage", Value: "true"}}},
			{Name: "hostResolver", Children: []*parser.Node{
				{Name: "cnameFlattening", Value: "true"},
				{Name: "resolvConfig", Value: "foobar"},
				{Name: "resolvDepth", Value: "42"}}},
			{Name: "log", Children: []*parser.Node{
				{Name: "filePath", Value: "foobar"},
				{Name: "format", Value: "foobar"},
				{Name: "level", Value: "foobar"}}},
			{Name: "metrics", Children: []*parser.Node{
				{Name: "datadog", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "pushInterval", Value: "10s"}}},
				{Name: "influxDB", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "database", Value: "foobar"},
					{Name: "password", Value: "foobar"},
					{Name: "protocol", Value: "foobar"},
					{Name: "pushInterval", Value: "10s"},
					{Name: "retentionPolicy", Value: "foobar"},
					{Name: "username", Value: "foobar"}}},
				{Name: "prometheus", Children: []*parser.Node{
					{Name: "buckets", Value: "42,42"},
					{Name: "entryPoint", Value: "foobar"},
					{Name: "middlewares", Value: "foobar,foobar"}}},
				{Name: "statsD", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "pushInterval", Value: "10s"}}}}},
			{Name: "ping", Children: []*parser.Node{
				{Name: "entryPoint", Value: "foobar"},
				{Name: "middlewares", Value: "foobar,foobar"}}},
			{Name: "providers", Children: []*parser.Node{
				{Name: "docker", Children: []*parser.Node{
					{Name: "constraints", Value: "foobar"},
					{Name: "defaultRule", Value: "foobar"},
					{Name: "endpoint", Value: "foobar"},
					{Name: "exposedByDefault", Value: "true"},
					{Name: "network", Value: "foobar"},
					{Name: "swarmMode", Value: "true"},
					{Name: "swarmModeRefreshSeconds", Value: "42"},
					{Name: "tls", Children: []*parser.Node{
						{Name: "ca", Value: "foobar"},
						{Name: "caOptional", Value: "true"},
						{Name: "cert", Value: "foobar"},
						{Name: "insecureSkipVerify", Value: "true"},
						{Name: "key", Value: "foobar"}}},
					{Name: "useBindPortIP", Value: "true"},
					{Name: "watch", Value: "true"}}},
				{Name: "file", Children: []*parser.Node{
					{Name: "debugLogGeneratedTemplate", Value: "true"},
					{Name: "directory", Value: "foobar"},
					{Name: "filename", Value: "foobar"},
					{Name: "watch", Value: "true"}}},
				{Name: "kubernetesCRD",
					Children: []*parser.Node{
						{Name: "certAuthFilePath", Value: "foobar"},
						{Name: "disablePassHostHeaders", Value: "true"},
						{Name: "endpoint", Value: "foobar"},
						{Name: "ingressClass", Value: "foobar"},
						{Name: "labelSelector", Value: "foobar"},
						{Name: "namespaces", Value: "foobar,foobar"},
						{Name: "token", Value: "foobar"}}},
				{Name: "kubernetesIngress", Children: []*parser.Node{
					{Name: "certAuthFilePath", Value: "foobar"},
					{Name: "disablePassHostHeaders", Value: "true"},
					{Name: "endpoint", Value: "foobar"},
					{Name: "ingressClass", Value: "foobar"},
					{Name: "ingressEndpoint", Children: []*parser.Node{
						{Name: "hostname", Value: "foobar"},
						{Name: "ip", Value: "foobar"},
						{Name: "publishedService", Value: "foobar"}}},
					{Name: "labelSelector", Value: "foobar"},
					{Name: "namespaces", Value: "foobar,foobar"},
					{Name: "token", Value: "foobar"}}},
				{Name: "marathon", Children: []*parser.Node{
					{Name: "basic", Children: []*parser.Node{
						{Name: "httpBasicAuthUser", Value: "foobar"},
						{Name: "httpBasicPassword", Value: "foobar"}}},
					{Name: "constraints", Value: "foobar"},
					{Name: "dcosToken", Value: "foobar"},
					{Name: "defaultRule", Value: "foobar"},
					{Name: "dialerTimeout", Value: "42"},
					{Name: "endpoint", Value: "foobar"},
					{Name: "exposedByDefault", Value: "true"},
					{Name: "forceTaskHostname", Value: "true"},
					{Name: "keepAlive", Value: "42"},
					{Name: "respectReadinessChecks", Value: "true"},
					{Name: "responseHeaderTimeout", Value: "42"},
					{Name: "tls", Children: []*parser.Node{
						{Name: "ca", Value: "foobar"},
						{Name: "caOptional", Value: "true"},
						{Name: "cert", Value: "foobar"},
						{Name: "insecureSkipVerify", Value: "true"},
						{Name: "key", Value: "foobar"}}},
					{Name: "tlsHandshakeTimeout", Value: "42"},
					{Name: "trace", Value: "true"},
					{Name: "watch", Value: "true"}}},
				{Name: "providersThrottleDuration", Value: "42"},
				{Name: "rancher", Children: []*parser.Node{
					{Name: "constraints", Value: "foobar"},
					{Name: "defaultRule", Value: "foobar"},
					{Name: "enableServiceHealthFilter", Value: "true"},
					{Name: "exposedByDefault", Value: "true"},
					{Name: "intervalPoll", Value: "true"},
					{Name: "prefix", Value: "foobar"},
					{Name: "refreshSeconds", Value: "42"},
					{Name: "watch", Value: "true"}}},
				{Name: "rest", Children: []*parser.Node{
					{Name: "entryPoint", Value: "foobar"}}}}},
			{Name: "serversTransport", Children: []*parser.Node{
				{Name: "forwardingTimeouts", Children: []*parser.Node{
					{Name: "dialTimeout", Value: "42"},
					{Name: "idleConnTimeout", Value: "42"},
					{Name: "responseHeaderTimeout", Value: "42"}}},
				{Name: "insecureSkipVerify", Value: "true"},
				{Name: "maxIdleConnsPerHost", Value: "42"},
				{Name: "rootCAs", Value: "foobar,foobar"}}},
			{Name: "tracing", Children: []*parser.Node{
				{Name: "datadog", Children: []*parser.Node{
					{Name: "bagagePrefixHeaderName", Value: "foobar"},
					{Name: "debug", Value: "true"},
					{Name: "globalTag", Value: "foobar"},
					{Name: "localAgentHostPort", Value: "foobar"},
					{Name: "parentIDHeaderName", Value: "foobar"},
					{Name: "prioritySampling", Value: "true"},
					{Name: "samplingPriorityHeaderName", Value: "foobar"},
					{Name: "traceIDHeaderName", Value: "foobar"}}},
				{Name: "haystack", Children: []*parser.Node{
					{Name: "globalTag", Value: "foobar"},
					{Name: "localAgentHost", Value: "foobar"},
					{Name: "localAgentPort", Value: "42"},
					{Name: "parentIDHeaderName", Value: "foobar"},
					{Name: "spanIDHeaderName", Value: "foobar"},
					{Name: "traceIDHeaderName", Value: "foobar"}}},
				{Name: "instana", Children: []*parser.Node{
					{Name: "localAgentHost", Value: "foobar"},
					{Name: "localAgentPort", Value: "42"},
					{Name: "logLevel", Value: "foobar"}}},
				{Name: "jaeger", Children: []*parser.Node{
					{Name: "gen128Bit", Value: "true"},
					{Name: "localAgentHostPort", Value: "foobar"},
					{Name: "propagation", Value: "foobar"},
					{Name: "samplingParam", Value: "42"},
					{Name: "samplingServerURL", Value: "foobar"},
					{Name: "samplingType", Value: "foobar"},
					{Name: "traceContextHeaderName", Value: "foobar"}}},
				{Name: "serviceName", Value: "foobar"},
				{Name: "spanNameLimit", Value: "42"},
				{Name: "zipkin", Children: []*parser.Node{
					{Name: "httpEndpoint", Value: "foobar"},
					{Name: "id128Bit", Value: "true"},
					{Name: "sameSpan", Value: "true"},
					{Name: "sampleRate", Value: "42"}}}}},
		},
	}

	assert.Equal(t, expected, node)
}

func Test_decodeFileToNode_Yaml(t *testing.T) {
	node, err := decodeFileToNode("./fixtures/sample.yml")
	require.NoError(t, err)

	expected := &parser.Node{
		Name: "traefik",
		Children: []*parser.Node{
			{Name: "accessLog", Children: []*parser.Node{
				{Name: "bufferingSize", Value: "42"},
				{Name: "fields", Children: []*parser.Node{
					{Name: "defaultMode", Value: "foobar"},
					{Name: "headers", Children: []*parser.Node{
						{Name: "defaultMode", Value: "foobar"},
						{Name: "names", Children: []*parser.Node{
							{Name: "name0", Value: "foobar"},
							{Name: "name1", Value: "foobar"}}}}},
					{Name: "names", Children: []*parser.Node{
						{Name: "name0", Value: "foobar"},
						{Name: "name1", Value: "foobar"}}}}},
				{Name: "filePath", Value: "foobar"},
				{Name: "filters", Children: []*parser.Node{
					{Name: "minDuration", Value: "42"},
					{Name: "retryAttempts", Value: "true"},
					{Name: "statusCodes", Value: "foobar,foobar"}}},
				{Name: "format", Value: "foobar"}}},
			{Name: "api", Children: []*parser.Node{
				{Name: "dashboard", Value: "true"},
				{Name: "entryPoint", Value: "foobar"},
				{Name: "middlewares", Value: "foobar,foobar"},
				{Name: "statistics", Children: []*parser.Node{
					{Name: "recentErrors", Value: "42"}}}}},
			{Name: "certificatesResolvers", Children: []*parser.Node{
				{Name: "default", Children: []*parser.Node{
					{Name: "acme",
						Children: []*parser.Node{
							{Name: "acmeLogging", Value: "true"},
							{Name: "caServer", Value: "foobar"},
							{Name: "dnsChallenge", Children: []*parser.Node{
								{Name: "delayBeforeCheck", Value: "42"},
								{Name: "disablePropagationCheck", Value: "true"},
								{Name: "provider", Value: "foobar"},
								{Name: "resolvers", Value: "foobar,foobar"},
							}},
							{Name: "email", Value: "foobar"},
							{Name: "entryPoint", Value: "foobar"},
							{Name: "httpChallenge", Children: []*parser.Node{
								{Name: "entryPoint", Value: "foobar"}}},
							{Name: "keyType", Value: "foobar"},
							{Name: "storage", Value: "foobar"},
							{Name: "tlsChallenge"},
						},
					},
				}},
			}},
			{Name: "entryPoints", Children: []*parser.Node{
				{Name: "EntryPoint0", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "forwardedHeaders", Children: []*parser.Node{
						{Name: "insecure", Value: "true"},
						{Name: "trustedIPs", Value: "foobar,foobar"}}},
					{Name: "proxyProtocol", Children: []*parser.Node{
						{Name: "insecure", Value: "true"},
						{Name: "trustedIPs", Value: "foobar,foobar"}}},
					{Name: "transport", Children: []*parser.Node{
						{Name: "lifeCycle", Children: []*parser.Node{
							{Name: "graceTimeOut", Value: "42"},
							{Name: "requestAcceptGraceTimeout", Value: "42"}}},
						{Name: "respondingTimeouts", Children: []*parser.Node{
							{Name: "idleTimeout", Value: "42"},
							{Name: "readTimeout", Value: "42"},
							{Name: "writeTimeout", Value: "42"}}}}}}}}},
			{Name: "global", Children: []*parser.Node{
				{Name: "checkNewVersion", Value: "true"},
				{Name: "sendAnonymousUsage", Value: "true"}}},
			{Name: "hostResolver", Children: []*parser.Node{
				{Name: "cnameFlattening", Value: "true"},
				{Name: "resolvConfig", Value: "foobar"},
				{Name: "resolvDepth", Value: "42"}}},
			{Name: "log", Children: []*parser.Node{
				{Name: "filePath", Value: "foobar"},
				{Name: "format", Value: "foobar"},
				{Name: "level", Value: "foobar"}}},
			{Name: "metrics", Children: []*parser.Node{
				{Name: "datadog", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "pushInterval", Value: "10s"}}},
				{Name: "influxDB", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "database", Value: "foobar"},
					{Name: "password", Value: "foobar"},
					{Name: "protocol", Value: "foobar"},
					{Name: "pushInterval", Value: "10s"},
					{Name: "retentionPolicy", Value: "foobar"},
					{Name: "username", Value: "foobar"}}},
				{Name: "prometheus", Children: []*parser.Node{
					{Name: "buckets", Value: "42,42"},
					{Name: "entryPoint", Value: "foobar"},
					{Name: "middlewares", Value: "foobar,foobar"}}},
				{Name: "statsD", Children: []*parser.Node{
					{Name: "address", Value: "foobar"},
					{Name: "pushInterval", Value: "10s"}}}}},
			{Name: "ping", Children: []*parser.Node{
				{Name: "entryPoint", Value: "foobar"},
				{Name: "middlewares", Value: "foobar,foobar"}}},
			{Name: "providers", Children: []*parser.Node{
				{Name: "docker", Children: []*parser.Node{
					{Name: "constraints", Value: "foobar"},
					{Name: "defaultRule", Value: "foobar"},
					{Name: "endpoint", Value: "foobar"},
					{Name: "exposedByDefault", Value: "true"},
					{Name: "network", Value: "foobar"},
					{Name: "swarmMode", Value: "true"},
					{Name: "swarmModeRefreshSeconds", Value: "42"},
					{Name: "tls", Children: []*parser.Node{
						{Name: "ca", Value: "foobar"},
						{Name: "caOptional", Value: "true"},
						{Name: "cert", Value: "foobar"},
						{Name: "insecureSkipVerify", Value: "true"},
						{Name: "key", Value: "foobar"}}},
					{Name: "useBindPortIP", Value: "true"},
					{Name: "watch", Value: "true"}}},
				{Name: "file", Children: []*parser.Node{
					{Name: "debugLogGeneratedTemplate", Value: "true"},
					{Name: "directory", Value: "foobar"},
					{Name: "filename", Value: "foobar"},
					{Name: "watch", Value: "true"}}},
				{Name: "kubernetesCRD",
					Children: []*parser.Node{
						{Name: "certAuthFilePath", Value: "foobar"},
						{Name: "disablePassHostHeaders", Value: "true"},
						{Name: "endpoint", Value: "foobar"},
						{Name: "ingressClass", Value: "foobar"},
						{Name: "labelSelector", Value: "foobar"},
						{Name: "namespaces", Value: "foobar,foobar"},
						{Name: "token", Value: "foobar"}}},
				{Name: "kubernetesIngress", Children: []*parser.Node{
					{Name: "certAuthFilePath", Value: "foobar"},
					{Name: "disablePassHostHeaders", Value: "true"},
					{Name: "endpoint", Value: "foobar"},
					{Name: "ingressClass", Value: "foobar"},
					{Name: "ingressEndpoint", Children: []*parser.Node{
						{Name: "hostname", Value: "foobar"},
						{Name: "ip", Value: "foobar"},
						{Name: "publishedService", Value: "foobar"}}},
					{Name: "labelSelector", Value: "foobar"},
					{Name: "namespaces", Value: "foobar,foobar"},
					{Name: "token", Value: "foobar"}}},
				{Name: "marathon", Children: []*parser.Node{
					{Name: "basic", Children: []*parser.Node{
						{Name: "httpBasicAuthUser", Value: "foobar"},
						{Name: "httpBasicPassword", Value: "foobar"}}},
					{Name: "constraints", Value: "foobar"},
					{Name: "dcosToken", Value: "foobar"},
					{Name: "defaultRule", Value: "foobar"},
					{Name: "dialerTimeout", Value: "42"},
					{Name: "endpoint", Value: "foobar"},
					{Name: "exposedByDefault", Value: "true"},
					{Name: "forceTaskHostname", Value: "true"},
					{Name: "keepAlive", Value: "42"},
					{Name: "respectReadinessChecks", Value: "true"},
					{Name: "responseHeaderTimeout", Value: "42"},
					{Name: "tls", Children: []*parser.Node{
						{Name: "ca", Value: "foobar"},
						{Name: "caOptional", Value: "true"},
						{Name: "cert", Value: "foobar"},
						{Name: "insecureSkipVerify", Value: "true"},
						{Name: "key", Value: "foobar"}}},
					{Name: "tlsHandshakeTimeout", Value: "42"},
					{Name: "trace", Value: "true"},
					{Name: "watch", Value: "true"}}},
				{Name: "providersThrottleDuration", Value: "42"},
				{Name: "rancher", Children: []*parser.Node{
					{Name: "constraints", Value: "foobar"},
					{Name: "defaultRule", Value: "foobar"},
					{Name: "enableServiceHealthFilter", Value: "true"},
					{Name: "exposedByDefault", Value: "true"},
					{Name: "intervalPoll", Value: "true"},
					{Name: "prefix", Value: "foobar"},
					{Name: "refreshSeconds", Value: "42"},
					{Name: "watch", Value: "true"}}},
				{Name: "rest", Children: []*parser.Node{
					{Name: "entryPoint", Value: "foobar"}}}}},
			{Name: "serversTransport", Children: []*parser.Node{
				{Name: "forwardingTimeouts", Children: []*parser.Node{
					{Name: "dialTimeout", Value: "42"},
					{Name: "idleConnTimeout", Value: "42"},
					{Name: "responseHeaderTimeout", Value: "42"}}},
				{Name: "insecureSkipVerify", Value: "true"},
				{Name: "maxIdleConnsPerHost", Value: "42"},
				{Name: "rootCAs", Value: "foobar,foobar"}}},
			{Name: "tracing", Children: []*parser.Node{
				{Name: "datadog", Children: []*parser.Node{
					{Name: "bagagePrefixHeaderName", Value: "foobar"},
					{Name: "debug", Value: "true"},
					{Name: "globalTag", Value: "foobar"},
					{Name: "localAgentHostPort", Value: "foobar"},
					{Name: "parentIDHeaderName", Value: "foobar"},
					{Name: "prioritySampling", Value: "true"},
					{Name: "samplingPriorityHeaderName", Value: "foobar"},
					{Name: "traceIDHeaderName", Value: "foobar"}}},
				{Name: "haystack", Children: []*parser.Node{
					{Name: "globalTag", Value: "foobar"},
					{Name: "localAgentHost", Value: "foobar"},
					{Name: "localAgentPort", Value: "42"},
					{Name: "parentIDHeaderName", Value: "foobar"},
					{Name: "spanIDHeaderName", Value: "foobar"},
					{Name: "traceIDHeaderName", Value: "foobar"}}},
				{Name: "instana", Children: []*parser.Node{
					{Name: "localAgentHost", Value: "foobar"},
					{Name: "localAgentPort", Value: "42"},
					{Name: "logLevel", Value: "foobar"}}},
				{Name: "jaeger", Children: []*parser.Node{
					{Name: "gen128Bit", Value: "true"},
					{Name: "localAgentHostPort", Value: "foobar"},
					{Name: "propagation", Value: "foobar"},
					{Name: "samplingParam", Value: "42"},
					{Name: "samplingServerURL", Value: "foobar"},
					{Name: "samplingType", Value: "foobar"},
					{Name: "traceContextHeaderName", Value: "foobar"}}},
				{Name: "serviceName", Value: "foobar"},
				{Name: "spanNameLimit", Value: "42"},
				{Name: "zipkin", Children: []*parser.Node{
					{Name: "httpEndpoint", Value: "foobar"},
					{Name: "id128Bit", Value: "true"},
					{Name: "sameSpan", Value: "true"},
					{Name: "sampleRate", Value: "42"}}}}},
		},
	}

	assert.Equal(t, expected, node)
}
