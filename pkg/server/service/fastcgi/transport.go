package fastcgi

import (
	"bytes"
	"fmt"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Transport struct {
	config         *dynamic.FastCGI
	clients        map[string]*Client
	serverSoftware string
}

func NewRoundTripper(cfg *dynamic.FastCGI, serverName string) *Transport {
	setDefaultTimeout(&cfg.DialTimeout, 60*time.Second)
	setDefaultTimeout(&cfg.WriteTimeout, 60*time.Second)
	setDefaultTimeout(&cfg.ReadTimeout, 60*time.Second)
	setDefaultTimeout(&cfg.AcquireConnTimeout, 4*time.Second)
	if cfg.SplitPathRegex == "" {
		cfg.SplitPathRegex = `^(.+\.php)(/.+)$`
	}

	return &Transport{
		config:         cfg,
		clients:        make(map[string]*Client),
		serverSoftware: serverName,
	}
}

func setDefaultTimeout(duration *ptypes.Duration, value time.Duration) {
	if *duration == 0 {
		*duration = ptypes.Duration(value)
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Disallow null bytes in the request path, because
	// PHP upstreams may do bad things, like execute a
	// non-PHP file as PHP code.
	if strings.Contains(req.URL.Path, "\x00") {
		return makeErrResponse(req, http.StatusBadRequest, "invalid request path"), nil
	}

	envVars, err := t.buildEnvVars(req)
	if err != nil {
		return nil, err
	}

	client, ok := t.clients[req.URL.Host]
	if !ok {
		network := "tcp"
		client, err = NewClient(
			network, req.URL.Host,
			t.config.MaxConns,
			time.Duration(t.config.WriteTimeout),
			time.Duration(t.config.ReadTimeout),
			time.Duration(t.config.DialTimeout),
			time.Duration(t.config.AcquireConnTimeout),
			t.config.LogStderr,
		)
		if err != nil {
			return nil, err
		}
		t.clients[req.URL.Host] = client
	}

	stdout, err := client.Do(&Request{
		params:     envVars,
		body:       req.Body,
		role:       FastCgiRoleResponder,
		httpMethod: req.Method,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to make FastCGI request: %w", err)
	}

	return makeResponse(stdout)
}

func makeResponse(stdout *ConnReadCloser) (*http.Response, error) {
	// emulates nginx FastCGI proxy behavior:
	// - discards http status line
	// - looking for status in 'Status' header
	// - missing status header = 200 OK
	tp := textproto.NewReader(stdout.reader)
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil && err != io.EOF {
		return nil, err
	}
	resp := new(http.Response)
	resp.Header = http.Header(mimeHeader)
	resp.Body = stdout

	if resp.Header.Get("Status") != "" {
		statusNumber, statusInfo, statusIsCut := strings.Cut(resp.Header.Get("Status"), " ")
		resp.StatusCode, err = strconv.Atoi(statusNumber)
		if err != nil {
			return nil, err
		}
		if statusIsCut {
			resp.Status = statusInfo
		}
	} else {
		resp.StatusCode = http.StatusOK
	}

	resp.ContentLength, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		resp.ContentLength = -1
		resp.TransferEncoding = []string{"chunked"}
	}

	return resp, nil
}

func makeErrResponse(req *http.Request, code int, message string) *http.Response {
	bodyBytes := []byte(message)
	bodyReader := io.NopCloser(bytes.NewReader(bodyBytes))
	lengthStr := strconv.Itoa(len(bodyBytes))

	resp := &http.Response{
		Status:        fmt.Sprintf("%d %s", code, http.StatusText(code)),
		StatusCode:    code,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		Header:        make(http.Header),
		Body:          bodyReader,
		ContentLength: int64(len(bodyBytes)),
		Request:       req,
	}
	if len(bodyBytes) > 0 {
		resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
		resp.Header.Set("Content-Length", lengthStr)
	}

	return resp
}

func (t *Transport) splitPathInfo(req *http.Request) (scriptName, pathInfo string, err error) {
	scriptName = req.URL.Path

	re, err := regexp.Compile(t.config.SplitPathRegex)
	if err != nil {
		return
	}
	matches := re.FindStringSubmatch(req.URL.Path)
	if len(matches) > 1 {
		scriptName = matches[1]
	}
	if len(matches) > 2 {
		pathInfo = matches[2]
	}
	if scriptName != "" && !strings.HasPrefix(scriptName, "/") {
		scriptName = "/" + scriptName
	}
	if pathInfo != "" && !strings.HasPrefix(pathInfo, "/") {
		pathInfo = "/" + pathInfo
	}

	return scriptName, pathInfo, nil
}

func (t *Transport) buildEnvVars(req *http.Request) (env, error) {
	host, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, err
	}

	reqHost, reqPort, err := net.SplitHostPort(req.Host)
	if err != nil {
		if !strings.Contains(err.Error(), "missing port in address") {
			return nil, err
		}
		reqPort = "80"
		if req.TLS != nil {
			reqPort = "443"
		}
	}

	rootPath := filepath.Clean(t.config.Root)
	if !filepath.IsAbs(rootPath) {
		rootPath, err = filepath.Abs(rootPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute root dir path: %w", err)
		}
	}

	if t.config.ResolveSymlink {
		rootPath, err = filepath.EvalSymlinks(rootPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlink: %w", err)
		}
	}

	scriptName, pathInfo, err := t.splitPathInfo(req)
	if err != nil {
		return nil, fmt.Errorf("failed to match script name and path info with regex: %w", err)
	}

	contentLength := "0"
	if req.ContentLength > 0 {
		contentLength = strconv.FormatInt(req.ContentLength, 10)
	}

	envVars := env{
		"AUTH_TYPE":         "",
		"CONTENT_LENGTH":    contentLength,
		"CONTENT_TYPE":      req.Header.Get("Content-Type"),
		"GATEWAY_INTERFACE": "CGI/1.1",
		"PATH_INFO":         pathInfo,
		"QUERY_STRING":      req.URL.RawQuery,
		"REMOTE_ADDR":       host,
		"REMOTE_HOST":       host,
		"REMOTE_PORT":       port,
		"REMOTE_IDENT":      "",
		"REMOTE_USER":       "",
		"REQUEST_METHOD":    req.Method,
		"SCRIPT_NAME":       scriptName,
		"SERVER_NAME":       reqHost,
		"SERVER_PORT":       reqPort,
		"SERVER_PROTOCOL":   req.Proto,
		"SERVER_SOFTWARE":   t.serverSoftware,

		"DOCUMENT_ROOT": rootPath,
		"HTTP_HOST":     req.Host,
		"REQUEST_URI":   req.URL.RequestURI(),
	}

	if scriptName != "" {
		envVars["SCRIPT_FILENAME"] = filepath.Join(rootPath, scriptName[1:]) // scriptName always has leading '/'
	}
	if pathInfo != "" {
		envVars["PATH_TRANSLATED"] = filepath.Join(rootPath, pathInfo[1:]) // pathInfo always has leading '/'
	}

	for key, val := range t.config.Env {
		envVars[key] = val
	}

	replacer := strings.NewReplacer(" ", "_", "-", "_")
	for hKey, hVal := range req.Header {
		key := fmt.Sprintf("HTTP_%s", replacer.Replace(hKey))
		key = strings.ToUpper(key)

		// skip header if present in envVars
		if _, ok := envVars[key]; ok {
			continue
		}
		envVars[key] = strings.Join(hVal, ", ")
	}

	return envVars, nil
}
