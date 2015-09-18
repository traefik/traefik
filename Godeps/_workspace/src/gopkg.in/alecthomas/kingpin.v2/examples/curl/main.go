// A curl-like HTTP command-line client.
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
)

var (
	timeout = kingpin.Flag("timeout", "Set connection timeout.").Short('t').Default("5s").Duration()
	headers = HTTPHeader(kingpin.Flag("headers", "Add HTTP headers to the request.").Short('H').PlaceHolder("HEADER=VALUE"))

	get         = kingpin.Command("get", "GET a resource.")
	getFlag     = get.Flag("test", "Test flag").Bool()
	getURL      = get.Command("url", "Retrieve a URL.")
	getURLURL   = getURL.Arg("url", "URL to GET.").Required().URL()
	getFile     = get.Command("file", "Retrieve a file.")
	getFileFile = getFile.Arg("file", "File to retrieve.").Required().ExistingFile()

	post           = kingpin.Command("post", "POST a resource.")
	postData       = post.Flag("data", "Key-value data to POST").Short('d').PlaceHolder("KEY:VALUE").StringMap()
	postBinaryFile = post.Flag("data-binary", "File with binary data to POST.").File()
	postURL        = post.Arg("url", "URL to POST to.").Required().URL()
)

type HTTPHeaderValue http.Header

func (h HTTPHeaderValue) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected HEADER=VALUE got '%s'", value)
	}
	(http.Header)(h).Add(parts[0], parts[1])
	return nil
}

func (h HTTPHeaderValue) String() string {
	return ""
}

func HTTPHeader(s kingpin.Settings) (target *http.Header) {
	target = &http.Header{}
	s.SetValue((*HTTPHeaderValue)(target))
	return
}

func applyRequest(req *http.Request) error {
	req.Header = *headers
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func apply(method string, url string) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	return applyRequest(req)
}

func applyPOST() error {
	req, err := http.NewRequest("POST", (*postURL).String(), nil)
	if err != nil {
		return err
	}
	if len(*postData) > 0 {
		for key, value := range *postData {
			req.Form.Set(key, value)
		}
	} else if postBinaryFile != nil {
		if headers.Get("Content-Type") != "" {
			headers.Set("Content-Type", "application/octet-stream")
		}
		req.Body = *postBinaryFile
	} else {
		return errors.New("--data or --data-binary must be provided to POST")
	}
	return applyRequest(req)
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("1.0").Author("Alec Thomas")
	kingpin.CommandLine.Help = "An example implementation of curl."
	switch kingpin.Parse() {
	case "get url":
		kingpin.FatalIfError(apply("GET", (*getURLURL).String()), "GET failed")

	case "post":
		kingpin.FatalIfError(applyPOST(), "POST failed")
	}
}
