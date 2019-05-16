package file

import (
	"testing"

	"github.com/containous/traefik/pkg/config/parser"
	"github.com/stretchr/testify/assert"
)

func Test_decodeFileToNode_compare(t *testing.T) {
	nodeToml, err := decodeFileToNode("./fixtures/sample.toml", "http", "tcp", "tls", "TLSOptions", "TLSStores")
	if err != nil {
		t.Fatal(err)
	}

	nodeYaml, err := decodeFileToNode("./fixtures/sample.yml")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, nodeToml, nodeYaml)
}

func Test_decodeFileToNode_Toml(t *testing.T) {
	node, err := decodeFileToNode("./fixtures/sample.toml", "http", "tcp", "tls", "TLSOptions", "TLSStores")
	if err != nil {
		t.Fatal(err)
	}

	expected := &parser.Node{
		Name: "traefik",
		Children: []*parser.Node{
			{Name: "ACME",
				Children: []*parser.Node{
					{Name: "ACMELogging", Value: "true"},
					{Name: "CAServer", Value: "foobar"},
					{Name: "DNSChallenge", Children: []*parser.Node{
						{Name: "DelayBeforeCheck", Value: "42"},
						{Name: "DisablePropagationCheck", Value: "true"},
						{Name: "Provider", Value: "foobar"},
						{Name: "Resolvers", Value: "foobar,foobar"},
					}},
					{Name: "Domains", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Main", Value: "foobar"},
							{Name: "SANs", Value: "foobar,foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Main", Value: "foobar"},
							{Name: "SANs", Value: "foobar,foobar"},
						}},
					}},
					{Name: "Email", Value: "foobar"},
					{Name: "EntryPoint", Value: "foobar"},
					{Name: "HTTPChallenge", Children: []*parser.Node{
						{Name: "EntryPoint", Value: "foobar"}}},
					{Name: "KeyType", Value: "foobar"},
					{Name: "OnHostRule", Value: "true"},
					{Name: "Storage", Value: "foobar"},
					{Name: "TLSChallenge"},
				},
			},
			{Name: "API", Children: []*parser.Node{
				{Name: "Dashboard", Value: "true"},
				{Name: "EntryPoint", Value: "foobar"},
				{Name: "Middlewares", Value: "foobar,foobar"},
				{Name: "Statistics", Children: []*parser.Node{
					{Name: "RecentErrors", Value: "42"}}}}},
			{Name: "AccessLog", Children: []*parser.Node{
				{Name: "BufferingSize", Value: "42"},
				{Name: "Fields", Children: []*parser.Node{
					{Name: "DefaultMode", Value: "foobar"},
					{Name: "Headers", Children: []*parser.Node{
						{Name: "DefaultMode", Value: "foobar"},
						{Name: "Names", Children: []*parser.Node{
							{Name: "name0", Value: "foobar"},
							{Name: "name1", Value: "foobar"}}}}},
					{Name: "Names", Children: []*parser.Node{
						{Name: "name0", Value: "foobar"},
						{Name: "name1", Value: "foobar"}}}}},
				{Name: "FilePath", Value: "foobar"},
				{Name: "Filters", Children: []*parser.Node{
					{Name: "MinDuration", Value: "42"},
					{Name: "RetryAttempts", Value: "true"},
					{Name: "StatusCodes", Value: "foobar,foobar"}}},
				{Name: "Format", Value: "foobar"}}},
			{Name: "EntryPoints", Children: []*parser.Node{
				{Name: "EntryPoint0", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "ForwardedHeaders", Children: []*parser.Node{
						{Name: "Insecure", Value: "true"},
						{Name: "TrustedIPs", Value: "foobar,foobar"}}},
					{Name: "ProxyProtocol", Children: []*parser.Node{
						{Name: "Insecure", Value: "true"},
						{Name: "TrustedIPs", Value: "foobar,foobar"}}},
					{Name: "Transport", Children: []*parser.Node{
						{Name: "LifeCycle", Children: []*parser.Node{
							{Name: "GraceTimeOut", Value: "42"},
							{Name: "RequestAcceptGraceTimeout", Value: "42"}}},
						{Name: "RespondingTimeouts", Children: []*parser.Node{
							{Name: "IdleTimeout", Value: "42"},
							{Name: "ReadTimeout", Value: "42"},
							{Name: "WriteTimeout", Value: "42"}}}}}}}}},
			{Name: "Global", Children: []*parser.Node{
				{Name: "CheckNewVersion", Value: "true"},
				{Name: "Debug", Value: "true"},
				{Name: "SendAnonymousUsage", Value: "true"}}},
			{Name: "HostResolver", Children: []*parser.Node{
				{Name: "CnameFlattening", Value: "true"},
				{Name: "ResolvConfig", Value: "foobar"},
				{Name: "ResolvDepth", Value: "42"}}},
			{Name: "Log", Children: []*parser.Node{
				{Name: "FilePath", Value: "foobar"},
				{Name: "Format", Value: "foobar"},
				{Name: "Level", Value: "foobar"}}},
			{Name: "Metrics", Children: []*parser.Node{
				{Name: "Datadog", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "PushInterval", Value: "10s"}}},
				{Name: "InfluxDB", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "Database", Value: "foobar"},
					{Name: "Password", Value: "foobar"},
					{Name: "Protocol", Value: "foobar"},
					{Name: "PushInterval", Value: "10s"},
					{Name: "RetentionPolicy", Value: "foobar"},
					{Name: "Username", Value: "foobar"}}},
				{Name: "Prometheus", Children: []*parser.Node{
					{Name: "Buckets", Value: "42,42"},
					{Name: "EntryPoint", Value: "foobar"},
					{Name: "Middlewares", Value: "foobar,foobar"}}},
				{Name: "StatsD", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "PushInterval", Value: "10s"}}}}},
			{Name: "Ping", Children: []*parser.Node{
				{Name: "EntryPoint", Value: "foobar"},
				{Name: "Middlewares", Value: "foobar,foobar"}}},
			{Name: "Providers", Children: []*parser.Node{
				{Name: "Docker", Children: []*parser.Node{
					{Name: "Constraints", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
					}},
					{Name: "DefaultRule", Value: "foobar"},
					{Name: "Endpoint", Value: "foobar"},
					{Name: "ExposedByDefault", Value: "true"},
					{Name: "Network", Value: "foobar"},
					{Name: "SwarmMode", Value: "true"},
					{Name: "SwarmModeRefreshSeconds", Value: "42"},
					{Name: "TLS", Children: []*parser.Node{
						{Name: "CA", Value: "foobar"},
						{Name: "CAOptional", Value: "true"},
						{Name: "Cert", Value: "foobar"},
						{Name: "InsecureSkipVerify", Value: "true"},
						{Name: "Key", Value: "foobar"}}},
					{Name: "UseBindPortIP", Value: "true"},
					{Name: "Watch", Value: "true"}}},
				{Name: "File", Children: []*parser.Node{
					{Name: "DebugLogGeneratedTemplate", Value: "true"},
					{Name: "Directory", Value: "foobar"},
					{Name: "Filename", Value: "foobar"},
					{Name: "TraefikFile", Value: "foobar"},
					{Name: "Watch", Value: "true"}}},
				{Name: "Kubernetes", Children: []*parser.Node{
					{Name: "CertAuthFilePath", Value: "foobar"},
					{Name: "DisablePassHostHeaders", Value: "true"},
					{Name: "Endpoint", Value: "foobar"},
					{Name: "IngressClass", Value: "foobar"},
					{Name: "IngressEndpoint", Children: []*parser.Node{
						{Name: "Hostname", Value: "foobar"},
						{Name: "IP", Value: "foobar"},
						{Name: "PublishedService", Value: "foobar"}}},
					{Name: "LabelSelector", Value: "foobar"},
					{Name: "Namespaces", Value: "foobar,foobar"},
					{Name: "Token", Value: "foobar"}}},
				{Name: "KubernetesCRD",
					Children: []*parser.Node{
						{Name: "CertAuthFilePath", Value: "foobar"},
						{Name: "DisablePassHostHeaders", Value: "true"},
						{Name: "Endpoint", Value: "foobar"},
						{Name: "IngressClass", Value: "foobar"},
						{Name: "LabelSelector", Value: "foobar"},
						{Name: "Namespaces", Value: "foobar,foobar"},
						{Name: "Token", Value: "foobar"}}},
				{Name: "Marathon", Children: []*parser.Node{
					{Name: "Basic", Children: []*parser.Node{
						{Name: "HTTPBasicAuthUser", Value: "foobar"},
						{Name: "HTTPBasicPassword", Value: "foobar"}}},
					{Name: "Constraints", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
					}},
					{Name: "DCOSToken", Value: "foobar"},
					{Name: "DefaultRule", Value: "foobar"},
					{Name: "DialerTimeout", Value: "42"},
					{Name: "Endpoint", Value: "foobar"},
					{Name: "ExposedByDefault", Value: "true"},
					{Name: "FilterMarathonConstraints", Value: "true"},
					{Name: "ForceTaskHostname", Value: "true"},
					{Name: "KeepAlive", Value: "42"},
					{Name: "RespectReadinessChecks", Value: "true"},
					{Name: "ResponseHeaderTimeout", Value: "42"},
					{Name: "TLS", Children: []*parser.Node{
						{Name: "CA", Value: "foobar"},
						{Name: "CAOptional", Value: "true"},
						{Name: "Cert", Value: "foobar"},
						{Name: "InsecureSkipVerify", Value: "true"},
						{Name: "Key", Value: "foobar"}}},
					{Name: "TLSHandshakeTimeout", Value: "42"},
					{Name: "Trace", Value: "true"},
					{Name: "Watch", Value: "true"}}},
				{Name: "ProvidersThrottleDuration", Value: "42"},
				{Name: "Rancher", Children: []*parser.Node{
					{Name: "Constraints", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
					}},
					{Name: "DefaultRule", Value: "foobar"},
					{Name: "EnableServiceHealthFilter", Value: "true"},
					{Name: "ExposedByDefault", Value: "true"},
					{Name: "IntervalPoll", Value: "true"},
					{Name: "Prefix", Value: "foobar"},
					{Name: "RefreshSeconds", Value: "42"},
					{Name: "Watch", Value: "true"}}},
				{Name: "Rest", Children: []*parser.Node{
					{Name: "EntryPoint", Value: "foobar"}}}}},
			{Name: "ServersTransport", Children: []*parser.Node{
				{Name: "ForwardingTimeouts", Children: []*parser.Node{
					{Name: "DialTimeout", Value: "42"},
					{Name: "ResponseHeaderTimeout", Value: "42"}}},
				{Name: "InsecureSkipVerify", Value: "true"},
				{Name: "MaxIdleConnsPerHost", Value: "42"},
				{Name: "RootCAs", Value: "foobar,foobar"}}},
			{Name: "Tracing", Children: []*parser.Node{
				{Name: "Backend", Value: "foobar"},
				{Name: "DataDog", Children: []*parser.Node{
					{Name: "BagagePrefixHeaderName", Value: "foobar"},
					{Name: "Debug", Value: "true"},
					{Name: "GlobalTag", Value: "foobar"},
					{Name: "LocalAgentHostPort", Value: "foobar"},
					{Name: "ParentIDHeaderName", Value: "foobar"},
					{Name: "PrioritySampling", Value: "true"},
					{Name: "SamplingPriorityHeaderName", Value: "foobar"},
					{Name: "TraceIDHeaderName", Value: "foobar"}}},
				{Name: "Instana", Children: []*parser.Node{
					{Name: "LocalAgentHost", Value: "foobar"},
					{Name: "LocalAgentPort", Value: "42"},
					{Name: "LogLevel", Value: "foobar"}}},
				{Name: "Jaeger", Children: []*parser.Node{
					{Name: "Gen128Bit", Value: "true"},
					{Name: "LocalAgentHostPort", Value: "foobar"},
					{Name: "Propagation", Value: "foobar"},
					{Name: "SamplingParam", Value: "42"},
					{Name: "SamplingServerURL", Value: "foobar"},
					{Name: "SamplingType", Value: "foobar"},
					{Name: "TraceContextHeaderName", Value: "foobar"}}},
				{Name: "ServiceName", Value: "foobar"},
				{Name: "SpanNameLimit", Value: "42"},
				{Name: "Zipkin", Children: []*parser.Node{
					{Name: "Debug", Value: "true"},
					{Name: "HTTPEndpoint", Value: "foobar"},
					{Name: "ID128Bit", Value: "true"},
					{Name: "SameSpan", Value: "true"},
					{Name: "SampleRate", Value: "42"}}}}}},
	}

	assert.Equal(t, expected, node)
}

func Test_decodeFileToNode_Yaml(t *testing.T) {
	node, err := decodeFileToNode("./fixtures/sample.yml")
	if err != nil {
		t.Fatal(err)
	}

	expected := &parser.Node{
		Name: "traefik",
		Children: []*parser.Node{
			{Name: "ACME",
				Children: []*parser.Node{
					{Name: "ACMELogging", Value: "true"},
					{Name: "CAServer", Value: "foobar"},
					{Name: "DNSChallenge", Children: []*parser.Node{
						{Name: "DelayBeforeCheck", Value: "42"},
						{Name: "DisablePropagationCheck", Value: "true"},
						{Name: "Provider", Value: "foobar"},
						{Name: "Resolvers", Value: "foobar,foobar"},
					}},
					{Name: "Domains", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Main", Value: "foobar"},
							{Name: "SANs", Value: "foobar,foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Main", Value: "foobar"},
							{Name: "SANs", Value: "foobar,foobar"},
						}},
					}},
					{Name: "Email", Value: "foobar"},
					{Name: "EntryPoint", Value: "foobar"},
					{Name: "HTTPChallenge", Children: []*parser.Node{
						{Name: "EntryPoint", Value: "foobar"}}},
					{Name: "KeyType", Value: "foobar"},
					{Name: "OnHostRule", Value: "true"},
					{Name: "Storage", Value: "foobar"},
					{Name: "TLSChallenge"},
				},
			},
			{Name: "API", Children: []*parser.Node{
				{Name: "Dashboard", Value: "true"},
				{Name: "EntryPoint", Value: "foobar"},
				{Name: "Middlewares", Value: "foobar,foobar"},
				{Name: "Statistics", Children: []*parser.Node{
					{Name: "RecentErrors", Value: "42"}}}}},
			{Name: "AccessLog", Children: []*parser.Node{
				{Name: "BufferingSize", Value: "42"},
				{Name: "Fields", Children: []*parser.Node{
					{Name: "DefaultMode", Value: "foobar"},
					{Name: "Headers", Children: []*parser.Node{
						{Name: "DefaultMode", Value: "foobar"},
						{Name: "Names", Children: []*parser.Node{
							{Name: "name0", Value: "foobar"},
							{Name: "name1", Value: "foobar"}}}}},
					{Name: "Names", Children: []*parser.Node{
						{Name: "name0", Value: "foobar"},
						{Name: "name1", Value: "foobar"}}}}},
				{Name: "FilePath", Value: "foobar"},
				{Name: "Filters", Children: []*parser.Node{
					{Name: "MinDuration", Value: "42"},
					{Name: "RetryAttempts", Value: "true"},
					{Name: "StatusCodes", Value: "foobar,foobar"}}},
				{Name: "Format", Value: "foobar"}}},
			{Name: "EntryPoints", Children: []*parser.Node{
				{Name: "EntryPoint0", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "ForwardedHeaders", Children: []*parser.Node{
						{Name: "Insecure", Value: "true"},
						{Name: "TrustedIPs", Value: "foobar,foobar"}}},
					{Name: "ProxyProtocol", Children: []*parser.Node{
						{Name: "Insecure", Value: "true"},
						{Name: "TrustedIPs", Value: "foobar,foobar"}}},
					{Name: "Transport", Children: []*parser.Node{
						{Name: "LifeCycle", Children: []*parser.Node{
							{Name: "GraceTimeOut", Value: "42"},
							{Name: "RequestAcceptGraceTimeout", Value: "42"}}},
						{Name: "RespondingTimeouts", Children: []*parser.Node{
							{Name: "IdleTimeout", Value: "42"},
							{Name: "ReadTimeout", Value: "42"},
							{Name: "WriteTimeout", Value: "42"}}}}}}}}},
			{Name: "Global", Children: []*parser.Node{
				{Name: "CheckNewVersion", Value: "true"},
				{Name: "Debug", Value: "true"},
				{Name: "SendAnonymousUsage", Value: "true"}}},
			{Name: "HostResolver", Children: []*parser.Node{
				{Name: "CnameFlattening", Value: "true"},
				{Name: "ResolvConfig", Value: "foobar"},
				{Name: "ResolvDepth", Value: "42"}}},
			{Name: "Log", Children: []*parser.Node{
				{Name: "FilePath", Value: "foobar"},
				{Name: "Format", Value: "foobar"},
				{Name: "Level", Value: "foobar"}}},
			{Name: "Metrics", Children: []*parser.Node{
				{Name: "Datadog", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "PushInterval", Value: "10s"}}},
				{Name: "InfluxDB", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "Database", Value: "foobar"},
					{Name: "Password", Value: "foobar"},
					{Name: "Protocol", Value: "foobar"},
					{Name: "PushInterval", Value: "10s"},
					{Name: "RetentionPolicy", Value: "foobar"},
					{Name: "Username", Value: "foobar"}}},
				{Name: "Prometheus", Children: []*parser.Node{
					{Name: "Buckets", Value: "42,42"},
					{Name: "EntryPoint", Value: "foobar"},
					{Name: "Middlewares", Value: "foobar,foobar"}}},
				{Name: "StatsD", Children: []*parser.Node{
					{Name: "Address", Value: "foobar"},
					{Name: "PushInterval", Value: "10s"}}}}},
			{Name: "Ping", Children: []*parser.Node{
				{Name: "EntryPoint", Value: "foobar"},
				{Name: "Middlewares", Value: "foobar,foobar"}}},
			{Name: "Providers", Children: []*parser.Node{
				{Name: "Docker", Children: []*parser.Node{
					{Name: "Constraints", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
					}},
					{Name: "DefaultRule", Value: "foobar"},
					{Name: "Endpoint", Value: "foobar"},
					{Name: "ExposedByDefault", Value: "true"},
					{Name: "Network", Value: "foobar"},
					{Name: "SwarmMode", Value: "true"},
					{Name: "SwarmModeRefreshSeconds", Value: "42"},
					{Name: "TLS", Children: []*parser.Node{
						{Name: "CA", Value: "foobar"},
						{Name: "CAOptional", Value: "true"},
						{Name: "Cert", Value: "foobar"},
						{Name: "InsecureSkipVerify", Value: "true"},
						{Name: "Key", Value: "foobar"}}},
					{Name: "UseBindPortIP", Value: "true"},
					{Name: "Watch", Value: "true"}}},
				{Name: "File", Children: []*parser.Node{
					{Name: "DebugLogGeneratedTemplate", Value: "true"},
					{Name: "Directory", Value: "foobar"},
					{Name: "Filename", Value: "foobar"},
					{Name: "TraefikFile", Value: "foobar"},
					{Name: "Watch", Value: "true"}}},
				{Name: "Kubernetes", Children: []*parser.Node{
					{Name: "CertAuthFilePath", Value: "foobar"},
					{Name: "DisablePassHostHeaders", Value: "true"},
					{Name: "Endpoint", Value: "foobar"},
					{Name: "IngressClass", Value: "foobar"},
					{Name: "IngressEndpoint", Children: []*parser.Node{
						{Name: "Hostname", Value: "foobar"},
						{Name: "IP", Value: "foobar"},
						{Name: "PublishedService", Value: "foobar"}}},
					{Name: "LabelSelector", Value: "foobar"},
					{Name: "Namespaces", Value: "foobar,foobar"},
					{Name: "Token", Value: "foobar"}}},
				{Name: "KubernetesCRD",
					Children: []*parser.Node{
						{Name: "CertAuthFilePath", Value: "foobar"},
						{Name: "DisablePassHostHeaders", Value: "true"},
						{Name: "Endpoint", Value: "foobar"},
						{Name: "IngressClass", Value: "foobar"},
						{Name: "LabelSelector", Value: "foobar"},
						{Name: "Namespaces", Value: "foobar,foobar"},
						{Name: "Token", Value: "foobar"}}},
				{Name: "Marathon", Children: []*parser.Node{
					{Name: "Basic", Children: []*parser.Node{
						{Name: "HTTPBasicAuthUser", Value: "foobar"},
						{Name: "HTTPBasicPassword", Value: "foobar"}}},
					{Name: "Constraints", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
					}},
					{Name: "DCOSToken", Value: "foobar"},
					{Name: "DefaultRule", Value: "foobar"},
					{Name: "DialerTimeout", Value: "42"},
					{Name: "Endpoint", Value: "foobar"},
					{Name: "ExposedByDefault", Value: "true"},
					{Name: "FilterMarathonConstraints", Value: "true"},
					{Name: "ForceTaskHostname", Value: "true"},
					{Name: "KeepAlive", Value: "42"},
					{Name: "RespectReadinessChecks", Value: "true"},
					{Name: "ResponseHeaderTimeout", Value: "42"},
					{Name: "TLS", Children: []*parser.Node{
						{Name: "CA", Value: "foobar"},
						{Name: "CAOptional", Value: "true"},
						{Name: "Cert", Value: "foobar"},
						{Name: "InsecureSkipVerify", Value: "true"},
						{Name: "Key", Value: "foobar"}}},
					{Name: "TLSHandshakeTimeout", Value: "42"},
					{Name: "Trace", Value: "true"},
					{Name: "Watch", Value: "true"}}},
				{Name: "ProvidersThrottleDuration", Value: "42"},
				{Name: "Rancher", Children: []*parser.Node{
					{Name: "Constraints", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "Key", Value: "foobar"},
							{Name: "MustMatch", Value: "true"},
							{Name: "Value", Value: "foobar"},
						}},
					}},
					{Name: "DefaultRule", Value: "foobar"},
					{Name: "EnableServiceHealthFilter", Value: "true"},
					{Name: "ExposedByDefault", Value: "true"},
					{Name: "IntervalPoll", Value: "true"},
					{Name: "Prefix", Value: "foobar"},
					{Name: "RefreshSeconds", Value: "42"},
					{Name: "Watch", Value: "true"}}},
				{Name: "Rest", Children: []*parser.Node{
					{Name: "EntryPoint", Value: "foobar"}}}}},
			{Name: "ServersTransport", Children: []*parser.Node{
				{Name: "ForwardingTimeouts", Children: []*parser.Node{
					{Name: "DialTimeout", Value: "42"},
					{Name: "ResponseHeaderTimeout", Value: "42"}}},
				{Name: "InsecureSkipVerify", Value: "true"},
				{Name: "MaxIdleConnsPerHost", Value: "42"},
				{Name: "RootCAs", Value: "foobar,foobar"}}},
			{Name: "Tracing", Children: []*parser.Node{
				{Name: "Backend", Value: "foobar"},
				{Name: "DataDog", Children: []*parser.Node{
					{Name: "BagagePrefixHeaderName", Value: "foobar"},
					{Name: "Debug", Value: "true"},
					{Name: "GlobalTag", Value: "foobar"},
					{Name: "LocalAgentHostPort", Value: "foobar"},
					{Name: "ParentIDHeaderName", Value: "foobar"},
					{Name: "PrioritySampling", Value: "true"},
					{Name: "SamplingPriorityHeaderName", Value: "foobar"},
					{Name: "TraceIDHeaderName", Value: "foobar"}}},
				{Name: "Instana", Children: []*parser.Node{
					{Name: "LocalAgentHost", Value: "foobar"},
					{Name: "LocalAgentPort", Value: "42"},
					{Name: "LogLevel", Value: "foobar"}}},
				{Name: "Jaeger", Children: []*parser.Node{
					{Name: "Gen128Bit", Value: "true"},
					{Name: "LocalAgentHostPort", Value: "foobar"},
					{Name: "Propagation", Value: "foobar"},
					{Name: "SamplingParam", Value: "42"},
					{Name: "SamplingServerURL", Value: "foobar"},
					{Name: "SamplingType", Value: "foobar"},
					{Name: "TraceContextHeaderName", Value: "foobar"}}},
				{Name: "ServiceName", Value: "foobar"},
				{Name: "SpanNameLimit", Value: "42"},
				{Name: "Zipkin", Children: []*parser.Node{
					{Name: "Debug", Value: "true"},
					{Name: "HTTPEndpoint", Value: "foobar"},
					{Name: "ID128Bit", Value: "true"},
					{Name: "SameSpan", Value: "true"},
					{Name: "SampleRate", Value: "42"}}}}}},
	}

	assert.Equal(t, expected, node)
}
