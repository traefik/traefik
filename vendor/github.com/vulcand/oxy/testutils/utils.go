package testutils

import (
	"crypto/tls"
	"fmt"
	"github.com/vulcand/oxy/utils"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func NewHandler(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func NewResponder(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	}))
}

// ParseURI is the version of url.ParseRequestURI that panics if incorrect, helpful to shorten the tests
func ParseURI(uri string) *url.URL {
	out, err := url.ParseRequestURI(uri)
	if err != nil {
		panic(err)
	}
	return out
}

type ReqOpts struct {
	Host    string
	Method  string
	Body    string
	Headers http.Header
	Auth    *utils.BasicAuth
}

type ReqOption func(o *ReqOpts) error

func Method(m string) ReqOption {
	return func(o *ReqOpts) error {
		o.Method = m
		return nil
	}
}

func Host(h string) ReqOption {
	return func(o *ReqOpts) error {
		o.Host = h
		return nil
	}
}

func Body(b string) ReqOption {
	return func(o *ReqOpts) error {
		o.Body = b
		return nil
	}
}

func Header(name, val string) ReqOption {
	return func(o *ReqOpts) error {
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}
		o.Headers.Add(name, val)
		return nil
	}
}

func Headers(h http.Header) ReqOption {
	return func(o *ReqOpts) error {
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}
		utils.CopyHeaders(o.Headers, h)
		return nil
	}
}

func BasicAuth(username, password string) ReqOption {
	return func(o *ReqOpts) error {
		o.Auth = &utils.BasicAuth{
			Username: username,
			Password: password,
		}
		return nil
	}
}

func MakeRequest(url string, opts ...ReqOption) (*http.Response, []byte, error) {
	o := &ReqOpts{}
	for _, s := range opts {
		if err := s(o); err != nil {
			return nil, nil, err
		}
	}

	if o.Method == "" {
		o.Method = "GET"
	}
	request, _ := http.NewRequest(o.Method, url, strings.NewReader(o.Body))
	if o.Headers != nil {
		utils.CopyHeaders(request.Header, o.Headers)
	}

	if o.Auth != nil {
		request.Header.Set("Authorization", o.Auth.String())
	}

	if len(o.Host) != 0 {
		request.Host = o.Host
	}

	var tr *http.Transport
	if strings.HasPrefix(url, "https") {
		tr = &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		tr = &http.Transport{
			DisableKeepAlives: true,
		}
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("No redirects")
		},
	}
	response, err := client.Do(request)
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		return response, bodyBytes, err
	}
	return response, nil, err
}

func Get(url string, opts ...ReqOption) (*http.Response, []byte, error) {
	opts = append(opts, Method("GET"))
	return MakeRequest(url, opts...)
}
