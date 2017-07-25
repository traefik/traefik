package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	api "gopkg.in/ns1/ns1-go.v2/rest"
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
	fmt.Println(header)
	fmt.Printf("%#v \n", v)
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func main() {
	zones, _, err := client.Zones.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, z := range zones {
		b, _ := json.MarshalIndent(z, "", "  ")
		fmt.Println(string(b))
	}

	// Construct/Create a zone.
	zone := "mylinktest.com"

	z := dns.NewZone(zone)
	z.NxTTL = 3600
	_, err = client.Zones.Create(z)
	if err != nil {
		// Ignore if zone already exists
		if err != api.ErrZoneExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", z, err)
		}
	}

	// Construct source record.
	sourceRec := dns.NewRecord(zone, "src", "A")

	ans1 := dns.NewAv4Answer("1.1.1.1")
	ans1.Meta.Up = true
	ans1.Meta.Country = []string{"US"}

	ans2 := dns.NewAv4Answer("1.1.1.1")
	ans2.Meta.Up = false
	ans2.Meta.Country = []string{"CA"}

	sourceRec.AddAnswer(ans1)
	sourceRec.AddAnswer(ans2)

	geotarget := filter.NewGeotargetCountry()
	selFirstN := filter.NewSelFirstN(1)

	sourceRec.AddFilter(geotarget)
	sourceRec.AddFilter(selFirstN)
	prettyPrint("source record:", sourceRec)

	// Create source record
	_, err = client.Records.Create(sourceRec)
	if err != nil {
		// Ignore if record already exists
		if err != api.ErrRecordExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", sourceRec, err)
		}
	}

	// Construct linked record
	linkedRec := dns.NewRecord(zone, "linked", "A")
	linkedRec.LinkTo(sourceRec.Domain)
	prettyPrint("linked record:", linkedRec)

	// Create linked record
	_, err = client.Records.Create(linkedRec)
	if err != nil {
		// Ignore if record already exists
		if err != api.ErrRecordExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", linkedRec, err)
		}
	}
}
