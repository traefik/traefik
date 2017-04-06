// Package namecheap implements a DNS provider for solving the DNS-01
// challenge using namecheap DNS.
package namecheap

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
)

// Notes about namecheap's tool API:
// 1. Using the API requires registration. Once registered, use your account
//    name and API key to access the API.
// 2. There is no API to add or modify a single DNS record. Instead you must
//    read the entire list of records, make modifications, and then write the
//    entire updated list of records.  (Yuck.)
// 3. Namecheap's DNS updates can be slow to propagate. I've seen them take
//    as long as an hour.
// 4. Namecheap requires you to whitelist the IP address from which you call
//    its APIs. It also requires all API calls to include the whitelisted IP
//    address as a form or query string value. This code uses a namecheap
//    service to query the client's IP address.

var (
	debug          = false
	defaultBaseURL = "https://api.namecheap.com/xml.response"
	getIPURL       = "https://dynamicdns.park-your-domain.com/getip"
	httpClient     = http.Client{Timeout: 60 * time.Second}
)

// DNSProvider is an implementation of the ChallengeProviderTimeout interface
// that uses Namecheap's tool API to manage TXT records for a domain.
type DNSProvider struct {
	baseURL  string
	apiUser  string
	apiKey   string
	clientIP string
}

// NewDNSProvider returns a DNSProvider instance configured for namecheap.
// Credentials must be passed in the environment variables: NAMECHEAP_API_USER
// and NAMECHEAP_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	apiUser := os.Getenv("NAMECHEAP_API_USER")
	apiKey := os.Getenv("NAMECHEAP_API_KEY")
	return NewDNSProviderCredentials(apiUser, apiKey)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for namecheap.
func NewDNSProviderCredentials(apiUser, apiKey string) (*DNSProvider, error) {
	if apiUser == "" || apiKey == "" {
		return nil, fmt.Errorf("Namecheap credentials missing")
	}

	clientIP, err := getClientIP()
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		baseURL:  defaultBaseURL,
		apiUser:  apiUser,
		apiKey:   apiKey,
		clientIP: clientIP,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Namecheap can sometimes take a long time to complete an
// update, so wait up to 60 minutes for the update to propagate.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Minute, 15 * time.Second
}

// host describes a DNS record returned by the Namecheap DNS gethosts API.
// Namecheap uses the term "host" to refer to all DNS records that include
// a host field (A, AAAA, CNAME, NS, TXT, URL).
type host struct {
	Type    string `xml:",attr"`
	Name    string `xml:",attr"`
	Address string `xml:",attr"`
	MXPref  string `xml:",attr"`
	TTL     string `xml:",attr"`
}

// apierror describes an error record in a namecheap API response.
type apierror struct {
	Number      int    `xml:",attr"`
	Description string `xml:",innerxml"`
}

// getClientIP returns the client's public IP address. It uses namecheap's
// IP discovery service to perform the lookup.
func getClientIP() (addr string, err error) {
	resp, err := httpClient.Get(getIPURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	clientIP, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if debug {
		fmt.Println("Client IP:", string(clientIP))
	}
	return string(clientIP), nil
}

// A challenge repesents all the data needed to specify a dns-01 challenge
// to lets-encrypt.
type challenge struct {
	domain   string
	key      string
	keyFqdn  string
	keyValue string
	tld      string
	sld      string
	host     string
}

// newChallenge builds a challenge record from a domain name, a challenge
// authentication key, and a map of available TLDs.
func newChallenge(domain, keyAuth string, tlds map[string]string) (*challenge, error) {
	domain = acme.UnFqdn(domain)
	parts := strings.Split(domain, ".")

	// Find the longest matching TLD.
	longest := -1
	for i := len(parts); i > 0; i-- {
		t := strings.Join(parts[i-1:], ".")
		if _, found := tlds[t]; found {
			longest = i - 1
		}
	}
	if longest < 1 {
		return nil, fmt.Errorf("Invalid domain name '%s'", domain)
	}

	tld := strings.Join(parts[longest:], ".")
	sld := parts[longest-1]

	var host string
	if longest >= 1 {
		host = strings.Join(parts[:longest-1], ".")
	}

	key, keyValue, _ := acme.DNS01Record(domain, keyAuth)

	return &challenge{
		domain:   domain,
		key:      "_acme-challenge." + host,
		keyFqdn:  key,
		keyValue: keyValue,
		tld:      tld,
		sld:      sld,
		host:     host,
	}, nil
}

// setGlobalParams adds the namecheap global parameters to the provided url
// Values record.
func (d *DNSProvider) setGlobalParams(v *url.Values, cmd string) {
	v.Set("ApiUser", d.apiUser)
	v.Set("ApiKey", d.apiKey)
	v.Set("UserName", d.apiUser)
	v.Set("ClientIp", d.clientIP)
	v.Set("Command", cmd)
}

// getTLDs requests the list of available TLDs from namecheap.
func (d *DNSProvider) getTLDs() (tlds map[string]string, err error) {
	values := make(url.Values)
	d.setGlobalParams(&values, "namecheap.domains.getTldList")

	reqURL, _ := url.Parse(d.baseURL)
	reqURL.RawQuery = values.Encode()

	resp, err := httpClient.Get(reqURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("getHosts HTTP error %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type GetTldsResponse struct {
		XMLName xml.Name   `xml:"ApiResponse"`
		Errors  []apierror `xml:"Errors>Error"`
		Result  []struct {
			Name string `xml:",attr"`
		} `xml:"CommandResponse>Tlds>Tld"`
	}

	var gtr GetTldsResponse
	if err := xml.Unmarshal(body, &gtr); err != nil {
		return nil, err
	}
	if len(gtr.Errors) > 0 {
		return nil, fmt.Errorf("Namecheap error: %s [%d]",
			gtr.Errors[0].Description, gtr.Errors[0].Number)
	}

	tlds = make(map[string]string)
	for _, t := range gtr.Result {
		tlds[t.Name] = t.Name
	}
	return tlds, nil
}

// getHosts reads the full list of DNS host records using the Namecheap API.
func (d *DNSProvider) getHosts(ch *challenge) (hosts []host, err error) {
	values := make(url.Values)
	d.setGlobalParams(&values, "namecheap.domains.dns.getHosts")
	values.Set("SLD", ch.sld)
	values.Set("TLD", ch.tld)

	reqURL, _ := url.Parse(d.baseURL)
	reqURL.RawQuery = values.Encode()

	resp, err := httpClient.Get(reqURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("getHosts HTTP error %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type GetHostsResponse struct {
		XMLName xml.Name   `xml:"ApiResponse"`
		Status  string     `xml:"Status,attr"`
		Errors  []apierror `xml:"Errors>Error"`
		Hosts   []host     `xml:"CommandResponse>DomainDNSGetHostsResult>host"`
	}

	var ghr GetHostsResponse
	if err = xml.Unmarshal(body, &ghr); err != nil {
		return nil, err
	}
	if len(ghr.Errors) > 0 {
		return nil, fmt.Errorf("Namecheap error: %s [%d]",
			ghr.Errors[0].Description, ghr.Errors[0].Number)
	}

	return ghr.Hosts, nil
}

// setHosts writes the full list of DNS host records using the Namecheap API.
func (d *DNSProvider) setHosts(ch *challenge, hosts []host) error {
	values := make(url.Values)
	d.setGlobalParams(&values, "namecheap.domains.dns.setHosts")
	values.Set("SLD", ch.sld)
	values.Set("TLD", ch.tld)

	for i, h := range hosts {
		ind := fmt.Sprintf("%d", i+1)
		values.Add("HostName"+ind, h.Name)
		values.Add("RecordType"+ind, h.Type)
		values.Add("Address"+ind, h.Address)
		values.Add("MXPref"+ind, h.MXPref)
		values.Add("TTL"+ind, h.TTL)
	}

	resp, err := httpClient.PostForm(d.baseURL, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("setHosts HTTP error %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type SetHostsResponse struct {
		XMLName xml.Name   `xml:"ApiResponse"`
		Status  string     `xml:"Status,attr"`
		Errors  []apierror `xml:"Errors>Error"`
		Result  struct {
			IsSuccess string `xml:",attr"`
		} `xml:"CommandResponse>DomainDNSSetHostsResult"`
	}

	var shr SetHostsResponse
	if err := xml.Unmarshal(body, &shr); err != nil {
		return err
	}
	if len(shr.Errors) > 0 {
		return fmt.Errorf("Namecheap error: %s [%d]",
			shr.Errors[0].Description, shr.Errors[0].Number)
	}
	if shr.Result.IsSuccess != "true" {
		return fmt.Errorf("Namecheap setHosts failed.")
	}

	return nil
}

// addChallengeRecord adds a DNS challenge TXT record to a list of namecheap
// host records.
func (d *DNSProvider) addChallengeRecord(ch *challenge, hosts *[]host) {
	host := host{
		Name:    ch.key,
		Type:    "TXT",
		Address: ch.keyValue,
		MXPref:  "10",
		TTL:     "120",
	}

	// If there's already a TXT record with the same name, replace it.
	for i, h := range *hosts {
		if h.Name == ch.key && h.Type == "TXT" {
			(*hosts)[i] = host
			return
		}
	}

	// No record was replaced, so add a new one.
	*hosts = append(*hosts, host)
}

// removeChallengeRecord removes a DNS challenge TXT record from a list of
// namecheap host records. Return true if a record was removed.
func (d *DNSProvider) removeChallengeRecord(ch *challenge, hosts *[]host) bool {
	// Find the challenge TXT record and remove it if found.
	for i, h := range *hosts {
		if h.Name == ch.key && h.Type == "TXT" {
			*hosts = append((*hosts)[:i], (*hosts)[i+1:]...)
			return true
		}
	}

	return false
}

// Present installs a TXT record for the DNS challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	tlds, err := d.getTLDs()
	if err != nil {
		return err
	}

	ch, err := newChallenge(domain, keyAuth, tlds)
	if err != nil {
		return err
	}

	hosts, err := d.getHosts(ch)
	if err != nil {
		return err
	}

	d.addChallengeRecord(ch, &hosts)

	if debug {
		for _, h := range hosts {
			fmt.Printf(
				"%-5.5s %-30.30s %-6s %-70.70s\n",
				h.Type, h.Name, h.TTL, h.Address)
		}
	}

	return d.setHosts(ch, hosts)
}

// CleanUp removes a TXT record used for a previous DNS challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	tlds, err := d.getTLDs()
	if err != nil {
		return err
	}

	ch, err := newChallenge(domain, keyAuth, tlds)
	if err != nil {
		return err
	}

	hosts, err := d.getHosts(ch)
	if err != nil {
		return err
	}

	if removed := d.removeChallengeRecord(ch, &hosts); !removed {
		return nil
	}

	return d.setHosts(ch, hosts)
}
