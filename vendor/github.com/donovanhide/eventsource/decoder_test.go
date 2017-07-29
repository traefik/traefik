package eventsource

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		rawInput     string
		wantedEvents []*publication
	}{
		{
			rawInput:     "event: eventName\ndata: {\"sample\":\"value\"}\n\n",
			wantedEvents: []*publication{{event: "eventName", data: "{\"sample\":\"value\"}"}},
		},
		{
			// the newlines should not be parsed as empty event
			rawInput:     "\n\n\nevent: event1\n\n\n\n\nevent: event2\n\n",
			wantedEvents: []*publication{{event: "event1"}, {event: "event2"}},
		},
	}

	for _, test := range tests {
		decoder := NewDecoder(strings.NewReader(test.rawInput))
		i := 0
		for {
			event, err := decoder.Decode()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Unexpected error on decoding event: %s", err)
			}

			if !reflect.DeepEqual(event, test.wantedEvents[i]) {
				t.Fatalf("Parsed event %+v does not equal wanted event %+v", event, test.wantedEvents[i])
			}
			i++
		}
		if i != len(test.wantedEvents) {
			t.Fatalf("Unexpected number of events: %d does not equal wanted: %d", i, len(test.wantedEvents))
		}
	}
}
