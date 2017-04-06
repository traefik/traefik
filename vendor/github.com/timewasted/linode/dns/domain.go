package dns

import (
	"errors"
	"strconv"
	"strings"

	"github.com/timewasted/linode"
)

type (
	// Domain represents a domain.
	Domain struct {
		AXFR_IPs         string `json:"AXFR_IPS"`
		Description      string `json:"DESCRIPTION"`
		Domain           string `json:"DOMAIN"`
		DomainID         int    `json:"DOMAINID"`
		Expire_Sec       int    `json:"EXPIRE_SEC"`
		LPM_DisplayGroup string `json:"LPM_DISPLAYGROUP"`
		Master_IPs       string `json:"MASTER_IPS"`
		Refresh_Sec      int    `json:"REFRESH_SEC"`
		Retry_Sec        int    `json:"RETRY_SEC"`
		SOA_Email        string `json:"SOA_EMAIL"`
		Status           int    `json:"STATUS"`
		TTL_Sec          int    `json:"TTL_SEC"`
		Type             string `json:"TYPE"`
	}
	// DomainResponse represents the response to a create, update, or
	// delete domain API call.
	DomainResponse struct {
		DomainID int `json:"DomainID"`
	}
)

// DeleteDomain executes the "domain.delete" API call.  This will delete the
// domain specified by domainID.
func (d *DNS) DeleteDomain(domainID int) (*DomainResponse, error) {
	params := linode.Parameters{
		"DomainID": strconv.Itoa(domainID),
	}
	var response *DomainResponse
	_, err := d.linode.Request("domain.delete", params, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetDomain returns the specified domain.  This search is not case-sensitive.
func (d *DNS) GetDomain(domain string) (*Domain, error) {
	list, err := d.GetDomains(nil)
	if err != nil {
		return nil, err
	}

	for _, d := range list {
		if strings.EqualFold(d.Domain, domain) {
			return d, nil
		}
	}

	return nil, linode.NewError(errors.New("dns: requested domain not found"))
}
