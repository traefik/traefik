package namesilo

// AddAccountFundsParams Parameters for operation addAccountFunds.
type AddAccountFundsParams struct {
	Amount    string `url:"amount"`
	PaymentID string `url:"payment_id"`
}

// AddAutoRenewalParams Parameters for operation addAutoRenewal.
type AddAutoRenewalParams struct {
	Domain string `url:"domain"` // Required
}

// AddPrivacyParams Parameters for operation addPrivacy.
type AddPrivacyParams struct {
	Domain string `url:"domain"` // Required
}

// AddRegisteredNameServerParams Parameters for operation addRegisteredNameServer.
type AddRegisteredNameServerParams struct {
	Domain  string `url:"domain"`   // required
	NewHost string `url:"new_host"` // Required
	IP1     string `url:"ip1"`      // Required

	IP2  string `url:"ip2"`  // Optional
	IP3  string `url:"ip3"`  // Optional
	IP4  string `url:"ip4"`  // Optional
	IP5  string `url:"ip5"`  // Optional
	IP6  string `url:"ip6"`  // Optional
	IP7  string `url:"ip7"`  // Optional
	IP8  string `url:"ip8"`  // Optional
	IP9  string `url:"ip9"`  // Optional
	IP10 string `url:"ip10"` // Optional
	IP11 string `url:"ip11"` // Optional
	IP12 string `url:"ip12"` // Optional
	IP13 string `url:"ip13"` // Optional
}

// ChangeNameServersParams Parameters for operation changeNameServers.
type ChangeNameServersParams struct {
	Domains string `url:"domain"` // Required (A comma-delimited list of up to 200 domains)

	NameServer1 string `url:"ns1"` // Required
	NameServer2 string `url:"ns2"` // Required

	NameServer3  string `url:"ns3"`
	NameServer4  string `url:"ns4"`
	NameServer5  string `url:"ns5"`
	NameServer6  string `url:"ns6"`
	NameServer7  string `url:"ns7"`
	NameServer8  string `url:"ns8"`
	NameServer9  string `url:"ns9"`
	NameServer10 string `url:"ns10"`
	NameServer11 string `url:"ns11"`
	NameServer12 string `url:"ns12"`
	NameServer13 string `url:"ns13"`
}

// CheckRegisterAvailabilityParams Parameters for operation checkRegisterAvailability.
type CheckRegisterAvailabilityParams struct {
	Domains string `url:"domains"` // Required (A comma-delimited list of domains to check)
}

// CheckTransferAvailabilityParams Parameters for operation checkTransferAvailability.
type CheckTransferAvailabilityParams struct {
	Domains string `url:"domains"` // Required (A comma-delimited list of domains to check)
}

// CheckTransferStatusParams Parameters for operation checkTransferStatus.
type CheckTransferStatusParams struct {
	Domain string `url:"domain"` // Required
}

// ConfigureEmailForwardParams Parameters for operation configureEmailForward.
type ConfigureEmailForwardParams struct {
	Domain   string `url:"domain"`   // Required
	Email    string `url:"email"`    // Required
	Forward1 string `url:"forward1"` // Required

	Forward2 string `url:"forward12"` // Optional
	Forward3 string `url:"forward13"` // Optional
	Forward4 string `url:"forward14"` // Optional
	Forward5 string `url:"forward15"` // Optional
}

// ContactAddParams Parameters for operation contactAdd.
type ContactAddParams struct {
	FirstName                     string `url:"fn"` // Contact Information
	LastName                      string `url:"ln"` // Contact Information
	MailingAddress                string `url:"ad"` // Contact Information
	MailingCity                   string `url:"cy"` // Contact Information
	MailingStateProvinceTerritory string `url:"st"` // Contact Information
	MailingZipPostalCode          string `url:"zp"` // Contact Information
	MailingCountry                string `url:"ct"` // Contact Information
	EmailAddress                  string `url:"em"` // Contact Information
	PhoneNumber                   string `url:"ph"` // Contact Information

	Company         string `url:"cp"`  // Contact Information
	MailingAddress2 string `url:"ad2"` // Contact Information
	Fax             string `url:"fx"`  // Contact Information

	USNexusCategory      string `url:"usnc"` // Contact Information
	USApplicationPurpose string `url:"usap"` // Contact Information

	CIRALegalForm        string `url:"calf"` // CIRA
	CIRALanguage         string `url:"caln"` // CIRA
	CIRAAgreementVersion string `url:"caag"` // CIRA
	CIRAWHOISDisplay     string `url:"cawd"` // CIRA
}

// ContactDeleteParams Parameters for operation contactDelete.
type ContactDeleteParams struct {
	ContactID string `url:"contact_id"`
}

// ContactDomainAssociateParams Parameters for operation contactDomainAssociate.
type ContactDomainAssociateParams struct {
	Domain string `url:"domain"` // Required

	Registrant     string `url:"registrant"`     // Optional
	Administrative string `url:"administrative"` // Optional
	Billing        string `url:"billing"`        // Optional
	Technical      string `url:"technical"`      // Optional

	ContactID string `url:"contact_id"` // Contact ID
}

// ContactListParams Parameters for operation contactList.
type ContactListParams struct {
	ContactID string `url:"contact_id"` // Optional
}

// ContactUpdateParams Parameters for operation contactUpdate.
type ContactUpdateParams struct {
	FirstName                     string `url:"fn"` // Contact Information
	LastName                      string `url:"ln"` // Contact Information
	MailingAddress                string `url:"ad"` // Contact Information
	MailingCity                   string `url:"cy"` // Contact Information
	MailingStateProvinceTerritory string `url:"st"` // Contact Information
	MailingZipPostalCode          string `url:"zp"` // Contact Information
	MailingCountry                string `url:"ct"` // Contact Information
	EmailAddress                  string `url:"em"` // Contact Information
	PhoneNumber                   string `url:"ph"` // Contact Information

	Company         string `url:"cp"`  // Contact Information
	MailingAddress2 string `url:"ad2"` // Contact Information
	Fax             string `url:"fx"`  // Contact Information

	USNexusCategory      string `url:"usnc"` // Contact Information
	USApplicationPurpose string `url:"usap"` // Contact Information

	CIRALegalForm        string `url:"calf"` // CIRA
	CIRALanguage         string `url:"caln"` // CIRA
	CIRAAgreementVersion string `url:"caag"` // CIRA
	CIRAWHOISDisplay     string `url:"cawd"` // CIRA
}

// DeleteEmailForwardParams Parameters for operation deleteEmailForward.
type DeleteEmailForwardParams struct {
	Domain string `url:"domain"` // Required
	Email  string `url:"email"`  // Required
}

// DeleteRegisteredNameServerParams Parameters for operation deleteRegisteredNameServer.
type DeleteRegisteredNameServerParams struct {
	Domain      string `url:"domain"`       // required
	CurrentHost string `url:"current_host"` // Required
}

// DnsAddRecordParams Parameters for operation dnsAddRecord.
type DnsAddRecordParams struct {
	Domain string `url:"domain"` // Required

	Type     string `url:"rrtype"` // Possible values are "A", "AAAA", "CNAME", "MX" and "TXT"
	Host     string `url:"rrhost"`
	Value    string `url:"rrvalue"`
	Distance int    `url:"rrdistance"`
	TTL      int    `url:"rrttl"`
}

// DnsDeleteRecordParams Parameters for operation dnsDeleteRecord.
type DnsDeleteRecordParams struct {
	Domain string `url:"domain"` // Required

	ID string `url:"rrid"`
}

// DnsListRecordsParams Parameters for operation dnsListRecords.
type DnsListRecordsParams struct {
	Domain string `url:"domain"` // Required
}

// DnsSecAddRecordParams Parameters for operation dnsSecAddRecord.
type DnsSecAddRecordParams struct {
	Domain string `url:"domain"` // Required

	Digest     string `url:"digest"`
	KeyTag     string `url:"keyTag"`
	DigestType string `url:"digestType"`
	Alg        string `url:"alg"`
}

// DnsSecDeleteRecordParams Parameters for operation dnsSecDeleteRecord.
type DnsSecDeleteRecordParams struct {
	Domain string `url:"domain"` // Required

	Digest     string `url:"digest"`
	KeyTag     string `url:"keyTag"`
	DigestType string `url:"digestType"`
	Alg        string `url:"alg"`
}

// DnsSecListRecordsParams Parameters for operation dnsSecListRecords.
type DnsSecListRecordsParams struct {
	Domain string `url:"domain"` // Required
}

// DnsUpdateRecordParams Parameters for operation dnsUpdateRecord.
type DnsUpdateRecordParams struct {
	Domain string `url:"domain"` // Required

	ID       string `url:"rrid"`
	Host     string `url:"rrhost"`
	Value    string `url:"rrvalue"`
	Distance int    `url:"rrdistance"`
	TTL      int    `url:"rrttl"`
}

// DomainForwardParams Parameters for operation domainForward.
type DomainForwardParams struct {
	Domain   string `url:"domain"`   // Required
	Protocol string `url:"protocol"` // Required
	Address  string `url:"address"`  // Required
	Method   string `url:"method"`   // Required

	MetaTitle       string `url:"meta_title"`       // Optional
	MetaDescription string `url:"meta_description"` // Optional
	MetaKeywords    string `url:"meta_keywords"`    // Optional

}

// DomainForwardSubDomainParams Parameters for operation domainForwardSubDomain.
type DomainForwardSubDomainParams struct {
	Domain    string `url:"domain"`     // Required
	SubDomain string `url:"sub_domain"` // Required
	Protocol  string `url:"protocol"`   // Required
	Address   string `url:"address"`    // Required
	Method    string `url:"method"`     // Required

	MetaTitle       string `url:"meta_title"`       // Optional
	MetaDescription string `url:"meta_description"` // Optional
	MetaKeywords    string `url:"meta_keywords"`    // Optional
}

// DomainForwardSubDomainDeleteParams Parameters for operation domainForwardSubDomainDelete.
type DomainForwardSubDomainDeleteParams struct {
	Domain    string `url:"domain"`     // Required
	SubDomain string `url:"sub_domain"` // Required
}

// DomainLockParams Parameters for operation domainLock.
type DomainLockParams struct {
	Domain string `url:"domain"` // Required
}

// DomainUnlockParams Parameters for operation domainUnlock.
type DomainUnlockParams struct {
	Domain string `url:"domain"` // Required
}

// EmailVerificationParams Parameters for operation emailVerification.
type EmailVerificationParams struct {
	Email string `url:"email"` // Required
}

// GetAccountBalanceParams Parameters for operation getAccountBalance.
type GetAccountBalanceParams struct{}

// GetDomainInfoParams Parameters for operation getDomainInfo.
type GetDomainInfoParams struct {
	Domain string `url:"domain"` // Required
}

// GetPricesParams Parameters for operation getPrices.
type GetPricesParams struct {
	RetailPrices        string `url:"retail_prices"`        // Required
	RegistrationDomains string `url:"registration_domains"` // Required
}

// ListDomainsParams Parameters for operation listDomains.
type ListDomainsParams struct {
	Portfolio string `url:"portfolio"` // Optional
}

// ListEmailForwardsParams Parameters for operation listEmailForwards.
type ListEmailForwardsParams struct {
	Domain string `url:"domain"` // Required
}

// ListOrdersParams Parameters for operation listOrders.
type ListOrdersParams struct{}

// ListRegisteredNameServersParams Parameters for operation listRegisteredNameServers.
type ListRegisteredNameServersParams struct {
	Domain string `url:"domain"` // required
}

// MarketplaceActiveSalesOverviewParams Parameters for operation marketplaceActiveSalesOverview.
type MarketplaceActiveSalesOverviewParams struct{}

// MarketplaceAddOrModifySaleParams Parameters for operation marketplaceAddOrModifySale.
type MarketplaceAddOrModifySaleParams struct {
	Domain   string `url:"domain"`    // Required
	Action   string `url:"action"`    // Required
	SaleType string `url:"sale_type"` // Required

	Reserve                string `url:"reserve"`                   // Optional
	ShowReserve            int32  `url:"show_reserve"`              // Optional
	BuyNow                 string `url:"buy_now"`                   // Optional
	PaymentPlanOffered     int32  `url:"payment_plan_offered"`      // Optional
	PaymentPlanMonths      int32  `url:"payment_plan_months"`       // Optional
	PaymentPlanDownPayment string `url:"payment_plan_down_payment"` // Optional
	EndDate                string `url:"end_date"`                  // Optional
	EndDateUseMaximum      int32  `url:"end_date_use_maximum"`      // Optional
	NotifyBuyers           int32  `url:"notify_buyers"`             // Optional
	Category1              string `url:"category1"`                 // Optional
	Description            string `url:"description"`               // Optional
	UseForSaleLandingPage  int32  `url:"use_for_sale_landing_page"` // Optional
	MpUseOurNameservers    int32  `url:"mp_use_our_nameservers"`    // Optional
	Password               string `url:"password"`                  // Optional
	CancelSale             int32  `url:"cancel_sale"`               // Optional
}

// MarketplaceLandingPageUpdateParams Parameters for operation marketplaceLandingPageUpdate.
type MarketplaceLandingPageUpdateParams struct {
	Domain string `url:"domain"` // Required

	MpTemplate         int32  `url:"mp_template"`            // Optional
	MpBgcolor          string `url:"mp_bgcolor"`             // Optional
	MpTextcolor        string `url:"mp_textcolor"`           // Optional
	MpShowBuyNow       int32  `url:"mp_show_buy_now"`        // Optional
	MpShowMoreInfo     int32  `url:"mp_show_more_info"`      // Optional
	MpShowRenewalPrice int32  `url:"mp_show_renewal_price"`  // Optional
	MpShowOtherForSale int32  `url:"mp_show_other_for_sale"` // Optional
	MpOtherDomainLinks string `url:"mp_other_domain_links"`  // Optional
	MpMessage          string `url:"mp_message"`             // Optional
}

// ModifyRegisteredNameServerParams Parameters for operation modifyRegisteredNameServer.
type ModifyRegisteredNameServerParams struct {
	Domain      string `url:"domain"`       // required
	CurrentHost string `url:"current_host"` // Required
	NewHost     string `url:"new_host"`     // Required
	IP1         string `url:"ip1"`          // Required

	IP2  string `url:"ip2"`  // Optional
	IP3  string `url:"ip3"`  // Optional
	IP4  string `url:"ip4"`  // Optional
	IP5  string `url:"ip5"`  // Optional
	IP6  string `url:"ip6"`  // Optional
	IP7  string `url:"ip7"`  // Optional
	IP8  string `url:"ip8"`  // Optional
	IP9  string `url:"ip9"`  // Optional
	IP10 string `url:"ip10"` // Optional
	IP11 string `url:"ip11"` // Optional
	IP12 string `url:"ip12"` // Optional
	IP13 string `url:"ip13"` // Optional
}

// OrderDetailsParams Parameters for operation orderDetails.
type OrderDetailsParams struct {
	OrderNumber int `url:"order_number"`
}

// PortfolioAddParams Parameters for operation portfolioAdd.
type PortfolioAddParams struct {
	Portfolio string `url:"portfolio"` // Required
}

// PortfolioDeleteParams Parameters for operation portfolioDelete.
type PortfolioDeleteParams struct {
	Portfolio string `url:"portfolio"` // Required
}

// PortfolioDomainAssociateParams Parameters for operation portfolioDomainAssociate.
type PortfolioDomainAssociateParams struct {
	Portfolio string `url:"portfolio"` // Required
	Domains   string `url:"domains"`   // Required (Comma-delimited list)
}

// PortfolioListParams Parameters for operation portfolioList.
type PortfolioListParams struct{}

// RegisterDomainParams Parameters for operation registerDomain.
type RegisterDomainParams struct {
	Domain string `url:"domain"` // Required
	Years  int32  `url:"years"`  // Required

	PaymentID string `url:"payment_id"` // Optional
	Private   int32  `url:"private"`    // Optional
	AutoRenew int32  `url:"auto_renew"` // Optional
	Portfolio string `url:"portfolio"`  // Optional
	Coupon    string `url:"coupon"`     // Optional

	NameServer1  string `url:"ns1"`
	NameServer2  string `url:"ns2"`
	NameServer3  string `url:"ns3"`
	NameServer4  string `url:"ns4"`
	NameServer5  string `url:"ns5"`
	NameServer6  string `url:"ns6"`
	NameServer7  string `url:"ns7"`
	NameServer8  string `url:"ns8"`
	NameServer9  string `url:"ns9"`
	NameServer10 string `url:"ns10"`
	NameServer11 string `url:"ns11"`
	NameServer12 string `url:"ns12"`
	NameServer13 string `url:"ns13"`

	FirstName                     string `url:"fn"` // Contact Information
	LastName                      string `url:"ln"` // Contact Information
	MailingAddress                string `url:"ad"` // Contact Information
	MailingCity                   string `url:"cy"` // Contact Information
	MailingStateProvinceTerritory string `url:"st"` // Contact Information
	MailingZipPostalCode          string `url:"zp"` // Contact Information
	MailingCountry                string `url:"ct"` // Contact Information
	EmailAddress                  string `url:"em"` // Contact Information
	PhoneNumber                   string `url:"ph"` // Contact Information

	Company         string `url:"cp"`  // Contact Information
	MailingAddress2 string `url:"ad2"` // Contact Information
	Fax             string `url:"fx"`  // Contact Information

	USNexusCategory      string `url:"usnc"` // Contact Information
	USApplicationPurpose string `url:"usap"` // Contact Information

	ContactID string `url:"contact_id"` // Contact ID
}

// RegisterDomainDropParams Parameters for operation registerDomainDrop.
type RegisterDomainDropParams struct {
	Domain string `url:"domain"` // Required
	Years  int32  `url:"years"`  // Required

	Private   int32 `url:"private"`    // Optional
	AutoRenew int32 `url:"auto_renew"` // Optional
}

// RegistrantVerificationStatusParams Parameters for operation registrantVerificationStatus.
type RegistrantVerificationStatusParams struct{}

// RemoveAutoRenewalParams Parameters for operation removeAutoRenewal.
type RemoveAutoRenewalParams struct {
	Domain string `url:"domain"` // Required
}

// RemovePrivacyParams Parameters for operation removePrivacy.
type RemovePrivacyParams struct {
	Domain string `url:"domain"` // Required
}

// RenewDomainParams Parameters for operation renewDomain.
type RenewDomainParams struct {
	Domain string `url:"domain"` // Required
	Years  int32  `url:"years"`  // Required

	PaymentID string `url:"payment_id"` // Optional
	Coupon    string `url:"coupon"`     // Optional
}

// RetrieveAuthCodeParams Parameters for operation retrieveAuthCode.
type RetrieveAuthCodeParams struct {
	Domain string `url:"domain"` // Required
}

// TransferDomainParams Parameters for operation transferDomain.
type TransferDomainParams struct {
	Domain string `url:"domain"` // Required

	PaymentID string `url:"payment_id"` // Optional
	Auth      string `url:"auth"`       // Optional
	Private   int32  `url:"private"`    // Optional
	AutoRenew int32  `url:"auto_renew"` // Optional
	Portfolio string `url:"portfolio"`  // Optional
	Coupon    string `url:"coupon"`     // Optional

	FirstName                     string `url:"fn"` // Contact Information
	LastName                      string `url:"ln"` // Contact Information
	MailingAddress                string `url:"ad"` // Contact Information
	MailingCity                   string `url:"cy"` // Contact Information
	MailingStateProvinceTerritory string `url:"st"` // Contact Information
	MailingZipPostalCode          string `url:"zp"` // Contact Information
	MailingCountry                string `url:"ct"` // Contact Information
	EmailAddress                  string `url:"em"` // Contact Information
	PhoneNumber                   string `url:"ph"` // Contact Information

	Company         string `url:"cp"`  // Contact Information
	MailingAddress2 string `url:"ad2"` // Contact Information
	Fax             string `url:"fx"`  // Contact Information

	USNexusCategory      string `url:"usnc"` // Contact Information
	USApplicationPurpose string `url:"usap"` // Contact Information

	ContactID string `url:"contact_id"` // Contact ID
}

// TransferUpdateChangeEPPCodeParams Parameters for operation transferUpdateChangeEPPCode.
type TransferUpdateChangeEPPCodeParams struct {
	Domain string `url:"domain"` // Required
	Auth   string `url:"auth"`   // Required
}

// TransferUpdateResendAdminEmailParams Parameters for operation transferUpdateResendAdminEmail.
type TransferUpdateResendAdminEmailParams struct {
	Domain string `url:"domain"` // Required
}

// TransferUpdateResubmitToRegistryParams Parameters for operation transferUpdateResubmitToRegistry.
type TransferUpdateResubmitToRegistryParams struct {
	Domain string `url:"domain"` // Required
}
