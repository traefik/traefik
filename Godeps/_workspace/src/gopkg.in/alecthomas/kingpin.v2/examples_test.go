package kingpin

import (
	"fmt"
	"net/http"
	"strings"
)

type HTTPHeaderValue http.Header

func (h *HTTPHeaderValue) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected HEADER:VALUE got '%s'", value)
	}
	(*http.Header)(h).Add(parts[0], parts[1])
	return nil
}

func (h *HTTPHeaderValue) String() string {
	return ""
}

func HTTPHeader(s Settings) (target *http.Header) {
	target = new(http.Header)
	s.SetValue((*HTTPHeaderValue)(target))
	return
}

// This example ilustrates how to define custom parsers. HTTPHeader
// cumulatively parses each encountered --header flag into a http.Header struct.
func ExampleValue() {
	var (
		curl    = New("curl", "transfer a URL")
		headers = HTTPHeader(curl.Flag("headers", "Add HTTP headers to the request.").Short('H').PlaceHolder("HEADER:VALUE"))
	)

	curl.Parse([]string{"-H Content-Type:application/octet-stream"})
	for key, value := range *headers {
		fmt.Printf("%s = %s\n", key, value)
	}
}
