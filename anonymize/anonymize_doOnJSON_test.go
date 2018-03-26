package anonymize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_doOnJSON(t *testing.T) {
	baseConfiguration := `
{
 "GraceTimeOut": 10000000000,
 "Debug": false,
 "CheckNewVersion": true,
 "AccessLogsFile": "",
 "TraefikLogsFile": "",
 "LogLevel": "ERROR",
 "EntryPoints": {
  "http": {
   "Network": "",
   "Address": ":80",
   "TLS": null,
   "Redirect": {
    "EntryPoint": "https",
    "Regex": "",
    "Replacement": ""
   },
   "Auth": null,
   "Compress": false
  },
  "https": {
   "Address": ":443",
   "TLS": {
    "MinVersion": "",
    "CipherSuites": null,
    "Certificates": null,
    "ClientCAFiles": null
   },
   "Redirect": null,
   "Auth": null,
   "Compress": false
  }
 },
 "Cluster": null,
 "Constraints": [],
 "ACME": {
  "Email": "foo@bar.com",
  "Domains": [
   {
    "Main": "foo@bar.com",
    "SANs": null
   },
   {
    "Main": "foo@bar.com",
    "SANs": null
   }
  ],
  "Storage": "",
  "StorageFile": "/acme/acme.json",
  "OnDemand": true,
  "OnHostRule": true,
  "CAServer": "",
  "EntryPoint": "https",
  "DNSProvider": "",
  "DelayDontCheckDNS": 0,
  "ACMELogging": false,
  "TLSConfig": null
 },
 "DefaultEntryPoints": [
  "https",
  "http"
 ],
 "ProvidersThrottleDuration": 2000000000,
 "MaxIdleConnsPerHost": 200,
 "IdleTimeout": 180000000000,
 "InsecureSkipVerify": false,
 "Retry": null,
 "HealthCheck": {
  "Interval": 30000000000
 },
 "Docker": null,
 "File": null,
 "Web": null,
 "Marathon": null,
 "Consul": null,
 "ConsulCatalog": null,
 "Etcd": null,
 "Zookeeper": null,
 "Boltdb": null,
 "Kubernetes": null,
 "Mesos": null,
 "Eureka": null,
 "ECS": null,
 "Rancher": null,
 "DynamoDB": null,
 "ConfigFile": "/etc/traefik/traefik.toml"
}
`
	expectedConfiguration := `
{
 "GraceTimeOut": 10000000000,
 "Debug": false,
 "CheckNewVersion": true,
 "AccessLogsFile": "",
 "TraefikLogsFile": "",
 "LogLevel": "ERROR",
 "EntryPoints": {
  "http": {
   "Network": "",
   "Address": ":80",
   "TLS": null,
   "Redirect": {
    "EntryPoint": "https",
    "Regex": "",
    "Replacement": ""
   },
   "Auth": null,
   "Compress": false
  },
  "https": {
   "Address": ":443",
   "TLS": {
    "MinVersion": "",
    "CipherSuites": null,
    "Certificates": null,
    "ClientCAFiles": null
   },
   "Redirect": null,
   "Auth": null,
   "Compress": false
  }
 },
 "Cluster": null,
 "Constraints": [],
 "ACME": {
  "Email": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "Domains": [
   {
    "Main": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "SANs": null
   },
   {
    "Main": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "SANs": null
   }
  ],
  "Storage": "",
  "StorageFile": "/acme/acme.json",
  "OnDemand": true,
  "OnHostRule": true,
  "CAServer": "",
  "EntryPoint": "https",
  "DNSProvider": "",
  "DelayDontCheckDNS": 0,
  "ACMELogging": false,
  "TLSConfig": null
 },
 "DefaultEntryPoints": [
  "https",
  "http"
 ],
 "ProvidersThrottleDuration": 2000000000,
 "MaxIdleConnsPerHost": 200,
 "IdleTimeout": 180000000000,
 "InsecureSkipVerify": false,
 "Retry": null,
 "HealthCheck": {
  "Interval": 30000000000
 },
 "Docker": null,
 "File": null,
 "Web": null,
 "Marathon": null,
 "Consul": null,
 "ConsulCatalog": null,
 "Etcd": null,
 "Zookeeper": null,
 "Boltdb": null,
 "Kubernetes": null,
 "Mesos": null,
 "Eureka": null,
 "ECS": null,
 "Rancher": null,
 "DynamoDB": null,
 "ConfigFile": "/etc/traefik/traefik.toml"
}
`
	anomConfiguration := doOnJSON(baseConfiguration)

	if anomConfiguration != expectedConfiguration {
		t.Errorf("Got %s, want %s.", anomConfiguration, expectedConfiguration)
	}
}

func Test_doOnJSON_simple(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name: "email",
			input: `{
				"email1": "goo@example.com",
				"email2": "foo.bargoo@example.com",
				"email3": "foo.bargoo@example.com.us"
			}`,
			expectedOutput: `{
				"email1": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"email2": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"email3": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
			}`,
		},
		{
			name: "url",
			input: `{
				"URL": "foo domain.com foo",
				"URL": "foo sub.domain.com foo",
				"URL": "foo sub.sub.domain.com foo",
				"URL": "foo sub.sub.sub.domain.com.us foo"
			}`,
			expectedOutput: `{
				"URL": "foo xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx foo",
				"URL": "foo xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx foo",
				"URL": "foo xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx foo",
				"URL": "foo xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx foo"
			}`,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			output := doOnJSON(test.input)
			assert.Equal(t, test.expectedOutput, output)
		})
	}
}
