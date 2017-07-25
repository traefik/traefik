package eventsource_test

import (
	"fmt"
	"github.com/donovanhide/eventsource"
	"net"
	"net/http"
)

func ExampleErrorHandlingStream() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return
	}
	defer listener.Close()
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Something wrong.", 500)
	})
	go http.Serve(listener, nil)

	_, err = eventsource.Subscribe("http://127.0.0.1:8080/stream", "")
	if err != nil {
		if serr, ok := err.(eventsource.SubscriptionError); ok {
			fmt.Printf("Status code: %d\n", serr.Code)
			fmt.Printf("Message: %s\n", serr.Message)
		} else {
			fmt.Println("failed to subscribe")
		}
	}

	// Output:
	// Status code: 500
	// Message: Something wrong.
}
