package webhook

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/dnsimple/dnsimple-go/dnsimple"
)

var regexpUUID = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

func TestParseGenericEvent(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "hosted", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2016-02-07T14:46:29.142Z", "expires_on": null, "updated_at": "2016-02-07T14:46:29.142Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": null}}, "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "name": "domain.create", "api_version": "v2", "request_identifier": "096bfc29-2bf0-40c6-991b-f03b1f8521f1"}`

	event := &GenericEvent{}
	err := ParseGenericEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.create", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}

	data := event.Data.(map[string]interface{})
	if want, got := "example.com", data["domain"].(map[string]interface{})["name"]; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	if err != nil {
		t.Fatalf("Parse returned error when parsing: %v", err)
	}
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseContactEvent_Contact_Create(t *testing.T) {
	payload := `{"data": {"contact": {"id": 1, "fax": "+39 339 1111111", "city": "Rome", "label": "Webhook", "phone": "+39 339 0000000", "country": "IT", "address1": "Some Street", "address2": "", "job_title": "Developer", "last_name": "Contact", "account_id": 1010, "created_at": "2016-02-13T13:11:29.388Z", "first_name": "Example", "updated_at": "2016-02-13T13:11:29.388Z", "postal_code": "12037", "email": "example@example.com", "state_province": "Italy", "organization_name": "Company"}}, "name": "contact.create", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "3be0422c-8ca2-44d9-95d6-9f045b938781"}
`

	event := &ContactEvent{}
	err := ParseContactEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "contact.create", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "Webhook", event.Contact.Label; want != got {
		t.Errorf("ParseEvent Contact.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ContactEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseContactEvent_Contact_Update(t *testing.T) {
	payload := `{"data": {"contact": {"id": 1, "fax": "+39 339 1111111", "city": "Rome", "label": "Webhook", "phone": "+39 339 0000000", "country": "IT", "address1": "Some Street", "address2": "", "job_title": "Developer", "last_name": "Contact", "account_id": 1010, "created_at": "2016-02-13T13:11:29.388Z", "first_name": "Example", "updated_at": "2016-02-13T13:11:29.388Z", "postal_code": "12037", "email": "example@example.com", "state_province": "Italy", "organization_name": "Company"}}, "name": "contact.update", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "3be0422c-8ca2-44d9-95d6-9f045b938781"}
`

	event := &ContactEvent{}
	err := ParseContactEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "contact.update", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "Webhook", event.Contact.Label; want != got {
		t.Errorf("ParseEvent Contact.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ContactEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseContactEvent_Contact_Delete(t *testing.T) {
	payload := `{"data": {"contact": {"id": 1, "fax": "+39 339 1111111", "city": "Rome", "label": "Webhook", "phone": "+39 339 0000000", "country": "IT", "address1": "Some Street", "address2": "", "job_title": "Developer", "last_name": "Contact", "account_id": 1010, "created_at": "2016-02-13T13:11:29.388Z", "first_name": "Example", "updated_at": "2016-02-13T13:11:29.388Z", "postal_code": "12037", "email": "example@example.com", "state_province": "Italy", "organization_name": "Company"}}, "name": "contact.delete", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "3be0422c-8ca2-44d9-95d6-9f045b938781"}
`

	event := &ContactEvent{}
	err := ParseContactEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "contact.delete", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "Webhook", event.Contact.Label; want != got {
		t.Errorf("ParseEvent Contact.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ContactEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_AutoRenewalEnable(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2013-05-17T12:58:57.459Z", "expires_on": "2016-05-17", "updated_at": "2016-02-13T12:33:22.723Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 11}}, "name": "domain.auto_renewal_enable", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "91d47480-c2ce-411c-ac95-b5b54f346bff"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.auto_renewal_enable", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_AutoRenewalDisable(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2013-05-17T12:58:57.459Z", "expires_on": "2016-05-17", "updated_at": "2016-02-13T12:33:22.723Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 11}}, "name": "domain.auto_renewal_disable", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "91d47480-c2ce-411c-ac95-b5b54f346bff"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.auto_renewal_disable", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_Create(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "hosted", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2016-02-07T14:46:29.142Z", "expires_on": null, "updated_at": "2016-02-07T14:46:29.142Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": null}}, "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "name": "domain.create", "api_version": "v2", "request_identifier": "096bfc29-2bf0-40c6-991b-f03b1f8521f1"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.create", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_Delete(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "hosted", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2016-02-07T14:46:29.142Z", "expires_on": null, "updated_at": "2016-02-07T14:46:29.142Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": null}}, "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "name": "domain.delete", "api_version": "v2", "request_identifier": "3e625f1c-3e8b-48fc-9326-9489f4b60e52"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.delete", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_Register(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2016-02-24T21:53:38.878Z", "expires_on": "2017-02-24", "updated_at": "2016-02-24T22:22:27.025Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 2}}, "name": "domain.register", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "8c92b76f-125d-43c0-8e72-b911e4bdbd96"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.register", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_Renew(t *testing.T) {
	payload := `{"data": {"auto": true, "domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": true, "created_at": "2014-04-01T08:37:15.729Z", "expires_on": "2017-04-01", "updated_at": "2016-03-04T07:40:02.334Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 2}}, "name": "domain.renew", "actor": {"id": "system", "entity": "dnsimple", "pretty": "support@dnsimple.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "9e8e47ef-f303-4455-b496-875f70ab5c00"}
`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.renew", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if event.Auto != true {
		t.Errorf("ParseEvent auto expected to be %v", true)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_DelegationChange(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2016-01-16T16:08:50.649Z", "expires_on": "2018-01-16", "updated_at": "2016-03-24T20:30:05.895Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 2}, "name_servers": ["ns1.dnsimple.com", "ns2.dnsimple.com"]}, "name": "domain.delegation_change", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "a07b97ac-6275-4e15-92dd-1d45881f7a2c"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))

	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.delegation_change", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent RequestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}
	if want, got := (&dnsimple.Delegation{"ns1.dnsimple.com", "ns2.dnsimple.com"}), event.Delegation; !reflect.DeepEqual(want, got) {
		t.Errorf("ParseEvent Delegation expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_RegistrantChange(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2016-01-16T16:08:50.649Z", "expires_on": "2018-01-16", "updated_at": "2016-03-24T20:30:05.895Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 2}, "registrant": {"id": 2, "fax": "+39 339 1111111", "city": "Rome", "label": "Webhook", "phone": "+39 339 0000000", "country": "IT", "address1": "Some Street", "address2": "", "job_title": "Developer", "last_name": "Contact", "account_id": 1010, "created_at": "2016-02-13T13:11:29.388Z", "first_name": "Example", "updated_at": "2016-02-13T13:11:29.388Z", "postal_code": "12037", "email": "example@example.com", "state_province": "Italy", "organization_name": "Company"}}, "name": "domain.registrant_change", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "0391e4e2-7614-41bf-a7bd-7ba01232e434"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))

	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.registrant_change", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent RequestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}
	if want, got := "Webhook", event.Registrant.Label; want != got {
		t.Errorf("ParseEvent Registrant.Label expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_ResolutionDisable(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2013-05-17T12:58:57.459Z", "expires_on": "2016-05-17", "updated_at": "2016-03-22T09:17:06.313Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 2}}, "name": "domain.resolution_disable", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "b6eb3eae-8f3b-476a-997a-564d830df015"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))

	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.resolution_disable", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent RequestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_ResolutionEnable(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2013-05-17T12:58:57.459Z", "expires_on": "2016-05-17", "updated_at": "2016-03-22T09:17:06.313Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 2}}, "name": "domain.resolution_enable", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "b6eb3eae-8f3b-476a-997a-564d830df015"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))

	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.resolution_enable", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent RequestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_TokenReset(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": false, "created_at": "2013-05-17T12:58:57.459Z", "expires_on": "2016-05-17", "updated_at": "2016-02-07T23:26:16.368Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 11549}}, "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "name": "domain.token_reset", "api_version": "v2", "request_identifier": "33537afb-0e99-49ec-b69e-93ffcc3db763"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.token_reset", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseDomainEvent_Domain_Transfer(t *testing.T) {
	//payload := `{"data": {"domain": {"id": 6637, "name": "example.com", "state": "hosted", "token": "domain-token", "account_id": 24, "auto_renew": false, "created_at": "2016-03-24T21:03:49.392Z", "expires_on": null, "updated_at": "2016-03-24T21:03:49.392Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 409}}, "name": "domain.transfer:started", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "49901af0-569e-4acd-900f-6edf0ebc123c"}`
	payload := `{"data": {"domain": {"id": 6637, "name": "example.com", "state": "hosted", "token": "domain-token", "account_id": 24, "auto_renew": false, "created_at": "2016-03-24T21:03:49.392Z", "expires_on": null, "updated_at": "2016-03-24T21:03:49.392Z", "unicode_name": "example.com", "private_whois": false, "registrant_id": 409}}, "name": "domain.transfer", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "49901af0-569e-4acd-900f-6edf0ebc123c"}`

	event := &DomainEvent{}
	err := ParseDomainEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "domain.transfer", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*DomainEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseEmailForwardEvent_EmailForward_Create(t *testing.T) {
	payload := `{"data": {"email_forward": {"id": 1, "to": "example@example.io", "from": "hello@example.com", "domain_id": 2, "created_at": "2016-03-24T19:40:09.357Z", "updated_at": "2016-03-24T19:40:09.357Z"}}, "name": "email_forward.create", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "f5f33be8-7074-4fa1-a296-4ddd9003c4a4"}`

	event := &EmailForwardEvent{}
	err := ParseEmailForwardEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "email_forward.create", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "hello@example.com", event.EmailForward.From; want != got {
		t.Errorf("ParseEvent EmailForward.From expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*EmailForwardEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseEmailForwardEvent_EmailForward_Delete(t *testing.T) {
	payload := `{"data": {"email_forward": {"id": 1, "to": "example@example.io", "from": "hello@example.com", "domain_id": 2, "created_at": "2016-03-24T19:40:09.357Z", "updated_at": "2016-03-24T19:40:09.357Z"}}, "name": "email_forward.delete", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "f5f33be8-7074-4fa1-a296-4ddd9003c4a4"}`

	event := &EmailForwardEvent{}
	err := ParseEmailForwardEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "email_forward.delete", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "hello@example.com", event.EmailForward.From; want != got {
		t.Errorf("ParseEvent EmailForward.From expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*EmailForwardEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseWebhookEvent_Webhook_Create(t *testing.T) {
	payload := `{"data": {"webhook": {"id": 25, "url": "https://webhook.test"}}, "name": "webhook.create", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "d6362e1f-310b-4009-a29d-ce76c849d32c"}`

	event := &WebhookEvent{}
	err := ParseWebhookEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "webhook.create", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "https://webhook.test", event.Webhook.URL; want != got {
		t.Errorf("ParseEvent Webhook.URL expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*WebhookEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseWebhookEvent_Webhook_Delete(t *testing.T) {
	payload := `{"data": {"webhook": {"id": 23, "url": "https://webhook.test"}}, "name": "webhook.delete", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "756bad5c-b432-43be-821a-2f4c4f285d19"}`

	event := &WebhookEvent{}
	err := ParseWebhookEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "webhook.delete", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "https://webhook.test", event.Webhook.URL; want != got {
		t.Errorf("ParseEvent Webhook.URL expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*WebhookEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseWhoisPrivacyEvent_WhoisPrivacy_Disable(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": true, "created_at": "2016-01-17T17:10:41.187Z", "expires_on": "2017-01-17", "updated_at": "2016-01-17T17:11:19.797Z", "unicode_name": "example.com", "private_whois": true, "registrant_id": 2}, "whois_privacy": {"id": 3, "enabled": true, "domain_id": 1, "created_at": "2016-01-17T17:10:50.713Z", "expires_on": "2017-01-17", "updated_at": "2016-03-20T16:45:57.409Z"}}, "name": "whois_privacy.disable", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "e3861a08-a771-4049-abc4-715a3f7b7d6f"}`

	event := &WhoisPrivacyEvent{}
	err := ParseWhoisPrivacyEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "whois_privacy.disable", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}
	if want, got := 3, event.WhoisPrivacy.ID; want != got {
		t.Errorf("ParseEvent WhoisPrivacy.ID expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*WhoisPrivacyEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseWhoisPrivacyEvent_WhoisPrivacy_Enable(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": true, "created_at": "2016-01-17T17:10:41.187Z", "expires_on": "2017-01-17", "updated_at": "2016-01-17T17:11:19.797Z", "unicode_name": "example.com", "private_whois": true, "registrant_id": 2}, "whois_privacy": {"id": 3, "enabled": true, "domain_id": 1, "created_at": "2016-01-17T17:10:50.713Z", "expires_on": "2017-01-17", "updated_at": "2016-03-20T16:45:57.409Z"}}, "name": "whois_privacy.enable", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "e3861a08-a771-4049-abc4-715a3f7b7d6f"}`

	event := &WhoisPrivacyEvent{}
	err := ParseWhoisPrivacyEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "whois_privacy.enable", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}
	if want, got := 3, event.WhoisPrivacy.ID; want != got {
		t.Errorf("ParseEvent WhoisPrivacy.ID expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*WhoisPrivacyEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseWhoisPrivacyEvent_WhoisPrivacy_Purchase(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": true, "created_at": "2016-01-17T17:10:41.187Z", "expires_on": "2017-01-17", "updated_at": "2016-01-17T17:11:19.797Z", "unicode_name": "example.com", "private_whois": true, "registrant_id": 2}, "whois_privacy": {"id": 3, "enabled": true, "domain_id": 1, "created_at": "2016-01-17T17:10:50.713Z", "expires_on": "2017-01-17", "updated_at": "2016-03-20T16:45:57.409Z"}}, "name": "whois_privacy.purchase", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "e3861a08-a771-4049-abc4-715a3f7b7d6f"}`

	event := &WhoisPrivacyEvent{}
	err := ParseWhoisPrivacyEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "whois_privacy.purchase", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}
	if want, got := 3, event.WhoisPrivacy.ID; want != got {
		t.Errorf("ParseEvent WhoisPrivacy.ID expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*WhoisPrivacyEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseWhoisPrivacyEvent_WhoisPrivacy_Renew(t *testing.T) {
	payload := `{"data": {"domain": {"id": 1, "name": "example.com", "state": "registered", "token": "domain-token", "account_id": 1010, "auto_renew": true, "created_at": "2016-01-17T17:10:41.187Z", "expires_on": "2017-01-17", "updated_at": "2016-01-17T17:11:19.797Z", "unicode_name": "example.com", "private_whois": true, "registrant_id": 2}, "whois_privacy": {"id": 3, "enabled": true, "domain_id": 1, "created_at": "2016-01-17T17:10:50.713Z", "expires_on": "2017-01-17", "updated_at": "2016-03-20T16:45:57.409Z"}}, "name": "whois_privacy.renew", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "e3861a08-a771-4049-abc4-715a3f7b7d6f"}`

	event := &WhoisPrivacyEvent{}
	err := ParseWhoisPrivacyEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "whois_privacy.renew", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Domain.Name; want != got {
		t.Errorf("ParseEvent Domain.Name expected to be %v, got %v", want, got)
	}
	if want, got := 3, event.WhoisPrivacy.ID; want != got {
		t.Errorf("ParseEvent WhoisPrivacy.ID expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*WhoisPrivacyEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseZoneEvent_Zone_Create(t *testing.T) {
	payload := `{"data": {"zone": {"id": 1, "name": "example.com", "reverse": false, "account_id": 1010, "created_at": "2016-03-24T20:08:54.109Z", "updated_at": "2016-03-24T20:08:54.109Z"}}, "name": "zone.create", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "1f4d7325-92a3-4504-a29d-0f664dfe7356"}`

	event := &ZoneEvent{}
	err := ParseZoneEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "zone.create", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Zone.Name; want != got {
		t.Errorf("ParseEvent Zone.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ZoneEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseZoneEvent_Zone_Delete(t *testing.T) {
	payload := `{"data": {"zone": {"id": 1, "name": "example.com", "reverse": false, "account_id": 1010, "created_at": "2016-03-24T20:08:54.109Z", "updated_at": "2016-03-24T20:08:54.109Z"}}, "name": "zone.delete", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "1f4d7325-92a3-4504-a29d-0f664dfe7356"}`

	event := &ZoneEvent{}
	err := ParseZoneEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "zone.delete", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "example.com", event.Zone.Name; want != got {
		t.Errorf("ParseEvent Zone.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ZoneEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseZoneRecordEvent_ZoneRecord_Create(t *testing.T) {
	payload := `{"data": {"zone_record": {"id": 1, "ttl": 60, "name": "_frame", "type": "TXT", "content": "https://dnsimple.com/", "zone_id": "example.com", "priority": null, "parent_id": null, "created_at": "2016-02-22T21:06:48.957Z", "updated_at": "2016-02-22T21:23:22.503Z", "system_record": false}}, "name": "zone_record.create", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "8f6cd405-2c87-453b-8b95-7a296982e4b8"}`

	event := &ZoneRecordEvent{}
	err := ParseZoneRecordEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "zone_record.create", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "_frame", event.ZoneRecord.Name; want != got {
		t.Errorf("ParseEvent ZoneRecord.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ZoneRecordEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseZoneRecordEvent_ZoneRecord_Update(t *testing.T) {
	payload := `{"data": {"zone_record": {"id": 1, "ttl": 60, "name": "_frame", "type": "TXT", "content": "https://dnsimple.com/", "zone_id": "example.com", "priority": null, "parent_id": null, "created_at": "2016-02-22T21:06:48.957Z", "updated_at": "2016-02-22T21:23:22.503Z", "system_record": false}}, "name": "zone_record.update", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "8f6cd405-2c87-453b-8b95-7a296982e4b8"}`

	event := &ZoneRecordEvent{}
	err := ParseZoneRecordEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "zone_record.update", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "_frame", event.ZoneRecord.Name; want != got {
		t.Errorf("ParseEvent ZoneRecord.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ZoneRecordEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}

func TestParseZoneRecordEvent_ZoneRecord_Delete(t *testing.T) {
	payload := `{"data": {"zone_record": {"id": 1, "ttl": 60, "name": "_frame", "type": "TXT", "content": "https://dnsimple.com/", "zone_id": "example.com", "priority": null, "parent_id": null, "created_at": "2016-02-22T21:06:48.957Z", "updated_at": "2016-02-22T21:23:22.503Z", "system_record": false}}, "name": "zone_record.delete", "actor": {"id": "1", "entity": "user", "pretty": "example@example.com"}, "account": {"id": 1010, "display": "User", "identifier": "user"}, "api_version": "v2", "request_identifier": "8f6cd405-2c87-453b-8b95-7a296982e4b8"}`

	event := &ZoneRecordEvent{}
	err := ParseZoneRecordEvent(event, []byte(payload))
	if err != nil {
		t.Fatalf("ParseEvent returned error: %v", err)
	}

	if want, got := "zone_record.delete", event.Name; want != got {
		t.Errorf("ParseEvent name expected to be %v, got %v", want, got)
	}
	if !regexpUUID.MatchString(event.RequestID) {
		t.Errorf("ParseEvent requestID expected to be an UUID, got %v", event.RequestID)
	}
	if want, got := "_frame", event.ZoneRecord.Name; want != got {
		t.Errorf("ParseEvent ZoneRecord.Name expected to be %v, got %v", want, got)
	}

	parsedEvent, err := Parse([]byte(payload))
	_, ok := parsedEvent.(*ZoneRecordEvent)
	if !ok {
		t.Fatalf("Parse returned error when typecasting: %v", err)
	}
}
