package integration

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/integration/try"
)

// Test for the issue described in GitHub issue about chunked responses with keep-alive
func (s *HTTPSuite) TestChunkedResponseWithKeepAlive(t *testing.T) {
	// Backend server that sends chunked response
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "text/plain")

		// Write the headers first
		w.WriteHeader(http.StatusOK)

		if flusher, ok := w.(http.Flusher); ok {
			// Send a data chunk
			_, err := w.Write([]byte("data chunk"))
			require.NoError(t, err)
			flusher.Flush()

			// Send the final empty chunk (this is where the issue occurs)
			// The empty chunk should be properly forwarded to the client
			flusher.Flush()
		}
	}))
	defer backend.Close()

	// Traefik configuration
	file := s.adaptFile("fixtures/proxy/simple.toml", struct {
		BackendURL string
	}{
		BackendURL: backend.URL,
	})

	s.traefikCmd(withConfigFile(file))

	// Wait for Traefik to start
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(t, err)

	// Create a client that uses keep-alive (like HAProxy would)
	client := &http.Client{
		Transport: &http.Transport{
			// Enable keep-alive
			DisableKeepAlives: false,
			// Force HTTP/1.1
			ForceAttemptHTTP2: false,
		},
	}

	// Make multiple requests to test keep-alive behavior
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8000/", nil)
		require.NoError(t, err)

		// Set Connection: keep-alive explicitly
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		require.NoError(t, err, "Request %d failed", i)

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read body for request %d", i)
		err = resp.Body.Close()
		require.NoError(t, err, "Failed to close body for request %d", i)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "data chunk", string(body))

		// The connection should remain open (not closed by Traefik)
		assert.NotEqual(t, "close", resp.Header.Get("Connection"))
	}
}

// Test raw TCP connection to verify chunked encoding behavior
func (s *HTTPSuite) TestChunkedResponseRawTCP(t *testing.T) {
	// Backend server that manually writes chunked response
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Manually construct chunked response
		conn, buf, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Fatalf("Failed to hijack connection: %v", err)
		}
		defer conn.Close()

		// Write HTTP response headers
		headers := "HTTP/1.1 200 OK\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"Content-Type: text/plain\r\n" +
			"\r\n"

		_, err = buf.WriteString(headers)
		require.NoError(t, err)
		err = buf.Flush()
		require.NoError(t, err)

		// Write data chunk
		dataChunk := "data chunk"
		chunkSize := fmt.Sprintf("%x\r\n", len(dataChunk))
		_, err = buf.WriteString(chunkSize)
		require.NoError(t, err)
		_, err = buf.WriteString(dataChunk)
		require.NoError(t, err)
		_, err = buf.WriteString("\r\n")
		require.NoError(t, err)
		err = buf.Flush()

		// Write final empty chunk
		_, err = buf.WriteString("0\r\n\r\n")
		require.NoError(t, err)
		err = buf.Flush()
		require.NoError(t, err)
	}))
	defer backend.Close()

	// Traefik configuration
	file := s.adaptFile("fixtures/proxy/simple.toml", struct {
		BackendURL string
	}{
		BackendURL: backend.URL,
	})

	s.traefikCmd(withConfigFile(file))

	// Wait for Traefik to start
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(t, err)

	// Connect directly to Traefik and send raw HTTP request
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	require.NoError(t, err)
	defer conn.Close()

	// Send HTTP request with keep-alive
	request := "GET / HTTP/1.1\r\n" +
		"Host: 127.0.0.1:8000\r\n" +
		"Connection: keep-alive\r\n" +
		"\r\n"

	_, err = conn.Write([]byte(request))
	require.NoError(t, err)

	// Read response
	reader := bufio.NewReader(conn)
	// Read status line
	statusLine, _, err := reader.ReadLine()
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(statusLine, []byte("HTTP/1.1 200")))

	// Read headers
	var headers []string
	for {
		line, _, err := reader.ReadLine()
		require.NoError(t, err)
		if len(line) == 0 {
			break // Empty line indicates end of headers
		}
		headers = append(headers, string(line))
	}

	// Verify Transfer-Encoding header is present
	hasChunkedEncoding := false
	for _, header := range headers {
		if strings.Contains(strings.ToLower(header), "transfer-encoding: chunked") {
			hasChunkedEncoding = true
			break
		}
	}
	assert.True(t, hasChunkedEncoding, "Response should have chunked transfer encoding")
	// Read chunked body
	var body bytes.Buffer
	for {
		// Read chunk size line
		chunkSizeLine, _, err := reader.ReadLine()
		require.NoError(t, err)

		chunkSizeStr := strings.TrimSpace(string(chunkSizeLine))
		if chunkSizeStr == "0" {
			// Final chunk - read trailing headers (should be just \r\n)
			_, _, err := reader.ReadLine()
			require.NoError(t, err)
			break
		}

		// Parse chunk size
		var chunkSize int
		_, err = fmt.Sscanf(chunkSizeStr, "%x", &chunkSize)
		require.NoError(t, err)
		// Read chunk data
		chunkData := make([]byte, chunkSize)
		_, err = io.ReadFull(reader, chunkData)
		require.NoError(t, err)
		body.Write(chunkData)

		// Read trailing \r\n
		_, _, err = reader.ReadLine()
		require.NoError(t, err)
	}

	assert.Equal(t, "data chunk", body.String())

	// Connection should still be alive for keep-alive
	// Try to send another request on the same connection
	_, err = conn.Write([]byte(request))
	require.NoError(t, err)
	// We should be able to read another response
	statusLine2, _, err := reader.ReadLine()
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(statusLine2, []byte("HTTP/1.1 200")))
}
