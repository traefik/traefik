package domain

import (
	"fmt"
	"net"

	"github.com/transip/gotransip"
	"github.com/transip/gotransip/util"
)

const (
	serviceName    string = "DomainService"
	dnsServiceName string = "DnsService"
)

// Domain represents a Transip_Domain object
// as described at https://api.transip.nl/docs/transip.nl/class-Transip_Domain.html
type Domain struct {
	Name              string         `xml:"name"`
	Nameservers       []Nameserver   `xml:"nameservers>item"`
	Contacts          []WhoisContact `xml:"contacts>item"`
	DNSEntries        []DNSEntry     `xml:"dnsEntries>item"`
	Branding          Branding       `xml:"branding"`
	AuthorizationCode string         `xml:"authCode"`
	IsLocked          bool           `xml:"isLocked"`
	RegistrationDate  util.XMLTime   `xml:"registrationDate"`
	RenewalDate       util.XMLTime   `xml:"renewalDate"`
}

// EncodeParams returns Domain parameters ready to be used for constructing a signature
// the order of parameters added here has to match the order in the WSDL
// as described at http://api.transip.nl/wsdl/?service=DomainService
func (d Domain) EncodeParams(prm gotransip.ParamsContainer, prefix string) {
	if len(prefix) == 0 {
		prefix = fmt.Sprintf("%d", prm.Len())
	}

	prm.Add(fmt.Sprintf("%s[name]", prefix), d.Name)

	// nameservers
	Nameservers(d.Nameservers).EncodeParams(prm, fmt.Sprintf("%s[nameservers]", prefix))

	// contacts
	WhoisContacts(d.Contacts).EncodeParams(prm, fmt.Sprintf("%s[contacts]", prefix))

	// dnsEntries
	DNSEntries(d.DNSEntries).EncodeParams(prm, fmt.Sprintf("%s[dnsEntries]", prefix))

	// branding
	d.Branding.EncodeParams(prm, fmt.Sprintf("%s[branding]", prefix))

	prm.Add(fmt.Sprintf("%s[authCode]", prefix), d.AuthorizationCode)
	prm.Add(fmt.Sprintf("%s[isLocked]", prefix), d.IsLocked)
	prm.Add(fmt.Sprintf("%s[registrationDate]", prefix), d.RegistrationDate.Format("2006-01-02"))
	prm.Add(fmt.Sprintf("%s[renewalDate]", prefix), d.RenewalDate.Format("2006-01-02"))
}

// EncodeArgs returns Domain XML body ready to be passed in the SOAP call
func (d Domain) EncodeArgs(key string) string {
	output := fmt.Sprintf(`<%s xsi:type="ns1:Domain">
	<name xsi:type="xsd:string">%s</name>`, key, d.Name) + "\n"

	output += Nameservers(d.Nameservers).EncodeArgs("nameservers") + "\n"
	output += WhoisContacts(d.Contacts).EncodeArgs("contacts") + "\n"
	output += DNSEntries(d.DNSEntries).EncodeArgs("dnsEntries") + "\n"
	output += d.Branding.EncodeArgs("branding") + "\n"

	output += fmt.Sprintf(`	<authCode xsi:type="xsd:string">%s</authCode>
	<isLocked xsi:type="xsd:boolean">%t</isLocked>
	<registrationDate xsi:type="xsd:string">%s</registrationDate>
	<renewalDate xsi:type="xsd:string">%s</renewalDate>`,
		d.AuthorizationCode, d.IsLocked,
		d.RegistrationDate.Format("2006-01-02"), d.RenewalDate.Format("2006-01-02"),
	) + "\n"

	return fmt.Sprintf("%s</%s>", output, key)
}

// Capability represents the possible capabilities a TLD can have
type Capability string

var (
	// CapabilityRequiresAuthCode defines this TLD requires an auth code
	// to be transferred
	CapabilityRequiresAuthCode Capability = "requiresAuthCode"
	// CapabilityCanRegister defines this TLD can be registered
	CapabilityCanRegister Capability = "canRegister"
	// CapabilityCanTransferWithOwnerChange defines this TLD can be transferred
	// with change of ownership
	CapabilityCanTransferWithOwnerChange Capability = "canTransferWithOwnerChange"
	// CapabilityCanTransferWithoutOwnerChange defines this TLD can be
	// transferred without change of ownership
	CapabilityCanTransferWithoutOwnerChange Capability = "canTransferWithoutOwnerChange"
	// CapabilityCanSetLock defines this TLD allows to be locked
	CapabilityCanSetLock Capability = "canSetLock"
	// CapabilityCanSetOwner defines this TLD supports setting an owner
	CapabilityCanSetOwner Capability = "canSetOwner"
	// CapabilityCanSetContacts defines this TLD supports setting contacts
	CapabilityCanSetContacts Capability = "canSetContacts"
	// CapabilityCanSetNameservers defines this TLD supports setting nameservers
	CapabilityCanSetNameservers Capability = "canSetNameservers"
)

// TLD represents a Transip_Tld object as described at
// https://api.transip.nl/docs/transip.nl/class-Transip_Tld.html
type TLD struct {
	Name                     string       `xml:"name"`
	Price                    float64      `xml:"price"`
	RenewalPrice             float64      `xml:"renewalPrice"`
	Capabilities             []Capability `xml:"capabilities>item"`
	RegistrationPeriodLength int64        `xml:"registrationPeriodLength"`
	CancelTimeFrame          int64        `xml:"cancelTimeFrame"`
}

// Nameserver represents a Transip_Nameserver object as described at
// https://api.transip.nl/docs/transip.nl/class-Transip_Nameserver.html
type Nameserver struct {
	Hostname    string `xml:"hostname"`
	IPv4Address net.IP `xml:"ipv4"`
	IPv6Address net.IP `xml:"ipv6"`
}

// Nameservers is just an array of Nameserver
// basically only here so it can implement paramsEncoder
type Nameservers []Nameserver

// EncodeParams returns Nameservers parameters ready to be used for constructing a signature
// the order of parameters added here has to match the order in the WSDL
// as described at http://api.transip.nl/wsdl/?service=DomainService
func (n Nameservers) EncodeParams(prm gotransip.ParamsContainer, prefix string) {
	if len(n) == 0 {
		prm.Add("anything", nil)
		return
	}

	if len(prefix) == 0 {
		prefix = fmt.Sprintf("%d", prm.Len())
	}

	for i, e := range n {
		var ipv4, ipv6 string
		if e.IPv4Address != nil {
			ipv4 = e.IPv4Address.String()
		}
		if e.IPv6Address != nil {
			ipv6 = e.IPv6Address.String()
		}
		prm.Add(fmt.Sprintf("%s[%d][hostname]", prefix, i), e.Hostname)
		prm.Add(fmt.Sprintf("%s[%d][ipv4]", prefix, i), ipv4)
		prm.Add(fmt.Sprintf("%s[%d][ipv6]", prefix, i), ipv6)
	}
}

// EncodeArgs returns Nameservers XML body ready to be passed in the SOAP call
func (n Nameservers) EncodeArgs(key string) string {
	output := fmt.Sprintf(`<%s SOAP-ENC:arrayType="ns1:Nameserver[%d]" xsi:type="ns1:ArrayOfNameserver">`, key, len(n)) + "\n"
	for _, e := range n {
		var ipv4, ipv6 string
		if e.IPv4Address != nil {
			ipv4 = e.IPv4Address.String()
		}
		if e.IPv6Address != nil {
			ipv6 = e.IPv6Address.String()
		}
		output += fmt.Sprintf(`	<item xsi:type="ns1:Nameserver">
		<hostname xsi:type="xsd:string">%s</hostname>
		<ipv4 xsi:type="xsd:string">%s</ipv4>
		<ipv6 xsi:type="xsd:string">%s</ipv6>
	</item>`, e.Hostname, ipv4, ipv6) + "\n"
	}

	return fmt.Sprintf("%s</%s>", output, key)
}

// DNSEntryType represents the possible types of DNS entries
type DNSEntryType string

var (
	// DNSEntryTypeA represents an A-record
	DNSEntryTypeA DNSEntryType = "A"
	// DNSEntryTypeAAAA represents an AAAA-record
	DNSEntryTypeAAAA DNSEntryType = "AAAA"
	// DNSEntryTypeCNAME represents a CNAME-record
	DNSEntryTypeCNAME DNSEntryType = "CNAME"
	// DNSEntryTypeMX represents an MX-record
	DNSEntryTypeMX DNSEntryType = "MX"
	// DNSEntryTypeNS represents an NS-record
	DNSEntryTypeNS DNSEntryType = "NS"
	// DNSEntryTypeTXT represents a TXT-record
	DNSEntryTypeTXT DNSEntryType = "TXT"
	// DNSEntryTypeSRV represents an SRV-record
	DNSEntryTypeSRV DNSEntryType = "SRV"
)

// DNSEntry represents a Transip_DnsEntry object as described at
// https://api.transip.nl/docs/transip.nl/class-Transip_DnsEntry.html
type DNSEntry struct {
	Name    string       `xml:"name"`
	TTL     int64        `xml:"expire"`
	Type    DNSEntryType `xml:"type"`
	Content string       `xml:"content"`
}

// DNSEntries is just an array of DNSEntry
// basically only here so it can implement paramsEncoder
type DNSEntries []DNSEntry

// EncodeParams returns DNSEntries parameters ready to be used for constructing a signature
// the order of parameters added here has to match the order in the WSDL
// as described at http://api.transip.nl/wsdl/?service=DomainService
func (d DNSEntries) EncodeParams(prm gotransip.ParamsContainer, prefix string) {
	if len(d) == 0 {
		prm.Add("anything", nil)
		return
	}

	if len(prefix) == 0 {
		prefix = fmt.Sprintf("%d", prm.Len())
	}

	for i, e := range d {
		prm.Add(fmt.Sprintf("%s[%d][name]", prefix, i), e.Name)
		prm.Add(fmt.Sprintf("%s[%d][expire]", prefix, i), fmt.Sprintf("%d", e.TTL))
		prm.Add(fmt.Sprintf("%s[%d][type]", prefix, i), string(e.Type))
		prm.Add(fmt.Sprintf("%s[%d][content]", prefix, i), e.Content)
	}
}

// EncodeArgs returns DNSEntries XML body ready to be passed in the SOAP call
func (d DNSEntries) EncodeArgs(key string) string {
	output := fmt.Sprintf(`<%s SOAP-ENC:arrayType="ns1:DnsEntry[%d]" xsi:type="ns1:ArrayOfDnsEntry">`, key, len(d)) + "\n"
	for _, e := range d {
		output += fmt.Sprintf(`	<item xsi:type="ns1:DnsEntry">
		<name xsi:type="xsd:string">%s</name>
		<expire xsi:type="xsd:int">%d</expire>
		<type xsi:type="xsd:string">%s</type>
		<content xsi:type="xsd:string">%s</content>
	</item>`, e.Name, e.TTL, e.Type, e.Content) + "\n"
	}

	return fmt.Sprintf("%s</%s>", output, key)
}

// DNSSecAlgorithm represents the possible types of DNSSec algorithms
type DNSSecAlgorithm int

const (
	// DNSSecAlgorithmDSA represents DSA
	DNSSecAlgorithmDSA DNSSecAlgorithm = iota + 3
	_
	// DNSSecAlgorithmRSASHA1 represents RSASHA1
	DNSSecAlgorithmRSASHA1
	// DNSSecAlgorithmDSANSEC3SHA1 represents DSANSEC3SHA1
	DNSSecAlgorithmDSANSEC3SHA1
	// DNSSecAlgorithmRSASHA1NSEC3SHA1 represents RSASHA1NSEC3SHA1
	DNSSecAlgorithmRSASHA1NSEC3SHA1
	// DNSSecAlgorithmRSASHA256 represents RSASHA256
	DNSSecAlgorithmRSASHA256
	// DNSSecAlgorithmRSASHA512 represents RSASHA512
	DNSSecAlgorithmRSASHA512 DNSSecAlgorithm = iota + 4
	_
	// DNSSecAlgorithmECCGOST represents ECCGOST
	DNSSecAlgorithmECCGOST
	// DNSSecAlgorithmECDSAP256SHA256 represents ECDSAP256SHA256
	DNSSecAlgorithmECDSAP256SHA256
	// DNSSecAlgorithmECDSAP384SHA384 represents ECDSAP384SHA384
	DNSSecAlgorithmECDSAP384SHA384
	// DNSSecAlgorithmED25519 represents ED25519
	DNSSecAlgorithmED25519
	// DNSSecAlgorithmED448 represents ED448
	DNSSecAlgorithmED448
)

// DNSSecFlag represents the possible types of DNSSec flags
type DNSSecFlag int

const (
	// DNSSecFlagNone means no flag is set
	DNSSecFlagNone DNSSecFlag = 0
	// DNSSecFlagZSK means this is a Zone Signing Key
	DNSSecFlagZSK DNSSecFlag = 256
	// DNSSecFlagKSK means this is a Key Signing Key
	DNSSecFlagKSK DNSSecFlag = 257
)

// DNSSecEntry represents a Transip_DnsSecEntry object as described at
// https://api.transip.nl/docs/transip.nl/class-Transip_DnsSecEntry.html
type DNSSecEntry struct {
	KeyTag    int             `xml:"keyTag"`
	Flags     DNSSecFlag      `xml:"flags"`
	Algorithm DNSSecAlgorithm `xml:"algorithm"`
	PublicKey string          `xml:"publicKey"`
}

// DNSSecEntries is just an array of DNSSecEntry
// basically only here so it can implement paramsEncoder
type DNSSecEntries []DNSSecEntry

// EncodeParams returns DNSSecEntries parameters ready to be used for constructing
// a signature
// the order of parameters added here has to match the order in the WSDL as
// described at http://api.transip.nl/wsdl/?service=DnsService
func (d DNSSecEntries) EncodeParams(prm gotransip.ParamsContainer, prefix string) {
	if len(d) == 0 {
		prm.Add("anything", nil)
		return
	}

	if len(prefix) == 0 {
		prefix = fmt.Sprintf("%d", prm.Len())
	}

	for i, e := range d {
		prm.Add(fmt.Sprintf("%s[%d][keyTag]", prefix, i), fmt.Sprintf("%d", e.KeyTag))
		prm.Add(fmt.Sprintf("%s[%d][flags]", prefix, i), fmt.Sprintf("%d", e.Flags))
		prm.Add(fmt.Sprintf("%s[%d][algorithm]", prefix, i), fmt.Sprintf("%d", e.Algorithm))
		prm.Add(fmt.Sprintf("%s[%d][publicKey]", prefix, i), e.PublicKey)
	}
}

// EncodeArgs returns Entries XML body ready to be passed in the SOAP call
func (d DNSSecEntries) EncodeArgs(key string) string {
	output := fmt.Sprintf(`<%s SOAP-ENC:arrayType="ns1:DnsSecEntry[%d]" xsi:type="ns1:ArrayOfDnsSecEntry">`, key, len(d)) + "\n"
	for _, e := range d {
		output += fmt.Sprintf(`	<item xsi:type="ns1:DnsSecEntry">
		<keyTag xsi:type="xsd:int">%d</keyTag>
		<flags xsi:type="xsd:int">%d</flags>
		<algorithm xsi:type="xsd:int">%d</algorithm>
		<publicKey xsi:type="xsd:string">%s</publicKey>
	</item>`, e.KeyTag, e.Flags, e.Algorithm, e.PublicKey) + "\n"
	}

	return fmt.Sprintf("%s</%s>", output, key)
}

// Status reflects the current status of a domain in a check result
type Status string

var (
	// StatusInYourAccount means he domain name is already in your account
	StatusInYourAccount Status = "inyouraccount"
	// StatusUnavailable means the domain name is currently unavailable and can not be registered due to unknown reasons.
	StatusUnavailable Status = "unavailable"
	// StatusNotFree means the domain name has already been registered
	StatusNotFree Status = "notfree"
	// StatusFree means the domain name is currently free, is available and can be registered
	StatusFree Status = "free"
	// StatusInternalPull means the domain name is currently registered at TransIP and is available to be pulled from another account to yours.
	StatusInternalPull Status = "internalpull"
	// StatusInternalPush means the domain name is currently registered at TransIP in your accounta and is available to be pushed to another account.
	StatusInternalPush Status = "internalpush"
)

// CheckResult represents a Transip_DomainCheckResult object as described at
// https://api.transip.nl/docs/transip.nl/class-Transip_DomainCheckResult.html
type CheckResult struct {
	DomainName string   `xml:"domainName"`
	Status     Status   `xml:"status"`
	Actions    []Action `xml:"actions>item"`
}

// Branding represents a Transip_DomainBranding object as described at
// https://api.transip.nl/docs/transip.nl/class-Transip_DomainBranding.html
type Branding struct {
	CompanyName     string `xml:"companyName"`
	SupportEmail    string `xml:"supportEmail"`
	CompanyURL      string `xml:"companyUrl"`
	TermsOfUsageURL string `xml:"termsOfUsageUrl"`
	BannerLine1     string `xml:"bannerLine1"`
	BannerLine2     string `xml:"bannerLine2"`
	BannerLine3     string `xml:"bannerLine3"`
}

// EncodeParams returns WhoisContacts parameters ready to be used for constructing a signature
// the order of parameters added here has to match the order in the WSDL
// as described at http://api.transip.nl/wsdl/?service=DomainService
func (b Branding) EncodeParams(prm gotransip.ParamsContainer, prefix string) {
	if len(prefix) == 0 {
		prefix = fmt.Sprintf("%d", prm.Len())
	}

	prm.Add(fmt.Sprintf("%s[companyName]", prefix), b.CompanyName)
	prm.Add(fmt.Sprintf("%s[supportEmail]", prefix), b.SupportEmail)
	prm.Add(fmt.Sprintf("%s[companyUrl]", prefix), b.CompanyURL)
	prm.Add(fmt.Sprintf("%s[termsOfUsageUrl]", prefix), b.TermsOfUsageURL)
	prm.Add(fmt.Sprintf("%s[bannerLine1]", prefix), b.BannerLine1)
	prm.Add(fmt.Sprintf("%s[bannerLine2]", prefix), b.BannerLine2)
	prm.Add(fmt.Sprintf("%s[bannerLine3]", prefix), b.BannerLine3)
}

// EncodeArgs returns Branding XML body ready to be passed in the SOAP call
func (b Branding) EncodeArgs(key string) string {
	return fmt.Sprintf(`<branding xsi:type="ns1:DomainBranding">
    <companyName xsi:type="xsd:string">%s</companyName>
    <supportEmail xsi:type="xsd:string">%s</supportEmail>
    <companyUrl xsi:type="xsd:string">%s</companyUrl>
    <termsOfUsageUrl xsi:type="xsd:string">%s</termsOfUsageUrl>
    <bannerLine1 xsi:type="xsd:string">%s</bannerLine1>
    <bannerLine2 xsi:type="xsd:string">%s</bannerLine2>
    <bannerLine3 xsi:type="xsd:string">%s</bannerLine3>
</branding>`, b.CompanyName, b.SupportEmail, b.CompanyURL, b.TermsOfUsageURL, b.BannerLine1, b.BannerLine2, b.BannerLine3)
}

// Action reflects the available actions to perform on a domain
type Action string

var (
	// ActionRegister registers a domain
	ActionRegister Action = "register"
	// ActionTransfer transfers a domain to another provider
	ActionTransfer Action = "transfer"
	// ActionInternalPull transfers a domain to another account at TransIP
	ActionInternalPull Action = "internalpull"
)

// ActionResult represents a Transip_DomainAction object as described at
// https://api.transip.nl/docs/transip.nl/class-Transip_DomainAction.html
type ActionResult struct {
	Name      string `xml:"name"`
	HasFailed bool   `xml:"hasFailed"`
	Message   string `xml:"message"`
}

// WhoisContact represents a TransIP_WhoisContact object
// as described at https://api.transip.nl/docs/transip.nl/class-Transip_WhoisContact.html
type WhoisContact struct {
	Type        string `xml:"type"`
	FirstName   string `xml:"firstName"`
	MiddleName  string `xml:"middleName"`
	LastName    string `xml:"lastName"`
	CompanyName string `xml:"companyName"`
	CompanyKvk  string `xml:"companyKvk"`
	CompanyType string `xml:"companyType"`
	Street      string `xml:"street"`
	Number      string `xml:"number"`
	PostalCode  string `xml:"postalCode"`
	City        string `xml:"city"`
	PhoneNumber string `xml:"phoneNumber"`
	FaxNumber   string `xml:"faxNumber"`
	Email       string `xml:"email"`
	Country     string `xml:"country"`
}

// WhoisContacts is just an array of WhoisContact
// basically only here so it can implement paramsEncoder
type WhoisContacts []WhoisContact

// EncodeParams returns WhoisContacts parameters ready to be used for constructing a signature
// the order of parameters added here has to match the order in the WSDL
// as described at http://api.transip.nl/wsdl/?service=DomainService
func (w WhoisContacts) EncodeParams(prm gotransip.ParamsContainer, prefix string) {
	if len(w) == 0 {
		prm.Add("anything", nil)
		return
	}

	if len(prefix) == 0 {
		prefix = fmt.Sprintf("%d", prm.Len())
	}

	for i, e := range w {
		prm.Add(fmt.Sprintf("%s[%d][type]", prefix, i), e.Type)
		prm.Add(fmt.Sprintf("%s[%d][firstName]", prefix, i), e.FirstName)
		prm.Add(fmt.Sprintf("%s[%d][middleName]", prefix, i), e.MiddleName)
		prm.Add(fmt.Sprintf("%s[%d][lastName]", prefix, i), e.LastName)
		prm.Add(fmt.Sprintf("%s[%d][companyName]", prefix, i), e.CompanyName)
		prm.Add(fmt.Sprintf("%s[%d][companyKvk]", prefix, i), e.CompanyKvk)
		prm.Add(fmt.Sprintf("%s[%d][companyType]", prefix, i), e.CompanyType)
		prm.Add(fmt.Sprintf("%s[%d][street]", prefix, i), e.Street)
		prm.Add(fmt.Sprintf("%s[%d][number]", prefix, i), e.Number)
		prm.Add(fmt.Sprintf("%s[%d][postalCode]", prefix, i), e.PostalCode)
		prm.Add(fmt.Sprintf("%s[%d][city]", prefix, i), e.City)
		prm.Add(fmt.Sprintf("%s[%d][phoneNumber]", prefix, i), e.PhoneNumber)
		prm.Add(fmt.Sprintf("%s[%d][faxNumber]", prefix, i), e.FaxNumber)
		prm.Add(fmt.Sprintf("%s[%d][email]", prefix, i), e.Email)
		prm.Add(fmt.Sprintf("%s[%d][country]", prefix, i), e.Country)
	}
}

// EncodeArgs returns WhoisContacts XML body ready to be passed in the SOAP call
func (w WhoisContacts) EncodeArgs(key string) string {
	output := fmt.Sprintf(`<%s SOAP-ENC:arrayType="ns1:WhoisContact[%d]" xsi:type="ns1:ArrayOfWhoisContact">`, key, len(w)) + "\n"
	for _, e := range w {
		output += fmt.Sprintf(`	<item xsi:type="ns1:WhoisContact">
		<type xsi:type="xsd:string">%s</type>
		<firstName xsi:type="xsd:string">%s</firstName>
		<middleName xsi:type="xsd:string">%s</middleName>
		<lastName xsi:type="xsd:string">%s</lastName>
		<companyName xsi:type="xsd:string">%s</companyName>
		<companyKvk xsi:type="xsd:string">%s</companyKvk>
		<companyType xsi:type="xsd:string">%s</companyType>
		<street xsi:type="xsd:string">%s</street>
		<number xsi:type="xsd:string">%s</number>
		<postalCode xsi:type="xsd:string">%s</postalCode>
		<city xsi:type="xsd:string">%s</city>
		<phoneNumber xsi:type="xsd:string">%s</phoneNumber>
		<faxNumber xsi:type="xsd:string">%s</faxNumber>
		<email xsi:type="xsd:string">%s</email>
		<country xsi:type="xsd:string">%s</country>
	</item>`, e.Type, e.FirstName, e.MiddleName, e.LastName, e.CompanyName,
			e.CompanyKvk, e.CompanyType, e.Street, e.Number, e.PostalCode, e.City,
			e.PhoneNumber, e.FaxNumber, e.Email, e.Country) + "\n"
	}

	return output + fmt.Sprintf("</%s>", key)
}
