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
	default:
		return nil, errors.New("Unable to handle request: " + url)
	}
}

func (c *mockWebClient) SetTransport(transport *http.Transport) {}

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
