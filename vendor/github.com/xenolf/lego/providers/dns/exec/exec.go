// Package exec implements a manual DNS provider which runs a program for
// adding/removing the DNS record.
//
// The file name of the external program is specified in the environment
// variable EXEC_PATH. When it is run by lego, three command-line parameters
// are passed to it: The action ("present" or "cleanup"), the fully-qualified domain
// name, the value for the record and the TTL.
//
// For example, requesting a certificate for the domain 'foo.example.com' can
// be achieved by calling lego as follows:
//
//  EXEC_PATH=./update-dns.sh \
//    lego --dns exec \
//    --domains foo.example.com \
//    --email invalid@example.com run
//
// It will then call the program './update-dns.sh' with like this:
//
//  ./update-dns.sh "present" "_acme-challenge.foo.example.com." "MsijOYZxqyjGnFGwhjrhfg-Xgbl5r68WPda0J9EgqqI" "120"
//
// The program then needs to make sure the record is inserted. When it returns
// an error via a non-zero exit code, lego aborts.
//
// When the record is to be removed again, the program is called with the first
// command-line parameter set to "cleanup" instead of "present".
package exec

import (
	"errors"
	"os"
	"os/exec"
	"strconv"

	"github.com/xenolf/lego/acme"
)

// DNSProvider adds and removes the record for the DNS challenge by calling a
// program with command-line parameters.
type DNSProvider struct {
	program string
}

// NewDNSProvider returns a new DNS provider which runs the program in the
// environment variable EXEC_PATH for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	s := os.Getenv("EXEC_PATH")
	if s == "" {
		return nil, errors.New("environment variable EXEC_PATH not set")
	}

	return &DNSProvider{program: s}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	cmd := exec.Command(d.program, "present", fqdn, value, strconv.Itoa(ttl))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	cmd := exec.Command(d.program, "cleanup", fqdn, value, strconv.Itoa(ttl))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
