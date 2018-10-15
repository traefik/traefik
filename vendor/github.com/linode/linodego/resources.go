package linodego

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/go-resty/resty"
)

const (
	stackscriptsName          = "stackscripts"
	imagesName                = "images"
	instancesName             = "instances"
	instanceDisksName         = "disks"
	instanceConfigsName       = "configs"
	instanceIPsName           = "ips"
	instanceSnapshotsName     = "snapshots"
	instanceVolumesName       = "instancevolumes"
	ipaddressesName           = "ipaddresses"
	ipv6poolsName             = "ipv6pools"
	ipv6rangesName            = "ipv6ranges"
	regionsName               = "regions"
	volumesName               = "volumes"
	kernelsName               = "kernels"
	typesName                 = "types"
	domainsName               = "domains"
	domainRecordsName         = "records"
	longviewName              = "longview"
	longviewclientsName       = "longviewclients"
	longviewsubscriptionsName = "longviewsubscriptions"
	nodebalancersName         = "nodebalancers"
	nodebalancerconfigsName   = "nodebalancerconfigs"
	nodebalancernodesName     = "nodebalancernodes"
	sshkeysName               = "sshkeys"
	ticketsName               = "tickets"
	accountName               = "account"
	eventsName                = "events"
	invoicesName              = "invoices"
	invoiceItemsName          = "invoiceitems"
	profileName               = "profile"
	managedName               = "managed"
	// notificationsName = "notifications"

	stackscriptsEndpoint          = "linode/stackscripts"
	imagesEndpoint                = "images"
	instancesEndpoint             = "linode/instances"
	instanceConfigsEndpoint       = "linode/instances/{{ .ID }}/configs"
	instanceDisksEndpoint         = "linode/instances/{{ .ID }}/disks"
	instanceSnapshotsEndpoint     = "linode/instances/{{ .ID }}/backups"
	instanceIPsEndpoint           = "linode/instances/{{ .ID }}/ips"
	instanceVolumesEndpoint       = "linode/instances/{{ .ID }}/volumes"
	ipaddressesEndpoint           = "network/ips"
	ipv6poolsEndpoint             = "network/ipv6/pools"
	ipv6rangesEndpoint            = "network/ipv6/ranges"
	regionsEndpoint               = "regions"
	volumesEndpoint               = "volumes"
	kernelsEndpoint               = "linode/kernels"
	typesEndpoint                 = "linode/types"
	domainsEndpoint               = "domains"
	domainRecordsEndpoint         = "domains/{{ .ID }}/records"
	longviewEndpoint              = "longview"
	longviewclientsEndpoint       = "longview/clients"
	longviewsubscriptionsEndpoint = "longview/subscriptions"
	nodebalancersEndpoint         = "nodebalancers"
	// @TODO we can't use these nodebalancer endpoints unless we include these templated fields
	// The API seems inconsistent about including parent IDs in objects, (compare instance configs to nb configs)
	// Parent IDs would be immutable for updates and are ignored in create requests ..
	// Should we include these fields in CreateOpts and UpdateOpts?
	nodebalancerconfigsEndpoint = "nodebalancers/{{ .ID }}/configs"
	nodebalancernodesEndpoint   = "nodebalancers/{{ .ID }}/configs/{{ .SecondID }}/nodes"
	sshkeysEndpoint             = "profile/sshkeys"
	ticketsEndpoint             = "support/tickets"
	accountEndpoint             = "account"
	eventsEndpoint              = "account/events"
	invoicesEndpoint            = "account/invoices"
	invoiceItemsEndpoint        = "account/invoices/{{ .ID }}/items"
	profileEndpoint             = "profile"
	managedEndpoint             = "managed"
	// notificationsEndpoint       = "account/notifications"
)

// Resource represents a linode API resource
type Resource struct {
	name             string
	endpoint         string
	isTemplate       bool
	endpointTemplate *template.Template
	R                func(ctx context.Context) *resty.Request
	PR               func(ctx context.Context) *resty.Request
}

// NewResource is the factory to create a new Resource struct. If it has a template string the useTemplate bool must be set.
func NewResource(client *Client, name string, endpoint string, useTemplate bool, singleType interface{}, pagedType interface{}) *Resource {
	var tmpl *template.Template

	if useTemplate {
		tmpl = template.Must(template.New(name).Parse(endpoint))
	}

	r := func(ctx context.Context) *resty.Request {
		return client.R(ctx).SetResult(singleType)
	}

	pr := func(ctx context.Context) *resty.Request {
		return client.R(ctx).SetResult(pagedType)
	}

	return &Resource{name, endpoint, useTemplate, tmpl, r, pr}
}

func (r Resource) render(data ...interface{}) (string, error) {
	if data == nil {
		return "", NewError("Cannot template endpoint with <nil> data")
	}
	out := ""
	buf := bytes.NewBufferString(out)

	var substitutions interface{}
	if len(data) == 1 {
		substitutions = struct{ ID interface{} }{data[0]}
	} else if len(data) == 2 {
		substitutions = struct {
			ID       interface{}
			SecondID interface{}
		}{data[0], data[1]}
	} else {
		return "", NewError("Too many arguments to render template (expected 1 or 2)")
	}
	if err := r.endpointTemplate.Execute(buf, substitutions); err != nil {
		return "", NewError(err)
	}
	return buf.String(), nil
}

// endpointWithID will return the rendered endpoint string for the resource with provided id
func (r Resource) endpointWithID(id ...int) (string, error) {
	if !r.isTemplate {
		return r.endpoint, nil
	}
	data := make([]interface{}, len(id))
	for i, v := range id {
		data[i] = v
	}
	return r.render(data...)
}

// Endpoint will return the non-templated endpoint string for resource
func (r Resource) Endpoint() (string, error) {
	if r.isTemplate {
		return "", NewError(fmt.Sprintf("Tried to get endpoint for %s without providing data for template", r.name))
	}
	return r.endpoint, nil
}
