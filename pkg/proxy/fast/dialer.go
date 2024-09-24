package fast

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const (
	schemeHTTP   = "http"
	schemeHTTPS  = "https"
	schemeSocks5 = "socks5"
)

type dialer interface {
	Dial(network, addr string) (c net.Conn, err error)
}

type dialerFunc func(network, addr string) (net.Conn, error)

func (d dialerFunc) Dial(network, addr string) (net.Conn, error) {
	return d(network, addr)
}

type dialerConfig struct {
	DialKeepAlive time.Duration
	DialTimeout   time.Duration
	ProxyURL      *url.URL
	HTTP          bool
	TLS           bool
}

func newDialer(cfg dialerConfig, tlsConfig *tls.Config) dialer {
	if cfg.ProxyURL == nil {
		return buildDialer(cfg, tlsConfig, cfg.TLS)
	}

	proxyDialer := buildDialer(cfg, tlsConfig, cfg.ProxyURL.Scheme == "https")
	proxyAddr := addrFromURL(cfg.ProxyURL)

	switch {
	case cfg.ProxyURL.Scheme == schemeSocks5:
		var auth *proxy.Auth
		if u := cfg.ProxyURL.User; u != nil {
			auth = &proxy.Auth{User: u.Username()}
			auth.Password, _ = u.Password()
		}

		// SOCKS5 implementation do not return errors.
		socksDialer, _ := proxy.SOCKS5("tcp", proxyAddr, auth, proxyDialer)
		return dialerFunc(func(network, targetAddr string) (net.Conn, error) {
			co, err := socksDialer.Dial("tcp", targetAddr)
			if err != nil {
				return nil, err
			}

			if cfg.TLS {
				c := &tls.Config{}
				if tlsConfig != nil {
					c = tlsConfig.Clone()
				}

				if c.ServerName == "" {
					host, _, _ := net.SplitHostPort(targetAddr)
					c.ServerName = host
				}

				return tls.Client(co, c), nil
			}

			return co, nil
		})
	case cfg.HTTP && !cfg.TLS:
		// Nothing to do the Proxy-Authorization header will be added by the ReverseProxy.

	default:
		hdr := make(http.Header)
		if u := cfg.ProxyURL.User; u != nil {
			username := u.Username()
			password, _ := u.Password()
			auth := username + ":" + password
			hdr.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
		}

		return dialerFunc(func(network, targetAddr string) (net.Conn, error) {
			conn, err := proxyDialer.Dial("tcp", proxyAddr)
			if err != nil {
				return nil, err
			}

			connectReq := &http.Request{
				Method: http.MethodConnect,
				URL:    &url.URL{Opaque: targetAddr},
				Host:   targetAddr,
				Header: hdr,
			}

			connectCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			didReadResponse := make(chan struct{}) // closed after CONNECT write+read is done or fails
			var resp *http.Response

			// Write the CONNECT request & read the response.
			go func() {
				defer close(didReadResponse)
				err = connectReq.Write(conn)
				if err != nil {
					return
				}
				// Okay to use and discard buffered reader here, because
				// TLS server will not speak until spoken to.
				br := bufio.NewReader(conn)
				resp, err = http.ReadResponse(br, connectReq)
			}()

			select {
			case <-connectCtx.Done():
				conn.Close()
				<-didReadResponse
				return nil, connectCtx.Err()
			case <-didReadResponse:
				// resp or err now set
			}
			if err != nil {
				conn.Close()
				return nil, err
			}

			if resp.StatusCode != http.StatusOK {
				_, statusText, ok := strings.Cut(resp.Status, " ")
				conn.Close()
				if !ok {
					return nil, errors.New("unknown status code")
				}

				return nil, errors.New(statusText)
			}

			c := &tls.Config{}
			if tlsConfig != nil {
				c = tlsConfig.Clone()
			}
			if c.ServerName == "" {
				host, _, _ := net.SplitHostPort(targetAddr)
				c.ServerName = host
			}

			return tls.Client(conn, c), nil
		})
	}
	return dialerFunc(func(network, addr string) (net.Conn, error) {
		return proxyDialer.Dial("tcp", proxyAddr)
	})
}

func buildDialer(cfg dialerConfig, tlsConfig *tls.Config, isTLS bool) dialer {
	dialer := &net.Dialer{
		Timeout:   cfg.DialTimeout,
		KeepAlive: cfg.DialKeepAlive,
	}

	if !isTLS {
		return dialer
	}

	return &tls.Dialer{
		NetDialer: dialer,
		Config:    tlsConfig,
	}
}

func addrFromURL(u *url.URL) string {
	addr := u.Host

	if u.Port() == "" {
		switch u.Scheme {
		case schemeHTTP:
			return addr + ":80"
		case schemeHTTPS:
			return addr + ":443"
		}
	}

	return addr
}
