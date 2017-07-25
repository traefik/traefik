package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

var tw = new(tabwriter.Writer)

func init() {
	tw.Init(os.Stdout, 0, 8, 2, '\t', 0)
}

type columns []interface{}

func tabsPrint(values columns, lengths []int) {
	if len(values) != len(lengths) {
		log.Fatalf("Internal error! Mismatch during tabbed line print. Values: %d, Lengths: %d\n", len(values), len(lengths))
	}

	for i, value := range values {
		format := "\t%s"
		if i == 0 {
			format = "%s"
		}
		fmt.Fprintf(tw, format, replaceCharacters(maxCharacters(fmt.Sprintf("%v", value), lengths[i])))
	}
	fmt.Fprintf(tw, "\n")
}

func tabsFlush() {
	tw.Flush()
}

func replaceCharacters(s string) string {
	s = strings.Replace(s, "\n", `\n`, -1)
	s = strings.Replace(s, "\r", `\r`, -1)
	s = strings.Replace(s, "\t", `\t`, -1)
	return s
}

func maxCharacters(input string, maxLength int) string {
	if len(input) > maxLength {
		input = input[:maxLength-2] + ".."
	}
	return input
}
