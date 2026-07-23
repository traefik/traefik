package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
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

			sticky := NewSticky(dynamic.Cookie{Name: "test"})

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
	sticky := NewSticky(cookieConfig)

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
	assert.WithinDuration(t, time.Now().Add(time.Duration(cookieConfig.Expires)*time.Second), cookie.Expires, time.Second)
	assert.Equal(t, "/foo", cookie.Path)
	assert.Equal(t, "foo.com", cookie.Domain)
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

func TestSticky_WriteStickyCookie_ExpiresAdvancesPerRequest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		expiration := time.Hour
		sticky := NewSticky(dynamic.Cookie{Name: "test", Expires: int(expiration / time.Second)})
		sticky.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

		wantExp1 := time.Now().Add(expiration)
		res1 := httptest.NewRecorder()
		require.NoError(t, sticky.WriteStickyCookie(res1, "first"))
		exp1 := res1.Result().Cookies()[0].Expires
		require.WithinDuration(t, wantExp1, exp1, time.Second)

		// Advance fake time beyond HTTP-date's one-second granularity.
		time.Sleep(2 * time.Second)

		wantExp2 := time.Now().Add(expiration)
		res2 := httptest.NewRecorder()
		require.NoError(t, sticky.WriteStickyCookie(res2, "first"))
		exp2 := res2.Result().Cookies()[0].Expires
		require.WithinDuration(t, wantExp2, exp2, time.Second)

		require.True(t, exp2.After(exp1))
	})
}

func TestSticky_WriteStickyCookie_NoExpiresIsSessionCookie(t *testing.T) {
	sticky := NewSticky(dynamic.Cookie{Name: "test"})
	sticky.AddHandler("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

	res := httptest.NewRecorder()
	require.NoError(t, sticky.WriteStickyCookie(res, "first"))

	// Without an expiry configured the cookie stays a session cookie, so the
	// per-request Expires branch must not add an Expires attribute.
	assert.True(t, res.Result().Cookies()[0].Expires.IsZero())
}
