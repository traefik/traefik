package middlewares

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	guuid "github.com/satori/go.uuid"
	"github.com/serialx/hashring"
	"github.com/vulcand/oxy/forward"
)

// StickinessParsed holds the parsed representation of a Stickiness object
type StickinessParsed struct {
	BackendName      string
	CookieEncryptKey string
	CookieName       string
	Rules            []*template.Template
	UseCookie        bool
	UseIP            bool
	UseRules         bool
}

// NewStickinessParsed creates a new StickinessParsed instance
func NewStickinessParsed(s *types.Stickiness, backendName string, cookieName string) *StickinessParsed {
	var useCookie, useIP, useRules bool
	rules := make([]*template.Template, 0)
	useDefault := true
	if len(s.Rules) > 0 {
		useRules = true
		useDefault = false
		for i, rule := range s.Rules {
			t, err := template.New(guuid.NewV4().String()).Funcs(sprig.FuncMap()).Parse(rule)
			if err != nil {
				log.Errorf("Backend %s: failed to parse sticky rule with index %d, error: %s", backendName, i, err)
				t = nil
			}
			rules = append(rules, t)
		}
	}
	if s.IP {
		useIP = true
		useDefault = false
	}
	if s.Cookie || useDefault {
		useCookie = true
	}

	return &StickinessParsed{
		BackendName:      backendName,
		CookieEncryptKey: s.CookieEncryptKey,
		CookieName:       cookieName,
		Rules:            rules,
		UseCookie:        useCookie,
		UseIP:            useIP,
		UseRules:         useRules,
	}
}

// StickyBackendHandler inspects criteria such as Client IP address and uses
// consistent hashing to route a request to a specific backend server
type StickyBackendHandler struct {
	lb   healthcheck.LoadBalancer
	next http.Handler
	sp   *StickinessParsed
}

// NewStickyBackendHandler creates a new StickyBackendHandler instance
func NewStickyBackendHandler(lb healthcheck.LoadBalancer, next http.Handler, sp *StickinessParsed) *StickyBackendHandler {
	return &StickyBackendHandler{lb: lb, next: next, sp: sp}
}

// ServeHTTP inspects the request for a sticky criteria and inserts the
// sticky header if a criteria is met.  It then invokes the next handler
// in the middleware chain.
func (h *StickyBackendHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	var server string
	var key string
	requestCookie, requestCookieErr := r.Cookie(h.sp.CookieName)
	stripSetCookie := true

	if h.sp.UseRules {
		for i, rule := range h.sp.Rules {
			if rule != nil {
				var b bytes.Buffer
				var w = bufio.NewWriter(&b)
				err := rule.Execute(w, r)
				if err != nil {
					log.Errorf("Backend %s: failed to execute sticky rule with index %d, error: %s", h.sp.BackendName, i, err)
				}
				w.Flush()
				key = strings.TrimSpace(b.String())
				if key != "" {
					break
				}
			}
		}
	}

	if key == "" && h.sp.UseIP {
		var clientIP string
		if xForwardedFor := r.Header.Get(forward.XForwardedFor); xForwardedFor != "" {
			clientIP = strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
		} else {
			clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		}
		if clientIP != "" {
			key = "clientIP: " + clientIP
		}
	}

	if key != "" {
		servers := h.lb.Servers()
		if len(servers) > 0 {
			sortedServers := make([]string, len(servers))
			for i, s := range servers {
				sortedServers[i] = s.String()
			}
			sort.Strings(sortedServers)
			ring := hashring.New(sortedServers)
			server, _ = ring.GetNode(key)
		}
	}

	if server == "" && h.sp.UseCookie {
		stripSetCookie = false
		if h.sp.CookieEncryptKey != "" && requestCookieErr == nil {
			server = aesDecryptString(h.sp.CookieEncryptKey, requestCookie.Value)
		}
	}

	if server != "" {
		if requestCookieErr == nil {
			oldCookieString := requestCookie.String()
			requestCookie.Value = server
			newCookieString := requestCookie.String()
			if cookieHeaders, ok := r.Header["Cookie"]; ok {
				for i, cookieHeader := range cookieHeaders {
					cookieHeaders[i] = strings.Replace(cookieHeader, oldCookieString, newCookieString, -1)
				}
			}
		} else {
			r.AddCookie(&http.Cookie{Name: h.sp.CookieName, Value: server})
		}
	}

	h.next.ServeHTTP(newStickyResponseWriter(h, rw, r, stripSetCookie), r)
}

func (h *StickyBackendHandler) findSetCookie(header http.Header) (*http.Cookie, int) {
	setCookieHeaders, ok := header["Set-Cookie"]

	if ok {
		setCookieRequest := &http.Request{
			Header: make(http.Header),
		}
		for i, setCookieHeader := range setCookieHeaders {
			setCookieRequest.Header.Set("Cookie", setCookieHeader)
			setCookie, err := setCookieRequest.Cookie(h.sp.CookieName)
			if err == nil {
				return setCookie, i
			}
		}
	}

	return nil, -1
}

type stickyResponseWriter struct {
	h              *StickyBackendHandler
	r              *http.Request
	rw             http.ResponseWriter
	stripSetCookie bool
}

func newStickyResponseWriter(h *StickyBackendHandler, rw http.ResponseWriter, r *http.Request, stripSetCookie bool) *stickyResponseWriter {
	return &stickyResponseWriter{h: h, r: r, rw: rw, stripSetCookie: stripSetCookie}
}

func (s *stickyResponseWriter) Header() http.Header {
	return s.rw.Header()
}

func (s *stickyResponseWriter) Write(bytes []byte) (int, error) {
	return s.rw.Write(bytes)
}

func (s *stickyResponseWriter) WriteHeader(status int) {
	setCookie, setCookieIndex := s.h.findSetCookie(s.rw.Header())
	if setCookieIndex >= 0 {
		requestCookie, requestCookieErr := s.r.Cookie(s.h.sp.CookieName)
		if requestCookieErr == nil {
			log.Debugf("Sticky backend requested was %s, actual backend %s", requestCookie.Value, setCookie.Value)
		}

		if s.stripSetCookie {
			// remove the set-cookie header as it does not need to be passed to the user agent
			s.rw.Header()["Set-Cookie"] = append(s.rw.Header()["Set-Cookie"][:setCookieIndex], s.rw.Header()["Set-Cookie"][setCookieIndex+1:]...)
		} else if s.h.sp.CookieEncryptKey != "" {
			// encrypt the set-cookie value
			setCookiePlainText := setCookie.String()
			setCookie.Value = aesEncryptString(s.h.sp.CookieEncryptKey, setCookie.Value)
			setCookieEncrypted := setCookie.String()
			s.rw.Header()["Set-Cookie"][setCookieIndex] = strings.Replace(s.rw.Header()["Set-Cookie"][setCookieIndex], setCookiePlainText, setCookieEncrypted, -1)
		}
	}
	s.rw.WriteHeader(status)
}

// encrypt/decrypt string adapted from https://gist.github.com/manishtpatel/8222606

// normalize key to 32 bytes required by AES functions
func aesKeyFromString(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// encrypt string to base64 crypto using AES
func aesEncryptString(key string, text string) string {
	plaintext := []byte(text)

	block, _ := aes.NewCipher(aesKeyFromString(key))

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	io.ReadFull(rand.Reader, iv)

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// decrypt from base64 to decrypted string
func aesDecryptString(key string, cryptoText string) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, _ := aes.NewCipher(aesKeyFromString(key))

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		// the encrypted string may have been altered; return encrypted string
		return cryptoText
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}
