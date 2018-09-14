// Package doapi : DO APIクライアントモジュール
package doapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	HmacSHA1          = "HmacSHA1"
	HmacSHA256        = "HmacSHA256"
	SignatureVersion2 = "2"
	APIVersion        = "20140601"
	EndpointJSON      = "https://do.api.iij.jp/"
	// EndpointJSON = "http://localhost:9999/"
	TimeLayout      = "2006-01-02T15:04:05Z"
	PostContentType = "application/json"
)

// API の呼び出し先に関連する構造
type API struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	SignMethod string
	Expires    time.Duration
	Insecure   bool
}

// NewAPI API構造体のコンストラクタ
func NewAPI(accesskey, secretkey string) *API {
	dur, _ := time.ParseDuration("1h")
	return &API{AccessKey: accesskey,
		SecretKey:  secretkey,
		Endpoint:   EndpointJSON,
		SignMethod: HmacSHA256,
		Expires:    dur,
	}
}

func convert1(r byte) string {
	passchar := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_~.-"
	if strings.ContainsRune(passchar, rune(r)) {
		return string(r)
	}
	return fmt.Sprintf("%%%02X", r)
}

// CustomEscape escape string
func CustomEscape(v string) string {
	res := ""
	for _, c := range []byte(v) {
		res += convert1(c)
	}
	return res
}

// String2Sign get string to calculate signature
func String2Sign(method string, header http.Header, param url.URL) string {
	var keys []string
	ctflag := false
	for k := range header {
		hdr := strings.ToLower(k)
		if strings.HasPrefix(hdr, "x-iijapi-") {
			keys = append(keys, hdr)
		} else if hdr == "content-type" || hdr == "content-md5" {
			keys = append(keys, hdr)
			ctflag = true
		}
	}
	sort.Strings(keys)
	var target []string
	target = append(target, method)
	target = append(target, "")
	if !ctflag {
		target = append(target, "")
	}
	for _, k := range keys {
		if k == "content-type" || k == "content-md5" {
			target = append(target, header.Get(k))
		} else {
			target = append(target, k+":"+header.Get(k))
		}
	}
	target = append(target, param.Path)
	return strings.Join(target, "\n")
}

// Sign get signature string
func (a API) Sign(method string, header http.Header, param url.URL, signmethod string) http.Header {
	header.Set("x-iijapi-Expire", time.Now().Add(a.Expires).UTC().Format(TimeLayout))
	header.Set("x-iijapi-SignatureMethod", signmethod)
	header.Set("x-iijapi-SignatureVersion", SignatureVersion2)
	tgtstr := String2Sign(method, header, param)
	var hfn func() hash.Hash
	switch signmethod {
	case HmacSHA1:
		hfn = sha1.New
	case HmacSHA256:
		hfn = sha256.New
	}
	mac := hmac.New(hfn, []byte(a.SecretKey))
	io.WriteString(mac, tgtstr)
	macstr := mac.Sum(nil)
	header.Set("Authorization", "IIJAPI "+a.AccessKey+":"+base64.StdEncoding.EncodeToString(macstr))
	return header
}

// Get : low-level Get
func (a API) Get(param url.URL) (resp *http.Response, err error) {
	return a.PostSome("GET", param, nil)
}

// PostSome : low-level Call
func (a API) PostSome(method string, param url.URL, body interface{}) (resp *http.Response, err error) {
	cl := http.Client{}
	if a.Insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cl.Transport = tr
	}
	log.Debug("param", param)
	var buf *bytes.Buffer
	if body != nil {
		var bufb []byte
		bufb, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		if len(bufb) > 2 {
			log.Debug("call with body", method, string(bufb))
			buf = bytes.NewBuffer(bufb)
		} else {
			// string(bufb)=="{}"
			log.Debug("call without body(empty)", method)
			buf = bytes.NewBufferString("")
			body = nil
		}
	} else {
		log.Debug("call without body(nil)", method)
		buf = bytes.NewBufferString("")
	}
	req, err := http.NewRequest(method, param.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Add("content-type", PostContentType)
	}
	req.Header = a.Sign(method, req.Header, param, HmacSHA256)
	log.Debug("sign", req.Header)
	resp, err = cl.Do(req)
	return
}
