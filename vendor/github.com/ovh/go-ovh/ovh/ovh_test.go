package ovh

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// In case you wonder, these are real *revoked* credentials
	MockApplicationKey    = "TDPKJdwZwAQPwKX2"
	MockApplicationSecret = "9ufkBmLaTQ9nz5yMUlg79taH0GNnzDjk"
	MockConsumerKey       = "5mBuy6SUQcRw2ZUxg0cG68BoDKpED4KY"

	MockTime = 1457018875
)

type SomeData struct {
	IntValue    int    `json:"i_val,omitempty"`
	StringValue string `json:"s_val,omitempty"`
}

//
// Utils
//

func initMockServer(InputRequest **http.Request, status int, responseBody string, requestBody *string) (*httptest.Server, *Client) {
	// Mock time
	getLocalTime = func() time.Time {
		return time.Unix(MockTime, 0)
	}

	// Mock hostname, in signature only
	getEndpointForSignature = func(c *Client) string {
		return "http://localhost"
	}

	// Create a fake API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Save input parameters
		*InputRequest = r
		defer r.Body.Close()

		if requestBody != nil {
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				*requestBody = string(reqBody[:])
			}
		}

		// Respond
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprintln(w, responseBody)
	}))

	// Create client
	client, _ := NewClient(ts.URL, MockApplicationKey, MockApplicationSecret, MockConsumerKey)
	client.timeDeltaDone = true

	return ts, client
}

func ensureHeaderPresent(t *testing.T, r *http.Request, name, value string) {
	val, present := r.Header[name]

	if !present {
		t.Fatalf("%s requests should include a %s header with %s value.", r.Method, name, value)
	}

	if val[0] != value {
		t.Fatalf("%s requests should include a %s header with %s value. Got %s", r.Method, name, value, val[0])
	}
}

func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + strings.ToLower(s[n:])
}

//
// Tests
//

func TestTime(t *testing.T) {
	// Init test
	var InputRequest *http.Request
	ts, client := initMockServer(&InputRequest, 200, fmt.Sprintf("%d", MockTime), nil)
	defer ts.Close()

	// Test
	serverTime, err := client.Time()
	if err != nil {
		t.Fatalf("Unexpected error while retrieving server time: %v\n", err)
	}

	// Validate
	if InputRequest.Method != "GET" || InputRequest.URL.String() != "/auth/time" {
		t.Fatalf("Time should be retrieved using GET /auth/time. Got %s %s", InputRequest.Method, InputRequest.URL.String())
	}
	if serverTime.Unix() != MockTime {
		t.Fatalf("Server time should be %d. Got %d", MockTime, serverTime.Unix())
	}
}

func TestPing(t *testing.T) {
	// Init test
	var InputRequest *http.Request
	ts, client := initMockServer(&InputRequest, 200, `0`, nil)
	defer ts.Close()

	// Test
	err := client.Ping()

	// Validate
	if err != nil {
		t.Fatalf("Unexpected error while pinging server: %v\n", err)
	}
}

func TestPingUnreachable(t *testing.T) {
	// Init test
	var InputRequest *http.Request
	ts, client := initMockServer(&InputRequest, 200, `0`, nil)
	defer ts.Close()

	// Test
	client.endpoint = "https://localhost:1/does not exist"
	err := client.Ping()

	// Validate
	if err == nil {
		t.Fatalf("Unexpected success while pinging server\n")
	}
}

// APIMethodTester applies the same sanity checks to all main Client method. It checks the
// request method, path, body and headers for both authenticated and unauthenticated variants
func APIMethodTester(t *testing.T, HTTPmethod string, body interface{}, expectedBody string, expectedSignature string) {
	// Init test
	var InputRequest *http.Request
	var InputRequestBody string
	ts, client := initMockServer(&InputRequest, 200, `"success"`, &InputRequestBody)
	defer ts.Close()

	// Prepare method name
	needAuth := expectedSignature != ""
	HTTPmethod = strings.ToUpper(HTTPmethod)
	methodName := Capitalize(HTTPmethod)
	if !needAuth {
		methodName += "UnAuth"
	}

	// Prepare method arguments
	var res interface{}
	var arguments []reflect.Value
	arguments = append(arguments, reflect.ValueOf("/some/resource"))
	if body != nil {
		arguments = append(arguments, reflect.ValueOf(body))
	}
	arguments = append(arguments, reflect.ValueOf(&res))

	// Get method to test
	method := reflect.ValueOf(client).MethodByName(methodName)
	if !method.IsValid() {
		t.Fatalf("Client should suport %s method\n", methodName)
	}
	ret := method.Call(arguments)
	if !ret[0].IsNil() {
		t.Fatalf("Unexpected error while retrieving server time: %v\n", ret[0])
	}

	// Log request details to help debugging
	t.Logf("Request: %s %s. Authenticated=%v", InputRequest.Method, InputRequest.URL.String(), needAuth)
	for key, value := range InputRequest.Header {
		t.Logf("\tHEADER: key=%v, value=%v\n", key, value)
	}

	// Validate Method
	if InputRequest.Method != HTTPmethod || InputRequest.URL.String() != "/some/resource" {
		t.Fatalf("%s should trigger a %s /some/resource request. Got %s %s", methodName, HTTPmethod, InputRequest.Method, InputRequest.URL.String())
	}

	// Validate Body
	if body != nil && expectedBody != InputRequestBody {
		t.Fatalf("%s /some/resource should have '%s' body. Got '%s'", methodName, expectedBody, InputRequestBody)
	}

	// Validate Headers
	ensureHeaderPresent(t, InputRequest, "Accept", "application/json")
	ensureHeaderPresent(t, InputRequest, "X-Ovh-Application", MockApplicationKey)

	if body != nil {
		ensureHeaderPresent(t, InputRequest, "Content-Type", "application/json;charset=utf-8")
	}

	if needAuth {
		ensureHeaderPresent(t, InputRequest, "X-Ovh-Timestamp", strconv.Itoa(MockTime))
		ensureHeaderPresent(t, InputRequest, "X-Ovh-Consumer", MockConsumerKey)
		ensureHeaderPresent(t, InputRequest, "X-Ovh-Signature", expectedSignature)
	}

}

func TestAllAPIMethods(t *testing.T) {
	body := SomeData{
		IntValue:    42,
		StringValue: "Hello World!",
	}

	APIMethodTester(t, "GET", nil, "", "$1$8a21169b341aa23e82192e07457ca978006b1ba9")
	APIMethodTester(t, "GET", nil, "", "")
	APIMethodTester(t, "DELETE", nil, "", "$1$f4571312a04a4c75188509e75c40581ca6bb6d7a")
	APIMethodTester(t, "DELETE", nil, "", "")
	APIMethodTester(t, "POST", body, `{"i_val":42,"s_val":"Hello World!"}`, "$1$6549d84e65be72f4ec0d7b6d7eaa19554a265990")
	APIMethodTester(t, "POST", body, `{"i_val":42,"s_val":"Hello World!"}`, "")
	APIMethodTester(t, "PUT", body, `{"i_val":42,"s_val":"Hello World!"}`, "$1$983e2a9a213c99211edd0b32715ac1ace1a6a0ea")
	APIMethodTester(t, "PUT", body, `{"i_val":42,"s_val":"Hello World!"}`, "")
}

// Mock ReadCloser, always failing
type ErrorCloseReader struct{}

func (ErrorCloseReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("ErrorReader")
}
func (ErrorCloseReader) Close() error {
	return nil
}

func TestGetResponse(t *testing.T) {
	var err error
	var apiInt int
	mockClient := Client{}

	// Nominal
	err = mockClient.getResponse(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(`42`)),
	}, &apiInt)
	if err != nil {
		t.Fatalf("Client.getResponse should be able to decode int when status is 200. Got %v", err)
	}

	// Nominal: empty body
	err = mockClient.getResponse(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(``)),
	}, nil)
	if err != nil {
		t.Fatalf("getResponse should not return an error when reponse is empty or target type is nil. Got %v", err)
	}

	// Error
	err = mockClient.getResponse(&http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(strings.NewReader(`{"code": 400, "message": "Ooops..."}`)),
	}, &apiInt)
	if err == nil {
		t.Fatalf("Client.getResponse should be able to decode an error when status is 400")
	}
	if _, ok := err.(*APIError); !ok {
		t.Fatalf("Client.getResponse error should be an APIError when status is 400. Got '%s' of type %s", err, reflect.TypeOf(err))
	}

	// Error: body read error
	err = mockClient.getResponse(&http.Response{
		Body: ErrorCloseReader{},
	}, nil)
	if err == nil {
		t.Fatalf("getResponse should return an error when failing to read HTTP Response body. %v", err)
	}

	// Error: HTTP Error + broken json
	err = mockClient.getResponse(&http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(strings.NewReader(`{"code": 400, "mes`)),
	}, nil)
	if err == nil {
		t.Fatalf("getResponse should return an error when failing to decode HTTP Response body. %v", err)
	}

	// Error with QueryID
	responseHeaders := http.Header{}
	responseHeaders.Add("X-Ovh-QueryID", "FR.ws-8.5860f657.4632.0180")
	err = mockClient.getResponse(&http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(strings.NewReader(`{"code": 400, "message": "Ooops..."}`)),
		Header:     responseHeaders,
	}, &apiInt)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Client.getResponse error should be an APIError when status is 400 and header QueryID is found. Got '%s' of type %s", err, reflect.TypeOf(err))
	}
	if apiErr.QueryID != "FR.ws-8.5860f657.4632.0180" {
		t.Fatalf("APIError should be filled with a correct QueryID. Got '%s' instead of '%s'", apiErr.QueryID, "FR.ws-8.5860f657.4632.0180")
	}
}

func TestConstructors(t *testing.T) {
	// Nominal: full constructor
	client, err := NewClient("ovh-eu", MockApplicationKey, MockApplicationSecret, MockConsumerKey)
	if err != nil {
		t.Fatalf("NewClient should not return an error in the nominal case. Got: %v", err)
	}
	if client.Client == nil {
		t.Fatalf("client.Client should be a valid HTTP client")
	}
	if client.AppKey != MockApplicationKey {
		t.Fatalf("client.AppKey should be '%s'. Got '%s'", MockApplicationKey, client.AppKey)
	}
	if client.AppSecret != MockApplicationSecret {
		t.Fatalf("client.AppSecret should be '%s'. Got '%s'", MockApplicationSecret, client.AppSecret)
	}
	if client.ConsumerKey != MockConsumerKey {
		t.Fatalf("client.ConsumerKey should be '%s'. Got '%s'", MockConsumerKey, client.ConsumerKey)
	}

	// Nominal: Endpoint constructor
	os.Setenv("OVH_APPLICATION_KEY", MockApplicationKey)
	os.Setenv("OVH_APPLICATION_SECRET", MockApplicationSecret)
	os.Setenv("OVH_CONSUMER_KEY", MockConsumerKey)

	client, err = NewEndpointClient("ovh-eu")
	if err != nil {
		t.Fatalf("NewEndpointClient should not return an error in the nominal case. Got: %v", err)
	}
	if client.Client == nil {
		t.Fatalf("client.Client should be a valid HTTP client")
	}
	if client.AppKey != MockApplicationKey {
		t.Fatalf("client.AppKey should be '%s'. Got '%s'", MockApplicationKey, client.AppKey)
	}
	if client.AppSecret != MockApplicationSecret {
		t.Fatalf("client.AppSecret should be '%s'. Got '%s'", MockApplicationSecret, client.AppSecret)
	}
	if client.ConsumerKey != MockConsumerKey {
		t.Fatalf("client.ConsumerKey should be '%s'. Got '%s'", MockConsumerKey, client.ConsumerKey)
	}

	// Nominal: Default constructor
	os.Setenv("OVH_ENDPOINT", "ovh-eu")

	client, err = NewDefaultClient()
	if err != nil {
		t.Fatalf("NewEndpointClient should not return an error in the nominal case. Got: %v", err)
	}
	if client.Client == nil {
		t.Fatalf("client.Client should be a valid HTTP client")
	}
	if client.endpoint != "https://eu.api.ovh.com/1.0" {
		t.Fatalf("client.Endpoint should be 'https://eu.api.ovh.com/1.0'. Got '%s'", client.endpoint)
	}

	// Clear
	os.Unsetenv("OVH_ENDPOINT")
	os.Unsetenv("OVH_APPLICATION_KEY")
	os.Unsetenv("OVH_APPLICATION_SECRET")
	os.Unsetenv("OVH_CONSUMER_KEY")

	// Error: missing Endpoint
	_, err = NewClient("", MockApplicationKey, MockApplicationSecret, MockConsumerKey)
	if err == nil {
		t.Fatalf("NewClient should return an error when missing Endpoint")
	}
	// Error: missing ApplicationKey
	_, err = NewClient("ovh-eu", "", MockApplicationSecret, MockConsumerKey)
	if err == nil {
		t.Fatalf("NewClient should return an error when missing ApplicationKey")
	}
	// Error: missing ApplicationSecret
	_, err = NewClient("ovh-eu", MockConsumerKey, "", MockConsumerKey)
	if err == nil {
		t.Fatalf("NewClient should return an error when missing ApplicationSecret")
	}
}

func TestGetTimeDelta(t *testing.T) {
	MockDelta := 747

	// Init test
	var InputRequest *http.Request
	ts, client := initMockServer(&InputRequest, 200, fmt.Sprintf("%d", int(time.Now().Unix())-MockDelta), nil)
	defer ts.Close()

	// Test
	client.timeDeltaDone = false
	delta, err := client.getTimeDelta()

	if err != nil {
		t.Fatalf("getTimeDelta should not return an error. Got %v", err)
	}
	// Hack: take races into account, avoid mocking whole earth
	if math.Abs(float64(delta/time.Second-time.Duration(MockDelta))) > 2 {
		t.Fatalf("getTimeDelta should return a delta of %d. Got %d", time.Duration(MockDelta)*time.Second, delta)
	}
}
