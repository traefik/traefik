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
	client = api.NewClient(httpClient, api.SetAPIKey(k))
}

func prettyPrint(header string, v interface{}) {
	fmt.Println(header)
	fmt.Printf("%#v \n", v)
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func main() {
	// Create a monitoring job
	tcpJob := &monitor.Job{
		Type: "tcp",
		Name: "myhost.com:443 Monitor",
		Config: monitor.Config{
			"host": "1.2.3.4",
			"port": 443,
			"send": "HEAD / HTTP/1.0\r\n\r\n",
			"ssl":  true,
		},
		Regions:        []string{"lga", "sjc"},
		Frequency:      10,
		Active:         true,
		Policy:         "quorum",
		RegionScope:    "fixed",
		NotifyRegional: false,
		NotifyFailback: true,
		NotifyRepeat:   0,
		NotifyDelay:    65,
	}

	_, err := client.Jobs.Create(tcpJob)
	if err != nil {
		log.Fatal(err)
	}

	// Check a monitoring jobs history
	slgs, _, err := client.Jobs.History(tcpJob.ID,
		api.SetTimeParam("start", time.Now().Add(-24*time.Hour)),
		api.SetTimeParam("end", time.Now()),
		api.SetBoolParam("exact", true),
	)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.MarshalIndent(slgs, "", "  ")
	fmt.Printf("Status Logs: %s\n", string(b))
}
