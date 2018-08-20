package cbreaker

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
)

// SideEffect a side effect
type SideEffect interface {
	Exec() error
}

// Webhook Web hook
type Webhook struct {
	URL     string
	Method  string
	Headers http.Header
	Form    url.Values
	Body    []byte
}

// WebhookSideEffect a web hook side effect
type WebhookSideEffect struct {
	w Webhook

	log *log.Logger
}

// NewWebhookSideEffectsWithLogger creates a new WebhookSideEffect
func NewWebhookSideEffectsWithLogger(w Webhook, l *log.Logger) (*WebhookSideEffect, error) {
	if w.Method == "" {
		return nil, fmt.Errorf("Supply method")
	}
	_, err := url.Parse(w.URL)
	if err != nil {
		return nil, err
	}

	return &WebhookSideEffect{w: w, log: l}, nil
}

// NewWebhookSideEffect creates a new WebhookSideEffect
func NewWebhookSideEffect(w Webhook) (*WebhookSideEffect, error) {
	return NewWebhookSideEffectsWithLogger(w, log.StandardLogger())
}

func (w *WebhookSideEffect) getBody() io.Reader {
	if len(w.w.Form) != 0 {
		return strings.NewReader(w.w.Form.Encode())
	}
	if len(w.w.Body) != 0 {
		return bytes.NewBuffer(w.w.Body)
	}
	return nil
}

// Exec execute the side effect
func (w *WebhookSideEffect) Exec() error {
	r, err := http.NewRequest(w.w.Method, w.w.URL, w.getBody())
	if err != nil {
		return err
	}
	if len(w.w.Headers) != 0 {
		utils.CopyHeaders(r.Header, w.w.Headers)
	}
	if len(w.w.Form) != 0 {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	re, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if re.Body != nil {
		defer re.Body.Close()
	}
	body, err := ioutil.ReadAll(re.Body)
	if err != nil {
		return err
	}
	w.log.Debugf("%v got response: (%s): %s", w, re.Status, string(body))
	return nil
}
