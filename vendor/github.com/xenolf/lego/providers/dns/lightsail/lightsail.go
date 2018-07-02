// Package lightsail implements a DNS provider for solving the DNS-01 challenge
// using AWS Lightsail DNS.
package lightsail

import (
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/xenolf/lego/acme"
)

const (
	maxRetries = 5
)

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	client *lightsail.Lightsail
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
// Lightsail service.
//
// AWS Credentials are automatically detected in the following locations
// and prioritized in the following order:
// 1. Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY,
//     [AWS_SESSION_TOKEN], [DNS_ZONE]
// 2. Shared credentials file (defaults to ~/.aws/credentials)
// 3. Amazon EC2 IAM role
//
// public hosted zone via the FQDN.
//
// See also: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk
func NewDNSProvider() (*DNSProvider, error) {
	r := customRetryer{}
	r.NumMaxRetries = maxRetries
	config := request.WithRetryer(aws.NewConfig(), r)
	client := lightsail.New(session.New(config))

	return &DNSProvider{
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters
func (r *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	value = `"` + value + `"`
	err := r.newTxtRecord(domain, fqdn, value)
	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (r *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	value = `"` + value + `"`
	params := &lightsail.DeleteDomainEntryInput{
		DomainName: aws.String(domain),
		DomainEntry: &lightsail.DomainEntry{
			Name:   aws.String(fqdn),
			Type:   aws.String("TXT"),
			Target: aws.String(value),
		},
	}
	_, err := r.client.DeleteDomainEntry(params)
	return err
}

func (r *DNSProvider) newTxtRecord(domain string, fqdn string, value string) error {
	params := &lightsail.CreateDomainEntryInput{
		DomainName: aws.String(domain),
		DomainEntry: &lightsail.DomainEntry{
			Name:   aws.String(fqdn),
			Target: aws.String(value),
			Type:   aws.String("TXT"),
		},
	}
	_, err := r.client.CreateDomainEntry(params)
	return err
}
