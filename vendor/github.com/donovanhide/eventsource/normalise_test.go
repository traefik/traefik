package eventsource

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

var (
	inputFormat  = "line1%sline2%sline3%s"
	endings      = []string{"\n", "\r\n", "\r"}
	suffixes     = []string{"\n", "\r\n", "\r", ""}
	descriptions = []string{"LF", "CRLF", "CR", "EOF"}
	expected     = []string{"line1", "line2", "line3"}
)

func Testnormaliser(t *testing.T) {
	for i, first := range endings {
		for j, second := range endings {
			for k, suffix := range suffixes {
				input := fmt.Sprintf(inputFormat, first, second, suffix)
				r := bufio.NewReader(newNormaliser(strings.NewReader(input)))
				for _, want := range expected {
					line, err := r.ReadString('\n')
					if err != nil && suffix != "" {
						t.Error("Unexpected error:", err)
					}
					line = strings.TrimSuffix(line, "\n")
					if line != want {
						expanded := fmt.Sprintf(inputFormat, descriptions[i], descriptions[j], descriptions[k])
						t.Errorf(`Using %s Expected: "%s" Got: "%s"`, expanded, want, line)
						t.Log([]byte(line))
					}
				}
				if _, err := r.ReadString('\n'); err == nil {
					t.Error("Expected EOF")
				}
			}
		}
	}
}
