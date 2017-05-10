package middlewares

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorPage(t *testing.T) {
	handler := NewErrorPagesHandler("Dummy_file")
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://example.com", nil)
	handler.ServeHTTP(rr, req, err)
	assert.Equal(t, rr.Code, http.StatusInternalServerError, "They should be equal")
	assert.Equal(t, string("Internal Server Error"), rr.Result().Status, "They should be equal")

	handler = NewErrorPagesHandler("testdata/error_page_testdata.html")
	rr = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "http://example.com", nil)
	handler.ServeHTTP(rr, req, err)
	responseData, _ := ioutil.ReadAll(rr.Result().Body)
	assert.Equal(t, string("Error Page Test Data\n"), string(responseData), "They should be equal")
}
