package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	api "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/monitor"
)

var client *api.Client

// Helper that initializes rest api client from environment variable.
func init() {
	k := os.Getenv("NS1_APIKEY")
	if k == "" {
		fmt.Println("NS1_APIKEY environment variable is not set, giving up")
	}

	httpClient := &http.Client{Timeout: time.Second * 10}
	// Adds logging to each http request.
	doer := api.Decorate(httpClient, api.Logging(log.New(os.Stdout, "", log.LstdFlags)))
	client = api.NewClient(doer, api.SetAPIKey(k))
}

func prettyPrint(header string, v interface{}) {
	fmt.Println(header)
	fmt.Printf("%#v \n", v)
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func main() {
	nl, _, err := client.Notifications.List()
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range nl {
		prettyPrint("notification:", n)
	}

	webhook := monitor.NewWebNotification("test.com/test")
	nList := monitor.NewNotifyList("my list", webhook)
	_, err = client.Notifications.Create(nList)
	if err != nil {
		log.Fatal(err)
	}
	prettyPrint("NotifyList:", nList)
}
