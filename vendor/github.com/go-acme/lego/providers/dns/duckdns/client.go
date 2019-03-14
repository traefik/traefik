package duckdns

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/miekg/dns"
)

// updateTxtRecord Update the domains TXT record
// To update the TXT record we just need to make one simple get request.
// In DuckDNS you only have one TXT record shared with the domain and all sub domains.
func (d *DNSProvider) updateTxtRecord(domain, token, txt string, clear bool) error {
	u, _ := url.Parse("https://www.duckdns.org/update")

	mainDomain := getMainDomain(domain)
	if len(mainDomain) == 0 {
		return fmt.Errorf("unable to find the main domain for: %s", domain)
	}

	query := u.Query()
	query.Set("domains", mainDomain)
	query.Set("token", token)
	query.Set("clear", strconv.FormatBool(clear))
	query.Set("txt", txt)
	u.RawQuery = query.Encode()

	response, err := d.config.HTTPClient.Get(u.String())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	body := string(bodyBytes)
	if body != "OK" {
		return fmt.Errorf("request to change TXT record for DuckDNS returned the following result (%s) this does not match expectation (OK) used url [%s]", body, u)
	}
	return nil
}

// DuckDNS only lets you write to your subdomain
// so it must be in format subdomain.duckdns.org
// not in format subsubdomain.subdomain.duckdns.org
// so strip off everything that is not top 3 levels
func getMainDomain(domain string) string {
	domain = dns01.UnFqdn(domain)

	split := dns.Split(domain)
	if strings.HasSuffix(strings.ToLower(domain), "duckdns.org") {
		if len(split) < 3 {
			return ""
		}

		firstSubDomainIndex := split[len(split)-3]
		return domain[firstSubDomainIndex:]
	}

	return domain[split[len(split)-1]:]
}
