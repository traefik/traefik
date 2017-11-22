package cbreaker

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
)

type SideEffect interface {
	Exec() error
}

type Webhook struct {
	URL     string
	Method  string
	Headers http.Header
	Form    url.Values
	Body    []byte
}

type WebhookSideEffect struct {
	w Webhook
}

func NewWebhookSideEffect(w Webhook) (*WebhookSideEffect, error) {
	if w.Method == "" {
		return nil, fmt.Errorf("Supply method")
	}
	_, err := url.Parse(w.URL)
	if err != nil {
		return nil, err
	}

	return &WebhookSideEffect{w: w}, nil
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
	log.Infof("%v got response: (%s): %s", w, re.Status, string(body))
	return nil
}
