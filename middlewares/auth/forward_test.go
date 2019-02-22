package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/forward"
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
	}, &tracing.Tracing{})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(middleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusForbidden, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "Forbidden\n", string(body), "they should be equal")
}

func TestForwardAuthSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-User", "user@example.com")
		w.Header().Add("X-Auth-Group", "group1")
		w.Header().Add("X-Auth-Group", "group2")
		w.Header().Set("X-Auth-Secret", "secret")
		fmt.Fprintln(w, "Success")
	}))
	defer server.Close()

	middleware, err := NewAuthenticator(&types.Auth{
		Forward: &types.Forward{
			Address:             server.URL,
			AuthResponseHeaders: []string{"X-Auth-User", "X-Auth-Group"},
		},
	}, &tracing.Tracing{})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "user@example.com", r.Header.Get("X-Auth-User"))
		assert.Equal(t, []string{"group1", "group2"}, r.Header["X-Auth-Group"])
		assert.Empty(t, r.Header.Get("X-Auth-Secret"))
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(middleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("X-Auth-Group", "admin_group")
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func TestForwardAuthRedirect(t *testing.T) {
	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://example.com/redirect-test", http.StatusFound)
	}))
	defer authTs.Close()

	authMiddleware, err := NewAuthenticator(&types.Auth{
		Forward: &types.Forward{
			Address: authTs.URL,
		},
	}, &tracing.Tracing{})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusFound, res.StatusCode, "they should be equal")

	location, err := res.Location()
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "http://example.com/redirect-test", location.String(), "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.NotEmpty(t, string(body), "there should be something in the body")
}

func TestForwardAuthRemoveHopByHopHeaders(t *testing.T) {
	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		for _, header := range forward.HopHeaders {
			if header == forward.TransferEncoding {
				headers.Add(header, "identity")
			} else {
				headers.Add(header, "test")
			}
		}

		http.Redirect(w, r, "http://example.com/redirect-test", http.StatusFound)
	}))
	defer authTs.Close()

	authMiddleware, err := NewAuthenticator(&types.Auth{
		Forward: &types.Forward{
			Address: authTs.URL,
		},
	}, &tracing.Tracing{})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusFound, res.StatusCode, "they should be equal")

	for _, header := range forward.HopHeaders {
		assert.Equal(t, "", res.Header.Get(header), "hop-by-hop header '%s' mustn't be set", header)
	}

	location, err := res.Location()
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "http://example.com/redirect-test", location.String(), "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.NotEmpty(t, string(body), "there should be something in the body")
}

func TestForwardAuthFailResponseHeaders(t *testing.T) {
	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := &http.Cookie{Name: "example", Value: "testing", Path: "/"}
		http.SetCookie(w, cookie)
		w.Header().Add("X-Foo", "bar")
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer authTs.Close()

	authMiddleware, err := NewAuthenticator(&types.Auth{
		Forward: &types.Forward{
			Address: authTs.URL,
		},
	}, &tracing.Tracing{})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	client := &http.Client{}
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusForbidden, res.StatusCode, "they should be equal")

	require.Len(t, res.Cookies(), 1)
	for _, cookie := range res.Cookies() {
		assert.Equal(t, "testing", cookie.Value, "they should be equal")
	}

	expectedHeaders := http.Header{
		"Content-Length":         []string{"10"},
		"Content-Type":           []string{"text/plain; charset=utf-8"},
		"X-Foo":                  []string{"bar"},
		"Set-Cookie":             []string{"example=testing; Path=/"},
		"X-Content-Type-Options": []string{"nosniff"},
	}

	assert.Len(t, res.Header, 6)
	for key, value := range expectedHeaders {
		assert.Equal(t, value, res.Header[key])
	}

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "Forbidden\n", string(body), "they should be equal")
}

func Test_writeHeader(t *testing.T) {
	testCases := []struct {
		name                      string
		headers                   map[string]string
		trustForwardHeader        bool
		emptyHost                 bool
		expectedHeaders           map[string]string
		checkForUnexpectedHeaders bool
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
		{
			name: "trust Forward Header with forwarded URI",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
				"X-Forwarded-Uri":  "/forward?q=1",
			},
			trustForwardHeader: true,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
				"X-Forwarded-Uri":  "/forward?q=1",
			},
		},
		{
			name: "not trust Forward Header with forward requested URI",
			headers: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "fii.bir",
				"X-Forwarded-Uri":  "/forward?q=1",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"Accept":           "application/json",
				"X-Forwarded-Host": "foo.bar",
				"X-Forwarded-Uri":  "/path?q=1",
			},
		}, {
			name: "trust Forward Header with forwarded request Method",
			headers: map[string]string{
				"X-Forwarded-Method": "OPTIONS",
			},
			trustForwardHeader: true,
			expectedHeaders: map[string]string{
				"X-Forwarded-Method": "OPTIONS",
			},
		},
		{
			name: "not trust Forward Header with forward request Method",
			headers: map[string]string{
				"X-Forwarded-Method": "OPTIONS",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"X-Forwarded-Method": "GET",
			},
		},
		{
			name: "remove hop-by-hop headers",
			headers: map[string]string{
				forward.Connection:         "Connection",
				forward.KeepAlive:          "KeepAlive",
				forward.ProxyAuthenticate:  "ProxyAuthenticate",
				forward.ProxyAuthorization: "ProxyAuthorization",
				forward.Te:                 "Te",
				forward.Trailers:           "Trailers",
				forward.TransferEncoding:   "TransferEncoding",
				forward.Upgrade:            "Upgrade",
				"X-CustomHeader":           "CustomHeader",
			},
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				"X-CustomHeader":     "CustomHeader",
				"X-Forwarded-Proto":  "http",
				"X-Forwarded-Host":   "foo.bar",
				"X-Forwarded-Uri":    "/path?q=1",
				"X-Forwarded-Method": "GET",
			},
			checkForUnexpectedHeaders: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/path?q=1", nil)
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			if test.emptyHost {
				req.Host = ""
			}

			forwardReq := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/path?q=1", nil)

			writeHeader(req, forwardReq, test.trustForwardHeader)

			actualHeaders := forwardReq.Header
			expectedHeaders := test.expectedHeaders
			for key, value := range expectedHeaders {
				assert.Equal(t, value, actualHeaders.Get(key))
				actualHeaders.Del(key)
			}
			if test.checkForUnexpectedHeaders {
				for key := range actualHeaders {
					assert.Fail(t, "Unexpected header found", key)
				}
			}
		})
	}
}
