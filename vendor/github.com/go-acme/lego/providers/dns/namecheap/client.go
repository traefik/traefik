package namecheap

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Record describes a DNS record returned by the Namecheap DNS gethosts API.
// Namecheap uses the term "host" to refer to all DNS records that include
// a host field (A, AAAA, CNAME, NS, TXT, URL).
type Record struct {
	Type    string `xml:",attr"`
	Name    string `xml:",attr"`
	Address string `xml:",attr"`
	MXPref  string `xml:",attr"`
	TTL     string `xml:",attr"`
}

// apiError describes an error record in a namecheap API response.
type apiError struct {
	Number      int    `xml:",attr"`
	Description string `xml:",innerxml"`
}

type setHostsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Status  string     `xml:"Status,attr"`
	Errors  []apiError `xml:"Errors>Error"`
	Result  struct {
		IsSuccess string `xml:",attr"`
	} `xml:"CommandResponse>DomainDNSSetHostsResult"`
}

type getHostsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Status  string     `xml:"Status,attr"`
	Errors  []apiError `xml:"Errors>Error"`
	Hosts   []Record   `xml:"CommandResponse>DomainDNSGetHostsResult>host"`
}

type getTldsResponse struct {
	XMLName xml.Name   `xml:"ApiResponse"`
	Errors  []apiError `xml:"Errors>Error"`
	Result  []struct {
		Name string `xml:",attr"`
	} `xml:"CommandResponse>Tlds>Tld"`
}

// getTLDs requests the list of available TLDs.
// https://www.namecheap.com/support/api/methods/domains/get-tld-list.aspx
func (d *DNSProvider) getTLDs() (map[string]string, error) {
	request, err := d.newRequestGet("namecheap.domains.getTldList")
	if err != nil {
		return nil, err
	}

	var gtr getTldsResponse
	err = d.do(request, &gtr)
	if err != nil {
		return nil, err
	}

	if len(gtr.Errors) > 0 {
		return nil, fmt.Errorf("%s [%d]", gtr.Errors[0].Description, gtr.Errors[0].Number)
	}

	tlds := make(map[string]string)
	for _, t := range gtr.Result {
		tlds[t.Name] = t.Name
	}
	return tlds, nil
}

// getHosts reads the full list of DNS host records.
// https://www.namecheap.com/support/api/methods/domains-dns/get-hosts.aspx
func (d *DNSProvider) getHosts(sld, tld string) ([]Record, error) {
	request, err := d.newRequestGet("namecheap.domains.dns.getHosts",
		addParam("SLD", sld),
		addParam("TLD", tld),
	)
	if err != nil {
		return nil, err
	}

	var ghr getHostsResponse
	err = d.do(request, &ghr)
	if err != nil {
		return nil, err
	}

	if len(ghr.Errors) > 0 {
		return nil, fmt.Errorf("%s [%d]", ghr.Errors[0].Description, ghr.Errors[0].Number)
	}

	return ghr.Hosts, nil
}

// setHosts writes the full list of DNS host records .
// https://www.namecheap.com/support/api/methods/domains-dns/set-hosts.aspx
func (d *DNSProvider) setHosts(sld, tld string, hosts []Record) error {
	req, err := d.newRequestPost("namecheap.domains.dns.setHosts",
		addParam("SLD", sld),
		addParam("TLD", tld),
		func(values url.Values) {
			for i, h := range hosts {
				ind := fmt.Sprintf("%d", i+1)
				values.Add("HostName"+ind, h.Name)
				values.Add("RecordType"+ind, h.Type)
				values.Add("Address"+ind, h.Address)
				values.Add("MXPref"+ind, h.MXPref)
				values.Add("TTL"+ind, h.TTL)
			}
		},
	)
	if err != nil {
		return err
	}

	var shr setHostsResponse
	err = d.do(req, &shr)
	if err != nil {
		return err
	}

	if len(shr.Errors) > 0 {
		return fmt.Errorf("%s [%d]", shr.Errors[0].Description, shr.Errors[0].Number)
	}
	if shr.Result.IsSuccess != "true" {
		return fmt.Errorf("setHosts failed")
	}

	return nil
}

func (d *DNSProvider) do(req *http.Request, out interface{}) error {
	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var body []byte
		body, err = readBody(resp)
		if err != nil {
			return fmt.Errorf("HTTP error %d [%s]: %v", resp.StatusCode, http.StatusText(resp.StatusCode), err)
		}
		return fmt.Errorf("HTTP error %d [%s]: %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	body, err := readBody(resp)
	if err != nil {
		return err
	}

	if err := xml.Unmarshal(body, out); err != nil {
		return err
	}

	return nil
}

func (d *DNSProvider) newRequestGet(cmd string, params ...func(url.Values)) (*http.Request, error) {
	query := d.makeQuery(cmd, params...)

	reqURL, err := url.Parse(d.config.BaseURL)
	if err != nil {
		return nil, err
	}

	reqURL.RawQuery = query.Encode()

	return http.NewRequest(http.MethodGet, reqURL.String(), nil)
}

func (d *DNSProvider) newRequestPost(cmd string, params ...func(url.Values)) (*http.Request, error) {
	query := d.makeQuery(cmd, params...)

	req, err := http.NewRequest(http.MethodPost, d.config.BaseURL, strings.NewReader(query.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (d *DNSProvider) makeQuery(cmd string, params ...func(url.Values)) url.Values {
	queryParams := make(url.Values)
	queryParams.Set("ApiUser", d.config.APIUser)
	queryParams.Set("ApiKey", d.config.APIKey)
	queryParams.Set("UserName", d.config.APIUser)
	queryParams.Set("Command", cmd)
	queryParams.Set("ClientIp", d.config.ClientIP)

	for _, param := range params {
		param(queryParams)
	}

	return queryParams
}

func addParam(key, value string) func(url.Values) {
	return func(values url.Values) {
		values.Set(key, value)
	}
}

func readBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	defer resp.Body.Close()

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return rawBody, nil
}
