package route53

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func TestRoute53TTL(t *testing.T) {

	m, err := testGetAndPreCheck()
	if err != nil {
		t.Skip(err.Error())
	}

	provider, err := NewDNSProvider()
	if err != nil {
		t.Fatalf("Fatal: %s", err.Error())
	}

	err = provider.Present(m["route53Domain"], "foo", "bar")
	if err != nil {
		t.Fatalf("Fatal: %s", err.Error())
	}
	// we need a separate R53 client here as the one in the DNS provider is
	// unexported.
	fqdn := "_acme-challenge." + m["route53Domain"] + "."
	svc := route53.New(session.New())
	zoneID, err := getHostedZoneID(fqdn, svc)
	if err != nil {
		provider.CleanUp(m["route53Domain"], "foo", "bar")
		t.Fatalf("Fatal: %s", err.Error())
	}
	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ListResourceRecordSets(params)
	if err != nil {
		provider.CleanUp(m["route53Domain"], "foo", "bar")
		t.Fatalf("Fatal: %s", err.Error())
	}

	for _, v := range resp.ResourceRecordSets {
		if *v.Name == fqdn && *v.Type == "TXT" && *v.TTL == 10 {
			provider.CleanUp(m["route53Domain"], "foo", "bar")
			return
		}
	}
	provider.CleanUp(m["route53Domain"], "foo", "bar")
	t.Fatalf("Could not find a TXT record for _acme-challenge.%s with a TTL of 10", m["route53Domain"])
}

func testGetAndPreCheck() (map[string]string, error) {
	m := map[string]string{
		"route53Key":    os.Getenv("AWS_ACCESS_KEY_ID"),
		"route53Secret": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"route53Region": os.Getenv("AWS_REGION"),
		"route53Domain": os.Getenv("R53_DOMAIN"),
	}
	for _, v := range m {
		if v == "" {
			return nil, fmt.Errorf("AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, and R53_DOMAIN are needed to run this test")
		}
	}
	return m, nil
}
