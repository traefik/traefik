package accesslog

import (
	"bytes"
	"regexp"
)

// ParseAccessLog parse line of access log and return a map with each fields
func ParseAccessLog(data string) (map[string]string, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`(\S+)`)                  // 1 - ClientHost
	buffer.WriteString(`\s-\s`)                  // - - Spaces
	buffer.WriteString(`(\S+)\s`)                // 2 - ClientUsername
	buffer.WriteString(`\[([^]]+)\]\s`)          // 3 - StartUTC
	buffer.WriteString(`"(\S*)\s?`)              // 4 - RequestMethod
	buffer.WriteString(`((?:[^"]*(?:\\")?)*)\s`) // 5 - RequestPath
	buffer.WriteString(`([^"]*)"\s`)             // 6 - RequestProtocol
	buffer.WriteString(`(\S+)\s`)                // 7 - OriginStatus
	buffer.WriteString(`(\S+)\s`)                // 8 - OriginContentSize
	buffer.WriteString(`("?\S+"?)\s`)            // 9 - Referrer
	buffer.WriteString(`("\S+")\s`)              // 10 - User-Agent
	buffer.WriteString(`(\S+)\s`)                // 11 - RequestCount
	buffer.WriteString(`("[^"]*"|-)\s`)          // 12 - FrontendName
	buffer.WriteString(`("[^"]*"|-)\s`)          // 13 - BackendURL
	buffer.WriteString(`(\S+)`)                  // 14 - Duration

	regex, err := regexp.Compile(buffer.String())
	if err != nil {
		return nil, err
	}

	submatch := regex.FindStringSubmatch(data)
	result := make(map[string]string)

	// Need to be > 13 to match CLF format
	if len(submatch) > 13 {
		result[ClientHost] = submatch[1]
		result[ClientUsername] = submatch[2]
		result[StartUTC] = submatch[3]
		result[RequestMethod] = submatch[4]
		result[RequestPath] = submatch[5]
		result[RequestProtocol] = submatch[6]
		result[OriginStatus] = submatch[7]
		result[OriginContentSize] = submatch[8]
		result[RequestRefererHeader] = submatch[9]
		result[RequestUserAgentHeader] = submatch[10]
		result[RequestCount] = submatch[11]
		result[FrontendName] = submatch[12]
		result[BackendURL] = submatch[13]
		result[Duration] = submatch[14]
	}

	return result, nil
}
