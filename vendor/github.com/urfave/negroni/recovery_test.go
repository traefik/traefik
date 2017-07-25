package negroni

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecovery(t *testing.T) {
	buff := bytes.NewBufferString("")
	recorder := httptest.NewRecorder()
	handlerCalled := false

	rec := NewRecovery()
	rec.Logger = log.New(buff, "[negroni] ", 0)
	rec.ErrorHandlerFunc = func(i interface{}) {
		handlerCalled = true
	}

	n := New()
	// replace log for testing
	n.Use(rec)
	n.UseHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		panic("here is a panic!")
	}))
	n.ServeHTTP(recorder, (*http.Request)(nil))
	expect(t, recorder.Header().Get("Content-Type"), "text/plain; charset=utf-8")
	expect(t, recorder.Code, http.StatusInternalServerError)
	expect(t, handlerCalled, true)
	refute(t, recorder.Body.Len(), 0)
	refute(t, len(buff.String()), 0)
}

func TestRecovery_noContentTypeOverwrite(t *testing.T) {
	recorder := httptest.NewRecorder()

	rec := NewRecovery()
	rec.Logger = log.New(bytes.NewBuffer([]byte{}), "[negroni] ", 0)

	n := New()
	n.Use(rec)
	n.UseHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		panic("here is a panic!")
	}))
	n.ServeHTTP(recorder, (*http.Request)(nil))
	expect(t, recorder.Header().Get("Content-Type"), "application/javascript; charset=utf-8")
}

func TestRecovery_callbackPanic(t *testing.T) {
	buff := bytes.NewBufferString("")
	recorder := httptest.NewRecorder()

	rec := NewRecovery()
	rec.Logger = log.New(buff, "[negroni] ", 0)
	rec.ErrorHandlerFunc = func(i interface{}) {
		panic("callback panic")
	}

	n := New()
	n.Use(rec)
	n.UseHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		panic("here is a panic!")
	}))
	n.ServeHTTP(recorder, (*http.Request)(nil))

	expect(t, strings.Contains(buff.String(), "callback panic"), true)
}
