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
	mjl, _, err := client.Jobs.List()
	if err != nil {
		log.Fatal(err)
	}
	for _, mj := range mjl {
		prettyPrint("monitor:", mj)
	}

	tcpJob := &monitor.Job{
		Type: "tcp",
		Name: "myhost.com:443 Monitor",
		Config: monitor.Config{
			"host": "1.2.3.4",
			"port": 443,
			"send": "HEAD / HTTP/1.0\r\n\r\n",
			"ssl":  true,
		},
		Rules: []*monitor.Rule{
			&monitor.Rule{
				Key:        "output",
				Value:      "200 OK",
				Comparison: "contains",
			},
			&monitor.Rule{
				Key:        "connect",
				Value:      200,
				Comparison: "<=",
			},
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

	_, err = client.Jobs.Create(tcpJob)
	if err != nil {
		log.Fatal(err)
	}
	prettyPrint("tcp job:", tcpJob)

	// Deactivate the job
	tcpJob.Deactivate()
	_, err = client.Jobs.Update(tcpJob)
	if err != nil {
		log.Fatal(err)
	}
	prettyPrint("tcp job deactivated:", tcpJob)

	// _, err = client.Jobs.Delete(tcpJob.ID)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
