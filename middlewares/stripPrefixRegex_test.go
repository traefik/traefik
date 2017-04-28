package middlewares

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStripPrefixRegex(t *testing.T) {

	handlerPath := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Path)
	})

	handler := NewStripPrefixRegex(handlerPath, []string{"/a/api/", "/b/{regex}/", "/c/{category}/{id:[0-9]+}/"})
	server := httptest.NewServer(handler)
	defer server.Close()

	tests := []struct {
		expectedCode     int
		expectedResponse string
		url              string
	}{
		{url: "/a/test", expectedCode: 404, expectedResponse: "404 page not found\n"},
		{url: "/a/api/test", expectedCode: 200, expectedResponse: "test"},

		{url: "/b/api/", expectedCode: 200, expectedResponse: ""},
		{url: "/b/api/test1", expectedCode: 200, expectedResponse: "test1"},
		{url: "/b/api2/test2", expectedCode: 200, expectedResponse: "test2"},

		{url: "/c/api/123/", expectedCode: 200, expectedResponse: ""},
		{url: "/c/api/123/test3", expectedCode: 200, expectedResponse: "test3"},
		{url: "/c/api/abc/test4", expectedCode: 404, expectedResponse: "404 page not found\n"},
	}

	for _, test := range tests {
		resp, err := http.Get(server.URL + test.url)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != test.expectedCode {
			t.Fatalf("Received non-%d response: %d\n", test.expectedCode, resp.StatusCode)
		}
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if test.expectedResponse != string(response) {
			t.Errorf("Expected '%s' :  '%s'\n", test.expectedResponse, response)
		}
	}

}
