package dnsimple

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	dnsimpleLiveTest bool
	dnsimpleToken    string
	dnsimpleBaseURL  string
	dnsimpleClient   *Client
)

func init() {
	dnsimpleToken = os.Getenv("DNSIMPLE_TOKEN")
	dnsimpleBaseURL = os.Getenv("DNSIMPLE_BASE_URL")

	// Prevent peoeple from wiping out their entire production account by mistake
	if dnsimpleBaseURL == "" {
		dnsimpleBaseURL = "https://api.sandbox.dnsimple.com"
	}

	if len(dnsimpleToken) > 0 {
		dnsimpleLiveTest = true
		dnsimpleClient = NewClient(NewOauthTokenCredentials(dnsimpleToken))
		dnsimpleClient.BaseURL = dnsimpleBaseURL
		dnsimpleClient.UserAgent = fmt.Sprintf("%v +livetest", dnsimpleClient.UserAgent)
	}
}

func TestLive_Whoami(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	whoamiResponse, err := dnsimpleClient.Identity.Whoami()
	if err != nil {
		t.Fatalf("Live Auth.Whoami() returned error: %v", err)
	}

	fmt.Printf("RateLimit: %v/%v until %v\n", whoamiResponse.RateLimitRemaining(), whoamiResponse.RateLimit(), whoamiResponse.RateLimitReset())
	whoami := whoamiResponse.Data
	fmt.Printf("Account: %+v\n", whoami.Account)
	fmt.Printf("User: %+v\n", whoami.User)
}

func TestLive_Domains(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	whoami, err := Whoami(dnsimpleClient)
	if err != nil {
		t.Fatalf("Live Whoami() returned error: %v", err)
	}

	accountID := whoami.Account.ID

	domainsResponse, err := dnsimpleClient.Domains.ListDomains(fmt.Sprintf("%v", accountID), nil)
	if err != nil {
		t.Fatalf("Live Domains.List() returned error: %v", err)
	}

	fmt.Printf("RateLimit: %v/%v until %v\n", domainsResponse.RateLimitRemaining(), domainsResponse.RateLimit(), domainsResponse.RateLimitReset())
	fmt.Printf("Domains: %+v\n", domainsResponse.Data)
}

func TestLive_Registration(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	whoami, err := Whoami(dnsimpleClient)
	if err != nil {
		t.Fatalf("Live Whoami() returned error: %v", err)
	}

	accountID := whoami.Account.ID

	// TODO: fetch the registrant randomly
	registerRequest := &DomainRegisterRequest{RegistrantID: 2}
	registrationResponse, err := dnsimpleClient.Registrar.RegisterDomain(fmt.Sprintf("%v", accountID), fmt.Sprintf("example-%v.com", time.Now().Unix()), registerRequest)
	if err != nil {
		t.Fatalf("Live Registrar.Register() returned error: %v", err)
	}

	fmt.Printf("RateLimit: %v/%v until %v\n", registrationResponse.RateLimitRemaining(), registrationResponse.RateLimit(), registrationResponse.RateLimitReset())
	fmt.Printf("Domain: %+v\n", registrationResponse.Data)
}

func TestLive_Webhooks(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	var err error
	var webhook *Webhook
	var webhookResponse *webhookResponse
	var webhooksResponse *webhooksResponse

	whoami, err := Whoami(dnsimpleClient)
	if err != nil {
		t.Fatalf("Live Auth.Whoami()/Domains.List() returned error: %v", err)
	}
	accountID := whoami.Account.ID

	webhooksResponse, err = dnsimpleClient.Webhooks.ListWebhooks(fmt.Sprintf("%v", accountID), nil)
	if err != nil {
		t.Fatalf("Live Webhooks.List() returned error: %v", err)
	}

	fmt.Printf("RateLimit: %v/%v until %v\n", webhooksResponse.RateLimitRemaining(), webhooksResponse.RateLimit(), webhooksResponse.RateLimitReset())
	fmt.Printf("Webhooks: %+v\n", webhooksResponse.Data)

	webhookAttributes := Webhook{URL: "https://livetest.test"}
	webhookResponse, err = dnsimpleClient.Webhooks.CreateWebhook(fmt.Sprintf("%v", accountID), webhookAttributes)
	if err != nil {
		t.Fatalf("Live Webhooks.Create() returned error: %v", err)
	}

	fmt.Printf("RateLimit: %v/%v until %v\n", webhooksResponse.RateLimitRemaining(), webhooksResponse.RateLimit(), webhooksResponse.RateLimitReset())
	fmt.Printf("Webhook: %+v\n", webhookResponse.Data)
	webhook = webhookResponse.Data

	webhookResponse, err = dnsimpleClient.Webhooks.DeleteWebhook(fmt.Sprintf("%v", accountID), webhook.ID)
	if err != nil {
		t.Fatalf("Live Webhooks.Delete(%v) returned error: %v", webhook.ID, err)
	}

	fmt.Printf("RateLimit: %v/%v until %v\n", webhooksResponse.RateLimitRemaining(), webhooksResponse.RateLimit(), webhooksResponse.RateLimitReset())
	webhook = webhookResponse.Data
}

func TestLive_Zones(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	whoami, err := Whoami(dnsimpleClient)
	if err != nil {
		t.Fatalf("Live Zones/Whoami() returned error: %v", err)
	}

	accountID := strconv.Itoa(whoami.Account.ID)

	domainResponse, err := dnsimpleClient.Domains.CreateDomain(fmt.Sprintf("%v", accountID), Domain{Name: fmt.Sprintf("example-%v.test", time.Now().Unix())})
	if err != nil {
		t.Fatalf("Live Zones/CreateZone() returned error: %v", err)
	}

	zoneName := domainResponse.Data.Name
	recordResponse, err := dnsimpleClient.Zones.CreateRecord(accountID, zoneName, ZoneRecord{Name: fmt.Sprintf("%v", time.Now().Unix()), Type: "TXT", Content: "Test"})
	if err != nil {
		t.Fatalf("Live Zones/CreateRecord() returned error: %v", err)
	}

	fmt.Printf("ZoneRecord: %+v\n", recordResponse.Data)
}

func TestLive_Error(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	whoami, err := Whoami(dnsimpleClient)
	if err != nil {
		t.Fatalf("Live Error/Whoami() returned error: %v", err)
	}

	_, err = dnsimpleClient.Registrar.RegisterDomain(fmt.Sprintf("%v", whoami.Account.ID), fmt.Sprintf("example-%v.test", time.Now().Unix()), &DomainRegisterRequest{})
	if err == nil {
		t.Fatalf("Live Error/RegisterDomain() expected to return error")
	}

	e := err.(*ErrorResponse)
	fmt.Println(e.Message)
}
