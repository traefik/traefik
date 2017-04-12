package egoscale

import (
	"encoding/json"
	"net/http"
)

type Client struct {
	client    *http.Client
	endpoint  string
	apiKey    string
	apiSecret string
}

type Error struct {
	ErrorCode   int    `json:"errorcode"`
	CSErrorCode int    `json:"cserrorcode"`
	ErrorText   string `json:"errortext"`
}

type StandardResponse struct {
	Success     string `json:"success"`
	DisplayText string `json:"displaytext"`
}

type Topology struct {
	Zones          map[string]string
	Images         map[string]map[int]string
	Profiles       map[string]string
	Keypairs       []string
	SecurityGroups map[string]string
	AffinityGroups map[string]string
}

type SecurityGroupRule struct {
	Cidr            string
	IcmpType        int
	IcmpCode        int
	Port            int
	Protocol        string
	SecurityGroupId string
        UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
}

type UserSecurityGroup struct {
        Group    string   `json:"group,omitempty"`
        Account  string   `json:"account,omitempty"`
}

type MachineProfile struct {
	Name            string
	SecurityGroups  []string
	Keypair         string
	Userdata        string
	ServiceOffering string
	Template        string
	Zone            string
	AffinityGroups  []string
}

type ListZonesResponse struct {
	Count int     `json:"count"`
	Zones []*Zone `json:"zone"`
}

type Zone struct {
	Allocationstate       string            `json:"allocationstate,omitempty"`
	Description           string            `json:"description,omitempty"`
	Displaytext           string            `json:"displaytext,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	Domainid              string            `json:"domainid,omitempty"`
	Domainname            string            `json:"domainname,omitempty"`
	Id                    string            `json:"id,omitempty"`
	Internaldns1          string            `json:"internaldns1,omitempty"`
	Internaldns2          string            `json:"internaldns2,omitempty"`
	Ip6dns1               string            `json:"ip6dns1,omitempty"`
	Ip6dns2               string            `json:"ip6dns2,omitempty"`
	Localstorageenabled   bool              `json:"localstorageenabled,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Networktype           string            `json:"networktype,omitempty"`
	Resourcedetails       map[string]string `json:"resourcedetails,omitempty"`
	Securitygroupsenabled bool              `json:"securitygroupsenabled,omitempty"`
	Vlan                  string            `json:"vlan,omitempty"`
	Zonetoken             string            `json:"zonetoken,omitempty"`
}

type ListServiceOfferingsResponse struct {
	Count            int                `json:"count"`
	ServiceOfferings []*ServiceOffering `json:"serviceoffering"`
}

type ServiceOffering struct {
	Cpunumber              int               `json:"cpunumber,omitempty"`
	Cpuspeed               int               `json:"cpuspeed,omitempty"`
	Displaytext            string            `json:"displaytext,omitempty"`
	Domain                 string            `json:"domain,omitempty"`
	Domainid               string            `json:"domainid,omitempty"`
	Hosttags               string            `json:"hosttags,omitempty"`
	Id                     string            `json:"id,omitempty"`
	Iscustomized           bool              `json:"iscustomized,omitempty"`
	Issystem               bool              `json:"issystem,omitempty"`
	Isvolatile             bool              `json:"isvolatile,omitempty"`
	Memory                 int               `json:"memory,omitempty"`
	Name                   string            `json:"name,omitempty"`
	Networkrate            int               `json:"networkrate,omitempty"`
	Serviceofferingdetails map[string]string `json:"serviceofferingdetails,omitempty"`
}

type ListTemplatesResponse struct {
	Count     int         `json:"count"`
	Templates []*Template `json:"template"`
}

type Template struct {
	Account               string            `json:"account,omitempty"`
	Accountid             string            `json:"accountid,omitempty"`
	Bootable              bool              `json:"bootable,omitempty"`
	Checksum              string            `json:"checksum,omitempty"`
	Created               string            `json:"created,omitempty"`
	CrossZones            bool              `json:"crossZones,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	Displaytext           string            `json:"displaytext,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	Domainid              string            `json:"domainid,omitempty"`
	Format                string            `json:"format,omitempty"`
	Hostid                string            `json:"hostid,omitempty"`
	Hostname              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	Id                    string            `json:"id,omitempty"`
	Isdynamicallyscalable bool              `json:"isdynamicallyscalable,omitempty"`
	Isextractable         bool              `json:"isextractable,omitempty"`
	Isfeatured            bool              `json:"isfeatured,omitempty"`
	Ispublic              bool              `json:"ispublic,omitempty"`
	Isready               bool              `json:"isready,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Ostypeid              string            `json:"ostypeid,omitempty"`
	Ostypename            string            `json:"ostypename,omitempty"`
	Passwordenabled       bool              `json:"passwordenabled,omitempty"`
	Project               string            `json:"project,omitempty"`
	Projectid             string            `json:"projectid,omitempty"`
	Removed               string            `json:"removed,omitempty"`
	Size                  int64             `json:"size,omitempty"`
	Sourcetemplateid      string            `json:"sourcetemplateid,omitempty"`
	Sshkeyenabled         bool              `json:"sshkeyenabled,omitempty"`
	Status                string            `json:"status,omitempty"`
	Zoneid                string            `json:"zoneid,omitempty"`
	Zonename              string            `json:"zonename,omitempty"`
}

type ListSSHKeyPairsResponse struct {
	Count       int           `json:"count"`
	SSHKeyPairs []*SSHKeyPair `json:"sshkeypair"`
}

type SSHKeyPair struct {
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
}

type ListAffinityGroupsResponse struct {
	Count          int              `json:"count"`
	AffinityGroups []*AffinityGroup `json:"affinitygroup"`
}

type AffinityGroup struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Id          string `json:"id,omitempty"`
	Domainid    string `json:"domainid,omitempty"`
	Domain      string `json:"domain,omitempty"`
	Account     string `json:"account,omitempty"`
}

type CreateAffinityGroupResponseWrapper struct {
	Wrapped AffinityGroup `json:"affinitygroup"`
}

type ListSecurityGroupsResponse struct {
	Count          int              `json:"count"`
	SecurityGroups []*SecurityGroup `json:"securitygroup"`
}

type SecurityGroup struct {
	Account      string `json:"account,omitempty"`
	Description  string `json:"description,omitempty"`
	Domain       string `json:"domain,omitempty"`
	Domainid     string `json:"domainid,omitempty"`
	Id           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Project      string `json:"project,omitempty"`
	Projectid    string `json:"projectid,omitempty"`
	IngressRules []struct {
		RuleId    string   `json:"ruleid,omitempty"`
		Protocol  string   `json:"protocol,omitempty"`
		StartPort int      `json:"startport,omitempty"`
		EndPort   int      `json:"endport,omitempty"`
		Cidr      string   `json:"cidr,omitempty"`
		IcmpCode  int      `json:"icmpcode,omitempty"`
		IcmpType  int      `json:"icmptype,omitempty"`
		Tags      []string `json:"tags,omitempty"`
	} `json:"ingressrule,omitempty"`
	EgressRules []struct {
		RuleId    string   `json:"ruleid,omitempty"`
		Protocol  string   `json:"protocol,omitempty"`
		StartPort int      `json:"startport,omitempty"`
		EndPort   int      `json:"endport,omitempty"`
		Cidr      string   `json:"cidr,omitempty"`
		IcmpCode  int      `json:"icmpcode,omitempty"`
		IcmpType  int      `json:"icmptype,omitempty"`
		Tags      []string `json:"tags,omitempty"`
	} `json:"egressrule,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

type CreateSecurityGroupResponseWrapper struct {
	Wrapped CreateSecurityGroupResponse `json:"securitygroup"`
}
type CreateSecurityGroupResponse struct {
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain,omitempty"`
	Domainid    string `json:"domainid,omitempty"`
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Project     string `json:"project,omitempty"`
	Projectid   string `json:"projectid,omitempty"`
}

type AuthorizeSecurityGroupIngressResponse struct {
	JobID             string `json:"jobid,omitempty"`
	Account           string `json:"account,omitempty"`
	Cidr              string `json:"cidr,omitempty"`
	Endport           int    `json:"endport,omitempty"`
	Icmpcode          int    `json:"icmpcode,omitempty"`
	Icmptype          int    `json:"icmptype,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
	Ruleid            string `json:"ruleid,omitempty"`
	Securitygroupname string `json:"securitygroupname,omitempty"`
	Startport         int    `json:"startport,omitempty"`
}

type AuthorizeSecurityGroupEgressResponse struct {
	JobID             string `json:"jobid,omitempty"`
	Account           string `json:"account,omitempty"`
	Cidr              string `json:"cidr,omitempty"`
	Endport           int    `json:"endport,omitempty"`
	Icmpcode          int    `json:"icmpcode,omitempty"`
	Icmptype          int    `json:"icmptype,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
	Ruleid            string `json:"ruleid,omitempty"`
	Securitygroupname string `json:"securitygroupname,omitempty"`
	Startport         int    `json:"startport,omitempty"`
}

type DeployVirtualMachineWrappedResponse struct {
	Wrapped DeployVirtualMachineResponse `json:"virtualmachine"`
}

type DeployVirtualMachineResponse struct {
	JobID         string `json:"jobid,omitempty"`
	Account       string `json:"account,omitempty"`
	Affinitygroup []struct {
		Account           string   `json:"account,omitempty"`
		Description       string   `json:"description,omitempty"`
		Domain            string   `json:"domain,omitempty"`
		Domainid          string   `json:"domainid,omitempty"`
		Id                string   `json:"id,omitempty"`
		Name              string   `json:"name,omitempty"`
		Type              string   `json:"type,omitempty"`
		VirtualmachineIds []string `json:"virtualmachineIds,omitempty"`
	} `json:"affinitygroup,omitempty"`
	Cpunumber             int               `json:"cpunumber,omitempty"`
	Cpuspeed              int               `json:"cpuspeed,omitempty"`
	Cpuused               string            `json:"cpuused,omitempty"`
	Created               string            `json:"created,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	Diskioread            int64             `json:"diskioread,omitempty"`
	Diskiowrite           int64             `json:"diskiowrite,omitempty"`
	Diskkbsread           int64             `json:"diskkbsread,omitempty"`
	Diskkbswrite          int64             `json:"diskkbswrite,omitempty"`
	Displayname           string            `json:"displayname,omitempty"`
	Displayvm             bool              `json:"displayvm,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	Domainid              string            `json:"domainid,omitempty"`
	Forvirtualnetwork     bool              `json:"forvirtualnetwork,omitempty"`
	Group                 string            `json:"group,omitempty"`
	Groupid               string            `json:"groupid,omitempty"`
	Guestosid             string            `json:"guestosid,omitempty"`
	Haenable              bool              `json:"haenable,omitempty"`
	Hostid                string            `json:"hostid,omitempty"`
	Hostname              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	Id                    string            `json:"id,omitempty"`
	Instancename          string            `json:"instancename,omitempty"`
	Isdynamicallyscalable bool              `json:"isdynamicallyscalable,omitempty"`
	Isodisplaytext        string            `json:"isodisplaytext,omitempty"`
	Isoid                 string            `json:"isoid,omitempty"`
	Isoname               string            `json:"isoname,omitempty"`
	Keypair               string            `json:"keypair,omitempty"`
	Memory                int               `json:"memory,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Networkkbsread        int64             `json:"networkkbsread,omitempty"`
	Networkkbswrite       int64             `json:"networkkbswrite,omitempty"`
	Nic                   []struct {
		Broadcasturi string   `json:"broadcasturi,omitempty"`
		Gateway      string   `json:"gateway,omitempty"`
		Id           string   `json:"id,omitempty"`
		Ipaddress    string   `json:"ipaddress,omitempty"`
		Isdefault    bool     `json:"isdefault,omitempty"`
		Isolationuri string   `json:"isolationuri,omitempty"`
		Macaddress   string   `json:"macaddress,omitempty"`
		Netmask      string   `json:"netmask,omitempty"`
		Networkid    string   `json:"networkid,omitempty"`
		Networkname  string   `json:"networkname,omitempty"`
		Secondaryip  []string `json:"secondaryip,omitempty"`
		Traffictype  string   `json:"traffictype,omitempty"`
		Type         string   `json:"type,omitempty"`
	} `json:"nic,omitempty"`
	Password            string `json:"password,omitempty"`
	Passwordenabled     bool   `json:"passwordenabled,omitempty"`
	Project             string `json:"project,omitempty"`
	Projectid           string `json:"projectid,omitempty"`
	Publicip            string `json:"publicip,omitempty"`
	Publicipid          string `json:"publicipid,omitempty"`
	Rootdeviceid        int64  `json:"rootdeviceid,omitempty"`
	Rootdevicetype      string `json:"rootdevicetype,omitempty"`
	Serviceofferingid   string `json:"serviceofferingid,omitempty"`
	Serviceofferingname string `json:"serviceofferingname,omitempty"`
	Servicestate        string `json:"servicestate,omitempty"`
	State               string `json:"state,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type QueryAsyncJobResultResponse struct {
	Accountid       string          `json:"accountid,omitempty"`
	Cmd             string          `json:"cmd,omitempty"`
	Created         string          `json:"created,omitempty"`
	Jobinstanceid   string          `json:"jobinstanceid,omitempty"`
	Jobinstancetype string          `json:"jobinstancetype,omitempty"`
	Jobprocstatus   int             `json:"jobprocstatus,omitempty"`
	Jobresult       json.RawMessage `json:"jobresult,omitempty"`
	Jobresultcode   int             `json:"jobresultcode,omitempty"`
	Jobresulttype   string          `json:"jobresulttype,omitempty"`
	Jobstatus       int             `json:"jobstatus,omitempty"`
	Userid          string          `json:"userid,omitempty"`
}

type ListVirtualMachinesResponse struct {
	Count           int               `json:"count"`
	VirtualMachines []*VirtualMachine `json:"virtualmachine"`
}

type VirtualMachine struct {
	Account               string            `json:"account,omitempty"`
	Cpunumber             int               `json:"cpunumber,omitempty"`
	Cpuspeed              int               `json:"cpuspeed,omitempty"`
	Cpuused               string            `json:"cpuused,omitempty"`
	Created               string            `json:"created,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	Diskioread            int64             `json:"diskioread,omitempty"`
	Diskiowrite           int64             `json:"diskiowrite,omitempty"`
	Diskkbsread           int64             `json:"diskkbsread,omitempty"`
	Diskkbswrite          int64             `json:"diskkbswrite,omitempty"`
	Displayname           string            `json:"displayname,omitempty"`
	Displayvm             bool              `json:"displayvm,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	Domainid              string            `json:"domainid,omitempty"`
	Forvirtualnetwork     bool              `json:"forvirtualnetwork,omitempty"`
	Group                 string            `json:"group,omitempty"`
	Groupid               string            `json:"groupid,omitempty"`
	Guestosid             string            `json:"guestosid,omitempty"`
	Haenable              bool              `json:"haenable,omitempty"`
	Hostid                string            `json:"hostid,omitempty"`
	Hostname              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	Id                    string            `json:"id,omitempty"`
	Instancename          string            `json:"instancename,omitempty"`
	Isdynamicallyscalable bool              `json:"isdynamicallyscalable,omitempty"`
	Isodisplaytext        string            `json:"isodisplaytext,omitempty"`
	Isoid                 string            `json:"isoid,omitempty"`
	Isoname               string            `json:"isoname,omitempty"`
	Keypair               string            `json:"keypair,omitempty"`
	Memory                int               `json:"memory,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Networkkbsread        int64             `json:"networkkbsread,omitempty"`
	Networkkbswrite       int64             `json:"networkkbswrite,omitempty"`
	Nic                   []struct {
		Broadcasturi string   `json:"broadcasturi,omitempty"`
		Gateway      string   `json:"gateway,omitempty"`
		Id           string   `json:"id,omitempty"`
		Ip6address   string   `json:"ip6address,omitempty"`
		Ip6cidr      string   `json:"ip6cidr,omitempty"`
		Ip6gateway   string   `json:"ip6gateway,omitempty"`
		Ipaddress    string   `json:"ipaddress,omitempty"`
		Isdefault    bool     `json:"isdefault,omitempty"`
		Isolationuri string   `json:"isolationuri,omitempty"`
		Macaddress   string   `json:"macaddress,omitempty"`
		Netmask      string   `json:"netmask,omitempty"`
		Networkid    string   `json:"networkid,omitempty"`
		Networkname  string   `json:"networkname,omitempty"`
		Secondaryip  []struct {
			Id		string `json:"id,omitempty"`
			IpAddress	string `json:"ipaddress,omitempty"`
		} `json:"secondaryip,omitempty"`
		Traffictype  string   `json:"traffictype,omitempty"`
		Type         string   `json:"type,omitempty"`
	} `json:"nic,omitempty"`
	Password            string `json:"password,omitempty"`
	Passwordenabled     bool   `json:"passwordenabled,omitempty"`
	Project             string `json:"project,omitempty"`
	Projectid           string `json:"projectid,omitempty"`
	Publicip            string `json:"publicip,omitempty"`
	Publicipid          string `json:"publicipid,omitempty"`
	Rootdeviceid        int64  `json:"rootdeviceid,omitempty"`
	Rootdevicetype      string `json:"rootdevicetype,omitempty"`
	SecurityGroups      []struct {
		Account     string `json:"account,omitempty"`
		Description string `json:"description,omitempty"`
		Id          string `json:"id,omitempty"`
		Name        string `json:"name,omitemtpy"`
		Tags        []string `json:"tags,omitempty"`
	} `json:"securitygroup,omitempty"`
	Serviceofferingid   string `json:"serviceofferingid,omitempty"`
	Serviceofferingname string `json:"serviceofferingname,omitempty"`
	Servicestate        string `json:"servicestate,omitempty"`
	State               string `json:"state,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type StartVirtualMachineResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type StopVirtualMachineResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type DestroyVirtualMachineResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type RebootVirtualMachineResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type CreateSSHKeyPairWrappedResponse struct {
	Wrapped CreateSSHKeyPairResponse `json:"keypair,omitempty"`
}

type CreateSSHKeyPairResponse struct {
	Privatekey string `json:"privatekey,omitempty"`
}

type RemoveIpFromNicResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type AddIpToNicResponse struct {
	Id string `json:"id"`
	IpAddress string `json:"ipaddress"`
	NetworkId string `json:"networkid"`
	NicId string `json:"nicid"`
	VmId string `json:"virtualmachineid"`
}

type CreateAffinityGroupResponse struct {
	JobId string `json:"jobid,omitempty"`
}

type DeleteAffinityGroupResponse struct {
	JobId string `json:"jobid,omitempty"`
}

type DeleteSSHKeyPairResponse struct {
	Privatekey string `json:"privatekey,omitempty"`
}

type DNSDomain struct {
	Id             int64  `json:"id"`
	UserId         int64  `json:"user_id"`
	RegistrantId   int64  `json:"registrant_id,omitempty"`
	Name           string `json:"name"`
	UnicodeName    string `json:"unicode_name"`
	Token          string `json:"token"`
	State          string `json:"state"`
	Language       string `json:"language,omitempty"`
	Lockable       bool   `json:"lockable"`
	AutoRenew      bool   `json:"auto_renew"`
	WhoisProtected bool   `json:"whois_protected"`
	RecordCount    int64  `json:"record_count"`
	ServiceCount   int64  `json:"service_count"`
	ExpiresOn      string `json:"expires_on,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type DNSDomainCreateRequest struct {
	Domain struct {
		Name string `json:"name"`
	} `json:"domain"`
}

type DNSRecord struct {
	Id         int64  `json:"id,omitempty"`
	DomainId   int64  `json:"domain_id,omitempty"`
	Name       string `json:"name"`
	Ttl        int    `json:"ttl,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Content    string `json:"content"`
	RecordType string `json:"record_type"`
	Prio       int    `json:"prio,omitempty"`
}

type DNSRecordResponse struct {
	Record DNSRecord `json:"record"`
}

type DNSError struct {
	Name []string `json:"name"`
}
