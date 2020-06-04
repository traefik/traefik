package auth

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	algorithm       = "algorithm"
	authorization   = "Authorization"
	nonce           = "nonce"
	opaque          = "opaque"
	qop             = "qop"
	realm           = "realm"
	wwwAuthenticate = "Www-Authenticate"
)

// DigestRequest is a client for digest authentication requests.
type digestRequest struct {
	client             *http.Client
	username, password string
	nonceCount         nonceCount
}

type nonceCount int

func (nc nonceCount) String() string {
	return fmt.Sprintf("%08x", int(nc))
}

var wanted = []string{algorithm, nonce, opaque, qop, realm}

// New makes a DigestRequest instance.
func newDigestRequest(username, password string, client *http.Client) *digestRequest {
	return &digestRequest{
		client:   client,
		username: username,
		password: password,
	}
}

// Do does requests as http.Do does.
func (r *digestRequest) Do(req *http.Request) (*http.Response, error) {
	parts, err := r.makeParts(req)
	if err != nil {
		return nil, err
	}

	if parts != nil {
		req.Header.Set(authorization, r.makeAuthorization(req, parts))
	}

	return r.client.Do(req)
}

func (r *digestRequest) makeParts(req *http.Request) (map[string]string, error) {
	authReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return nil, nil
	}

	if len(resp.Header[wwwAuthenticate]) == 0 {
		return nil, fmt.Errorf("headers do not have %s", wwwAuthenticate)
	}

	headers := strings.Split(resp.Header[wwwAuthenticate][0], ",")
	parts := make(map[string]string, len(wanted))
	for _, r := range headers {
		for _, w := range wanted {
			if strings.Contains(r, w) {
				parts[w] = strings.Split(r, `"`)[1]
			}
		}
	}

	if len(parts) != len(wanted) {
		return nil, fmt.Errorf("header is invalid: %+v", parts)
	}

	return parts, nil
}

func getMD5(texts []string) string {
	h := md5.New()
	_, _ = io.WriteString(h, strings.Join(texts, ":"))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *digestRequest) getNonceCount() string {
	r.nonceCount++
	return r.nonceCount.String()
}

func (r *digestRequest) makeAuthorization(req *http.Request, parts map[string]string) string {
	ha1 := getMD5([]string{r.username, parts[realm], r.password})
	ha2 := getMD5([]string{req.Method, req.URL.String()})
	cnonce := generateRandom(16)
	nc := r.getNonceCount()
	response := getMD5([]string{
		ha1,
		parts[nonce],
		nc,
		cnonce,
		parts[qop],
		ha2,
	})
	return fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s", algorithm=%s, qop=%s, nc=%s, cnonce="%s", response="%s", opaque="%s"`,
		r.username,
		parts[realm],
		parts[nonce],
		req.URL.String(),
		parts[algorithm],
		parts[qop],
		nc,
		cnonce,
		response,
		parts[opaque],
	)
}

// GenerateRandom generates random string.
func generateRandom(n int) string {
	b := make([]byte, 8)
	_, _ = io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:n]
}
