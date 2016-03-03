package main

import (
	"fmt"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Handler1: Hi there, I love %s!\n", r.URL.Path[1:])
}

func main() {
	fmt.Printf("Sample handler 1\n")
	http.HandleFunc("/", handler)
	fmt.Printf("handler1: Listening on :8081...\n")
	if er := http.ListenAndServe(":8081", nil); er != nil {
		fmt.Printf("handler1: Error from ListenAndServe: %s", er.Error())
		os.Exit(1)
	}
	fmt.Printf("sample1 How'd we get past listen and serve???\n")
}
