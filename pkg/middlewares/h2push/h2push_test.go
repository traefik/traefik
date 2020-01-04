package h2push

import (
	"net/http"
	"context"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/stretchr/testify/assert"
)

type dummyResponseWriter struct {
	headers map[string][]string
	pushed []string
}

func (rw *dummyResponseWriter) Header() http.Header {
	return rw.headers
}

func (rw *dummyResponseWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (rw *dummyResponseWriter) WriteHeader(statusCode int) {
}

func (rw *dummyResponseWriter) Push(target string, opts *http.PushOptions) error {
	rw.pushed = append(rw.pushed, target)
	
	return nil
}

func TestH2Push(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	h2push, _ := New(context.Background(), next, dynamic.H2Push{}, "test")
	dummyRW := &dummyResponseWriter {
		headers: map[string][]string {
			"Link": { "<script.js>; rel=preload; as=script" },
		},
		pushed: make([]string, 0, 1),
	}

	h2push.ServeHTTP(dummyRW, nil)

	assert.Len(t, dummyRW.pushed, 1)
	assert.Equal(t, "/script.js", dummyRW.pushed[0])
}

func Test_normalizePath(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		wantAbsolutePath string
	}{
		{
			name:             "Should append a forward slash when given a relative path",
			path:             "test.js",
			wantAbsolutePath: "/test.js",
		},
		{
			name:             "Should return the same path when given an absolute file path",
			path:             "/absolute.js",
			wantAbsolutePath: "/absolute.js",
		},
		{
			name:             "Should return the same path when given an absolute URL",
			path:             "http://example.com/file.js",
			wantAbsolutePath: "http://example.com/file.js",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAbsolutePath := normalizePath(tt.path); gotAbsolutePath != tt.wantAbsolutePath {
				t.Errorf("normalizePath() = %v, want %v", gotAbsolutePath, tt.wantAbsolutePath)
			}
		})
	}
}

func Test_parseLink(t *testing.T) {
	tests := []struct {
		name         string
		link         string
		wantFileName string
		wantRel      string
		wantKind     string
		wantErr      bool
	}{
		{
			name:         "Should return the link parts when given a valid Link header",
			link:         "<script.js>; rel=preload; as=script",
			wantFileName: "script.js",
			wantRel:      "preload",
			wantKind:     "script",
			wantErr:      false,
		},
		{
			name:    "Should return an error when given an invalid Link header",
			link:    "some invalid header",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFileName, gotRel, gotKind, err := parseLink(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFileName != tt.wantFileName {
				t.Errorf("parseLink() gotFileName = %v, want %v", gotFileName, tt.wantFileName)
			}
			if gotRel != tt.wantRel {
				t.Errorf("parseLink() gotRel = %v, want %v", gotRel, tt.wantRel)
			}
			if gotKind != tt.wantKind {
				t.Errorf("parseLink() gotKind = %v, want %v", gotKind, tt.wantKind)
			}
		})
	}
}
