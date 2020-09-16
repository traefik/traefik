package tracing

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/traefik/traefik/v2/pkg/log"
)

// TraceNameHashLength defines the number of characters to use from the head of the generated hash.
const TraceNameHashLength = 8

// OperationNameMaxLengthNumber defines the number of static characters in a Span Trace name:
// 8 chars for hash + 2 chars for '_'.
const OperationNameMaxLengthNumber = 10

func generateOperationName(prefix string, parts []string, sep string, spanLimit int) string {
	name := prefix + " " + strings.Join(parts, sep)

	maxLength := OperationNameMaxLengthNumber + len(prefix) + 1

	if spanLimit > 0 && len(name) > spanLimit {
		if spanLimit < maxLength {
			log.WithoutContext().Warnf("SpanNameLimit cannot be lesser than %d: falling back on %d, maxLength, maxLength+3", maxLength)
			spanLimit = maxLength + 3
		}

		limit := (spanLimit - maxLength) / 2

		var fragments []string
		for _, value := range parts {
			fragments = append(fragments, truncateString(value, limit))
		}
		fragments = append(fragments, computeHash(name))

		name = prefix + " " + strings.Join(fragments, sep)
	}

	return name
}

// truncateString reduces the length of the 'str' argument to 'num' - 3 and adds a '...' suffix to the tail.
func truncateString(str string, num int) string {
	text := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		text = str[0:num] + "..."
	}
	return text
}

// computeHash returns the first TraceNameHashLength character of the sha256 hash for 'name' argument.
func computeHash(name string) string {
	data := []byte(name)
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		// Impossible case
		log.WithoutContext().WithField("OperationName", name).Errorf("Failed to create Span name hash for %s: %v", name, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))[:TraceNameHashLength]
}
