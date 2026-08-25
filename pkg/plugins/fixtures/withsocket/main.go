package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/stealthrocket/net/wasip1"
)

var target string

// Built with
// GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -trimpath -o plugin.wasm ./main.go
func main() {
	target = os.Getenv("ECHO_ADDR")
	handler.HandleRequestFn = handleRequest
}

func handleRequest(req api.Request, resp api.Response) (next bool, reqCtx uint32) {
	uri := req.GetURI()
	msg := fmt.Sprintf("REQ-%s", uri)

	conn, err := wasip1.DialContext(context.Background(), "tcp", target)
	if err != nil {
		resp.SetStatusCode(500)
		resp.Body().WriteString("dial error: " + err.Error())
		return false, 0
	}
	defer conn.Close()

	if _, err = conn.Write([]byte(msg)); err != nil {
		resp.SetStatusCode(500)
		resp.Body().WriteString("write error: " + err.Error())
		return false, 0
	}

	buf := make([]byte, len(msg))
	if _, err = io.ReadFull(conn, buf); err != nil {
		resp.SetStatusCode(500)
		resp.Body().WriteString("read error: " + err.Error())
		return false, 0
	}

	got := string(buf)
	if got != msg {
		resp.SetStatusCode(409)
		resp.Body().WriteString(fmt.Sprintf("MISMATCH sent=%q got=%q", msg, got))
		return false, 0
	}

	resp.SetStatusCode(200)
	resp.Body().WriteString("OK")
	return false, 0
}
