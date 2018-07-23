package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/config/env"
)

// Config Provider configuration.
type Config struct {
	Program string
	Mode    string
}

// DNSProvider adds and removes the record for the DNS challenge by calling a
// program with command-line parameters.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider which runs the program in the
// environment variable EXEC_PATH for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("EXEC_PATH")
	if err != nil {
		return nil, fmt.Errorf("exec: %v", err)
	}

	return NewDNSProviderConfig(&Config{
		Program: values["EXEC_PATH"],
		Mode:    os.Getenv("EXEC_MODE"),
	})
}

// NewDNSProviderConfig returns a new DNS provider which runs the given configuration
// for adding and removing the DNS record.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration is nil")
	}

	return &DNSProvider{config: config}, nil
}

// NewDNSProviderProgram returns a new DNS provider which runs the given program
// for adding and removing the DNS record.
// Deprecated: use NewDNSProviderConfig instead
func NewDNSProviderProgram(program string) (*DNSProvider, error) {
	if len(program) == 0 {
		return nil, errors.New("the program is undefined")
	}

	return NewDNSProviderConfig(&Config{Program: program})
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	var args []string
	if d.config.Mode == "RAW" {
		args = []string{"present", "--", domain, token, keyAuth}
	} else {
		fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
		args = []string{"present", fqdn, value, strconv.Itoa(ttl)}
	}

	cmd := exec.Command(d.config.Program, args...)

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}

	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	var args []string
	if d.config.Mode == "RAW" {
		args = []string{"cleanup", "--", domain, token, keyAuth}
	} else {
		fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
		args = []string{"cleanup", fqdn, value, strconv.Itoa(ttl)}
	}

	cmd := exec.Command(d.config.Program, args...)

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}

	return err
}
