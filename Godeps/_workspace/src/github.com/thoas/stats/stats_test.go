package stats

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("bar"))
})

func TestSimple(t *testing.T) {
	s := New()

	res := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)

	s.Handler(testHandler).ServeHTTP(res, req)

	assert.Equal(t, res.Code, 200)
	assert.Equal(t, s.ResponseCounts, map[string]int{"200": 1})
}

func TestGetStats(t *testing.T) {
	s := New()

	var stats = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		stats := s.Data()

		b, _ := json.Marshal(stats)

		w.Write(b)
		w.WriteHeader(200)
		w.Header().Add("Content-Type", "application/json")
	})

	res := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)

	s.Handler(testHandler).ServeHTTP(res, req)

	res = httptest.NewRecorder()

	s.Handler(stats).ServeHTTP(res, req)

	assert.Equal(t, res.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}

	err := json.Unmarshal(res.Body.Bytes(), &data)

	assert.Nil(t, err)

	assert.Equal(t, data["total_count"].(float64), float64(1))
}
