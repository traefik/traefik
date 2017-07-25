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
	domain := "myzonetest.com"

	z := dns.NewZone(domain)
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

	// Update the zone.
	z.Retry = 5401
	_, err = client.Zones.Update(z)
	if err != nil {
		log.Fatal(err)
	}

	// Add an A record with a single static answer.
	orchidRec := dns.NewRecord(domain, "orchid", "A")
	orchidRec.AddAnswer(dns.NewAv4Answer("2.2.2.2"))
	_, err = client.Records.Create(orchidRec)
	if err != nil {
		switch {
		case err == api.ErrRecordExists:
			// Ignore if record already exists
			log.Printf("Create %s: %s \n", orchidRec, err)
		case err == api.ErrZoneMissing:
			log.Printf("Create %s: %s \n", orchidRec, err)
			return
		default:
			log.Fatal(err)
		}
	}

	orchidRec.TTL = 333
	_, err = client.Records.Update(orchidRec)
	if err != nil {
		switch {
		case err == api.ErrRecordExists:
			// Ignore if record already exists
			log.Printf("Update %s: %s \n", orchidRec, err)
		case err == api.ErrZoneMissing:
			log.Printf("Update %s: %s \n", orchidRec, err)
			return
		default:
			log.Fatal(err)
		}
	}

	fmt.Printf("%#v \n", orchidRec)
	bRec, _ := json.MarshalIndent(orchidRec, "", "  ")
	fmt.Println(string(bRec))

	// Add an A record with two static answers.
	honeyRec := dns.NewRecord(domain, "honey", "A")
	honeyRec.Answers = []*dns.Answer{
		dns.NewAv4Answer("1.2.3.4"),
		dns.NewAv4Answer("5.6.7.8"),
	}
	_, err = client.Records.Create(honeyRec)
	if err != nil {
		// Ignore if record already exists
		if err != api.ErrRecordExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", honeyRec, err)
		}
	}

	// Add a cname
	potRec := dns.NewRecord(domain, "pot", "CNAME")
	potRec.AddAnswer(dns.NewCNAMEAnswer("honey.test.com"))
	_, err = client.Records.Create(potRec)
	if err != nil {
		// Ignore if record already exists
		if err != api.ErrRecordExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", potRec, err)
		}
	}

	// Add a MX with two answers, priority 5 and 10
	mailRec := dns.NewRecord(domain, "mail", "MX")
	mailRec.Answers = []*dns.Answer{
		dns.NewMXAnswer(5, "mail1.test.com"),
		dns.NewMXAnswer(10, "mail2.test.com"),
	}
	_, err = client.Records.Create(mailRec)
	if err != nil {
		// Ignore if record already exists
		if err != api.ErrRecordExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", mailRec, err)
		}
	}

	// Add a AAAA, specify ttl of 300 seconds
	aaaaRec := dns.NewRecord(domain, "honey6", "AAAA")
	aaaaRec.TTL = 300
	aaaaRec.AddAnswer(dns.NewAv6Answer("2607:f8b0:4006:806::1010"))
	_, err = client.Records.Create(aaaaRec)
	if err != nil {
		// Ignore if record already exists
		if err != api.ErrRecordExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", aaaaRec, err)
		}
	}

	// Add an A record using full answer format to specify 2 answers with meta data.
	// ensure edns-client-subnet is in use, and add two filters: geotarget_country,
	// and select_first_n, which has a filter config option N set to 1.
	bumbleRec := dns.NewRecord(domain, "bumble", "A")

	usAns := dns.NewAv4Answer("1.1.1.1")
	usAns.Meta.Up = false
	usAns.Meta.Country = []string{"US"}

	fraAns := dns.NewAv4Answer("1.1.1.1")
	fraAns.Meta.Up = true
	fraAns.Meta.Country = []string{"FR"}

	bumbleRec.AddAnswer(usAns)
	bumbleRec.AddAnswer(fraAns)

	geotarget := filter.NewGeotargetCountry()
	selFirstN := filter.NewSelFirstN(1)

	bumbleRec.AddFilter(geotarget)
	bumbleRec.AddFilter(selFirstN)

	_, err = client.Records.Create(bumbleRec)
	if err != nil {
		// Ignore if record already exists
		if err != api.ErrRecordExists {
			log.Fatal(err)
		} else {
			log.Printf("Create %s: %s \n", bumbleRec, err)
		}
	}

	// _, err = client.Zones.Delete(domain)
	// if err != nil {
	// 	// Ignore if zone doesnt yet exist
	// 	if err != api.ErrZoneMissing {
	// 		log.Fatal(err)
	// 	} else {
	// 		log.Printf("Delete %s: %s \n", z, err)
	// 	}
	// }

}
