// Command zbase32-decode decodes zbase32 from arguments or lines of
// stdin.
//
// Usage:
//
//     zbase32-decode ZBASE32..
//     zbase32-decode <LINES_OF_ZBASE32
//
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tv42/zbase32"
)

var prog = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s ZBASE32..\n", prog)
	fmt.Fprintf(os.Stderr, "  %s <LINES_OF_ZBASE32\n", prog)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() > 0 {
		for _, input := range flag.Args() {
			decoded, err := zbase32.DecodeString(input)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := os.Stdout.Write(decoded); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := scanner.Text()
			decoded, err := zbase32.DecodeString(input)
			if err != nil {
				log.Fatalf("decoding input: %q: %v", input, err)
			}
			if _, err = os.Stdout.Write(decoded); err != nil {
				log.Fatal(err)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("reading standard input:", err)
		}
	}
}
