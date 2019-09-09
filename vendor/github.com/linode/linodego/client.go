package linodego

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/resty.v1"
)

const (
	// APIHost Linode API hostname
	APIHost = "api.linode.com"
	// APIHostVar environment var to check for alternate API URL
	APIHostVar = "LINODE_URL"
	// APIHostCert environment var containing path to CA cert to validate against
	APIHostCert = "LINODE_CA"
	// APIVersion Linode API version
	APIVersion = "v4"
	// APIProto connect to API with http(s)
	APIProto = "https"
	// Version of linodego
	Version = "0.10.0"
	// APIEnvVar environment var to check for API token
	APIEnvVar = "LINODE_TOKEN"
	// APISecondsPerPoll how frequently to poll for new Events or Status in WaitFor functions
	APISecondsPerPoll = 3
	// DefaultUserAgent is the default User-Agent sent in HTTP request headers
	DefaultUserAgent = "linodego " + Version + " https://github.com/linode/linodego"
)

var (
	envDebug = false
)

// Client is a wrapper around the Resty client
type Client struct {
	resty     *resty.Client
	userAgent string
	resources map[string]*Resource
	debug     bool

	millisecondsPerPoll time.Duration

	Images                *Resource
	InstanceDisks         *Resource
	InstanceConfigs       *Resource
	InstanceSnapshots     *Resource
	InstanceIPs           *Resource
	InstanceVolumes       *Resource
	InstanceStats         *Resource
	Instances             *Resource
	IPAddresses           *Resource
	IPv6Pools             *Resource
	IPv6Ranges            *Resource
	Regions               *Resource
	StackScripts          *Resource
	Volumes               *Resource
	Kernels               *Resource
	Types                 *Resource
	Domains               *Resource
	DomainRecords         *Resource
	Longview              *Resource
	LongviewClients       *Resource
	LongviewSubscriptions *Resource
	NodeBalancers         *Resource
	NodeBalancerConfigs   *Resource
	NodeBalancerNodes     *Resource
	SSHKeys               *Resource
	Tickets               *Resource
	Tokens                *Resource
	Token                 *Resource
	Account               *Resource
	AccountSettings       *Resource
	Invoices              *Resource
	InvoiceItems          *Resource
	Events                *Resource
	Notifications         *Resource
	OAuthClients          *Resource
	Profile               *Resource
	Managed               *Resource
	Tags                  *Resource
	Users                 *Resource
	Payments              *Resource
}

func init() {
	// Wether or not we will enable Resty debugging output
	if apiDebug, ok := os.LookupEnv("LINODE_DEBUG"); ok {
		if parsed, err := strconv.ParseBool(apiDebug); err == nil {
			envDebug = parsed
			log.Println("[INFO] LINODE_DEBUG being set to", envDebug)
		} else {
			log.Println("[WARN] LINODE_DEBUG should be an integer, 0 or 1")
		}
	}

}

// SetUserAgent sets a custom user-agent for HTTP requests
func (c *Client) SetUserAgent(ua string) *Client {
	c.userAgent = ua
	c.resty.SetHeader("User-Agent", c.userAgent)

	return c
}

// R wraps resty's R method
func (c *Client) R(ctx context.Context) *resty.Request {
	return c.resty.R().
		ExpectContentType("application/json").
		SetHeader("Content-Type", "application/json").
		SetContext(ctx).
		SetError(APIError{})
}

// SetDebug sets the debug on resty's client
func (c *Client) SetDebug(debug bool) *Client {
	c.debug = debug
	c.resty.SetDebug(debug)
	return c
}

// SetBaseURL sets the base URL of the Linode v4 API (https://api.linode.com/v4)
func (c *Client) SetBaseURL(url string) *Client {
	c.resty.SetHostURL(url)
	return c
}

func (c *Client) SetRootCertificate(path string) *Client {
	c.resty.SetRootCertificate(path)
	return c
}

// SetToken sets the API token for all requests from this client
// Only necessary if you haven't already provided an http client to NewClient() configured with the token.
func (c *Client) SetToken(token string) *Client {
	c.resty.SetHeader("Authorization", fmt.Sprintf("Bearer %s", token))
	return c
}

// SetPollDelay sets the number of milliseconds to wait between events or status polls.
// Affects all WaitFor* functions.
func (c *Client) SetPollDelay(delay time.Duration) *Client {
	c.millisecondsPerPoll = delay
	return c
}

// Resource looks up a resource by name
func (c Client) Resource(resourceName string) *Resource {
	selectedResource, ok := c.resources[resourceName]
	if !ok {
		log.Fatalf("Could not find resource named '%s', exiting.", resourceName)
	}
	return selectedResource
}

// NewClient factory to create new Client struct
func NewClient(hc *http.Client) (client Client) {
	if hc != nil {
		client.resty = resty.NewWithClient(hc)
	} else {
		client.resty = resty.New()
	}
	client.SetUserAgent(DefaultUserAgent)
	baseURL, baseURLExists := os.LookupEnv(APIHostVar)
	if baseURLExists {
		client.SetBaseURL(baseURL)
	} else {
		client.SetBaseURL(fmt.Sprintf("%s://%s/%s", APIProto, APIHost, APIVersion))
	}
	certPath, certPathExists := os.LookupEnv(APIHostCert)
	if certPathExists {
		cert, err := ioutil.ReadFile(certPath)
		if err != nil {
			log.Fatalf("[ERROR] Error when reading cert at %s: %s\n", certPath, err.Error())
		}
		client.SetRootCertificate(certPath)
		if envDebug {
			log.Printf("[DEBUG] Set API root certificate to %s with contents %s\n", certPath, cert)
		}
	}
	client.SetPollDelay(1000 * APISecondsPerPoll)

	resources := map[string]*Resource{
		stackscriptsName:          NewResource(&client, stackscriptsName, stackscriptsEndpoint, false, Stackscript{}, StackscriptsPagedResponse{}),
		imagesName:                NewResource(&client, imagesName, imagesEndpoint, false, Image{}, ImagesPagedResponse{}),
		instancesName:             NewResource(&client, instancesName, instancesEndpoint, false, Instance{}, InstancesPagedResponse{}),
		instanceDisksName:         NewResource(&client, instanceDisksName, instanceDisksEndpoint, true, InstanceDisk{}, InstanceDisksPagedResponse{}),
		instanceConfigsName:       NewResource(&client, instanceConfigsName, instanceConfigsEndpoint, true, InstanceConfig{}, InstanceConfigsPagedResponse{}),
		instanceSnapshotsName:     NewResource(&client, instanceSnapshotsName, instanceSnapshotsEndpoint, true, InstanceSnapshot{}, nil),
		instanceIPsName:           NewResource(&client, instanceIPsName, instanceIPsEndpoint, true, InstanceIP{}, nil),                           // really?
		instanceVolumesName:       NewResource(&client, instanceVolumesName, instanceVolumesEndpoint, true, nil, InstanceVolumesPagedResponse{}), // really?
		instanceStatsName:         NewResource(&client, instanceStatsName, instanceStatsEndpoint, true, InstanceStats{}, nil),
		ipaddressesName:           NewResource(&client, ipaddressesName, ipaddressesEndpoint, false, nil, IPAddressesPagedResponse{}), // really?
		ipv6poolsName:             NewResource(&client, ipv6poolsName, ipv6poolsEndpoint, false, nil, IPv6PoolsPagedResponse{}),       // really?
		ipv6rangesName:            NewResource(&client, ipv6rangesName, ipv6rangesEndpoint, false, IPv6Range{}, IPv6RangesPagedResponse{}),
		regionsName:               NewResource(&client, regionsName, regionsEndpoint, false, Region{}, RegionsPagedResponse{}),
		volumesName:               NewResource(&client, volumesName, volumesEndpoint, false, Volume{}, VolumesPagedResponse{}),
		kernelsName:               NewResource(&client, kernelsName, kernelsEndpoint, false, LinodeKernel{}, LinodeKernelsPagedResponse{}),
		typesName:                 NewResource(&client, typesName, typesEndpoint, false, LinodeType{}, LinodeTypesPagedResponse{}),
		domainsName:               NewResource(&client, domainsName, domainsEndpoint, false, Domain{}, DomainsPagedResponse{}),
		domainRecordsName:         NewResource(&client, domainRecordsName, domainRecordsEndpoint, true, DomainRecord{}, DomainRecordsPagedResponse{}),
		longviewName:              NewResource(&client, longviewName, longviewEndpoint, false, nil, nil), // really?
		longviewclientsName:       NewResource(&client, longviewclientsName, longviewclientsEndpoint, false, LongviewClient{}, LongviewClientsPagedResponse{}),
		longviewsubscriptionsName: NewResource(&client, longviewsubscriptionsName, longviewsubscriptionsEndpoint, false, LongviewSubscription{}, LongviewSubscriptionsPagedResponse{}),
		nodebalancersName:         NewResource(&client, nodebalancersName, nodebalancersEndpoint, false, NodeBalancer{}, NodeBalancerConfigsPagedResponse{}),
		nodebalancerconfigsName:   NewResource(&client, nodebalancerconfigsName, nodebalancerconfigsEndpoint, true, NodeBalancerConfig{}, NodeBalancerConfigsPagedResponse{}),
		nodebalancernodesName:     NewResource(&client, nodebalancernodesName, nodebalancernodesEndpoint, true, NodeBalancerNode{}, NodeBalancerNodesPagedResponse{}),
		notificationsName:         NewResource(&client, notificationsName, notificationsEndpoint, false, Notification{}, NotificationsPagedResponse{}),
		oauthClientsName:          NewResource(&client, oauthClientsName, oauthClientsEndpoint, false, OAuthClient{}, OAuthClientsPagedResponse{}),
		sshkeysName:               NewResource(&client, sshkeysName, sshkeysEndpoint, false, SSHKey{}, SSHKeysPagedResponse{}),
		ticketsName:               NewResource(&client, ticketsName, ticketsEndpoint, false, Ticket{}, TicketsPagedResponse{}),
		tokensName:                NewResource(&client, tokensName, tokensEndpoint, false, Token{}, TokensPagedResponse{}),
		accountName:               NewResource(&client, accountName, accountEndpoint, false, Account{}, nil),                         // really?
		accountSettingsName:       NewResource(&client, accountSettingsName, accountSettingsEndpoint, false, AccountSettings{}, nil), // really?
		eventsName:                NewResource(&client, eventsName, eventsEndpoint, false, Event{}, EventsPagedResponse{}),
		invoicesName:              NewResource(&client, invoicesName, invoicesEndpoint, false, Invoice{}, InvoicesPagedResponse{}),
		invoiceItemsName:          NewResource(&client, invoiceItemsName, invoiceItemsEndpoint, true, InvoiceItem{}, InvoiceItemsPagedResponse{}),
		profileName:               NewResource(&client, profileName, profileEndpoint, false, nil, nil), // really?
		managedName:               NewResource(&client, managedName, managedEndpoint, false, nil, nil), // really?
		tagsName:                  NewResource(&client, tagsName, tagsEndpoint, false, Tag{}, TagsPagedResponse{}),
		usersName:                 NewResource(&client, usersName, usersEndpoint, false, User{}, UsersPagedResponse{}),
		paymentsName:              NewResource(&client, paymentsName, paymentsEndpoint, false, Payment{}, PaymentsPagedResponse{}),
	}

	client.resources = resources

	client.SetDebug(envDebug)
	client.Images = resources[imagesName]
	client.StackScripts = resources[stackscriptsName]
	client.Instances = resources[instancesName]
	client.Regions = resources[regionsName]
	client.InstanceDisks = resources[instanceDisksName]
	client.InstanceConfigs = resources[instanceConfigsName]
	client.InstanceSnapshots = resources[instanceSnapshotsName]
	client.InstanceIPs = resources[instanceIPsName]
	client.InstanceVolumes = resources[instanceVolumesName]
	client.InstanceStats = resources[instanceStatsName]
	client.IPAddresses = resources[ipaddressesName]
	client.IPv6Pools = resources[ipv6poolsName]
	client.IPv6Ranges = resources[ipv6rangesName]
	client.Volumes = resources[volumesName]
	client.Kernels = resources[kernelsName]
	client.Types = resources[typesName]
	client.Domains = resources[domainsName]
	client.DomainRecords = resources[domainRecordsName]
	client.Longview = resources[longviewName]
	client.LongviewSubscriptions = resources[longviewsubscriptionsName]
	client.NodeBalancers = resources[nodebalancersName]
	client.NodeBalancerConfigs = resources[nodebalancerconfigsName]
	client.NodeBalancerNodes = resources[nodebalancernodesName]
	client.Notifications = resources[notificationsName]
	client.OAuthClients = resources[oauthClientsName]
	client.SSHKeys = resources[sshkeysName]
	client.Tickets = resources[ticketsName]
	client.Tokens = resources[tokensName]
	client.Account = resources[accountName]
	client.Events = resources[eventsName]
	client.Invoices = resources[invoicesName]
	client.Profile = resources[profileName]
	client.Managed = resources[managedName]
	client.Tags = resources[tagsName]
	client.Users = resources[usersName]
	client.Payments = resources[paymentsName]
	return
}

func copyBool(bPtr *bool) *bool {
	if bPtr == nil {
		return nil
	}
	var t = *bPtr
	return &t
}

func copyInt(iPtr *int) *int {
	if iPtr == nil {
		return nil
	}
	var t = *iPtr
	return &t
}

func copyString(sPtr *string) *string {
	if sPtr == nil {
		return nil
	}
	var t = *sPtr
	return &t
}

func copyTime(tPtr *time.Time) *time.Time {
	if tPtr == nil {
		return nil
	}
	var t = *tPtr
	return &t
}
