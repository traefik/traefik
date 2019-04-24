# Event API

The Instana Go Event API is a simple method to report custom events into your dashboard.  These events can range from generic change events to more severe events tied to a specific host or service.

Events are a versatile way to broaden the knowledge base that Instana has to continuously monitor, learn from and alert on your environment.

# Example: A High Latency Service Event

Suppose that we have a remote point of presence monitoring a specific public microservice (_games_) in Eastern Asia that is critical to our infrastructure for that local area.  When that monitor detects slow response times to it's queries, it reports a custom high latency event to Instana.

The following example shows how such Go code would send a _critical_ event on the service _games_.

```Go
package main

import (
	"time"
	"github.com/instana/golang-sensor"
)

func main() {
	instana.InitSensor(&instana.Options{Service:  service})

	go forever()
	select {}
}

func forever() {
	for {
		instana.SendServiceEvent("games",
			"Games High Latency", "Games - High latency from East Asia POP.",
			instana.SeverityCritical, 1*time.Second)
		time.Sleep(30 * time.Second)
	}
}
```

The critical event is reported to the Instana _Service Quality Engine_, it is logged to the dashboard and directly affects the state of the _games_ service:

![games_service_event](https://disznc.s3.amazonaws.com/Instana-Event-API-Service-Event-games-2017-07-18.png)
