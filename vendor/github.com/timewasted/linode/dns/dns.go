package dns

import (
	"strconv"

	"github.com/timewasted/linode"
)

type (
	// DNS represents the interface to the DNS portion of Linode's API.
	DNS struct {
		linode *linode.Linode
	}
)

// New returns a pointer to a new DNS object.
func New(apiKey string) *DNS {
	return &DNS{
		linode: linode.New(apiKey),
	}
}

// FromLinode returns a pointer to a new DNS object, using the provided Linode
// instance as backing.
func FromLinode(l *linode.Linode) *DNS {
	return &DNS{
		linode: l,
	}
}

// ToLinode returns a pointer to the internal Linode object.
func (d *DNS) ToLinode() *linode.Linode {
	return d.linode
}

// GetDomains executes the "domain.list" API call.  When domainID is nil, this
// will return a list of domains.  Otherwise, it will return only the domain
// specified by domainID.
func (d *DNS) GetDomains(domainId interface{}) ([]*Domain, error) {
	params := linode.Parameters{}
	if domainId != nil {
		id, ok := domainId.(int)
		if ok {
			params.Set("DomainID", strconv.Itoa(id))
		}
	}

	var list []*Domain
	_, err := d.linode.Request("domain.list", params, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetDomainResources executes the "domain.resource.list" API call.  This will
// return a list of domain resources associated with the specified domainID.
func (d *DNS) GetDomainResources(domainID int) ([]*Resource, error) {
	params := linode.Parameters{
		"DomainID": strconv.Itoa(domainID),
	}

	var list []*Resource
	_, err := d.linode.Request("domain.resource.list", params, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
