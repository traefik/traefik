package vegadns2client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Domain - struct containing a domain object
type Domain struct {
	Status   string `json:"status"`
	Domain   string `json:"domain"`
	DomainID int    `json:"domain_id"`
	OwnerID  int    `json:"owner_id"`
}

// DomainResponse - api response of a domain list
type DomainResponse struct {
	Status  string   `json:"status"`
	Total   int      `json:"total_domains"`
	Domains []Domain `json:"domains"`
}

// GetDomainID - returns the id for a domain
// Input: domain
// Output: int, err
func (vega *VegaDNSClient) GetDomainID(domain string) (int, error) {
	params := make(map[string]string)
	params["search"] = domain

	resp, err := vega.Send("GET", "domains", params)

	if err != nil {
		return -1, fmt.Errorf("Error sending GET to GetDomainID: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response from GET to GetDomainID: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("Got bad answer from VegaDNS on GetDomainID. Code: %d. Message: %s", resp.StatusCode, string(body))
	}

	answer := DomainResponse{}
	if err := json.Unmarshal(body, &answer); err != nil {
		return -1, fmt.Errorf("Error unmarshalling body from GetDomainID: %s", err)
	}
	log.Println(answer)
	for _, d := range answer.Domains {
		if d.Domain == domain {
			return d.DomainID, nil
		}
	}

	return -1, fmt.Errorf("Didnt find domain %s", domain)

}

// GetAuthZone retrieves the closest match to a given
// domain. Example: Given an argument "a.b.c.d.e", and a VegaDNS
// hosted domain of "c.d.e", GetClosestMatchingDomain will return
// "c.d.e".
func (vega *VegaDNSClient) GetAuthZone(fqdn string) (string, int, error) {
	fqdn = strings.TrimSuffix(fqdn, ".")
	numComponents := len(strings.Split(fqdn, "."))
	for i := 1; i < numComponents; i++ {
		tmpHostname := strings.SplitN(fqdn, ".", i)[i-1]
		log.Printf("tmpHostname for i = %d: %s\n", i, tmpHostname)
		if domainID, err := vega.GetDomainID(tmpHostname); err == nil {
			log.Printf("Found zone: %s\n\tShortened to %s\n", tmpHostname, strings.TrimSuffix(tmpHostname, "."))
			return strings.TrimSuffix(tmpHostname, "."), domainID, nil
		}
	}
	log.Println("Unable to find hosted zone in vegadns")
	return "", -1, fmt.Errorf("Unable to find auth zone for fqdn %s", fqdn)
}
