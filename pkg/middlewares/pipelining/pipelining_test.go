package pipelining

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

type recorderWithCloseNotify struct {
	*httptest.ResponseRecorder
}

func (r *recorderWithCloseNotify) CloseNotify() <-chan bool {
	panic("implement me")
}

func TestNew(t *testing.T) {
	testCases := []struct {
		desc                   string
		HTTPMethod             string
		implementCloseNotifier bool
	}{
		{
			desc:                   "should not implement CloseNotifier with GET method",
			HTTPMethod:             http.MethodGet,
			implementCloseNotifier: false,
		},
		{
			desc:                   "should implement CloseNotifier with PUT method",
			HTTPMethod:             http.MethodPut,
			implementCloseNotifier: true,
		},
		{
			desc:                   "should implement CloseNotifier with POST method",
			HTTPMethod:             http.MethodPost,
			implementCloseNotifier: true,
		},
		{
			desc:                   "should  not implement CloseNotifier with GET method",
			HTTPMethod:             http.MethodHead,
			implementCloseNotifier: false,
		},
		{
			desc:                   "should  not implement CloseNotifier with PROPFIND method",
			HTTPMethod:             "PROPFIND",
			implementCloseNotifier: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, ok := w.(http.CloseNotifier)
				assert.Equal(t, test.implementCloseNotifier, ok)
				w.WriteHeader(http.StatusOK)
			})
			handler := New(context.Background(), nextHandler, "pipe")

			req := httptest.NewRequest(test.HTTPMethod, "http://localhost", nil)

			handler.ServeHTTP(&recorderWithCloseNotify{httptest.NewRecorder()}, req)
		})
	}
}

// This test is an adapted version of net/http/httputil.Test1xxResponses test.
// This test is only here to guarantee that there would not be any regression in the future,
// because the pipelining middleware is already supporting informational headers.
func Test1xxResponses(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Link", "</style.css>; rel=preload; as=style")
		h.Add("Link", "</script.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusEarlyHints)

		h.Add("Link", "</foo.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusProcessing)

		_, _ = w.Write([]byte("Hello"))
	})

	pipe := New(context.Background(), next, "pipe")

	server := httptest.NewServer(pipe)
	t.Cleanup(server.Close)
	frontendClient := server.Client()

	checkLinkHeaders := func(t *testing.T, expected, got []string) {
		t.Helper()

		if len(expected) != len(got) {
			t.Errorf("Expected %d link headers; got %d", len(expected), len(got))
		}

		for i := range expected {
			if i >= len(got) {
				t.Errorf("Expected %q link header; got nothing", expected[i])

				continue
			}

			if expected[i] != got[i] {
				t.Errorf("Expected %q link header; got %q", expected[i], got[i])
			}
		}
	}

	var respCounter uint8
	trace := &httptrace.ClientTrace{
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			switch code {
			case http.StatusEarlyHints:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script"}, header["Link"])
			case http.StatusProcessing:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, header["Link"])
			default:
				t.Error("Unexpected 1xx response")
			}

			respCounter++

			return nil
		},
	}
	req, _ := http.NewRequestWithContext(httptrace.WithClientTrace(context.Background(), trace), http.MethodGet, server.URL, nil)

	res, err := frontendClient.Do(req)
	assert.Nil(t, err)

	defer res.Body.Close()

	if respCounter != 2 {
		t.Errorf("Expected 2 1xx responses; got %d", respCounter)
	}
	checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, res.Header["Link"])

	body, _ := io.ReadAll(res.Body)
	if string(body) != "Hello" {
		t.Errorf("Read body %q; want Hello", body)
	}
}
