package main

import (
    "fmt"
    "net/http"
    "os"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Handler2: Hi there, I love %s!\n", r.URL.Path[1:])
}

func main() {
    fmt.Printf("Sample handler 2\n")
    http.HandleFunc("/", handler)
    fmt.Printf("Listening on :8082...\n")
    if er := http.ListenAndServe(":8082", nil); er != nil {
        fmt.Printf("handler2: Error from ListenAndServe: %s", er.Error())
	os.Exit(1)
	}
    fmt.Printf("Handler2: How'd we get past listen and serve???\n")
}
