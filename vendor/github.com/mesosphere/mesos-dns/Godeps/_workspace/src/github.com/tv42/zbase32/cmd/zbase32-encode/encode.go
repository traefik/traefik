// Command zbase32-encode encodes its standard input as zbase32.
//
// Usage:
//
//     zbase32-encode <FILE
//
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/tv42/zbase32"
)

var prog = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s <FILE\n", prog)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 0 {
		usage()
		os.Exit(1)
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	encoded := make([]byte, zbase32.EncodedLen(len(data)))
	n := zbase32.Encode(encoded, data)
	encoded = encoded[:n]
	encoded = append(encoded, '\n')
	if _, err := os.Stdout.Write(encoded); err != nil {
		log.Fatal(err)
	}
}
