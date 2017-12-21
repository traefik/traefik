package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/containous/traefik/types"
)

func TestStickyBackendHandler(t *testing.T) {
	const cookieName = "traefikSticky"
	const numRequests = 2
	const requestCountHeader = "X-Request-Count"

	tests := []struct {
		amountServer int
		seedCookie   string
		headers      map[string]string
		stickiness   *types.Stickiness
		targets      []string
	}{
		{
			amountServer: 2,
			stickiness:   &types.Stickiness{},
		},
		{
			amountServer: 2,
			stickiness: &types.Stickiness{
				CookieEncryptKey: "abc123",
			},
		},
		{
			amountServer: 3,
			stickiness: &types.Stickiness{
				IP: true,
			},
		},
		{
			amountServer: 5,
			headers: map[string]string{
				"X-Forwarded-For": "172.17.0.1,192.168.1.1",
			},
			stickiness: &types.Stickiness{
				IP: true,
			},
		},
		{
			amountServer: 3,
			headers: map[string]string{
				"X-Backend": "abc123",
			},
			stickiness: &types.Stickiness{
				Rules: []string{
					`{{index .Header "X-Notfound" | join ""}}`,
					`{{index .Header "X-Backend" | join ""}}`,
				},
			},
		},
		{
			amountServer: 8,
			targets: []string{
				"http://localhost/isbn/12345/author",
				"http://localhost/isbn/12345/description",
				"http://localhost/isbn/12345/a",
				"http://localhost/isbn/12345/b",
				"http://localhost/isbn/12345/c",
			},
			stickiness: &types.Stickiness{
				Rules: []string{
					`{{regexFind "^/isbn/\\d+" .URL.Path}}`,
				},
			},
		},
		{
			amountServer: 3,
			stickiness: &types.Stickiness{
				IP: true,
				Rules: []string{
					`{{index .Header "X-Notfound" | join ""}}`,
				},
			},
		},
		{
			amountServer: 3,
			stickiness: &types.Stickiness{
				Cookie: true,
				Rules: []string{
					`{{index .Header "X-Notfound" | join ""}}`,
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		for _, seedCookie := range []string{"", "notfound"} {
			test.seedCookie = seedCookie
			t.Run(fmt.Sprintf("amount servers %d seed cookie %s", test.amountServer, test.seedCookie), func(t *testing.T) {
				t.Parallel()

				lb := newHealthCheckLoadBalancer(test.amountServer)
				nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get(requestCountHeader) == "0" {
						cookie := &http.Cookie{Name: cookieName, Value: lb.Servers()[0].String(), Path: "/"}
						http.SetCookie(w, cookie)
					}
					w.WriteHeader(http.StatusOK)
				})
				h := NewStickyBackendHandler(lb, nextHandler, NewStickinessParsed(test.stickiness, "/test/backend", cookieName))

				var stickyHost string
				var stickyCookie string
				recorder := httptest.NewRecorder()
				for i := 0; i < numRequests; i++ {
					if test.targets == nil || len(test.targets) == 0 {
						test.targets = []string{"http://localhost"}
					}
					for _, target := range test.targets {
						req := httptest.NewRequest(http.MethodGet, target, nil)

						// send a header with the request count
						req.Header.Set(requestCountHeader, strconv.Itoa(i))

						// set custom headers
						if test.headers != nil {
							for k, v := range test.headers {
								req.Header.Set(k, v)
							}
						}

						// logic to add cookie to the request
						var addCookieValue string
						if i == 0 {
							addCookieValue = test.seedCookie
						} else {
							addCookieValue = stickyCookie
						}
						if addCookieValue != "" {
							cookie := &http.Cookie{Name: cookieName, Value: addCookieValue, Path: "/"}
							req.AddCookie(cookie)
						}

						// process the request
						h.ServeHTTP(recorder, req)

						// did the middleware send a cookie?  if so, capture it
						var host string
						requestCookie, requestCookieErr := req.Cookie(cookieName)
						if requestCookieErr == nil {
							host = requestCookie.Value
						}

						// did the midleware return a set-cookie header?
						if i == 0 {
							setCookie, setCookieIndex := h.findSetCookie(recorder.Result().Header)
							if h.sp.UseCookie {
								if !h.sp.UseRules && !h.sp.UseIP && setCookieIndex < 0 {
									t.Errorf("Sticky cookie does not exist, should exist on sticky session response")
									return
								}
							} else {
								if setCookieIndex >= 0 {
									t.Errorf("Sticky cookie exists, should not exist on consistent hash response")
									return
								}
							}

							if setCookieIndex >= 0 {
								stickyCookie = setCookie.Value
								if test.stickiness.CookieEncryptKey != "" {
									host = aesDecryptString(test.stickiness.CookieEncryptKey, setCookie.Value)
								} else {
									host = setCookie.Value
								}
							}
						}

						// was the correct sticky host chosen?
						if i == 0 {
							if host == "" {
								t.Errorf("Sticky host was not set")
								return
							}
							found := false
							for _, lbServer := range lb.Servers() {
								if host == lbServer.String() {
									found = true
									break
								}
							}
							if found == false {
								t.Errorf("Sticky host %s was not in possible list of hosts", host)
								return
							}
							stickyHost = host
						} else if host != stickyHost {
							t.Errorf("Received host %s, wanted %s", host, stickyHost)
							return
						}
					}
				}

			})
		}
	}
}
