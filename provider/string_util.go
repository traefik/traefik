package provider

import "strings"

// SplitAndTrimString splits separatedString at the comma character and trims each
// piece, filtering out empty pieces. Returns the list of pieces or nil if the input
// did not contain a non-empty piece.
func SplitAndTrimString(base string) []string {
	var trimmedStrings []string

	for _, s := range strings.Split(base, ",") {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			trimmedStrings = append(trimmedStrings, s)
		}
	}

	return trimmedStrings
}
