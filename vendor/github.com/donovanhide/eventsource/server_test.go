package eventsource

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServerHandlerRespondsAfterClose(t *testing.T) {
	server := NewServer()
	httpServer := httptest.NewServer(server.Handler("test"))
	defer httpServer.Close()

	server.Close()
	responses := make(chan *http.Response)

	go func() {
		resp, err := http.Get(httpServer.URL)
		if err != nil {
			t.Fatalf("Unexpected error %s", err)
		}
		responses <- resp
	}()

	select {
	case resp := <-responses:
		if resp.StatusCode != 200 {
			t.Errorf("Received StatusCode %d, want 200", resp.StatusCode)
		}
	case <-time.After(250 * time.Millisecond):
		t.Errorf("Did not receive response in time")
	}
}
