package instana

import (
	"bytes"
	"io/ioutil"
	"strings"
)

const eumTemplate string = "eum.js"

// EumSnippet generates javascript code to initialize JavaScript agent
func EumSnippet(apiKey string, traceID string, meta map[string]string) string {

	if len(apiKey) == 0 || len(traceID) == 0 {
		return ""
	}

	b, err := ioutil.ReadFile(eumTemplate)

	if err != nil {
		return ""
	}

	var snippet = string(b)
	var metaBuffer bytes.Buffer

	snippet = strings.Replace(snippet, "$apiKey", apiKey, -1)
	snippet = strings.Replace(snippet, "$traceId", traceID, -1)

	for key, value := range meta {
		metaBuffer.WriteString("  ineum('meta', '" + key + "', '" + value + "');\n")
	}

	snippet = strings.Replace(snippet, "$meta", metaBuffer.String(), -1)

	return snippet
}
