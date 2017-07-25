package webhook

import (
	"github.com/dnsimple/dnsimple-go/dnsimple"
)

func switchEvent(name string, payload []byte) (Event, error) {
	var event Event

	switch name {
	case // account
		"account.update",                  // TODO
		"account.billing_settings_update", // TODO
		"account.payment_details_update",  // TODO
		"account.add_user",                // TODO
		"account.remove_user":             // TODO
		event = &AccountEvent{}
	//case // certificate
	//	"certificate.issue",
	//	"certificate.reissue",
	//	"certificate.remove_private_key":
	//	event = &CertificateEvent{}
	case // contact
		"contact.create",
		"contact.update",
		"contact.delete":
		event = &ContactEvent{}
	//case // dnssec
	//	"dnssec.create",
	//	"dnssec.delete":
	//	event = &DNSSEC{}
	case // domain
		"domain.auto_renewal_enable",
		"domain.auto_renewal_disable",
		"domain.create",
		"domain.delete",
		"domain.register",
		"domain.renew",
		"domain.delegation_change",
		"domain.registrant_change",
		"domain.resolution_disable",
		"domain.resolution_enable",
		"domain.token_reset",
		"domain.transfer":
		event = &DomainEvent{}
	case // email forward
		"email_forward.create",
		"email_forward.delete":
		event = &EmailForwardEvent{}
	//case // name servers
	//	"name_server.deregister",
	//	"name_server.register":
	//	event = &NameServerEvent{}
	//case // push
	//	"push.accept",
	//	"push.initiate",
	//	"push.reject":
	//	event = &PushEvent{}
	//case // secondary dns
	//	"secondary_dns.create",
	//	"secondary_dns.update",
	//	"secondary_dns.delete":
	//	event = &SecondaryDNSEvent{}
	//case // subscription
	//	"subscription.migrate",
	//	"subscription.subscribe",
	//	"subscription.unsubscribe":
	//	event = &SubscriptionEvent{}
	//case // template
	//	"template.create",
	//	"template.delete",
	//	"template.update":
	//	event = &TemplateEvent{}
	//case // template record
	//	"template_record.create",
	//	"template_record.delete":
	//	event = &TemplateRecordEvent{}
	//case // vanity
	//	"vanity.disable",
	//	"vanity.enable":
	//	event = &VanityEvent{}
	case // webhook
		"webhook.create",
		"webhook.delete":
		event = &WebhookEvent{}
	case // whois privacy
		"whois_privacy.disable",
		"whois_privacy.enable",
		"whois_privacy.purchase",
		"whois_privacy.renew":
		event = &WhoisPrivacyEvent{}
	case // zone
		"zone.create",
		"zone.delete":
		event = &ZoneEvent{}
	case // zone record
		"zone_record.create",
		"zone_record.update",
		"zone_record.delete":
		event = &ZoneRecordEvent{}
	default:
		event = &GenericEvent{}
	}

	return event, event.parse(payload)
}

//
// GenericEvent
//

// GenericEvent represents a generic event, where the data is a simple map of strings.
type GenericEvent struct {
	Event_Header
	Data interface{} `json:"data"`
}

func (e *GenericEvent) parse(payload []byte) error {
	e.payload = payload
	return unmashalEvent(payload, e)
}

// ParseGenericEvent unpacks the data into a GenericEvent.
func ParseGenericEvent(e *GenericEvent, payload []byte) error {
	return e.parse(payload)
}

//
// AccountEvent
//

// AccountEvent represents the base event sent for an account action.
type AccountEvent struct {
	Event_Header
	Data    *AccountEvent     `json:"data"`
	Account *dnsimple.Account `json:"account"`
}

// ParseAccountEvent unpacks the data into an AccountEvent.
func ParseAccountEvent(e *AccountEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *AccountEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}

//
// ContactEvent
//

// ContactEvent represents the base event sent for a contact action.
type ContactEvent struct {
	Event_Header
	Data    *ContactEvent     `json:"data"`
	Contact *dnsimple.Contact `json:"contact"`
}

// ParseContactEvent unpacks the data into a ContactEvent.
func ParseContactEvent(e *ContactEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *ContactEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}

//
// DomainEvent
//

// DomainEvent represents the base event sent for a domain action.
type DomainEvent struct {
	Event_Header
	Data       *DomainEvent         `json:"data"`
	Domain     *dnsimple.Domain     `json:"domain"`
	Registrant *dnsimple.Contact    `json:"registrant"`
	Delegation *dnsimple.Delegation `json:"name_servers"`
}

// ParseDomainEvent unpacks the payload into a DomainEvent.
func ParseDomainEvent(e *DomainEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *DomainEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}

//
// EmailForwardEvent
//

// EmailForwardEvent represents the base event sent for an email forward action.
type EmailForwardEvent struct {
	Event_Header
	Data         *EmailForwardEvent     `json:"data"`
	EmailForward *dnsimple.EmailForward `json:"email_forward"`
}

// ParseDomainEvent unpacks the payload into a EmailForwardEvent.
func ParseEmailForwardEvent(e *EmailForwardEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *EmailForwardEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}

//
// WebhookEvent
//

// WebhookEvent represents the base event sent for a webhook action.
type WebhookEvent struct {
	Event_Header
	Data    *WebhookEvent     `json:"data"`
	Webhook *dnsimple.Webhook `json:"webhook"`
}

// ParseWebhookEvent unpacks the data into a WebhookEvent.
func ParseWebhookEvent(e *WebhookEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *WebhookEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}

//
// WhoisPrivacyEvent
//

// WhoisPrivacyEvent represents the base event sent for a whois privacy action.
type WhoisPrivacyEvent struct {
	Event_Header
	Data         *WhoisPrivacyEvent     `json:"data"`
	Domain       *dnsimple.Domain       `json:"domain"`
	WhoisPrivacy *dnsimple.WhoisPrivacy `json:"whois_privacy"`
}

// ParseWhoisPrivacyEvent unpacks the data into a WhoisPrivacyEvent.
func ParseWhoisPrivacyEvent(e *WhoisPrivacyEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *WhoisPrivacyEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}

//
// ZoneEvent
//

// ZoneEvent represents the base event sent for a zone action.
type ZoneEvent struct {
	Event_Header
	Data *ZoneEvent     `json:"data"`
	Zone *dnsimple.Zone `json:"zone"`
}

// ParseZoneEvent unpacks the data into a ZoneEvent.
func ParseZoneEvent(e *ZoneEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *ZoneEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}

//
// ZoneRecordEvent
//

// ZoneRecordEvent represents the base event sent for a zone record action.
type ZoneRecordEvent struct {
	Event_Header
	Data       *ZoneRecordEvent     `json:"data"`
	ZoneRecord *dnsimple.ZoneRecord `json:"zone_record"`
}

// ParseZoneRecordEvent unpacks the data into a ZoneRecordEvent.
func ParseZoneRecordEvent(e *ZoneRecordEvent, payload []byte) error {
	return e.parse(payload)
}

func (e *ZoneRecordEvent) parse(payload []byte) error {
	e.payload, e.Data = payload, e
	return unmashalEvent(payload, e)
}
