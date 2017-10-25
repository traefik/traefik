// Package route53 implements a DNS provider for solving the DNS-01 challenge
// using AWS Route 53 DNS.
package route53

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/xenolf/lego/acme"
)

const (
	maxRetries = 5
	route53TTL = 10
)

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	client       *route53.Route53
	hostedZoneID string
}

// customRetryer implements the client.Retryer interface by composing the
// DefaultRetryer. It controls the logic for retrying recoverable request
// errors (e.g. when rate limits are exceeded).
type customRetryer struct {
	client.DefaultRetryer
}

// RetryRules overwrites the DefaultRetryer's method.
// It uses a basic exponential backoff algorithm that returns an initial
// delay of ~400ms with an upper limit of ~30 seconds which should prevent
// causing a high number of consecutive throttling errors.
// For reference: Route 53 enforces an account-wide(!) 5req/s query limit.
func (d customRetryer) RetryRules(r *request.Request) time.Duration {
	retryCount := r.RetryCount
	if retryCount > 7 {
		retryCount = 7
	}

	delay := (1 << uint(retryCount)) * (rand.Intn(50) + 200)
	return time.Duration(delay) * time.Millisecond
}

// NewDNSProvider returns a DNSProvider instance configured for the AWS
// Route 53 service.
//
// AWS Credentials are automatically detected in the following locations
// and prioritized in the following order:
// 1. Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY,
//    AWS_REGION, [AWS_SESSION_TOKEN]
// 2. Shared credentials file (defaults to ~/.aws/credentials)
// 3. Amazon EC2 IAM role
//
// If AWS_HOSTED_ZONE_ID is not set, Lego tries to determine the correct
// public hosted zone via the FQDN.
//
// See also: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk
func NewDNSProvider() (*DNSProvider, error) {
	hostedZoneID := os.Getenv("AWS_HOSTED_ZONE_ID")

	r := customRetryer{}
	r.NumMaxRetries = maxRetries
	config := request.WithRetryer(aws.NewConfig(), r)
	client := route53.New(session.New(config))

	return &DNSProvider{
		client:       client,
		hostedZoneID: hostedZoneID,
	}, nil
}

// Present creates a TXT record using the specified parameters
func (r *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	value = `"` + value + `"`
	return r.changeRecord("UPSERT", fqdn, value, route53TTL)
}

// CleanUp removes the TXT record matching the specified parameters
func (r *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	value = `"` + value + `"`
	return r.changeRecord("DELETE", fqdn, value, route53TTL)
}

func (r *DNSProvider) changeRecord(action, fqdn, value string, ttl int) error {
	hostedZoneID, err := r.getHostedZoneID(fqdn)
	if err != nil {
		return fmt.Errorf("Failed to determine Route 53 hosted zone ID: %v", err)
	}

	recordSet := newTXTRecordSet(fqdn, value, ttl)
	reqParams := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		ChangeBatch: &route53.ChangeBatch{
			Comment: aws.String("Managed by Lego"),
			Changes: []*route53.Change{
				{
					Action:            aws.String(action),
					ResourceRecordSet: recordSet,
				},
			},
		},
	}

	resp, err := r.client.ChangeResourceRecordSets(reqParams)
	if err != nil {
		return fmt.Errorf("Failed to change Route 53 record set: %v", err)
	}

	statusID := resp.ChangeInfo.Id

	return acme.WaitFor(120*time.Second, 4*time.Second, func() (bool, error) {
		reqParams := &route53.GetChangeInput{
			Id: statusID,
		}
		resp, err := r.client.GetChange(reqParams)
		if err != nil {
			return false, fmt.Errorf("Failed to query Route 53 change status: %v", err)
		}
		if *resp.ChangeInfo.Status == route53.ChangeStatusInsync {
			return true, nil
		}
		return false, nil
	})
}

func (r *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	if r.hostedZoneID != "" {
		return r.hostedZoneID, nil
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	// .DNSName should not have a trailing dot
	reqParams := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(acme.UnFqdn(authZone)),
	}
	resp, err := r.client.ListHostedZonesByName(reqParams)
	if err != nil {
		return "", err
	}

	var hostedZoneID string
	for _, hostedZone := range resp.HostedZones {
		// .Name has a trailing dot
		if !*hostedZone.Config.PrivateZone && *hostedZone.Name == authZone {
			hostedZoneID = *hostedZone.Id
			break
		}
	}

	if len(hostedZoneID) == 0 {
		return "", fmt.Errorf("Zone %s not found in Route 53 for domain %s", authZone, fqdn)
	}

	if strings.HasPrefix(hostedZoneID, "/hostedzone/") {
		hostedZoneID = strings.TrimPrefix(hostedZoneID, "/hostedzone/")
	}

	return hostedZoneID, nil
}

func newTXTRecordSet(fqdn, value string, ttl int) *route53.ResourceRecordSet {
	return &route53.ResourceRecordSet{
		Name: aws.String(fqdn),
		Type: aws.String("TXT"),
		TTL:  aws.Int64(int64(ttl)),
		ResourceRecords: []*route53.ResourceRecord{
			{Value: aws.String(value)},
		},
	}
}
