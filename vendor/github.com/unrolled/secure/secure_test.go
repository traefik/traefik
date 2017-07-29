package secure

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("bar"))
})

func TestNoConfig(t *testing.T) {
	s := New()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")
}

func TestNoAllowHosts(t *testing.T) {
	s := New(Options{
		AllowedHosts: []string{},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestGoodSingleAllowHosts(t *testing.T) {
	s := New(Options{
		AllowedHosts: []string{"www.example.com"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestBadSingleAllowHosts(t *testing.T) {
	s := New(Options{
		AllowedHosts: []string{"sub.example.com"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
}

func TestGoodSingleAllowHostsProxyHeaders(t *testing.T) {
	s := New(Options{
		AllowedHosts:      []string{"www.example.com"},
		HostsProxyHeaders: []string{"X-Proxy-Host"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "example-internal"
	req.Header.Set("X-Proxy-Host", "www.example.com")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestBadSingleAllowHostsProxyHeaders(t *testing.T) {
	s := New(Options{
		AllowedHosts:      []string{"sub.example.com"},
		HostsProxyHeaders: []string{"X-Proxy-Host"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "example-internal"
	req.Header.Set("X-Proxy-Host", "www.example.com")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
}

func TestGoodMultipleAllowHosts(t *testing.T) {
	s := New(Options{
		AllowedHosts: []string{"www.example.com", "sub.example.com"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "sub.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestBadMultipleAllowHosts(t *testing.T) {
	s := New(Options{
		AllowedHosts: []string{"www.example.com", "sub.example.com"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www3.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
}

func TestAllowHostsInDevMode(t *testing.T) {
	s := New(Options{
		AllowedHosts:  []string{"www.example.com", "sub.example.com"},
		IsDevelopment: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www3.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestBadHostHandler(t *testing.T) {
	s := New(Options{
		AllowedHosts: []string{"www.example.com", "sub.example.com"},
	})

	badHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "BadHost", http.StatusInternalServerError)
	})

	s.SetBadHostHandler(badHandler)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www3.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	// http.Error outputs a new line character with the response.
	expect(t, res.Body.String(), "BadHost\n")
}

func TestSSL(t *testing.T) {
	s := New(Options{
		SSLRedirect: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestSSLInDevMode(t *testing.T) {
	s := New(Options{
		SSLRedirect:   true,
		IsDevelopment: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestBasicSSL(t *testing.T) {
	s := New(Options{
		SSLRedirect: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://www.example.com/foo")
}

func TestBasicSSLWithHost(t *testing.T) {
	s := New(Options{
		SSLRedirect: true,
		SSLHost:     "secure.example.com",
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://secure.example.com/foo")
}

func TestBadProxySSL(t *testing.T) {
	s := New(Options{
		SSLRedirect: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://www.example.com/foo")
}

func TestCustomProxySSL(t *testing.T) {
	s := New(Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestCustomProxySSLInDevMode(t *testing.T) {
	s := New(Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		IsDevelopment:   true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "http")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestCustomProxyAndHostProxyHeadersWithRedirect(t *testing.T) {
	s := New(Options{
		HostsProxyHeaders: []string{"X-Forwarded-Host"},
		SSLRedirect:       true,
		SSLProxyHeaders:   map[string]string{"X-Forwarded-Proto": "http"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "example-internal"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")
	req.Header.Add("X-Forwarded-Host", "www.example.com")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://www.example.com/foo")
}

func TestCustomProxyAndHostSSL(t *testing.T) {
	s := New(Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		SSLHost:         "secure.example.com",
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestCustomBadProxyAndHostSSL(t *testing.T) {
	s := New(Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "superman"},
		SSLHost:         "secure.example.com",
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://secure.example.com/foo")
}

func TestCustomBadProxyAndHostSSLWithTempRedirect(t *testing.T) {
	s := New(Options{
		SSLRedirect:          true,
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "superman"},
		SSLHost:              "secure.example.com",
		SSLTemporaryRedirect: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusTemporaryRedirect)
	expect(t, res.Header().Get("Location"), "https://secure.example.com/foo")
}

func TestStsHeaderWithNoSSL(t *testing.T) {
	s := New(Options{
		STSSeconds: 315360000,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "")
}

func TestStsHeaderWithNoSSLButWithForce(t *testing.T) {
	s := New(Options{
		STSSeconds:     315360000,
		ForceSTSHeader: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000")
}

func TestStsHeaderWithNoSSLButWithForceAndIsDev(t *testing.T) {
	s := New(Options{
		STSSeconds:     315360000,
		ForceSTSHeader: true,
		IsDevelopment:  true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "")
}

func TestStsHeaderWithSSL(t *testing.T) {
	s := New(Options{
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:      315360000,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000")
}

func TestStsHeaderInDevMode(t *testing.T) {
	s := New(Options{
		STSSeconds:    315360000,
		IsDevelopment: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "")
}

func TestStsHeaderWithSubdomains(t *testing.T) {
	s := New(Options{
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000; includeSubdomains")
}

func TestStsHeaderWithPreload(t *testing.T) {
	s := New(Options{
		STSSeconds: 315360000,
		STSPreload: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000; preload")
}

func TestStsHeaderWithSubdomainsWithPreload(t *testing.T) {
	s := New(Options{
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000; includeSubdomains; preload")
}

func TestFrameDeny(t *testing.T) {
	s := New(Options{
		FrameDeny: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "DENY")
}

func TestCustomFrameValue(t *testing.T) {
	s := New(Options{
		CustomFrameOptionsValue: "SAMEORIGIN",
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "SAMEORIGIN")
}

func TestCustomFrameValueWithDeny(t *testing.T) {
	s := New(Options{
		FrameDeny:               true,
		CustomFrameOptionsValue: "SAMEORIGIN",
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "SAMEORIGIN")
}

func TestContentNosniff(t *testing.T) {
	s := New(Options{
		ContentTypeNosniff: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Content-Type-Options"), "nosniff")
}

func TestXSSProtection(t *testing.T) {
	s := New(Options{
		BrowserXssFilter: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-XSS-Protection"), "1; mode=block")
}

func TestCsp(t *testing.T) {
	s := New(Options{
		ContentSecurityPolicy: "default-src 'self'",
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func TestInlineSecure(t *testing.T) {
	s := New(Options{
		FrameDeny: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.HandlerFuncWithNext(w, r, nil)
		w.Write([]byte("bar"))
	})

	handler.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "DENY")
}

// https://developer.mozilla.org/en-US/docs/Web/Security/Public_Key_Pinning
const hpkp = `pin-sha256="cUPcTAZWKaASuYWhhneDttWpY3oBAkE3h2+soZS7sWs="; pin-sha256="M8HztCzM3elUxkcjR2S5P4hhyBNf6lHkmjAHKhpGPWE="; max-age=5184000; includeSubdomains; report-uri="https://www.example.net/hpkp-report"`

func TestHPKP(t *testing.T) {
	s := New(Options{
		PublicKey: hpkp,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Public-Key-Pins"), hpkp)
}

func TestHPKPNotSet(t *testing.T) {
	s := New()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Public-Key-Pins"), "")
}

func TestHPKPInDevMode(t *testing.T) {
	s := New(Options{
		PublicKey:     hpkp,
		IsDevelopment: true,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Public-Key-Pins"), "")
}

func TestHPKPNonSSL(t *testing.T) {
	s := New(Options{
		PublicKey: hpkp,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Public-Key-Pins"), "")
}

func TestReferrer(t *testing.T) {
	s := New(Options{
		ReferrerPolicy: "same-origin",
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Referrer-Policy"), "same-origin")
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected [%v] (type %v) - Got [%v] (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
