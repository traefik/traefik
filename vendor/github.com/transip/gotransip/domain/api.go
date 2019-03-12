package domain

import (
	"github.com/transip/gotransip"
)

// This file holds all DomainService methods directly ported from TransIP API

// BatchCheckAvailability checks the availability of multiple domains
func BatchCheckAvailability(c gotransip.Client, domainNames []string) ([]CheckResult, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "batchCheckAvailability",
	}
	sr.AddArgument("domainNames", domainNames)

	var v struct {
		V []CheckResult `xml:"item"`
	}

	err := c.Call(sr, &v)
	return v.V, err
}

// CheckAvailability returns the availability status of a domain.
func CheckAvailability(c gotransip.Client, domainName string) (Status, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "checkAvailability",
	}
	sr.AddArgument("domainName", domainName)

	var v Status
	err := c.Call(sr, &v)
	return v, err
}

// GetWhois returns the whois of a domain name
func GetWhois(c gotransip.Client, domainName string) (string, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getWhois",
	}
	sr.AddArgument("domainName", domainName)

	var v string
	err := c.Call(sr, &v)
	return v, err
}

// GetDomainNames returns list with domain names or error when this failed
func GetDomainNames(c gotransip.Client) ([]string, error) {
	var d = struct {
		D []string `xml:"item"`
	}{}
	err := c.Call(gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getDomainNames",
	}, &d)

	return d.D, err
}

// GetInfo returns Domain for given name or error when this failed
func GetInfo(c gotransip.Client, domainName string) (Domain, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getInfo",
	}
	sr.AddArgument("domainName", domainName)

	var d Domain
	err := c.Call(sr, &d)

	return d, err
}

// BatchGetInfo returns array of Domain for given name or error when this failed
func BatchGetInfo(c gotransip.Client, domainNames []string) ([]Domain, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "batchGetInfo",
	}
	sr.AddArgument("domainNames", domainNames)

	var d = struct {
		D []Domain `xml:"item"`
	}{}
	err := c.Call(sr, &d)

	return d.D, err
}

// GetAuthCode returns the Auth code for a domainName
func GetAuthCode(c gotransip.Client, domainName string) (string, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getAuthCode",
	}
	sr.AddArgument("domainName", domainName)

	var v string
	err := c.Call(sr, &v)
	return v, err
}

// GetIsLocked returns the lock status for a domainName
func GetIsLocked(c gotransip.Client, domainName string) (bool, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getIsLocked",
	}
	sr.AddArgument("domainName", domainName)

	var v bool
	err := c.Call(sr, &v)
	return v, err
}

// Register registers a domain name and will automatically create and sign a proposition for it
func Register(c gotransip.Client, domain Domain) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "register",
	}
	sr.AddArgument("domain", domain)

	return c.Call(sr, nil)
}

// Cancel cancels a domain name, will automatically create and sign a cancellation document
func Cancel(c gotransip.Client, domainName string, endTime gotransip.CancellationTime) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "cancel",
	}
	sr.AddArgument("domainName", domainName)
	sr.AddArgument("endTime", string(endTime))

	return c.Call(sr, nil)
}

// TransferWithOwnerChange transfers a domain with changing the owner
func TransferWithOwnerChange(c gotransip.Client, domain, authCode string) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "transferWithOwnerChange",
	}
	sr.AddArgument("domain", domain)
	sr.AddArgument("authCode", authCode)

	return c.Call(sr, nil)
}

// TransferWithoutOwnerChange transfers a domain without changing the owner
func TransferWithoutOwnerChange(c gotransip.Client, domain, authCode string) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "transferWithoutOwnerChange",
	}
	sr.AddArgument("domain", domain)
	sr.AddArgument("authCode", authCode)

	return c.Call(sr, nil)
}

// SetNameservers starts a nameserver change for this domain, will replace all
// existing nameservers with the new nameservers
func SetNameservers(c gotransip.Client, domainName string, nameservers Nameservers) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "setNameservers",
	}
	sr.AddArgument("domainName", domainName)
	sr.AddArgument("nameservers", nameservers)

	return c.Call(sr, nil)
}

// SetLock locks this domain
func SetLock(c gotransip.Client, domainName string) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "setLock",
	}
	sr.AddArgument("domainName", domainName)

	return c.Call(sr, nil)
}

// UnsetLock unlocks this domain
func UnsetLock(c gotransip.Client, domainName string) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "unsetLock",
	}
	sr.AddArgument("domainName", domainName)

	return c.Call(sr, nil)
}

// SetDNSEntries sets the DnsEntries for this Domain, will replace all existing
// dns entries with the new entries
func SetDNSEntries(c gotransip.Client, domainName string, dnsEntries DNSEntries) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "setDnsEntries",
	}
	sr.AddArgument("domainName", domainName)
	sr.AddArgument("dnsEntries", dnsEntries)

	return c.Call(sr, nil)
}

// SetOwner starts an owner change of a domain
func SetOwner(c gotransip.Client, domainName, registrantWhoisContact WhoisContact) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "setOwner",
	}
	sr.AddArgument("domainName", domainName)
	// make sure contact is of type registrant
	registrantWhoisContact.Type = "registrant"
	sr.AddArgument("registrantWhoisContact", registrantWhoisContact)

	return c.Call(sr, nil)
}

// SetContacts starts a contact change of a domain, this will replace all existing contacts
func SetContacts(c gotransip.Client, domainName, contacts WhoisContacts) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "setContacts",
	}
	sr.AddArgument("domainName", domainName)
	sr.AddArgument("contacts", contacts)

	return c.Call(sr, nil)
}

// GetAllTLDInfos returns slice with TLD objects or error when this failed
func GetAllTLDInfos(c gotransip.Client) ([]TLD, error) {
	var d = struct {
		TLD []TLD `xml:"item"`
	}{}
	err := c.Call(gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getAllTldInfos",
	}, &d)

	return d.TLD, err
}

// GetTldInfo returns info about a specific TLD
func GetTldInfo(c gotransip.Client, tldName string) (TLD, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getTldInfo",
	}
	sr.AddArgument("tldName", tldName)

	var v TLD
	err := c.Call(sr, &v)
	return v, err
}

// GetCurrentDomainAction returns info about the action this domain is currently running
func GetCurrentDomainAction(c gotransip.Client, domainName string) (ActionResult, error) {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "getCurrentDomainAction",
	}
	sr.AddArgument("domainName", domainName)

	var v ActionResult
	err := c.Call(sr, &v)
	return v, err
}

// RetryCurrentDomainActionWithNewData retries a failed domain action with new
// domain data. The Domain.Name field must contain the name of the Domain. The
// Nameservers, Contacts, DNSEntries fields contain the new data for this domain.
func RetryCurrentDomainActionWithNewData(c gotransip.Client, domain Domain) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "retryCurrentDomainActionWithNewData",
	}
	sr.AddArgument("domain", domain)

	return c.Call(sr, nil)
}

// RetryTransferWithDifferentAuthCode retries a transfer action with a new authcode
func RetryTransferWithDifferentAuthCode(c gotransip.Client, domain Domain, newAuthCode string) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "retryTransferWithDifferentAuthCode",
	}
	sr.AddArgument("domain", domain)
	sr.AddArgument("newAuthCode", newAuthCode)

	return c.Call(sr, nil)
}

// CancelDomainAction cancels a failed domain action
func CancelDomainAction(c gotransip.Client, domain string) error {
	sr := gotransip.SoapRequest{
		Service: serviceName,
		Method:  "cancelDomainAction",
	}
	sr.AddArgument("domain", domain)

	return c.Call(sr, nil)
}
