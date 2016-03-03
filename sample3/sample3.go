package main

import (
	"fmt"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Handler3: Hi there, I love %s!\n", r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handler)
	if er := http.ListenAndServe(":8083", nil); er != nil {
		fmt.Printf("handler3: Error from ListenAndServe: %s", er.Error())
		os.Exit(1)
	}
}
