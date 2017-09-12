package servicefabric

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

type mockWebClient struct {
}

func (c *mockWebClient) Get(url string) (resp *http.Response, err error) {
	switch url {
	case "Test/Applications/?api-version=1.0":
		body := `{"ContinuationToken":"",
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
	default:
		return nil, errors.New("Unable to handle request: " + url)
	}
}

func (c *mockWebClient) Transport(transport *http.Transport) {}

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

func setupClient() Client {
	webClient := &mockWebClient{}
	sfClient, _ := NewClient(
		webClient,
		"Test",
		"1.0",
		"",
		"",
		"")
	return sfClient
}

func TestGetApplications(t *testing.T) {
	expected := &ApplicationsData{
		ContinuationToken: nil,
		Items: []ApplicationItem{
			ApplicationItem{
				HealthState: "Ok",
				ID:          "TestApplication",
				Name:        "fabric:/TestApplication",
				Parameters: []*struct {
					Key   string `json:"Key"`
					Value string `json:"Value"`
				}{
					&struct {
						Key   string `json:"Key"`
						Value string `json:"Value"`
					}{"Param1", "Value1"},
					&struct {
						Key   string `json:"Key"`
						Value string `json:"Value"`
					}{"Param2", "Value2"},
				},
				Status:      "Ready",
				TypeName:    "TestApplicationType",
				TypeVersion: "1.0.0",
			},
		},
	}
	sfClient := setupClient()
	actual, err := sfClient.GetApplications()
	if err != nil {
		t.Errorf("Exception thrown %v", err)
	}
	isEqual := reflect.DeepEqual(expected, actual)
	if !isEqual {
		t.Error("actual != expected")
	}
}

func TestGetServices(t *testing.T) {
	expected := &ServicesData{
		ContinuationToken: nil,
		Items: []ServiceItem{
			ServiceItem{
				HasPersistedState: true,
				HealthState:       "Ok",
				ID:                "TestApplication/TestService",
				IsServiceGroup:    false,
				ManifestVersion:   "1.0.0",
				Name:              "fabric:/TestApplication/TestService",
				ServiceKind:       "Stateful",
				ServiceStatus:     "Active",
				TypeName:          "TestServiceType",
			},
		},
	}
	sfClient := setupClient()
	actual, err := sfClient.GetServices("TestApplication")
	if err != nil {
		t.Errorf("Exception thrown %v", err)
	}
	isEqual := reflect.DeepEqual(expected, actual)
	if !isEqual {
		t.Error("actual != expected")
	}
}

func TestGetPartitions(t *testing.T) {
	expected := &PartitionsData{
		ContinuationToken: nil,
		Items: []PartitionData{
			PartitionData{
				CurrentConfigurationEpoch: struct {
					ConfigurationVersion string `json:"ConfigurationVersion"`
					DataLossVersion      string `json:"DataLossVersion"`
				}{
					ConfigurationVersion: "12884901891",
					DataLossVersion:      "131496928071680379",
				},
				HealthState:       "Ok",
				MinReplicaSetSize: 3,
				PartitionInformation: struct {
					HighKey              string `json:"HighKey"`
					ID                   string `json:"Id"`
					LowKey               string `json:"LowKey"`
					ServicePartitionKind string `json:"ServicePartitionKind"`
				}{
					HighKey:              "9223372036854775807",
					ID:                   "bce46a8c-b62d-4996-89dc-7ffc00a96902",
					LowKey:               "-9223372036854775808",
					ServicePartitionKind: "Int64Range",
				},
				PartitionStatus:      "Ready",
				ServiceKind:          "Stateful",
				TargetReplicaSetSize: 3,
			},
		},
	}
	sfClient := setupClient()
	actual, err := sfClient.GetPartitions("TestApplication", "TestApplication/TestService")
	if err != nil {
		t.Errorf("Exception thrown %v", err)
	}
	isEqual := reflect.DeepEqual(expected, actual)
	if !isEqual {
		t.Error("actual != expected")
	}
}

func TestGetReplicas(t *testing.T) {
	expected := &ReplicasData{
		ContinuationToken: nil,
		Items: []ReplicaItem{
			ReplicaItem{
				ReplicaItemBase: &ReplicaItemBase{
					Address:                      "{\"Endpoints\":{\"\":\"localhost:30001+bce46a8c-b62d-4996-89dc-7ffc00a96902-131496928082309293\"}}",
					HealthState:                  "Ok",
					LastInBuildDurationInSeconds: "1",
					NodeName:                     "_Node_0",
					ReplicaRole:                  "Primary",
					ReplicaStatus:                "Ready",
					ServiceKind:                  "Stateful",
				},
				ID: "131496928082309293",
			},
		},
	}
	sfClient := setupClient()
	actual, err := sfClient.GetReplicas("TestApplication", "TestApplication/TestService", "bce46a8c-b62d-4996-89dc-7ffc00a96902")
	if err != nil {
		t.Errorf("Exception thrown %v", err)
	}
	isEqual := reflect.DeepEqual(expected, actual)
	if !isEqual {
		t.Error("actual != expected")
	}
}

func TestGetInstances(t *testing.T) {
	expected := &InstancesData{
		ContinuationToken: nil,
		Items: []InstanceItem{
			InstanceItem{
				ReplicaItemBase: &ReplicaItemBase{
					Address:                      "{\"Endpoints\":{\"\":\"http:\\/\\/localhost:8081\"}}",
					HealthState:                  "Ok",
					LastInBuildDurationInSeconds: "3",
					NodeName:                     "_Node_0",
					ReplicaStatus:                "Ready",
					ServiceKind:                  "Stateless",
				},
				ID: "131497042182378182",
			},
		},
	}
	sfClient := setupClient()
	actual, err := sfClient.GetInstances("TestApplication", "TestApplication/TestService", "824091ba-fa32-4e9c-9e9c-71738e018312")
	if err != nil {
		t.Errorf("Exception thrown %v", err)
	}
	isEqual := reflect.DeepEqual(expected, actual)
	if !isEqual {
		t.Error("actual != expected")
	}
}
