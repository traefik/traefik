package dynamic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/mitchellh/hashstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestJSONPointer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		tokens   []string
		expected string
	}{
		{
			desc:     "no token is the whole document",
			expected: "",
		},
		{
			desc:     "tokens are joined with slashes",
			tokens:   []string{"plugin", "my-plugin", "0"},
			expected: "/plugin/my-plugin/0",
		},
		{
			desc:     "slashes and tildes in tokens are escaped",
			tokens:   []string{"a/b", "m~n", "~1"},
			expected: "/a~1b/m~0n/~01",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, JSONPointer(test.tokens...))
		})
	}
}

func TestMiddleware_ResolveURNRefs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc        string
		middleware  *Middleware
		expected    *Middleware
		expectedErr string
	}{
		{
			desc: "nested plugin values are set",
			middleware: &Middleware{
				Plugin: map[string]PluginConf{
					"my-plugin": {
						"token": "urn:k8s:secret:creds:token",
						"nested": map[string]any{
							"a/b":  "urn:k8s:secret:creds:a",
							"list": []any{"keep", "urn:k8s:secret:creds:b"},
						},
						"count": float64(1),
					},
				},
				URNRefs: map[string]string{
					"/plugin/my-plugin/token":         "token-value",
					"/plugin/my-plugin/nested/a~1b":   "a-value",
					"/plugin/my-plugin/nested/list/1": "b-value",
				},
			},
			expected: &Middleware{
				Plugin: map[string]PluginConf{
					"my-plugin": {
						"token": "token-value",
						"nested": map[string]any{
							"a/b":  "a-value",
							"list": []any{"keep", "b-value"},
						},
						"count": float64(1),
					},
				},
				URNRefs: map[string]string{
					"/plugin/my-plugin/token":         "token-value",
					"/plugin/my-plugin/nested/a~1b":   "a-value",
					"/plugin/my-plugin/nested/list/1": "b-value",
				},
			},
		},
		{
			desc: "escaped tilde followed by 1 is decoded once",
			middleware: &Middleware{
				Plugin:  map[string]PluginConf{"my-plugin": {"~1": "urn:k8s:secret:creds:token"}},
				URNRefs: map[string]string{"/plugin/my-plugin/~01": "token-value"},
			},
			expected: &Middleware{
				Plugin:  map[string]PluginConf{"my-plugin": {"~1": "token-value"}},
				URNRefs: map[string]string{"/plugin/my-plugin/~01": "token-value"},
			},
		},
		{
			desc: "users, TLS key and Redis credentials are set",
			middleware: &Middleware{
				// The provider splits one Secret key into one user per line, so the users share the URN of the key.
				BasicAuth:   &BasicAuth{Users: Users{"urn:k8s:secret:auth:users", "urn:k8s:secret:auth:users"}, Realm: "realm"},
				ForwardAuth: &ForwardAuth{Address: "http://auth", TLS: &ClientTLS{CA: "ca", Key: "urn:k8s:secret:tls:tls.key"}},
				RateLimit: &RateLimit{Redis: &Redis{
					Password: "urn:k8s:secret:redis:password",
					TLS:      &types.ClientTLS{Key: "urn:k8s:secret:redis-tls:tls.key"},
				}},
				URNRefs: map[string]string{
					"/basicAuth/users/0":        "user1:hash1",
					"/basicAuth/users/1":        "user2:hash2",
					"/forwardAuth/tls/key":      "key",
					"/rateLimit/redis/password": "password",
					"/rateLimit/redis/tls/key":  "redis-key",
				},
			},
			expected: &Middleware{
				BasicAuth:   &BasicAuth{Users: Users{"user1:hash1", "user2:hash2"}, Realm: "realm"},
				ForwardAuth: &ForwardAuth{Address: "http://auth", TLS: &ClientTLS{CA: "ca", Key: "key"}},
				RateLimit: &RateLimit{Redis: &Redis{
					Password: "password",
					TLS:      &types.ClientTLS{Key: "redis-key"},
				}},
				URNRefs: map[string]string{
					"/basicAuth/users/0":        "user1:hash1",
					"/basicAuth/users/1":        "user2:hash2",
					"/forwardAuth/tls/key":      "key",
					"/rateLimit/redis/password": "password",
					"/rateLimit/redis/tls/key":  "redis-key",
				},
			},
		},
		{
			desc: "pointer without leading slash",
			middleware: &Middleware{
				BasicAuth: &BasicAuth{Users: Users{"urn"}},
				URNRefs:   map[string]string{"basicAuth/users/0": "value"},
			},
			expectedErr: `resolving URN reference "basicAuth/users/0": JSON Pointer must start with a slash`,
		},
		{
			desc: "unknown field",
			middleware: &Middleware{
				BasicAuth: &BasicAuth{Users: Users{"urn"}},
				URNRefs:   map[string]string{"/basicAuth/unknown": "value"},
			},
			expectedErr: `resolving URN reference "/basicAuth/unknown": unknown field "unknown"`,
		},
		{
			desc: "nil value on the path",
			middleware: &Middleware{
				ForwardAuth: &ForwardAuth{Address: "http://auth"},
				URNRefs:     map[string]string{"/forwardAuth/tls/key": "value"},
			},
			expectedErr: `resolving URN reference "/forwardAuth/tls/key": nil value on the path`,
		},
		{
			desc: "unknown map key",
			middleware: &Middleware{
				Plugin:  map[string]PluginConf{"my-plugin": {"token": "urn"}},
				URNRefs: map[string]string{"/plugin/my-plugin/other": "value"},
			},
			expectedErr: `resolving URN reference "/plugin/my-plugin/other": unknown key "other"`,
		},
		{
			desc: "array index with a leading zero",
			middleware: &Middleware{
				BasicAuth: &BasicAuth{Users: Users{"urn"}},
				URNRefs:   map[string]string{"/basicAuth/users/00": "value"},
			},
			expectedErr: `resolving URN reference "/basicAuth/users/00": invalid array index "00"`,
		},
		{
			desc: "array index out of range",
			middleware: &Middleware{
				BasicAuth: &BasicAuth{Users: Users{"urn"}},
				URNRefs:   map[string]string{"/basicAuth/users/1": "value"},
			},
			expectedErr: `resolving URN reference "/basicAuth/users/1": invalid array index "1"`,
		},
		{
			desc: "value that is not a string",
			middleware: &Middleware{
				Plugin:  map[string]PluginConf{"my-plugin": {"count": float64(1)}},
				URNRefs: map[string]string{"/plugin/my-plugin/count": "value"},
			},
			expectedErr: `resolving URN reference "/plugin/my-plugin/count": value of type float64 is not a string`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			original := test.middleware.DeepCopy()

			resolved, err := test.middleware.ResolveURNRefs()

			assert.Equal(t, original, test.middleware)

			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, resolved)
		})
	}
}

func TestMiddleware_ResolveURNRefs_withoutRefs(t *testing.T) {
	t.Parallel()

	middleware := &Middleware{BasicAuth: &BasicAuth{Users: Users{"user:hash"}}}

	resolved, err := middleware.ResolveURNRefs()
	require.NoError(t, err)

	assert.Same(t, middleware, resolved)
}

func Test_swapStringAt(t *testing.T) {
	t.Parallel()

	type item struct {
		Name  string `json:"name"`
		NoTag string
	}

	testCases := []struct {
		desc        string
		value       any
		pointer     string
		expected    any
		expectedOld string
		expectedErr string
	}{
		{
			desc:        "struct held by an interface is set",
			value:       map[string]any{"item": item{Name: "urn"}},
			pointer:     "/item/name",
			expected:    map[string]any{"item": item{Name: "value"}},
			expectedOld: "urn",
		},
		{
			desc:        "array index with a plus sign",
			value:       &[]string{"urn", "urn"},
			pointer:     "/+1",
			expectedErr: `invalid array index "+1"`,
		},
		{
			desc:        "empty token does not match a field without JSON name",
			value:       &item{},
			pointer:     "/",
			expectedErr: `unknown field ""`,
		},
		{
			desc:        "map key that is not a string",
			value:       map[int]string{1: "urn"},
			pointer:     "/1",
			expectedErr: "unsupported map key of type int on the path",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tokens, err := parseJSONPointer(test.pointer)
			require.NoError(t, err)

			old, err := swapStringAt(reflect.ValueOf(test.value), tokens, "value")
			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedOld, old)
			assert.Equal(t, test.expected, test.value)
		})
	}
}

func TestMiddleware_RedactURNRefs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		pluginConf PluginConf
		urnRefs    map[string]string
		err        error
		expected   string
	}{
		{
			desc:       "double quoted value is replaced with its double quoted URN",
			pluginConf: PluginConf{"token": "urn:k8s:secret:creds:token"},
			urnRefs:    map[string]string{"/plugin/my-plugin/token": "s3cr3t"},
			err:        errors.New(`strconv.ParseInt: parsing "s3cr3t": invalid syntax`),
			expected:   `strconv.ParseInt: parsing "urn:k8s:secret:creds:token": invalid syntax`,
		},
		{
			desc:       "double quoted value with escaped characters is replaced with its double quoted URN",
			pluginConf: PluginConf{"token": "urn:k8s:secret:creds:token"},
			urnRefs:    map[string]string{"/plugin/my-plugin/token": `s3"cr\3t`},
			err:        fmt.Errorf("parsing %q: invalid syntax", `s3"cr\3t`),
			expected:   `parsing "urn:k8s:secret:creds:token": invalid syntax`,
		},
		{
			desc:       "single quoted value is replaced with its single quoted URN",
			pluginConf: PluginConf{"token": "urn:k8s:secret:creds:token"},
			urnRefs:    map[string]string{"/plugin/my-plugin/token": "s3cr3t"},
			err:        errors.New("'token' expected type 'int', got unconvertible type 'string', value: 's3cr3t'"),
			expected:   "'token' expected type 'int', got unconvertible type 'string', value: 'urn:k8s:secret:creds:token'",
		},
		{
			desc:       "value that is not quoted is not replaced",
			pluginConf: PluginConf{"password": "urn:k8s:secret:redis:password"},
			urnRefs:    map[string]string{"/plugin/my-plugin/password": "redis"},
			err:        errors.New("dial tcp: lookup redis: no such host"),
			expected:   "dial tcp: lookup redis: no such host",
		},
		{
			desc:       "value that contains another value is replaced first",
			pluginConf: PluginConf{"a": "urn:k8s:secret:creds:a", "b": "urn:k8s:secret:creds:b"},
			urnRefs:    map[string]string{"/plugin/my-plugin/a": "ab", "/plugin/my-plugin/b": "ab'c"},
			err:        errors.New("value: 'ab'c'"),
			expected:   "value: 'urn:k8s:secret:creds:b'",
		},
		{
			desc:     "empty value is ignored",
			urnRefs:  map[string]string{"/a": ""},
			err:      errors.New("error"),
			expected: "error",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			middleware := &Middleware{URNRefs: test.urnRefs}
			if test.pluginConf != nil {
				middleware.Plugin = map[string]PluginConf{"my-plugin": test.pluginConf}
			}

			assert.EqualError(t, middleware.RedactURNRefs(test.err), test.expected)
		})
	}
}

func TestMiddleware_RedactURNRefs_withoutValue(t *testing.T) {
	t.Parallel()

	middleware := &Middleware{URNRefs: map[string]string{"/basicAuth/users/0": "user:hash"}}
	err := fmt.Errorf("building: %w", errors.ErrUnsupported)

	assert.Same(t, err, middleware.RedactURNRefs(err))
}

func TestMiddleware_URNRefs_json(t *testing.T) {
	t.Parallel()

	middleware := &Middleware{
		BasicAuth: &BasicAuth{Users: Users{"urn:k8s:secret:auth:users"}},
		URNRefs:   map[string]string{"/basicAuth/users/0": "user:hash"},
	}

	data, err := json.Marshal(middleware)
	require.NoError(t, err)

	assert.JSONEq(t, `{"basicAuth":{"users":["urn:k8s:secret:auth:users"]}}`, string(data))
}

func TestMiddleware_URNRefs_hash(t *testing.T) {
	t.Parallel()

	hash := func(value string) uint64 {
		h, err := hashstructure.Hash(&Middleware{URNRefs: map[string]string{"/basicAuth/users/0": value}}, nil)
		require.NoError(t, err)

		return h
	}

	// The providers skip a configuration with an unchanged hash, so a new Secret value must change the hash.
	assert.NotEqual(t, hash("user:hash1"), hash("user:hash2"))
}
