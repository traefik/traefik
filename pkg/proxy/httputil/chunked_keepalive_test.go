package httputil

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestChunkedResponseWithKeepAlive tests that streaming responses work correctly
// with keep-alive connections, specifically addressing the Go issue #40747
func TestChunkedResponseWithKeepAlive(t *testing.T) {
	// Create a backend server that sends streaming response
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Don't set Content-Length to make this a streaming response
		rw.Header().Set("Content-Type", "text/plain")

		// Write response in chunks
		flusher := rw.(http.Flusher)
		_, err := rw.Write([]byte("chunk1\n"))
		require.NoError(t, err)
		flusher.Flush()

		time.Sleep(10 * time.Millisecond) // Small delay to simulate real server behavior

		_, err = rw.Write([]byte("chunk2\n"))
		require.NoError(t, err)
		flusher.Flush()

		// The response ends here
	}))
	defer backend.Close()

	// Create proxy with our fixed configuration
	target, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("Failed to parse backend URL: %v", err)
	}

	proxy := buildSingleHostProxy(
		target,
		false,                 // passHostHeader
		false,                 // preservePath
		time.Duration(0),      // flushInterval
		http.DefaultTransport, // roundTripper
		nil,                   // bufferPool
	)

	// Create a test server with our proxy
	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	// Test with keep-alive connection (this is the problematic case)
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: false, // Enable keep-alive
			MaxIdleConns:      10,
			IdleConnTimeout:   30 * time.Second,
		},
	}

	// Make multiple requests to test keep-alive behavior and ensure no connection errors
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("request_%d", i), func(t *testing.T) {
			req, err := http.NewRequest("GET", proxyServer.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Explicitly request keep-alive connection
			req.Header.Set("Connection", "keep-alive")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Read the response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// This is the key test - we should not get connection errors
				if strings.Contains(err.Error(), "use of closed network connection") {
					t.Errorf("Got connection error that should be fixed: %v", err)
				} else {
					t.Fatalf("Failed to read response body: %v", err)
				}
			}

			// Verify the response content
			expectedBody := "chunk1\nchunk2\n"
			if string(body) != expectedBody {
				t.Errorf("Expected body %q, got %q", expectedBody, string(body))
			}

			// The response should be successful
			if resp.StatusCode != 200 {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

// TestChunkedResponseRawTCP tests the fix at a lower level using raw TCP connections
// to ensure the connection isn't closed inappropriately
func TestChunkedResponseRawTCP(t *testing.T) { // Create a backend server that sends chunked response
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't set Content-Length to force chunked encoding
		w.Header().Set("Content-Type", "text/plain")
		// Do NOT set Transfer-Encoding manually - Go will set it automatically

		flusher := w.(http.Flusher)
		_, err := w.Write([]byte("test chunk"))
		require.NoError(t, err)
		flusher.Flush()
	}))
	defer backend.Close()
	// Create proxy
	target, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("Failed to parse backend URL: %v", err)
	}

	proxy := buildSingleHostProxy(
		target,
		false,
		false,
		time.Duration(0),
		http.DefaultTransport,
		nil,
	)

	// Create a test server with our proxy
	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	// Parse the proxy server URL to get host and port
	proxyURL := proxyServer.URL
	proxyURL = strings.TrimPrefix(proxyURL, "http://")

	// Connect using raw TCP to test connection handling
	conn, err := net.Dial("tcp", proxyURL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send HTTP request with keep-alive
	request := "GET / HTTP/1.1\r\n" +
		"Host: " + proxyURL + "\r\n" +
		"Connection: keep-alive\r\n" +
		"\r\n"

	_, err = conn.Write([]byte(request))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	// Read status line
	statusLine, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read status line: %v", err)
	}

	if !bytes.HasPrefix(statusLine, []byte("HTTP/1.1 200")) {
		t.Errorf("Expected 200 OK, got %s", statusLine)
	}
	// Read headers until empty line
	var transferEncoding string
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			t.Fatalf("Failed to read header: %v", err)
		}

		if len(line) == 0 {
			break // End of headers
		}

		if bytes.HasPrefix(bytes.ToLower(line), []byte("transfer-encoding:")) {
			transferEncoding = string(line)
		}
	}

	// Verify chunked encoding
	if !strings.Contains(strings.ToLower(transferEncoding), "chunked") {
		t.Errorf("Expected chunked transfer encoding, got %s", transferEncoding)
	}
	// Read the chunked body
	// First chunk size
	chunkSizeLine, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read chunk size: %v", err)
	}

	if len(chunkSizeLine) == 0 {
		t.Error("Empty chunk size line")
	}
	// Read chunk data
	chunkData, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read chunk data: %v", err)
	}

	if !bytes.Equal(chunkData, []byte("test chunk")) {
		t.Errorf("Expected 'test chunk', got %s", chunkData)
	}
	// Read final chunk (should be "0")
	finalChunkSize, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read final chunk size: %v", err)
	}

	if !bytes.Equal(finalChunkSize, []byte("0")) {
		t.Errorf("Expected final chunk size '0', got %s", finalChunkSize)
	}

	// The connection should still be alive after this point
	// Try to send another request on the same connection
	request2 := "GET / HTTP/1.1\r\n" +
		"Host: " + proxyURL + "\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	_, err = conn.Write([]byte(request2))
	if err != nil {
		t.Errorf("Connection was closed prematurely, failed to send second request: %v", err)
	}
}
