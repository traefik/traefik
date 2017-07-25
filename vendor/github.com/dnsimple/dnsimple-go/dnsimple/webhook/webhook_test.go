package webhook

import (
	"testing"
)

func TestParse(t *testing.T) {
	payload := `{"data": {"webhook": {"id": 25, "url": "https://webhook.test"}}, "name": "webhook.create", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "d6362e1f-310b-4009-a29d-ce76c849d32c"}`

	event, err := Parse([]byte(payload))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if want, got := "webhook.create", event.EventName(); want != got {
		t.Errorf("Parse event Name expected to be %v, got %v", want, got)
	}

	eventAccount := event.EventHeader().Account
	if want, got := "User", eventAccount.Display; want != got {
		t.Errorf("Parse event Account.Display expected to be %v, got %v", want, got)
	}

	_, ok := event.(*WebhookEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}
