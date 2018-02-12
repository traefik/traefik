package egoscale

const (
	// UserAccount represents a User
	UserAccount = iota
	// AdminAccount represents an Admin
	AdminAccount
	// DomainAdminAccount represents a Domain Admin
	DomainAdminAccount
)

// AccountType represents the type of an Account
//
// http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/4.8/accounts.html#accounts-users-and-domains
type AccountType int64

// Account provides the detailed account information
type Account struct {
	ID                        string      `json:"id"`
	AccountType               AccountType `json:"accounttype,omitempty"`
	AccountDetails            string      `json:"accountdetails,omitempty"`
	CPUAvailable              string      `json:"cpuavailable,omitempty"`
	CPULimit                  string      `json:"cpulimit,omitempty"`
	CPUTotal                  int64       `json:"cputotal,omitempty"`
	DefaultZoneID             string      `json:"defaultzoneid,omitempty"`
	Domain                    string      `json:"domain,omitempty"`
	DomainID                  string      `json:"domainid,omitempty"`
	EIPLimit                  string      `json:"eiplimit,omitempty"`
	Groups                    []string    `json:"groups,omitempty"`
	IPAvailable               string      `json:"ipavailable,omitempty"`
	IPLimit                   string      `json:"iplimit,omitempty"`
	IPTotal                   int64       `json:"iptotal,omitempty"`
	IsDefault                 bool        `json:"isdefault,omitempty"`
	MemoryAvailable           string      `json:"memoryavailable,omitempty"`
	MemoryLimit               string      `json:"memorylimit,omitempty"`
	MemoryTotal               int64       `json:"memorytotal,omitempty"`
	Name                      string      `json:"name,omitempty"`
	NetworkAvailable          string      `json:"networkavailable,omitempty"`
	NetworkDomain             string      `json:"networkdomain,omitempty"`
	NetworkLimit              string      `json:"networklimit,omitempty"`
	NetworkTotal              int16       `json:"networktotal,omitempty"`
	PrimaryStorageAvailable   string      `json:"primarystorageavailable,omitempty"`
	PrimaryStorageLimit       string      `json:"primarystoragelimit,omitempty"`
	PrimaryStorageTotal       int64       `json:"primarystoragetotal,omitempty"`
	ProjectAvailable          string      `json:"projectavailable,omitempty"`
	ProjectLimit              string      `json:"projectlimit,omitempty"`
	ProjectTotal              int64       `json:"projecttotal,omitempty"`
	SecondaryStorageAvailable string      `json:"secondarystorageavailable,omitempty"`
	SecondaryStorageLimit     string      `json:"secondarystoragelimit,omitempty"`
	SecondaryStorageTotal     int64       `json:"secondarystoragetotal,omitempty"`
	SnapshotAvailable         string      `json:"snapshotavailable,omitempty"`
	SnapshotLimit             string      `json:"snapshotlimit,omitempty"`
	SnapshotTotal             int64       `json:"snapshottotal,omitempty"`
	State                     string      `json:"state,omitempty"`
	TemplateAvailable         string      `json:"templateavailable,omitempty"`
	TemplateLimit             string      `json:"templatelimit,omitempty"`
	TemplateTotal             int64       `json:"templatetotal,omitempty"`
	VMAvailable               string      `json:"vmavailable,omitempty"`
	VMLimit                   string      `json:"vmlimit,omitempty"`
	VMTotal                   int64       `json:"vmtotal,omitempty"`
	VolumeAvailable           string      `json:"volumeavailable,omitempty"`
	VolumeLimit               string      `json:"volumelimit,omitempty"`
	VolumeTotal               int64       `json:"volumetotal,omitempty"`
	VPCAvailable              string      `json:"vpcavailable,omitempty"`
	VPCLimit                  string      `json:"vpclimit,omitempty"`
	VPCTotal                  int64       `json:"vpctotal,omitempty"`
	User                      []User      `json:"user,omitempty"`
}

// ListAccounts represents a query to display the accounts
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAccounts.html
type ListAccounts struct {
	AccountType       int64  `json:"accounttype,omitempty"`
	DomainID          string `json:"domainid,omitempty"`
	ID                string `json:"id,omitempty"`
	IsCleanUpRequired bool   `json:"iscleanuprequired,omitempty"`
	IsRecursive       bool   `json:"isrecursive,omitempty"`
	Keyword           string `json:"keyword,omitempty"`
	ListAll           bool   `json:"listall,omitempty"`
	Page              int    `json:"page,omitempty"`
	PageSize          int    `json:"pagesize,omitempty"`
	State             string `json:"state,omitempty"`
}

func (*ListAccounts) name() string {
	return "listAccounts"
}

func (*ListAccounts) response() interface{} {
	return new(ListAccountsResponse)
}

// ListAccountsResponse represents a list of accounts
type ListAccountsResponse struct {
	Count   int       `json:"count"`
	Account []Account `json:"account"`
}
