package fasthttp

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

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
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

type dialerFunc func(network, addr string) (c net.Conn, err error)

func (d dialerFunc) Dial(network, addr string) (c net.Conn, err error) {
	return d(network, addr)
}

// ProxyBuilder handles the connection pools for the FastHTTP proxies.
type ProxyBuilder struct {
	// lock isn't needed because ProxyBuilder is not called concurrently.
	pools map[string]map[string]*connPool
	proxy func(*http.Request) (*url.URL, error)
}

// NewProxyBuilder creates a new ProxyBuilder.
func NewProxyBuilder() *ProxyBuilder {
	return &ProxyBuilder{
		pools: make(map[string]map[string]*connPool),
		proxy: http.ProxyFromEnvironment,
	}
}

// Delete deletes the round-tripper corresponding to the given dynamic.HTTPClientConfig.
func (r *ProxyBuilder) Delete(cfgName string) {
	delete(r.pools, cfgName)
}

// Build builds a new ReverseProxy with the given configuration.
func (r *ProxyBuilder) Build(cfgName string, cfg *dynamic.HTTPClientConfig, tlsConfig *tls.Config, targetURL *url.URL) (http.Handler, error) {
	proxyURL, err := r.proxy(&http.Request{URL: targetURL})
	if err != nil {
		return nil, err
	}

	var responseHeaderTimeout time.Duration
	if cfg.ForwardingTimeouts != nil {
		responseHeaderTimeout = time.Duration(cfg.ForwardingTimeouts.ResponseHeaderTimeout)
	}
	pool := r.getPool(cfgName, cfg, tlsConfig, targetURL, proxyURL)
	return NewReverseProxy(targetURL, proxyURL, cfg.PassHostHeader, responseHeaderTimeout, pool)
}

func (r *ProxyBuilder) getPool(cfgName string, config *dynamic.HTTPClientConfig, tlsConfig *tls.Config, targetURL *url.URL, proxyURL *url.URL) *connPool {
	pool, ok := r.pools[cfgName]
	if !ok {
		pool = make(map[string]*connPool)
		r.pools[cfgName] = pool
	}

	if connPool, ok := pool[targetURL.String()]; ok {
		return connPool
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if config.ForwardingTimeouts != nil {
		dialer.Timeout = time.Duration(config.ForwardingTimeouts.DialTimeout)
	}

	idleConnTimeout := time.Duration(dynamic.DefaultIdleConnTimeout)
	if config.ForwardingTimeouts != nil {
		idleConnTimeout = time.Duration(config.ForwardingTimeouts.IdleConnTimeout)
	}

	dialFn := getDialFn(targetURL, proxyURL, tlsConfig, config)

	connPool := NewConnPool(config.MaxIdleConnsPerHost, idleConnTimeout, dialFn)

	r.pools[cfgName][targetURL.String()] = connPool
	return connPool
}

func getDialFn(targetURL *url.URL, proxyURL *url.URL, tlsConfig *tls.Config, config *dynamic.HTTPClientConfig) func() (net.Conn, error) {
	targetAddr := addrFromURL(targetURL)

	if proxyURL == nil {
		return func() (net.Conn, error) {
			d := getDialer(targetURL.Scheme, tlsConfig, config)
			return d.Dial("tcp", targetAddr)
		}
	}

	proxyDialer := getDialer(proxyURL.Scheme, tlsConfig, config)
	proxyAddr := addrFromURL(proxyURL)

	switch {
	case proxyURL.Scheme == schemeSocks5:
		var auth *proxy.Auth
		if u := proxyURL.User; u != nil {
			auth = &proxy.Auth{User: u.Username()}
			auth.Password, _ = u.Password()
		}

		// SOCKS5 implementation do not return errors.
		socksDialer, _ := proxy.SOCKS5("tcp", proxyAddr, auth, proxyDialer)
		return func() (net.Conn, error) {
			co, err := socksDialer.Dial("tcp", targetAddr)
			if err != nil {
				return nil, err
			}

			if targetURL.Scheme == schemeHTTPS {
				c := &tls.Config{}
				if tlsConfig != nil {
					c = tlsConfig.Clone()
				}

				if c.ServerName == "" {
					c.ServerName = targetURL.Hostname()
				}
				return tls.Client(co, c), nil
			}
			return co, nil
		}

	case targetURL.Scheme == schemeHTTP:
		// Nothing to do the Proxy-Authorization header will be added by the ReverseProxy.

	case targetURL.Scheme == schemeHTTPS:
		hdr := make(http.Header)
		if u := proxyURL.User; u != nil {
			username := u.Username()
			password, _ := u.Password()
			auth := username + ":" + password
			hdr.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
		}

		return func() (net.Conn, error) {
			conn, err := proxyDialer.Dial("tcp", proxyAddr)
			if err != nil {
				return nil, err
			}

			connectReq := &http.Request{
				Method: http.MethodConnect,
				URL:    &url.URL{Opaque: targetAddr},
				Host:   targetURL.Host,
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
				_, text, ok := strings.Cut(resp.Status, " ")
				conn.Close()
				if !ok {
					return nil, errors.New("unknown status code")
				}
				return nil, errors.New(text)
			}

			if targetURL.Scheme == schemeHTTPS {
				c := &tls.Config{}
				if tlsConfig != nil {
					c = tlsConfig.Clone()
				}

				if c.ServerName == "" {
					c.ServerName = targetURL.Hostname()
				}
				return tls.Client(conn, c), nil
			}
			return conn, nil
		}
	}

	return func() (net.Conn, error) {
		return proxyDialer.Dial("tcp", proxyAddr)
	}
}

func getDialer(scheme string, tlsConfig *tls.Config, cfg *dynamic.HTTPClientConfig) dialer {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if cfg.ForwardingTimeouts != nil {
		dialer.Timeout = time.Duration(cfg.ForwardingTimeouts.DialTimeout)
	}

	if scheme == schemeHTTPS && tlsConfig != nil {
		return dialerFunc(func(network, addr string) (c net.Conn, err error) {
			return tls.DialWithDialer(dialer, network, addr, tlsConfig)
		})
	}
	return dialer
}

func addrFromURL(u *url.URL) string {
	addr := u.Host

	if u.Port() == "" {
		if u.Scheme == schemeHTTP {
			return addr + ":80"
		}
		if u.Scheme == schemeHTTPS {
			return addr + ":443"
		}
	}

	return addr
}
