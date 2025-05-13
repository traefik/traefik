package integration

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"golang.org/x/net/http2"
	"golang.org/x/net/websocket"
)

// WebsocketSuite tests suite.
type WebsocketSuite struct{ BaseSuite }

func TestWebsocketSuite(t *testing.T) {
	suite.Run(t, new(WebsocketSuite))
}

func (s *WebsocketSuite) TestBase() {
	upgrader := gorillawebsocket.Upgrader{} // use default options

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))

	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", nil)
	require.NoError(s.T(), err)

	err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
	require.NoError(s.T(), err)

	_, msg, err := conn.ReadMessage()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "OK", string(msg))
}

func (s *WebsocketSuite) TestWrongOrigin() {
	upgrader := gorillawebsocket.Upgrader{} // use default options

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))

	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	config, err := websocket.NewConfig("ws://127.0.0.1:8000/ws", "ws://127.0.0.1:800")
	assert.NoError(s.T(), err)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	require.NoError(s.T(), err)
	_, err = websocket.NewClient(config, conn)
	assert.ErrorContains(s.T(), err, "bad status")
}

func (s *WebsocketSuite) TestOrigin() {
	// use default options
	upgrader := gorillawebsocket.Upgrader{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))

	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	config, err := websocket.NewConfig("ws://127.0.0.1:8000/ws", "ws://127.0.0.1:8000")
	assert.NoError(s.T(), err)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	assert.NoError(s.T(), err)
	client, err := websocket.NewClient(config, conn)
	require.NoError(s.T(), err)

	n, err := client.Write([]byte("OK"))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, n)

	msg := make([]byte, 2)
	n, err = client.Read(msg)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, n)
	assert.Equal(s.T(), "OK", string(msg))
}

func (s *WebsocketSuite) TestWrongOriginIgnoredByServer() {
	upgrader := gorillawebsocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))

	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	config, err := websocket.NewConfig("ws://127.0.0.1:8000/ws", "ws://127.0.0.1:80")
	assert.NoError(s.T(), err)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	require.NoError(s.T(), err)
	client, err := websocket.NewClient(config, conn)
	require.NoError(s.T(), err)

	n, err := client.Write([]byte("OK"))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, n)

	msg := make([]byte, 2)
	n, err = client.Read(msg)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, n)
	assert.Equal(s.T(), "OK", string(msg))
}

func (s *WebsocketSuite) TestSSLTermination() {
	upgrader := gorillawebsocket.Upgrader{} // use default options

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	file := s.adaptFile("fixtures/websocket/config_https.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	// Add client self-signed cert
	roots := x509.NewCertPool()
	certContent, err := os.ReadFile("./resources/tls/local.cert")
	require.NoError(s.T(), err)
	roots.AppendCertsFromPEM(certContent)
	gorillawebsocket.DefaultDialer.TLSClientConfig = &tls.Config{
		RootCAs: roots,
	}
	conn, _, err := gorillawebsocket.DefaultDialer.Dial("wss://127.0.0.1:8000/ws", nil)
	require.NoError(s.T(), err)

	err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
	require.NoError(s.T(), err)

	_, msg, err := conn.ReadMessage()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "OK", string(msg))
}

func (s *WebsocketSuite) TestBasicAuth() {
	upgrader := gorillawebsocket.Upgrader{} // use default options

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		user, password, _ := r.BasicAuth()
		assert.Equal(s.T(), "traefiker", user)
		assert.Equal(s.T(), "secret", password)

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			err = conn.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	config, err := websocket.NewConfig("ws://127.0.0.1:8000/ws", "ws://127.0.0.1:8000")
	auth := "traefiker:secret"
	config.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))

	assert.NoError(s.T(), err)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	require.NoError(s.T(), err)
	client, err := websocket.NewClient(config, conn)
	require.NoError(s.T(), err)

	n, err := client.Write([]byte("OK"))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, n)

	msg := make([]byte, 2)
	n, err = client.Read(msg)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, n)
	assert.Equal(s.T(), "OK", string(msg))
}

func (s *WebsocketSuite) TestSpecificResponseFromBackend() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	_, resp, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", nil)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *WebsocketSuite) TestURLWithURLEncodedChar() {
	upgrader := gorillawebsocket.Upgrader{} // use default options

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "/ws/http%3A%2F%2Ftest", r.URL.EscapedPath())
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			err = conn.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))

	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws/http%3A%2F%2Ftest", nil)
	require.NoError(s.T(), err)

	err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
	require.NoError(s.T(), err)

	_, msg, err := conn.ReadMessage()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "OK", string(msg))
}

func (s *WebsocketSuite) TestSSLhttp2() {
	upgrader := gorillawebsocket.Upgrader{} // use default options

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))

	ts.TLS = &tls.Config{}
	ts.TLS.NextProtos = append(ts.TLS.NextProtos, `h2`, `http/1.1`)
	ts.StartTLS()

	file := s.adaptFile("fixtures/websocket/config_https.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: ts.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG", "--accesslog")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	// Add client self-signed cert
	roots := x509.NewCertPool()
	certContent, err := os.ReadFile("./resources/tls/local.cert")
	require.NoError(s.T(), err)
	roots.AppendCertsFromPEM(certContent)
	gorillawebsocket.DefaultDialer.TLSClientConfig = &tls.Config{
		RootCAs: roots,
	}
	conn, _, err := gorillawebsocket.DefaultDialer.Dial("wss://127.0.0.1:8000/echo", nil)
	require.NoError(s.T(), err)

	err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
	require.NoError(s.T(), err)

	_, msg, err := conn.ReadMessage()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "OK", string(msg))
}

func (s *WebsocketSuite) TestSettingEnableConnectProtocol() {
	file := s.adaptFile("fixtures/websocket/config_https.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: "http://127.0.0.1",
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG", "--accesslog")

	// Wait for traefik.
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	// Add client self-signed cert.
	roots := x509.NewCertPool()
	certContent, err := os.ReadFile("./resources/tls/local.cert")
	require.NoError(s.T(), err)

	roots.AppendCertsFromPEM(certContent)

	// Open a connection to inspect SettingsFrame.
	conn, err := tls.Dial("tcp", "127.0.0.1:8000", &tls.Config{
		RootCAs:    roots,
		NextProtos: []string{"h2"},
	})
	require.NoError(s.T(), err)

	framer := http2.NewFramer(nil, conn)
	frame, err := framer.ReadFrame()
	require.NoError(s.T(), err)

	fr, ok := frame.(*http2.SettingsFrame)
	require.True(s.T(), ok)

	_, ok = fr.Value(http2.SettingEnableConnectProtocol)
	assert.False(s.T(), ok)
}

func (s *WebsocketSuite) TestHeaderAreForwarded() {
	upgrader := gorillawebsocket.Upgrader{} // use default options

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "my-token", r.Header.Get("X-Token"))
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))

	file := s.adaptFile("fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	headers := http.Header{}
	headers.Add("X-Token", "my-token")
	conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", headers)

	require.NoError(s.T(), err)
	err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
	require.NoError(s.T(), err)

	_, msg, err := conn.ReadMessage()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "OK", string(msg))
}
