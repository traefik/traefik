package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/negroni"
)

func TestForwardAuthFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	middleware, err := NewAuthenticator(&types.Auth{
		Forward: &types.Forward{
			Address: server.URL,
		},
	})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(middleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusForbidden, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "Forbidden\n", string(body), "they should be equal")
}

func TestForwardAuthSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Success")
	}))
	defer server.Close()

	middleware, err := NewAuthenticator(&types.Auth{
		Forward: &types.Forward{
			Address: server.URL,
		},
	})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(middleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func Test_writeHeader(t *testing.T) {

	testCases := []struct {
		name               string
		headers            map[string]string
		trustForwardHeader bool
		emptyHost          bool
		expectedHeaders    map[string]string
	}{
		{
			name: "trust Forward Header",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
		},
		{
			name: "not trust Forward Header",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "foo.bar",
			},
		},
		{
			name: "trust Forward Header with empty Host",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: true,
			emptyHost:          true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
		},
		{
			name: "not trust Forward Header with empty Host",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
			},
			trustForwardHeader: false,
			emptyHost:          true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			if test.emptyHost {
				req.Host = ""
			}

			forwardReq := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)

			writeHeader(req, forwardReq, test.trustForwardHeader)

			for key, value := range test.expectedHeaders {
				assert.Equal(t, value, forwardReq.Header.Get(key))
			}
		})
	}
}
