package http2curl

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// CurlCommand contains exec.Command compatible slice + helpers
type CurlCommand struct {
	slice []string
}

// append appends a string to the CurlCommand
func (c *CurlCommand) append(newSlice ...string) {
	c.slice = append(c.slice, newSlice...)
}

// String returns a ready to copy/paste command
func (c *CurlCommand) String() string {
	slice := make([]string, len(c.slice))
	copy(slice, c.slice)
	for i := range slice {
		quoted := fmt.Sprintf("%q", slice[i])
		if strings.Contains(slice[i], " ") || len(quoted) != len(slice[i])+2 {
			slice[i] = quoted
		}
	}
	return strings.Join(slice, " ")
}

// nopCloser is used to create a new io.ReadCloser for req.Body
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

// GetCurlCommand returns a CurlCommand corresponding to an http.Request
func GetCurlCommand(req *http.Request) (*CurlCommand, error) {
	command := CurlCommand{}

	command.append("curl")

	command.append("-X", req.Method)

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = nopCloser{bytes.NewBuffer(body)}
		command.append("-d", fmt.Sprintf("%s", bytes.Trim(body, "\n")))
	}

	var keys []string

	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		command.append("-H", fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " ")))
	}

	command.append(fmt.Sprintf("'%v'", req.URL.String()))

	return &command, nil
}
