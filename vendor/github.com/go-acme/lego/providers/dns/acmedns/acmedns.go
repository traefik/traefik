// Package acmedns implements a DNS provider for solving DNS-01 challenges using Joohoi's acme-dns project.
// For more information see the ACME-DNS homepage: https://github.com/joohoi/acme-dns
package acmedns

import (
	"errors"
	"fmt"

	"github.com/cpu/goacmedns"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const (
	// envNamespace is the prefix for ACME-DNS environment variables.
	envNamespace = "ACME_DNS_"
	// apiBaseEnvVar is the environment variable name for the ACME-DNS API address
	// (e.g. https://acmedns.your-domain.com).
	apiBaseEnvVar = envNamespace + "API_BASE"
	// storagePathEnvVar is the environment variable name for the ACME-DNS JSON account data file.
	// A per-domain account will be registered/persisted to this file and used for TXT updates.
	storagePathEnvVar = envNamespace + "STORAGE_PATH"
)

// acmeDNSClient is an interface describing the goacmedns.Client functions the DNSProvider uses.
// It makes it easier for tests to shim a mock Client into the DNSProvider.
type acmeDNSClient interface {
	// UpdateTXTRecord updates the provided account's TXT record
	// to the given value or returns an error.
	UpdateTXTRecord(goacmedns.Account, string) error
	// RegisterAccount registers and returns a new account
	// with the given allowFrom restriction or returns an error.
	RegisterAccount([]string) (goacmedns.Account, error)
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface for
// an ACME-DNS server.
type DNSProvider struct {
	client  acmeDNSClient
	storage goacmedns.Storage
}

// NewDNSProvider creates an ACME-DNS provider using file based account storage.
// Its configuration is loaded from the environment by reading apiBaseEnvVar and storagePathEnvVar.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(apiBaseEnvVar, storagePathEnvVar)
	if err != nil {
		return nil, fmt.Errorf("acme-dns: %v", err)
	}

	client := goacmedns.NewClient(values[apiBaseEnvVar])
	storage := goacmedns.NewFileStorage(values[storagePathEnvVar], 0600)
	return NewDNSProviderClient(client, storage)
}

// NewDNSProviderClient creates an ACME-DNS DNSProvider with the given acmeDNSClient and goacmedns.Storage.
func NewDNSProviderClient(client acmeDNSClient, storage goacmedns.Storage) (*DNSProvider, error) {
	if client == nil {
		return nil, errors.New("ACME-DNS Client must be not nil")
	}

	if storage == nil {
		return nil, errors.New("ACME-DNS Storage must be not nil")
	}

	return &DNSProvider{
		client:  client,
		storage: storage,
	}, nil
}

// ErrCNAMERequired is returned by Present when the Domain indicated had no
// existing ACME-DNS account in the Storage and additional setup is required.
// The user must create a CNAME in the DNS zone for Domain that aliases FQDN
// to Target in order to complete setup for the ACME-DNS account that was created.
type ErrCNAMERequired struct {
	// The Domain that is being issued for.
	Domain string
	// The alias of the CNAME (left hand DNS label).
	FQDN string
	// The RDATA of the CNAME (right hand side, canonical name).
	Target string
}

// Error returns a descriptive message for the ErrCNAMERequired instance telling
// the user that a CNAME needs to be added to the DNS zone of c.Domain before
// the ACME-DNS hook will work. The CNAME to be created should be of the form:
//   {{ c.FQDN }} 	CNAME	{{ c.Target }}
func (e ErrCNAMERequired) Error() string {
	return fmt.Sprintf("acme-dns: new account created for %q. "+
		"To complete setup for %q you must provision the following "+
		"CNAME in your DNS zone and re-run this provider when it is "+
		"in place:\n"+
		"%s CNAME %s.",
		e.Domain, e.Domain, e.FQDN, e.Target)
}

// Present creates a TXT record to fulfill the DNS-01 challenge.
// If there is an existing account for the domain in the provider's storage
// then it will be used to set the challenge response TXT record with the ACME-DNS server and issuance will continue.
// If there is not an account for the given domain present in the DNSProvider storage
// one will be created and registered with the ACME DNS server and an ErrCNAMERequired error is returned.
// This will halt issuance and indicate to the user that a one-time manual setup is required for the domain.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	// Compute the challenge response FQDN and TXT value for the domain based
	// on the keyAuth.
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// Check if credentials were previously saved for this domain.
	account, err := d.storage.Fetch(domain)
	// Errors other than goacmeDNS.ErrDomainNotFound are unexpected.
	if err != nil && err != goacmedns.ErrDomainNotFound {
		return err
	}
	if err == goacmedns.ErrDomainNotFound {
		// The account did not exist. Create a new one and return an error
		// indicating the required one-time manual CNAME setup.
		return d.register(domain, fqdn)
	}

	// Update the acme-dns TXT record.
	return d.client.UpdateTXTRecord(account, value)
}

// CleanUp removes the record matching the specified parameters. It is not
// implemented for the ACME-DNS provider.
func (d *DNSProvider) CleanUp(_, _, _ string) error {
	// ACME-DNS doesn't support the notion of removing a record.
	// For users of ACME-DNS it is expected the stale records remain in-place.
	return nil
}

// register creates a new ACME-DNS account for the given domain.
// If account creation works as expected a ErrCNAMERequired error is returned describing
// the one-time manual CNAME setup required to complete setup of the ACME-DNS hook for the domain.
// If any other error occurs it is returned as-is.
func (d *DNSProvider) register(domain, fqdn string) error {
	// TODO(@cpu): Read CIDR whitelists from the environment
	newAcct, err := d.client.RegisterAccount(nil)
	if err != nil {
		return err
	}

	// Store the new account in the storage and call save to persist the data.
	err = d.storage.Put(domain, newAcct)
	if err != nil {
		return err
	}
	err = d.storage.Save()
	if err != nil {
		return err
	}

	// Stop issuance by returning an error.
	// The user needs to perform a manual one-time CNAME setup in their DNS zone
	// to complete the setup of the new account we created.
	return ErrCNAMERequired{
		Domain: domain,
		FQDN:   fqdn,
		Target: newAcct.FullDomain,
	}
}
