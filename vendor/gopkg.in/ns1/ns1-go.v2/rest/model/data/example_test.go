package data_test

import (
	"fmt"

	"gopkg.in/ns1/ns1-go.v2/rest/model/data"
)

func ExampleSource() {
	// Construct an NSONE API data source.
	source := data.NewSource("my api source", "nsone_v1")
	fmt.Println(source.ID) // will be empty string
	fmt.Println(source.Name)
	fmt.Println(source.Type)
	// Output:
	// my api source
	// nsone_v1
}

func ExampleFeed() {

	// Construct the london data feed.
	feed := data.NewFeed(
		"London Feed",
		data.Config{"label": "London-UK"})
	fmt.Println(feed.ID) // will be empty string
	fmt.Println(feed.Name)
	fmt.Println(feed.Config)
	// Output:
	// London Feed
	// map[label:London-UK]
}

func ExampleMeta() {
	feedID := "feed_id"

	meta := data.Meta{}
	meta.Priority = 1
	meta.Up = data.FeedPtr{FeedID: feedID}
	fmt.Println(meta.Connections) // will be nil
	fmt.Println(meta.Priority)
	fmt.Println(meta.Up)
	// Output:
	// <nil>
	// 1
	// {feed_id}
}

func ExampleRegions() {
	feedPtr := data.FeedPtr{FeedID: "feed_id"}

	regions := data.Regions{}
	// Set a regions' 'up' metavalue to false('down').
	regions["some_region"] = data.Region{
		Meta: data.Meta{Up: false},
	}
	// Set a regions' 'connections' metavalue to receive from a feed.
	regions["other_region"] = data.Region{
		Meta: data.Meta{Connections: feedPtr},
	}
	fmt.Println(regions["some_region"].Meta.Up)
	fmt.Println(regions["some_region"].Meta.Priority)
	fmt.Println(regions["other_region"].Meta.Connections)
	fmt.Println(regions["other_region"].Meta.Priority)
	// Output:
	// false
	// <nil>
	// {feed_id}
	// <nil>
}
