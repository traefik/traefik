package servicefabric

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

type mockHTTPClient struct {
}

func (c *mockHTTPClient) Get(url string) (resp *http.Response, err error) {
	switch url {
	case "Test/Applications/?api-version=1.0":
		body := `{"ContinuationToken":"00001234",
				  "Items":[
					  {"Id":"TestApplication",
					   "Name":"fabric:\/TestApplication",
					   "TypeName":"TestApplicationType",
					   "TypeVersion":"1.0.0",
					   "Status":"Ready",
					   "Parameters":
							   [{"Key":"Param1","Value":"Value1"},
							   {"Key":"Param2","Value":"Value2"}],
					   "HealthState":"Ok"}
					]}`
		return buildSuccessResponse(body), nil
	case "Test/Applications/?api-version=1.0&continue=00001234":
		body := `{"ContinuationToken":"",
				  "Items":[
					  {"Id":"TestApplication2",
					   "Name":"fabric:\/TestApplication2",
					   "TypeName":"TestApplication2Type",
					   "TypeVersion":"1.0.0",
					   "Status":"Ready",
					   "Parameters":
							   [{"Key":"Param1","Value":"Value1"},
							   {"Key":"Param2","Value":"Value2"}],
					   "HealthState":"Ok"}
					]}`
		return buildSuccessResponse(body), nil
	case "Test/Applications/TestApplication/$/GetServices?api-version=1.0":
		body := `{"ContinuationToken":"",
				  "Items":[
					  {"Id":"TestApplication\/TestService",
					   "ServiceKind": "Stateful",
					   "Name":"fabric:\/TestApplication\/TestService",
					   "TypeName":"TestServiceType",
					   "ManifestVersion":"1.0.0",
					   "HasPersistedState": true,
					   "HealthState": "Ok",
					   "ServiceStatus": "Active",
					   "IsServiceGroup": false}
					]}`
		return buildSuccessResponse(body), nil
	case "Test/Applications/TestApplication/$/GetServices/TestApplication/TestService/$/GetPartitions/?api-version=1.0":
		body := `{
			"ContinuationToken": "",
			"Items": [{
				"ServiceKind": "Stateful",
				"PartitionInformation": {
					"ServicePartitionKind": "Int64Range",
					"Id": "bce46a8c-b62d-4996-89dc-7ffc00a96902",
					"LowKey": "-9223372036854775808",
					"HighKey": "9223372036854775807"
				},
				"TargetReplicaSetSize": 3,
				"MinReplicaSetSize": 3,
				"HealthState": "Ok",
				"PartitionStatus": "Ready",
				"LastQuorumLossDurationInSeconds": "3",
				"CurrentConfigurationEpoch": {
					"ConfigurationVersion": "12884901891",
					"DataLossVersion": "131496928071680379"
				}
			}]
		}`
		return buildSuccessResponse(body), nil
	case "Test/Applications/TestApplication/$/GetServices/TestApplication/TestService/$/GetPartitions/bce46a8c-b62d-4996-89dc-7ffc00a96902/$/GetReplicas?api-version=1.0":
		body := `{
				"ContinuationToken": "",
				"Items": [{
					"ServiceKind": "Stateful",
					"ReplicaId": "131496928082309293",
					"ReplicaRole": "Primary",
					"ReplicaStatus": "Ready",
					"HealthState": "Ok",
					"Address": "{\"Endpoints\":{\"\":\"localhost:30001+bce46a8c-b62d-4996-89dc-7ffc00a96902-131496928082309293\"}}",
					"NodeName": "_Node_0",
					"LastInBuildDurationInSeconds": "1"
				}]
			}`
		return buildSuccessResponse(body), nil
	case "Test/Applications/TestApplication/$/GetServices/TestApplication/TestService/$/GetPartitions/824091ba-fa32-4e9c-9e9c-71738e018312/$/GetReplicas?api-version=1.0":
		body := `{
				"ContinuationToken": "",
				"Items": [{
					"ServiceKind": "Stateless",
					"InstanceId": "131497042182378182",
					"ReplicaStatus": "Ready",
					"HealthState": "Ok",
					"Address": "{\"Endpoints\":{\"\":\"http:\\\/\\\/localhost:8081\"}}",
					"NodeName": "_Node_0",
					"LastInBuildDurationInSeconds": "3"
				}]
			}`
		return buildSuccessResponse(body), nil
	case "Test/ApplicationTypes/TestApplication/$/GetServiceTypes?api-version=1.0&ApplicationTypeVersion=1.0.0":
		body := `[
			{
				"ServiceTypeDescription": {
					"IsStateful": true,
					"ServiceTypeName": "Test",
					"PlacementConstraints": "",
					"HasPersistedState": true,
					"Kind": "Stateful",
					"Extensions": [
						{
							"Key": "Test",
							"Value": "<Tests xmlns=\"http://schemas.microsoft.com/2015/03/fabact-no-schema\"><Test Key=\"key1\">value1</Test></Tests>"
						}
					],
					"LoadMetrics": [],
					"ServicePlacementPolicies": []
				},
				"ServiceManifestVersion": "1.0.0",
				"ServiceManifestName": "ServiceManifest",
				"IsServiceGroup": false
			}]`
		return buildSuccessResponse(body), nil
	case "Test/ApplicationTypes/TestApplication/$/GetServiceTypes?api-version=1.0&ApplicationTypeVersion=1.0.1":
		body := `[
			{
				"ServiceTypeDescription": {
					"IsStateful": true,
					"ServiceTypeName": "Test",
					"PlacementConstraints": "",
					"HasPersistedState": true,
					"Kind": "Stateful",
					"Extensions": [],
					"LoadMetrics": [],
					"ServicePlacementPolicies": []
				},
				"ServiceManifestVersion": "1.0.0",
				"ServiceManifestName": "ServiceManifest",
				"IsServiceGroup": false
			}]`
		return buildSuccessResponse(body), nil
	default:
		return nil, errors.New("Unable to handle request: " + url)
	}
}

func (c *mockHTTPClient) Transport(transport *http.Transport) {}

func buildSuccessResponse(body string) *http.Response {
	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       nil,
		Header:        make(http.Header, 0),
	}
}
