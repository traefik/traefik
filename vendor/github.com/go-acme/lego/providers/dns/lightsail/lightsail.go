// Package lightsail implements a DNS provider for solving the DNS-01 challenge using AWS Lightsail DNS.
package lightsail

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const (
	maxRetries = 5
)

// customRetryer implements the client.Retryer interface by composing the DefaultRetryer.
// It controls the logic for retrying recoverable request errors (e.g. when rate limits are exceeded).
type customRetryer struct {
	client.DefaultRetryer
}

// RetryRules overwrites the DefaultRetryer's method.
// It uses a basic exponential backoff algorithm that returns an initial
// delay of ~400ms with an upper limit of ~30 seconds which should prevent
// causing a high number of consecutive throttling errors.
// For reference: Route 53 enforces an account-wide(!) 5req/s query limit.
func (c customRetryer) RetryRules(r *request.Request) time.Duration {
	retryCount := r.RetryCount
	if retryCount > 7 {
		retryCount = 7
	}

	delay := (1 << uint(retryCount)) * (rand.Intn(50) + 200)
	return time.Duration(delay) * time.Millisecond
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	DNSZone            string
	Region             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		DNSZone:            env.GetOrFile("DNS_ZONE"),
		PropagationTimeout: env.GetOrDefaultSecond("LIGHTSAIL_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("LIGHTSAIL_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		Region:             env.GetOrDefaultString("LIGHTSAIL_REGION", "us-east-1"),
	}
}

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	client *lightsail.Lightsail
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for the AWS Lightsail service.
//
// AWS Credentials are automatically detected in the following locations
// and prioritized in the following order:
// 1. Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY,
//     [AWS_SESSION_TOKEN], [DNS_ZONE], [LIGHTSAIL_REGION]
// 2. Shared credentials file (defaults to ~/.aws/credentials)
// 3. Amazon EC2 IAM role
//
// public hosted zone via the FQDN.
//
// See also: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(NewDefaultConfig())
}

// NewDNSProviderConfig return a DNSProvider instance configured for AWS Lightsail.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("lightsail: the configuration of the DNS provider is nil")
	}

	retryer := customRetryer{}
	retryer.NumMaxRetries = maxRetries

	conf := aws.NewConfig().WithRegion(config.Region)
	sess, err := session.NewSession(request.WithRetryer(conf, retryer))
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		config: config,
		client: lightsail.New(sess),
	}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	err := d.newTxtRecord(fqdn, `"`+value+`"`)
	if err != nil {
		return fmt.Errorf("lightsail: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	params := &lightsail.DeleteDomainEntryInput{
		DomainName: aws.String(d.config.DNSZone),
		DomainEntry: &lightsail.DomainEntry{
			Name:   aws.String(fqdn),
			Type:   aws.String("TXT"),
			Target: aws.String(`"` + value + `"`),
		},
	}

	_, err := d.client.DeleteDomainEntry(params)
	if err != nil {
		return fmt.Errorf("lightsail: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) newTxtRecord(fqdn string, value string) error {
	params := &lightsail.CreateDomainEntryInput{
		DomainName: aws.String(d.config.DNSZone),
		DomainEntry: &lightsail.DomainEntry{
			Name:   aws.String(fqdn),
			Target: aws.String(value),
			Type:   aws.String("TXT"),
		},
	}
	_, err := d.client.CreateDomainEntry(params)
	return err
}
