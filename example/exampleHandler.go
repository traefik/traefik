/*
Simple program to start a web server on a specified port
*/
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	name string
	port int
	help *bool
)

func init() {
	flag.StringVar(&name, "n", "", "Name of handler for messages")
	flag.IntVar(&port, "p", 0, "Port number to listen")
	help = flag.Bool("h", false, "Displays help message")
}

func usage() {
	fmt.Printf("Usage: example -n name -p port \n")
	os.Exit(2)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s: Received query %s!\n", name, r.URL.Path[1:])
}

func main() {
	flag.Parse()
	if *help || len(name) == 0 || port <= 0 {
		usage()
	}
	http.HandleFunc("/", handler)
	fmt.Printf("%s: Listening on :%d...\n", name, port)
	if er := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); er != nil {
		fmt.Printf("%s: Error from ListenAndServe: %s", name, er.Error())
		os.Exit(1)
	}
	fmt.Printf("%s: How'd we get past listen and serve???\n", name)
}
