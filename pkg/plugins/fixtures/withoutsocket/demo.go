package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

type config struct {
	File string   `json:"file"`
	Envs []string `json:"envs"`
}

var cfg config

// Built with
// tinygo build -o plugin.wasm -scheduler=none --no-debug -target=wasi ./demo.go
func main() {
	err := json.Unmarshal(handler.Host.GetConfig(), &cfg)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config %v", err))
		os.Exit(1)
	}

	handler.HandleRequestFn = handleRequest
}

func handleRequest(req api.Request, resp api.Response) (next bool, reqCtx uint32) {
	var bodyContent []byte
	if cfg.File != "" {
		var err error
		bodyContent, err = os.ReadFile(cfg.File)
		if err != nil {
			resp.SetStatusCode(http.StatusInternalServerError)
			resp.Body().Write([]byte(fmt.Sprintf("error reading file %q: %w", cfg.File, err.Error())))
			return false, 0
		}
	}

	if len(cfg.Envs) > 0 {
		for _, env := range cfg.Envs {
			bodyContent = append(bodyContent, []byte(os.Getenv(env)+"\n")...)
		}
	}

	resp.Body().Write(bodyContent)
	return false, 0
}
