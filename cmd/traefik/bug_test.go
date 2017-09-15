package main

import (
	"testing"
)

func Test_anonymize(t *testing.T) {
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
   "Network": "",
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
   "Network": "",
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
  "Email": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "Domains": [
   {
    "Main": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "SANs": null
   },
   {
    "Main": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
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
	anomConfiguration := anonymize(baseConfiguration)

	if anomConfiguration != expectedConfiguration {
		t.Errorf("Got %s, want %s.", anomConfiguration, expectedConfiguration)
	}
}
