package dns_test

import (
	"encoding/json"
	"fmt"

	"gopkg.in/ns1/ns1-go.v2/rest/model/data"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
	"gopkg.in/ns1/ns1-go.v2/rest/model/filter"
)

func ExampleZone() {
	z := dns.NewZone("example.com")

	fmt.Println(z)
	// Output:
	// example.com
}

// Example references https://ns1.com/articles/primary-dns-with-ns1
func ExamplePrimaryZone() {
	// Secondary/slave dns server info.
	secondary := dns.ZoneSecondaryServer{
		IP:     "1.2.3.4",
		Port:   53,
		Notify: true,
	}

	// Construct the primary/master zone.
	domain := "masterzone.example"

	masterZone := dns.NewZone(domain)
	masterZone.MakePrimary(secondary)

	b, _ := json.MarshalIndent(masterZone, "", "  ")

	fmt.Println(string(b))
	// Output:
	// {
	//   "zone": "masterzone.example",
	//   "primary": {
	//     "enabled": true,
	//     "secondaries": [
	//       {
	//         "ip": "1.2.3.4",
	//         "port": 53,
	//         "notify": true
	//       }
	//     ]
	//   }
	// }

}

func ExampleRecord() {
	// Construct the A record
	record := dns.NewRecord("test.com", "a", "A")
	record.TTL = 300

	// Construct primary answer(higher priority)
	pAns := dns.NewAv4Answer("1.1.1.1")
	pAns.Meta.Priority = 1
	pAns.Meta.Up = data.FeedPtr{FeedID: "feed1_id"}

	// Construct secondary answer(lower priority)
	sAns := dns.NewAv4Answer("2.2.2.2")
	sAns.Meta.Priority = 2
	sAns.Meta.Up = data.FeedPtr{FeedID: "feed2_id"}

	// Add both answers to record
	record.AddAnswer(pAns)
	record.AddAnswer(sAns)

	// Construct and add both filters to the record(ORDER MATTERS)
	record.AddFilter(filter.NewUp())
	record.AddFilter(filter.NewSelFirstN(1))

	// Add region 'test' to record(set as down)
	record.Regions["test"] = data.Region{Meta: data.Meta{Up: false}}

	fmt.Println(record)
	fmt.Println(record.TTL)

	fmt.Println("Primary answer:")
	fmt.Println(record.Answers[0])
	fmt.Println(record.Answers[0].Meta.Priority)
	fmt.Println(record.Answers[0].Meta.Up)

	fmt.Println("Secondary answer:")
	fmt.Println(record.Answers[1])
	fmt.Println(record.Answers[1].Meta.Priority)
	fmt.Println(record.Answers[1].Meta.Up)

	fmt.Println("First Filter in Chain:")
	fmt.Println(record.Filters[0].Type)
	fmt.Println(record.Filters[0].Config)

	fmt.Println("Second Filter in Chain:")
	fmt.Println(record.Filters[1].Type)
	fmt.Println(record.Filters[1].Config)

	fmt.Println("Regions:")
	fmt.Println(record.Regions["test"].Meta.Up)

	// Output:
	// a.test.com A
	// 300
	// Primary answer:
	// 1.1.1.1
	// 1
	// {feed1_id}
	// Secondary answer:
	// 2.2.2.2
	// 2
	// {feed2_id}
	// First Filter in Chain:
	// up
	// map[]
	// Second Filter in Chain:
	// select_first_n
	// map[N:1]
	// Regions:
	// false
}

func ExampleRecordLink() {
	// Construct the src record
	srcRecord := dns.NewRecord("test.com", "a", "A")
	srcRecord.TTL = 300
	srcRecord.Meta.Priority = 2

	linkedRecord := dns.NewRecord("test.com", "l", "A")
	linkedRecord.LinkTo(srcRecord.Domain)
	fmt.Println(linkedRecord)
	fmt.Println(linkedRecord.Meta)
	fmt.Println(linkedRecord.Answers)
	// Output:
	// l.test.com A
	// <nil>
	// []
}
