package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func pointer[T any](v T) *T { return &v }

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
	sticky := NewSticky(dynamic.Cookie{
		Name:     "test",
		Secure:   true,
		HTTPOnly: true,
		SameSite: "none",
		MaxAge:   42,
		Path:     pointer("/foo"),
		Domain:   "foo.com",
	})

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
	assert.Equal(t, "/foo", cookie.Path)
	assert.Equal(t, "foo.com", cookie.Domain)
}
