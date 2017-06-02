package provider

import "strings"

// SplitAndTrimString splits separatedString at the comma character and trims each
// piece, filtering out empty pieces. Returns the list of pieces or nil if the input
// did not contain a non-empty piece.
func SplitAndTrimString(separatedString string) []string {
	listOfStrings := strings.Split(separatedString, ",")
	var trimmedListOfStrings []string
	for _, s := range listOfStrings {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			trimmedListOfStrings = append(trimmedListOfStrings, s)
		}
	}

	return trimmedListOfStrings
}
