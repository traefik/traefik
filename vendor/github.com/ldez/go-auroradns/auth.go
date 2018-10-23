package auroradns

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TokenTransport HTTP transport for API authentication
type TokenTransport struct {
	userID string
	key    string

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// NewTokenTransport Creates a  new TokenTransport
func NewTokenTransport(userID, key string) (*TokenTransport, error) {
	if userID == "" || key == "" {
		return nil, fmt.Errorf("credentials missing")
	}

	return &TokenTransport{userID: userID, key: key}, nil
}

// RoundTrip executes a single HTTP transaction
func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	enrichedReq := &http.Request{}
	*enrichedReq = *req

	enrichedReq.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		enrichedReq.Header[k] = append([]string(nil), s...)
	}

	if t.userID != "" && t.key != "" {
		timestamp := time.Now().UTC()

		fmtTime := timestamp.Format("20060102T150405Z")
		req.Header.Set("X-AuroraDNS-Date", fmtTime)

		token, err := newToken(t.userID, t.key, req.Method, req.URL.Path, timestamp)
		if err == nil {
			req.Header.Set("Authorization", fmt.Sprintf("AuroraDNSv1 %s", token))
		}
	}

	return t.transport().RoundTrip(enrichedReq)
}

// Wrap Wrap a HTTP client Transport with the TokenTransport
func (t *TokenTransport) Wrap(client *http.Client) *http.Client {
	backup := client.Transport
	t.Transport = backup
	client.Transport = t
	return client
}

// Client Creates a new HTTP client
func (t *TokenTransport) Client() *http.Client {
	return &http.Client{
		Transport: t,
		Timeout:   30 * time.Second,
	}
}

func (t *TokenTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// newToken generates a token for accessing a specific method of the API
func newToken(userID string, key string, method string, action string, timestamp time.Time) (string, error) {
	fmtTime := timestamp.Format("20060102T150405Z")
	message := strings.Join([]string{method, action, fmtTime}, "")

	signatureHmac := hmac.New(sha256.New, []byte(key))
	_, err := signatureHmac.Write([]byte(message))
	if err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(signatureHmac.Sum(nil))

	userIDAndSignature := fmt.Sprintf("%s:%s", userID, signature)

	token := base64.StdEncoding.EncodeToString([]byte(userIDAndSignature))

	return token, nil
}
