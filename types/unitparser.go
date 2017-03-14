package types

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// AsSI parses a string for its number. Suffixes are allowed that loosely follow SI rules: K, Ki, M, Mi.
// 'k' and 'K' are equivalent.
// Example: "2 KiB" returns 2048, "B", nil
func AsSI(value string) (int64, string, error) {
	if value == "" {
		return 0, "", fmt.Errorf("Blank value")
	}

	numEnd := len(value)
	for i, r := range value {
		if unicode.IsDigit(r) {
			numEnd = i + 1
		}
	}

	number := value[:numEnd]
	unit := strings.TrimSpace(value[numEnd:])

	if strings.HasPrefix(unit, "Ki") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1024, unit[2:], e
	}

	if strings.HasPrefix(strings.ToUpper(unit), "K") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1000, unit[1:], e
	}

	if strings.HasPrefix(unit, "Mi") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1024 * 1024, unit[2:], e
	}

	if strings.HasPrefix(unit, "M") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1000000, unit[1:], e
	}

	i, e := strconv.ParseInt(number, 10, 64)
	return i, "", e
}
