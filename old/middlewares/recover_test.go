package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urfave/negroni"
)

func TestRecoverHandler(t *testing.T) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		panic("I love panicing!")
	}
	recoverHandler := RecoverHandler(http.HandlerFunc(fn))
	server := httptest.NewServer(recoverHandler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Received non-%d response: %d\n", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestNegroniRecoverHandler(t *testing.T) {
	n := negroni.New()
	n.Use(NegroniRecoverHandler())
	panicHandler := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		panic("I love panicing!")
	}
	n.UseFunc(negroni.HandlerFunc(panicHandler))
	server := httptest.NewServer(n)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Received non-%d response: %d\n", http.StatusInternalServerError, resp.StatusCode)
	}
}
