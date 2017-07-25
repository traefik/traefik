// Example referencing https://ns1.com/articles/automated-failover
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	api "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/data"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
	"gopkg.in/ns1/ns1-go.v2/rest/model/filter"
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
	log.Println(header)
	log.Printf("%#v \n", v)
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Println(string(b))
}

// Define the type of update data we will send to the datasource.
// update maps feed labels to metadata.
type update map[string]data.Meta

func main() {
	// Create the zone(if it doesnt already exist).
	// all we need to create zone is the domain name.
	domain := "myfailover.com"
	z := dns.NewZone(domain)
	_, err := client.Zones.Create(z)
	if err != nil {
		log.Fatal(err)
	}

	// Construct an NSONE API data source.
	s := data.NewSource("my api source", "nsone_v1")
	prettyPrint("Data Source:", s)

	// Create the nsone_v1 api data source.
	// Note: this does not create the associated feeds.
	_, err = client.DataSources.Create(s)
	if err != nil {
		log.Fatal(err)
	}

	// Construct feeds which will drive the meta data for each answer.
	//  We'll use the id of these feeds when we connect the feeds to the
	//  answer meta below.
	feeds := map[string]*data.Feed{}

	// Construct the buffalo data feed.
	feeds["buf"] = data.NewFeed(
		"Buffalo Feed",
		data.Config{"label": "Buffalo-US"})

	// Construct the london data feed.
	feeds["lon"] = data.NewFeed(
		"London Feed",
		data.Config{"label": "London-UK"})

	// Create the buf/lon feeds through the rest api.
	for _, f := range feeds {
		prettyPrint("Data Feed:", f)
		_, err = client.DataFeeds.Create(s.ID, f)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Construct the A record with two answers(US, UK).
	r := dns.NewRecord(domain, "a", "A")

	pAns := dns.NewAv4Answer("1.1.1.1")          // Construct the PRIMARY answer(with BUFFALO feed).
	pAns.Meta.Priority = 1                       // (static) primary has higher priority
	pAns.Meta.Up = data.FeedPtr{feeds["buf"].ID} // (dynamic) connect the primary answer to the Buffalo feed.

	sAns := dns.NewAv4Answer("2.2.2.2")          // Construct the SECONDARY answer(with LONDON feed).
	sAns.Meta.Priority = 2                       // (static) secondary has lower priority.
	sAns.Meta.Up = data.FeedPtr{feeds["lon"].ID} // (dynamic) connect the secondary answer to the London feed.

	r.AddAnswer(pAns) // Add primary answer to record
	r.AddAnswer(sAns) // Add secondary answer to record

	// Construct and add both filters to the record(order matters).
	r.AddFilter(filter.NewUp())
	r.AddFilter(filter.NewSelFirstN(1))

	// Helper to show record in json before sending PUT
	prettyPrint("Record:", r)

	// Create the record with REST API
	_, err = client.Records.Create(r)
	if err != nil {
		log.Fatal(err)
	}

	// Flip which answer is 'Up' 5 times below.

	// Make an 'abort' goroutine for cancelling loop
	abort := make(chan struct{})
	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte
		abort <- struct{}{}
	}()

	// Every 5 sec, reverse which answer is 'up'.
	fmt.Println("Flipping answer every 5 sec. Press return to abort.")
	tick := time.Tick(5 * time.Second)
	var buffaloUp bool
	pool := z.NetworkPools[0]
	for countdown := 5; countdown > 0; countdown-- {
		log.Println(countdown)
		select {
		case <-tick:
			// Update the buffalo feed
			d := update{"Buffalo-US": data.Meta{Up: buffaloUp}}
			prettyPrint("Publishing:", d)
			_, err = client.DataSources.Publish(s.ID, d)
			if err != nil {
				log.Fatal(err)
			}

			if buffaloUp {
				log.Printf("'dig %s @dns1.%s.nsone.net' will respond with answer %s \n",
					r.Domain, pool, pAns)
			} else {
				log.Printf("'dig %s @dns1.%s.nsone.net' will respond with answer %s \n",
					r.Domain, pool, sAns)
			}
			// Toggle status for next update.
			buffaloUp = !buffaloUp

		case <-abort:
			log.Fatal("Aborted")
		}
	}
}
