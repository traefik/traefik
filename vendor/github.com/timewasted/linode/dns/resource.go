package dns

import (
	"errors"
	"strconv"
	"strings"

	"github.com/timewasted/linode"
)

type (
	// Resource represents a domain resource.
	Resource struct {
		DomainID   int         `json:"DOMAINID"`
		Name       string      `json:"NAME"`
		Port       interface{} `json:"PORT"`
		Priority   interface{} `json:"PRIORITY"`
		Protocol   string      `json:"PROTOCOL"`
		ResourceID int         `json:"RESOURCEID"`
		Target     string      `json:"TARGET"`
		TTL_Sec    int         `json:"TTL_SEC"`
		Type       string      `json:"TYPE"`
		Weight     interface{} `json:"WEIGHT"`
	}
	// ResourceResponse represents the response to a create, update, or
	// delete resource API call.
	ResourceResponse struct {
		ResourceID int `json:"ResourceID"`
	}
)

// CreateDomainResourceA executes the "domain.resource.create" API call.  This
// will create a new "A" resource using the specified parameters.
func (d *DNS) CreateDomainResourceA(domainID int, name, target string, ttlSeconds int) (*ResourceResponse, error) {
	return d.createDomainResource(linode.Parameters{
		"Type":     "A",
		"DomainID": strconv.Itoa(domainID),
		"Name":     name,
		"Target":   target,
		"TTL_Sec":  strconv.Itoa(ttlSeconds),
	})
}

// CreateDomainResourceAAAA executes the "domain.resource.create" API call.
// This will create a new "AAAA" resource using the specified parameters.
func (d *DNS) CreateDomainResourceAAAA(domainID int, name, target string, ttlSeconds int) (*ResourceResponse, error) {
	return d.createDomainResource(linode.Parameters{
		"Type":     "AAAA",
		"DomainID": strconv.Itoa(domainID),
		"Name":     name,
		"Target":   target,
		"TTL_Sec":  strconv.Itoa(ttlSeconds),
	})
}

// CreateDomainResourceCNAME executes the "domain.resource.create" API call.
// This will create a new "CNAME" resource using the specified parameters.
func (d *DNS) CreateDomainResourceCNAME(domainID int, name, target string, ttlSeconds int) (*ResourceResponse, error) {
	return d.createDomainResource(linode.Parameters{
		"Type":     "CNAME",
		"DomainID": strconv.Itoa(domainID),
		"Name":     name,
		"Target":   target,
		"TTL_Sec":  strconv.Itoa(ttlSeconds),
	})
}

// CreateDomainResourceMX executes the "domain.resource.create" API call.  This
// will create a new "MX" resource using the specified parameters.
func (d *DNS) CreateDomainResourceMX(domainID int, name, target string, priority, ttlSeconds int) (*ResourceResponse, error) {
	return d.createDomainResource(linode.Parameters{
		"Type":     "MX",
		"DomainID": strconv.Itoa(domainID),
		"Name":     name,
		"Target":   target,
		"Priority": strconv.Itoa(priority),
		"TTL_Sec":  strconv.Itoa(ttlSeconds),
	})
}

// CreateDomainResourceNS executes the "domain.resource.create" API call.  This
// will create a new "NS" resource using the specified parameters.
func (d *DNS) CreateDomainResourceNS(domainID int, name, target string, ttlSeconds int) (*ResourceResponse, error) {
	return d.createDomainResource(linode.Parameters{
		"Type":     "NS",
		"DomainID": strconv.Itoa(domainID),
		"Name":     name,
		"Target":   target,
		"TTL_Sec":  strconv.Itoa(ttlSeconds),
	})
}

// CreateDomainResourceSRV executes the "domain.resource.create" API call.  This
// will create a new "SRV" resource using the specified parameters.
func (d *DNS) CreateDomainResourceSRV(domainID int, name, target, protocol string, priority, ttlSeconds int) (*ResourceResponse, error) {
	// FIXME: This probably also needs weight and port.  Weight has a valid
	// range of 0-255, while port is 0-65535.
	return d.createDomainResource(linode.Parameters{
		"Type":     "SRV",
		"DomainID": strconv.Itoa(domainID),
		"Name":     name,
		"Target":   target,
		"Protocol": protocol,
		"Priority": strconv.Itoa(priority),
		"TTL_Sec":  strconv.Itoa(ttlSeconds),
	})
}

// CreateDomainResourceTXT executes the "domain.resource.create" API call.  This
// will create a new "TXT" resource using the specified parameters.
func (d *DNS) CreateDomainResourceTXT(domainID int, name, target string, ttlSeconds int) (*ResourceResponse, error) {
	return d.createDomainResource(linode.Parameters{
		"Type":     "TXT",
		"DomainID": strconv.Itoa(domainID),
		"Name":     name,
		"Target":   target,
		"TTL_Sec":  strconv.Itoa(ttlSeconds),
	})
}

// CreateDomainResource executes the "domain.resource.create" API call.  This
// will create a new resource using the values specified in the resource.
func (d *DNS) CreateDomainResource(r *Resource) (*ResourceResponse, error) {
	// Ensure that the resource has a name.
	if len(r.Name) == 0 {
		return nil, linode.NewError(errors.New("dns: creating a resource requires Name be specified"))
	}

	// Initialize parameters that are shared across resource types.
	params := linode.Parameters{
		"DomainID": strconv.Itoa(r.DomainID),
		"Name":     r.Name,
		"TTL_Sec":  strconv.Itoa(r.TTL_Sec),
	}

	// Ensure that the resource has a valid, supported type.
	r.Type = strings.ToUpper(r.Type)
	switch r.Type {
	case "A":
	case "AAAA":
	case "CNAME":
	case "MX":
	case "NS":
	case "TXT":
		// No further processing required for these types.
		break
	case "SRV":
		// Ensure that SRV has a protocol.
		if len(r.Protocol) == 0 {
			return nil, linode.NewError(errors.New("dns: creating a SRV resource requires Priority be specified"))
		}
		params.Set("Protocol", r.Protocol)
		break
	default:
		// Unsupported type.
		return nil, linode.NewError(errors.New("dns: can not create resource of unsupported type: " + r.Type))
	}
	params.Set("Type", r.Type)

	// Ensure that the resource has a valid target.
	if len(r.Target) == 0 {
		return nil, linode.NewError(errors.New("dns: creating a resource requires Target to be specified"))
	}
	params.Set("Target", r.Target)

	if r.Name == "MX" || r.Name == "SRV" {
		// If priority is defined, ensure that it's valid.
		if r.Priority != nil {
			priority, ok := r.Priority.(int)
			if !ok {
				return nil, linode.NewError(errors.New("dns: priority must be specified as an int"))
			}
			if priority < 0 || priority > 255 {
				return nil, linode.NewError(errors.New("dns: priority must be within the range of 0-255"))
			}
			r.Priority = priority
			params.Set("Priority", strconv.Itoa(priority))
		}
	}

	// Create the resource.
	return d.createDomainResource(params)
}

// createDomainResource executes the "domain.resource.create" API call.  This
// will create a resource using the specified parameters.
func (d *DNS) createDomainResource(params linode.Parameters) (*ResourceResponse, error) {
	var response *ResourceResponse
	_, err := d.linode.Request("domain.resource.create", params, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// DeleteDomainResource executes the "domain.resource.delete" API call.  This
// will delete the resource specified by resourceID under the domain specified
// by domainID.
func (d *DNS) DeleteDomainResource(domainID, resourceID int) (*ResourceResponse, error) {
	params := linode.Parameters{
		"DomainID":   strconv.Itoa(domainID),
		"ResourceID": strconv.Itoa(resourceID),
	}
	var response *ResourceResponse
	_, err := d.linode.Request("domain.resource.delete", params, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetResourceByType returns a list of domain resources that match the specified
// type.  This search is not case-sensitive.
func (d *DNS) GetResourcesByType(domainID int, res_type string) ([]*Resource, error) {
	resources, err := d.GetDomainResources(domainID)
	if err != nil {
		return nil, err
	}

	list := []*Resource{}
	for _, r := range resources {
		if strings.EqualFold(r.Type, res_type) {
			list = append(list, r)
		}
	}

	return list, nil
}
