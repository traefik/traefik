package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestSticky_StickyHandler(t *testing.T) {
	testCases := []struct {
		desc        string
		handlers    []string
		cookies     []*http.Cookie
		wantHandler string
		wantRewrite bool
	}{
		{
			desc:        "No previous cookie",
			handlers:    []string{"first"},
			wantHandler: "",
			wantRewrite: false,
		},
		{
			desc:     "Wrong previous cookie",
			handlers: []string{"first"},
			cookies: []*http.Cookie{
				{Name: "test", Value: sha256Hash("foo")},
			},
			wantHandler: "",
			wantRewrite: false,
		},
		{
			desc:     "Sha256 previous cookie",
			handlers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: sha256Hash("first")},
			},
			wantHandler: "first",
			wantRewrite: false,
		},
		{
			desc:     "Raw previous cookie",
			handlers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: "first"},
			},
			wantHandler: "first",
			wantRewrite: true,
		},
		{
			desc:     "Fnv previous cookie",
			handlers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: fnvHash("first")},
			},
			wantHandler: "first",
			wantRewrite: true,
		},
		{
			desc:     "Double fnv previous cookie",
			handlers: []string{"first", "second"},
			cookies: []*http.Cookie{
				{Name: "test", Value: fnvHash("first")},
			},
			wantHandler: "first",
			wantRewrite: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sticky := NewStickyCookie(dynamic.Cookie{Name: "test"})

			for _, handler := range test.handlers {
				sticky.AddHandler(handler, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, cookie := range test.cookies {
				req.AddCookie(cookie)
			}

			h, rewrite, err := sticky.StickyHandler(req)
			require.NoError(t, err)

			if test.wantHandler != "" {
				assert.NotNil(t, h)
				assert.Equal(t, test.wantHandler, h.Name)
			} else {
				assert.Nil(t, h)
			}
			assert.Equal(t, test.wantRewrite, rewrite)
		})
	}
}

func TestSticky_WriteStickyCookie(t *testing.T) {
	cookieConfig := dynamic.Cookie{
		Name:     "test",
		Secure:   true,
		HTTPOnly: true,
		SameSite: "none",
		MaxAge:   42,
		Expires:  10,
		Path:     new("/foo"),
		Domain:   "foo.com",
	}
	sticky := NewStickyCookie(cookieConfig)

	// Should return an error if the handler does not exist.
	res := httptest.NewRecorder()
	require.Error(t, sticky.WriteStickyCookie(res, "first"))

	// Should write the sticky cookie and use the sha256 hash.
	sticky.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

	res = httptest.NewRecorder()
	require.NoError(t, sticky.WriteStickyCookie(res, "first"))

	assert.Len(t, res.Result().Cookies(), 1)

	cookie := res.Result().Cookies()[0]

	assert.Equal(t, sha256Hash("first"), cookie.Value)
	assert.Equal(t, "test", cookie.Name)
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite)
	assert.Equal(t, 42, cookie.MaxAge)
	assert.WithinDuration(t, time.Now(), cookie.Expires, time.Duration(cookieConfig.Expires)*time.Second)
	assert.Equal(t, "/foo", cookie.Path)
	assert.Equal(t, "foo.com", cookie.Domain)
}

func TestStickyHeader_StickyHandler(t *testing.T) {
	testCases := []struct {
		desc            string
		handlers        []string
		absoluteTimeout int
		headerName      string
		headerValue     string
		wantHandler     string
		wantRewrite     bool
	}{
		{
			desc:        "No previous header",
			handlers:    []string{"first"},
			headerName:  "",
			headerValue: "",
			wantHandler: "",
			wantRewrite: false,
		},
		{
			desc:        "Wrong previous header value",
			handlers:    []string{"first"},
			headerName:  "X-Sticky-Session",
			headerValue: sha256Hash("foo"),
			wantHandler: "",
			wantRewrite: false,
		},
		{
			desc:        "Sha256 previous header",
			handlers:    []string{"first", "second"},
			headerName:  "X-Sticky-Session",
			headerValue: sha256Hash("first"),
			wantHandler: "first",
			wantRewrite: false,
		},
		{
			desc:        "Raw previous header",
			handlers:    []string{"first", "second"},
			headerName:  "X-Sticky-Session",
			headerValue: "first",
			wantHandler: "first",
			wantRewrite: true,
		},
		{
			desc:        "Fnv previous header",
			handlers:    []string{"first", "second"},
			headerName:  "X-Sticky-Session",
			headerValue: fnvHash("first"),
			wantHandler: "first",
			wantRewrite: true,
		},
		{
			desc:            "AbsoluteTimeout: valid, not yet expired",
			handlers:        []string{"first"},
			absoluteTimeout: 3600,
			headerName:      "X-Sticky-Session",
			headerValue:     sha256Hash("first") + "-" + strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			wantHandler:     "first",
			wantRewrite:     false,
		},
		{
			desc:            "AbsoluteTimeout: expired",
			handlers:        []string{"first"},
			absoluteTimeout: 3600,
			headerName:      "X-Sticky-Session",
			headerValue:     sha256Hash("first") + "-" + strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10),
			wantHandler:     "",
			wantRewrite:     false,
		},
		{
			desc:            "AbsoluteTimeout: no embedded expiry",
			handlers:        []string{"first"},
			absoluteTimeout: 3600,
			headerName:      "X-Sticky-Session",
			headerValue:     sha256Hash("first"),
			wantHandler:     "",
			wantRewrite:     false,
		},
		{
			desc:            "AbsoluteTimeout: malformed expiry",
			handlers:        []string{"first"},
			absoluteTimeout: 3600,
			headerName:      "X-Sticky-Session",
			headerValue:     sha256Hash("first") + "-not-a-timestamp",
			wantHandler:     "",
			wantRewrite:     false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sticky := NewStickyHeader(dynamic.Header{Name: "X-Sticky-Session", AbsoluteTimeout: test.absoluteTimeout})

			for _, handler := range test.handlers {
				sticky.AddHandler(handler, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.headerName != "" {
				req.Header.Set(test.headerName, test.headerValue)
			}

			h, rewrite, err := sticky.StickyHandler(req)
			require.NoError(t, err)

			if test.wantHandler != "" {
				assert.NotNil(t, h)
				assert.Equal(t, test.wantHandler, h.Name)
			} else {
				assert.Nil(t, h)
			}
			assert.Equal(t, test.wantRewrite, rewrite)
		})
	}
}

func TestStickyHeader_WriteStickyHeader(t *testing.T) {
	sticky := NewStickyHeader(dynamic.Header{Name: "X-Sticky-Session"})

	// Should return an error if the handler does not exist.
	res := httptest.NewRecorder()
	require.Error(t, sticky.WriteStickyHeader(res, "first"))

	// Should write the sticky header and use the sha256 hash.
	sticky.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

	res = httptest.NewRecorder()
	require.NoError(t, sticky.WriteStickyHeader(res, "first"))

	assert.Equal(t, sha256Hash("first"), res.Header().Get("X-Sticky-Session"))

	// When AbsoluteTimeout is set, the written value should embed an expiry, not the bare hash.
	stickyWithTimeout := NewStickyHeader(dynamic.Header{Name: "X-Sticky-Session", AbsoluteTimeout: 3600})
	stickyWithTimeout.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

	res = httptest.NewRecorder()
	require.NoError(t, stickyWithTimeout.WriteStickyHeader(res, "first"))

	assert.NotEqual(t, sha256Hash("first"), res.Header().Get("X-Sticky-Session"))
}

func TestStickyHeader_DefaultName(t *testing.T) {
	// When no name is provided, should use the default "X-Sticky-Session".
	sticky := NewStickyHeader(dynamic.Header{})

	sticky.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

	res := httptest.NewRecorder()
	require.NoError(t, sticky.WriteStickyHeader(res, "first"))

	assert.Equal(t, sha256Hash("first"), res.Header().Get("X-Sticky-Session"))
}

func TestNewSticky(t *testing.T) {
	testCases := []struct {
		desc     string
		config   *dynamic.Sticky
		wantNil  bool
		wantMode StickyMode
	}{
		{
			desc:    "nil config",
			config:  nil,
			wantNil: true,
		},
		{
			desc:    "empty config",
			config:  &dynamic.Sticky{},
			wantNil: true,
		},
		{
			desc: "cookie config",
			config: &dynamic.Sticky{
				Cookie: &dynamic.Cookie{Name: "test"},
			},
			wantNil:  false,
			wantMode: StickyModeCookie,
		},
		{
			desc: "header config",
			config: &dynamic.Sticky{
				Header: &dynamic.Header{Name: "X-Test"},
			},
			wantNil:  false,
			wantMode: StickyModeHeader,
		},
		{
			desc: "both configs - cookie takes precedence",
			config: &dynamic.Sticky{
				Cookie: &dynamic.Cookie{Name: "test"},
				Header: &dynamic.Header{Name: "X-Test"},
			},
			wantNil:  false,
			wantMode: StickyModeCookie,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sticky := NewSticky(test.config)

			if test.wantNil {
				assert.Nil(t, sticky)
			} else {
				require.NotNil(t, sticky)
				assert.Equal(t, test.wantMode, sticky.mode)
			}
		})
	}
}

func TestSticky_WriteStickyResponse(t *testing.T) {
	t.Run("cookie mode", func(t *testing.T) {
		sticky := NewSticky(&dynamic.Sticky{
			Cookie: &dynamic.Cookie{Name: "test"},
		})

		sticky.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

		res := httptest.NewRecorder()
		require.NoError(t, sticky.WriteStickyResponse(res, "first"))

		// Should write a cookie.
		assert.Len(t, res.Result().Cookies(), 1)
		assert.Equal(t, sha256Hash("first"), res.Result().Cookies()[0].Value)
	})

	t.Run("header mode", func(t *testing.T) {
		sticky := NewSticky(&dynamic.Sticky{
			Header: &dynamic.Header{Name: "X-Sticky"},
		})

		sticky.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

		res := httptest.NewRecorder()
		require.NoError(t, sticky.WriteStickyResponse(res, "first"))

		// Should write a header.
		assert.Equal(t, sha256Hash("first"), res.Header().Get("X-Sticky"))
		// Should not write a cookie.
		assert.Empty(t, res.Result().Cookies())
	})
}

func TestConvertSameSite_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected http.SameSite
	}{
		{"none", http.SameSiteNoneMode},
		{"None", http.SameSiteNoneMode},
		{"NONE", http.SameSiteNoneMode},
		{"lax", http.SameSiteLaxMode},
		{"Lax", http.SameSiteLaxMode},
		{"LAX", http.SameSiteLaxMode},
		{"strict", http.SameSiteStrictMode},
		{"Strict", http.SameSiteStrictMode},
		{"STRICT", http.SameSiteStrictMode},
		{"", http.SameSiteDefaultMode},
		{"invalid", http.SameSiteDefaultMode},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, convertSameSite(tt.input))
		})
	}
}
