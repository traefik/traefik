package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	//"strconv"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/forward"
	"io/ioutil"
	"strings"
	"time"
)

func TestMirror(t *testing.T) {
	t1 := false
	t2 := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Test Server")
		if r.Body != nil {
			body, _ := ioutil.ReadAll(r.Body)
			fmt.Fprintln(w, string(body))
		}
		fmt.Println("Test server request")
		t1 = true
	}))
	defer ts.Close()

	tsMirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Test Mirror Server")
		fmt.Println("Mirror server request")
		t2 = true
	}))
	defer tsMirror.Close()

	testMirror := &types.Mirror{Backend: "mirror"}
	testFrontend := &types.Frontend{Mirror: testMirror}
	//testMirror := &types.Backend{Servers: map[string]types.Server{
	//	testMirror.Backend: {URL: ts.URL},
	//}}
	testMirrorBackend := &types.Backend{Servers: map[string]types.Server{
		testMirror.Backend: {URL: tsMirror.URL},
	}}

	testMirrorHandler, err := NewMirrorMiddleware(testFrontend, testMirrorBackend)
	require.NoError(t, err)

	assert.Equal(t, testMirrorHandler.backendURL.String(), tsMirror.URL, "Should be equal")

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/test", nil)
	// Make sure this is set, otherwise the mirror forwarding will use the req.URL instead of req.RequestURI for the path.
	req.RequestURI = req.URL.RequestURI()
	require.NoError(t, err)

	//handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(w, "traefik")
	//})

	handler, _ := forward.New()
	n := negroni.New()
	n.Use(testMirrorHandler)
	n.UseHandler(handler)

	n.ServeHTTP(recorder, req)

	time.Sleep(5 * time.Second)

	assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status")
	assert.Contains(t, recorder.Body.String(), "Test Server")

	assert.True(t, t1, "Test server should be called")
	assert.True(t, t2, "Mirror server should be called")

	// ----

	body := strings.NewReader("Test Body")
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/test", body)

	req.RequestURI = req.URL.RequestURI()
	require.NoError(t, err)

	n.ServeHTTP(recorder, req)

	time.Sleep(5 * time.Second)

	assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status")
	assert.Contains(t, recorder.Body.String(), "Test Server")
	assert.Contains(t, recorder.Body.String(), "Test Body")

	assert.True(t, t1, "Test server should be called")
	assert.True(t, t2, "Mirror server should be called")
}
