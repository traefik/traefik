package integration

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	gorillawebsocket "github.com/gorilla/websocket"
	checker "github.com/vdemeester/shakers"
	"golang.org/x/net/websocket"
)

// WebsocketSuite
type WebsocketSuite struct{ BaseSuite }

func (suite *WebsocketSuite) TestBase(c *check.C) {
	var upgrader = gorillawebsocket.Upgrader{} // use default options

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

	file := suite.adaptFile(c, "fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	defer os.Remove(file)
	cmd, _ := suite.cmdTraefik(withConfigFile(file), "--debug")

	err := cmd.Start()
	c.Assert(err, check.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.BodyContains("127.0.0.1"))
	c.Assert(err, checker.IsNil)

	conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", nil)
	c.Assert(err, checker.IsNil)

	err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
	c.Assert(err, checker.IsNil)

	_, msg, err := conn.ReadMessage()
	c.Assert(err, checker.IsNil)
	c.Assert(string(msg), checker.Equals, "OK")
}

func (suite *WebsocketSuite) TestWrongOrigin(c *check.C) {
	var upgrader = gorillawebsocket.Upgrader{} // use default options

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

	file := suite.adaptFile(c, "fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	defer os.Remove(file)
	cmd, _ := suite.cmdTraefik(withConfigFile(file), "--debug")

	err := cmd.Start()
	c.Assert(err, check.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.BodyContains("127.0.0.1"))
	c.Assert(err, checker.IsNil)

	config, err := websocket.NewConfig("ws://127.0.0.1:8000/ws", "ws://127.0.0.1:800")
	c.Assert(err, check.IsNil)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	c.Assert(err, checker.IsNil)
	_, err = websocket.NewClient(config, conn)
	c.Assert(err, checker.NotNil)
	c.Assert(err, checker.ErrorMatches, "bad status")
}

func (suite *WebsocketSuite) TestOrigin(c *check.C) {
	// use default options
	var upgrader = gorillawebsocket.Upgrader{}

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

	file := suite.adaptFile(c, "fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	defer os.Remove(file)
	cmd, _ := suite.cmdTraefik(withConfigFile(file), "--debug")

	err := cmd.Start()
	c.Assert(err, check.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.BodyContains("127.0.0.1"))
	c.Assert(err, checker.IsNil)

	config, err := websocket.NewConfig("ws://127.0.0.1:8000/ws", "ws://127.0.0.1:8000")
	c.Assert(err, check.IsNil)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	c.Assert(err, check.IsNil)
	client, err := websocket.NewClient(config, conn)
	c.Assert(err, checker.IsNil)

	n, err := client.Write([]byte("OK"))
	c.Assert(err, checker.IsNil)
	c.Assert(n, checker.Equals, 2)

	msg := make([]byte, 2)
	n, err = client.Read(msg)
	c.Assert(err, checker.IsNil)
	c.Assert(n, checker.Equals, 2)
	c.Assert(string(msg), checker.Equals, "OK")

}

func (suite *WebsocketSuite) TestWrongOriginIgnoredByServer(c *check.C) {
	var upgrader = gorillawebsocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
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

	file := suite.adaptFile(c, "fixtures/websocket/config.toml", struct {
		WebsocketServer string
	}{
		WebsocketServer: srv.URL,
	})

	defer os.Remove(file)
	cmd, _ := suite.cmdTraefik(withConfigFile(file), "--debug")

	err := cmd.Start()
	c.Assert(err, check.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.BodyContains("127.0.0.1"))
	c.Assert(err, checker.IsNil)

	config, err := websocket.NewConfig("ws://127.0.0.1:8000/ws", "ws://127.0.0.1:80")
	c.Assert(err, check.IsNil)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	c.Assert(err, checker.IsNil)
	client, err := websocket.NewClient(config, conn)
	c.Assert(err, checker.IsNil)

	n, err := client.Write([]byte("OK"))
	c.Assert(err, checker.IsNil)
	c.Assert(n, checker.Equals, 2)

	msg := make([]byte, 2)
	n, err = client.Read(msg)
	c.Assert(err, checker.IsNil)
	c.Assert(n, checker.Equals, 2)
	c.Assert(string(msg), checker.Equals, "OK")

}
