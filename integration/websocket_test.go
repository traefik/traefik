package main

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-check/check"

	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/containous/traefik/integration/utils"
	"github.com/gorilla/websocket"
	checker "github.com/vdemeester/shakers"
)

// WebsocketSuite
type WebsocketSuite struct{ BaseSuite }

func (suite *WebsocketSuite) TestBase(c *check.C) {
	var upgrader = websocket.Upgrader{} // use default options

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
	err = utils.TryRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, func(res *http.Response) error {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if !strings.Contains(string(body), "127.0.0.1") {
			return errors.New("Incorrect traefik config")
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", nil)

	c.Assert(err, checker.IsNil)
	conn.WriteMessage(websocket.TextMessage, []byte("OK"))

	_, msg, err := conn.ReadMessage()
	c.Assert(err, checker.IsNil)

	c.Assert(string(msg), checker.Equals, "OK")

}
