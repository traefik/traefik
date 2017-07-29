package ovh

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

// Common helpers are in ovh_test.go

func TestNewCkRequest(t *testing.T) {
	const expectedRequest = `{"accessRules":[{"method":"GET","path":"/me"},{"method":"GET","path":"/xdsl/*"}]}`

	// Init test
	var InputRequest *http.Request
	var InputRequestBody string
	ts, client := initMockServer(&InputRequest, 200, `{
		"validationUrl":"https://validation.url",
		"ConsumerKey":"`+MockConsumerKey+`",
		"state":"pendingValidation"
	}`, &InputRequestBody)
	client.ConsumerKey = ""
	defer ts.Close()

	// Test
	ckRequest := client.NewCkRequest()
	ckRequest.AddRule("GET", "/me")
	ckRequest.AddRule("GET", "/xdsl/*")

	got, err := ckRequest.Do()

	// Validate
	if err != nil {
		t.Fatalf("CkRequest.Do() should not return an error. Got: %q", err)
	}
	if client.ConsumerKey != MockConsumerKey {
		t.Fatalf("CkRequest.Do() should set client.ConsumerKey to %s. Got %s", MockConsumerKey, client.ConsumerKey)
	}
	if got.ConsumerKey != MockConsumerKey {
		t.Fatalf("CkRequest.Do() should set CkValidationState.ConsumerKey to %s. Got %s", MockConsumerKey, got.ConsumerKey)
	}
	if got.ValidationURL == "" {
		t.Fatalf("CkRequest.Do() should set CkValidationState.ValidationURL")
	}
	if InputRequestBody != expectedRequest {
		t.Fatalf("CkRequest.Do() should issue '%s' request. Got %s", expectedRequest, InputRequestBody)
	}
	ensureHeaderPresent(t, InputRequest, "Accept", "application/json")
	ensureHeaderPresent(t, InputRequest, "X-Ovh-Application", MockApplicationKey)
}

func TestInvalidCkRequest(t *testing.T) {
	// Init test
	var InputRequest *http.Request
	var InputRequestBody string
	ts, client := initMockServer(&InputRequest, http.StatusForbidden, `{"message":"Invalid application key"}`, &InputRequestBody)
	client.ConsumerKey = ""
	defer ts.Close()

	// Test
	ckRequest := client.NewCkRequest()
	ckRequest.AddRule("GET", "/me")
	ckRequest.AddRule("GET", "/xdsl/*")

	_, err := ckRequest.Do()
	apiError, ok := err.(*APIError)

	// Validate
	if err == nil {
		t.Fatal("Expected an error, got none")
	}
	if !ok {
		t.Fatal("Expected error of type APIError")
	}
	if apiError.Code != http.StatusForbidden {
		t.Fatalf("Expected HTTP error 403. Got %d", apiError.Code)
	}
	if apiError.Message != "Invalid application key" {
		t.Fatalf("Expected API error message 'Invalid application key'. Got '%s'", apiError.Message)
	}
}

func TestAddRules(t *testing.T) {
	// Init test
	var InputRequest *http.Request
	var InputRequestBody string
	ts, client := initMockServer(&InputRequest, http.StatusForbidden, `{"message":"Invalid application key"}`, &InputRequestBody)
	client.ConsumerKey = ""
	defer ts.Close()

	// Test: allow all
	ckRequest := client.NewCkRequest()
	ckRequest.AddRecursiveRules(ReadWrite, "/")
	ExpectedRules := []AccessRule{
		AccessRule{Method: "GET", Path: "/*"},
		AccessRule{Method: "POST", Path: "/*"},
		AccessRule{Method: "PUT", Path: "/*"},
		AccessRule{Method: "DELETE", Path: "/*"},
	}
	if !reflect.DeepEqual(ckRequest.AccessRules, ExpectedRules) {
		t.Fatalf("Inserting recursive RW rules for / should generate %v. Got %v", ExpectedRules, ckRequest.AccessRules)
	}

	// Test: allow exactly /sms, RO
	ckRequest = client.NewCkRequest()
	ckRequest.AddRules(ReadOnly, "/sms")
	ExpectedRules = []AccessRule{
		AccessRule{Method: "GET", Path: "/sms"},
	}
	if !reflect.DeepEqual(ckRequest.AccessRules, ExpectedRules) {
		t.Fatalf("Inserting RO rule for /sms should generate %v. Got %v", ExpectedRules, ckRequest.AccessRules)
	}

	// Test: allow /sms/*, RW, no delete
	ckRequest = client.NewCkRequest()
	ckRequest.AddRecursiveRules(ReadWriteSafe, "/sms")
	ExpectedRules = []AccessRule{
		AccessRule{Method: "GET", Path: "/sms"},
		AccessRule{Method: "POST", Path: "/sms"},
		AccessRule{Method: "PUT", Path: "/sms"},

		AccessRule{Method: "GET", Path: "/sms/*"},
		AccessRule{Method: "POST", Path: "/sms/*"},
		AccessRule{Method: "PUT", Path: "/sms/*"},
	}
	if !reflect.DeepEqual(ckRequest.AccessRules, ExpectedRules) {
		t.Fatalf("Inserting recursive safe RW rule for /sms should generate %v. Got %v", ExpectedRules, ckRequest.AccessRules)
	}
}

func TestCkRequestString(t *testing.T) {
	ckValidationState := &CkValidationState{
		ConsumerKey:   "ck",
		State:         "pending",
		ValidationURL: "fakeURL",
	}

	expected := fmt.Sprintf("CK: \"ck\"\nStatus: \"pending\"\nValidation URL: \"fakeURL\"\n")
	got := fmt.Sprintf("%s", ckValidationState)

	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestCkRequestRedirection(t *testing.T) {
	client, _ := NewClient("endpoint", "appKey", "appSecret", "consumerKey")

	redirection := "http://localhost/api/auth/callback?token=123456"

	ckRequest := client.NewCkRequestWithRedirection(redirection)

	if ckRequest.Redirection != redirection {
		t.Fatalf("NewCkRequestWithRedirection should set ckRequest.Redirection")
	}
}
