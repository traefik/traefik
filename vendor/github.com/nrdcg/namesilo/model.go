package namesilo

import "encoding/xml"

// Request Base request representation.
type Request struct {
	Operation string `xml:"operation"`
	IP        string `xml:"ip"`
}

// Reply Base reply representation.
type Reply struct {
	Code   string `xml:"code"`
	Detail string `xml:"detail"`
}

// Operation was generated 2019-03-20 19:35:05.
type Operation struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// AddAccountFunds was generated 2019-03-20 19:35:05.
type AddAccountFunds struct {
	XMLName xml.Name             `xml:"namesilo"`
	Request Request              `xml:"request"`
	Reply   AddAccountFundsReply `xml:"reply"`
}

// AddAccountFundsReply A reply representation.
type AddAccountFundsReply struct {
	Reply
	NewBalance string `xml:"new_balance"`
}

// AddAutoRenewal was generated 2019-03-20 19:35:05.
type AddAutoRenewal struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// AddPrivacy was generated 2019-03-20 19:35:05.
type AddPrivacy struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// AddRegisteredNameServer was generated 2019-03-20 19:35:05.
type AddRegisteredNameServer struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// ChangeNameServers was generated 2019-03-20 19:35:05.
type ChangeNameServers struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// CheckRegisterAvailability was generated 2019-03-20 19:35:05.
type CheckRegisterAvailability struct {
	XMLName xml.Name                       `xml:"namesilo"`
	Request Request                        `xml:"request"`
	Reply   CheckRegisterAvailabilityReply `xml:"reply"`
}

// CheckRegisterAvailabilityReply A reply representation.
type CheckRegisterAvailabilityReply struct {
	Reply
	Available struct {
		Domain []string `xml:"domain"`
	} `xml:"available"`
	Unavailable struct {
		Domain string `xml:"domain"`
	} `xml:"unavailable"`
	Invalid struct {
		Domain string `xml:"domain"`
	} `xml:"invalid"`
}

// CheckTransferAvailability was generated 2019-03-20 19:35:05.
type CheckTransferAvailability struct {
	XMLName xml.Name                       `xml:"namesilo"`
	Request Request                        `xml:"request"`
	Reply   CheckTransferAvailabilityReply `xml:"reply"`
}

// CheckTransferAvailabilityReply A reply representation.
type CheckTransferAvailabilityReply struct {
	Reply
	Available struct {
		Domain []string `xml:"domain"`
	} `xml:"available"`
	Unavailable struct {
		Domain []struct {
			Name   string `xml:",chardata"`
			Reason string `xml:"reason,attr"`
		} `xml:"domain"`
	} `xml:"unavailable"`
}

// CheckTransferStatus was generated 2019-03-20 19:35:05.
type CheckTransferStatus struct {
	XMLName xml.Name                 `xml:"namesilo"`
	Request Request                  `xml:"request"`
	Reply   CheckTransferStatusReply `xml:"reply"`
}

// CheckTransferStatusReply A reply representation.
type CheckTransferStatusReply struct {
	Reply
	Date    string `xml:"date"`
	Status  string `xml:"status"`
	Message string `xml:"message"`
}

// ConfigureEmailForward was generated 2019-03-20 19:35:05.
type ConfigureEmailForward struct {
	XMLName xml.Name                   `xml:"namesilo"`
	Request Request                    `xml:"request"`
	Reply   ConfigureEmailForwardReply `xml:"reply"`
}

// ConfigureEmailForwardReply A reply representation.
type ConfigureEmailForwardReply struct {
	Reply
	Message string `xml:"message"`
}

// ContactAdd was generated 2019-03-20 19:35:05.
type ContactAdd struct {
	XMLName xml.Name        `xml:"namesilo"`
	Request Request         `xml:"request"`
	Reply   ContactAddReply `xml:"reply"`
}

// ContactAddReply A reply representation.
type ContactAddReply struct {
	Reply
	ContactID string `xml:"contact_id"`
}

// ContactDomainAssociate was generated 2019-03-20 19:35:05.
type ContactDomainAssociate struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// ContactList was generated 2019-03-20 19:35:05.
type ContactList struct {
	XMLName xml.Name         `xml:"namesilo"`
	Request Request          `xml:"request"`
	Reply   ContactListReply `xml:"reply"`
}

// ContactListReply A reply representation.
type ContactListReply struct {
	Reply
	Contact []Contact `xml:"contact"`
}

// Contact A contact representation.
type Contact struct {
	ContactID      string `xml:"contact_id"`
	DefaultProfile string `xml:"default_profile"`
	Nickname       string `xml:"nickname"`
	Company        string `xml:"company"`
	FirstName      string `xml:"first_name"`
	LastName       string `xml:"last_name"`
	Address        string `xml:"address"`
	Address2       string `xml:"address2"`
	City           string `xml:"city"`
	State          string `xml:"state"`
	Zip            string `xml:"zip"`
	Country        string `xml:"country"`
	Email          string `xml:"email"`
	Phone          string `xml:"phone"`
	Fax            string `xml:"fax"`
	Usnc           string `xml:"usnc"`
	Usap           string `xml:"usap"`
	Calf           string `xml:"calf"`
	Caln           string `xml:"caln"`
	Caag           string `xml:"caag"`
	Cawd           string `xml:"cawd"`
}

// ContactUpdate was generated 2019-03-20 19:35:05.
type ContactUpdate struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// ContactDelete was generated 2019-03-20 19:35:05.
type ContactDelete struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DeleteEmailForward was generated 2019-03-20 19:35:05.
type DeleteEmailForward struct {
	XMLName xml.Name                `xml:"namesilo"`
	Request Request                 `xml:"request"`
	Reply   DeleteEmailForwardReply `xml:"reply"`
}

// DeleteEmailForwardReply A reply representation.
type DeleteEmailForwardReply struct {
	Reply
	Message string `xml:"message"`
}

// DeleteRegisteredNameServer was generated 2019-03-20 19:35:05.
type DeleteRegisteredNameServer struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DnsAddRecord was generated 2019-03-20 19:35:05.
type DnsAddRecord struct {
	XMLName xml.Name          `xml:"namesilo"`
	Request Request           `xml:"request"`
	Reply   DnsAddRecordReply `xml:"reply"`
}

// DnsAddRecordReply A reply representation.
type DnsAddRecordReply struct {
	Reply
	RecordID string `xml:"record_id"`
}

// DnsDeleteRecord was generated 2019-03-20 19:35:05.
type DnsDeleteRecord struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DnsListRecords was generated 2019-03-20 19:35:05.
type DnsListRecords struct {
	XMLName xml.Name            `xml:"namesilo"`
	Request Request             `xml:"request"`
	Reply   DnsListRecordsReply `xml:"reply"`
}

// DnsListRecordsReply A reply representation.
type DnsListRecordsReply struct {
	Reply
	ResourceRecord []ResourceRecord `xml:"resource_record"`
}

// ResourceRecord A Resource Record representation.
type ResourceRecord struct {
	RecordID string `xml:"record_id"`
	Type     string `xml:"type"`
	Host     string `xml:"host"`
	Value    string `xml:"value"`
	TTL      string `xml:"ttl"`
	Distance string `xml:"distance"`
}

// DnsSecAddRecord was generated 2019-03-20 19:35:05.
type DnsSecAddRecord struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DnsSecDeleteRecord was generated 2019-03-20 19:35:05.
type DnsSecDeleteRecord struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DnsSecListRecords was generated 2019-03-20 19:35:05.
type DnsSecListRecords struct {
	XMLName xml.Name               `xml:"namesilo"`
	Request Request                `xml:"request"`
	Reply   DnsSecListRecordsReply `xml:"reply"`
}

// DnsSecListRecordsReply A reply representation.
type DnsSecListRecordsReply struct {
	Reply
	DsRecord []DsRecord `xml:"ds_record"`
}

// DsRecord A DsRecord representation.
type DsRecord struct {
	Digest     string `xml:"digest"`
	DigestType string `xml:"digest_type"`
	Algorithm  string `xml:"algorithm"`
	KeyTag     string `xml:"key_tag"`
}

// DnsUpdateRecord was generated 2019-03-20 19:35:05.
type DnsUpdateRecord struct {
	XMLName xml.Name             `xml:"namesilo"`
	Request Request              `xml:"request"`
	Reply   DnsUpdateRecordReply `xml:"reply"`
}

// DnsUpdateRecordReply A reply representation.
type DnsUpdateRecordReply struct {
	Reply
	RecordID string `xml:"record_id"`
}

// DomainForwardSubDomainDelete was generated 2019-03-20 19:35:05.
type DomainForwardSubDomainDelete struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DomainForwardSubDomain was generated 2019-03-20 19:35:05.
type DomainForwardSubDomain struct {
	XMLName xml.Name                    `xml:"namesilo"`
	Request Request                     `xml:"request"`
	Reply   DomainForwardSubDomainReply `xml:"reply"`
}

// DomainForwardSubDomainReply A reply representation.
type DomainForwardSubDomainReply struct {
	Reply
	Message string `xml:"message"`
}

// DomainForward was generated 2019-03-20 19:35:05.
type DomainForward struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DomainLock was generated 2019-03-20 19:35:05.
type DomainLock struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// DomainUnlock was generated 2019-03-20 19:35:05.
type DomainUnlock struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// EmailVerification was generated 2019-03-20 19:35:05.
type EmailVerification struct {
	XMLName xml.Name               `xml:"namesilo"`
	Request Request                `xml:"request"`
	Reply   EmailVerificationReply `xml:"reply"`
}

// EmailVerificationReply A reply representation.
type EmailVerificationReply struct {
	Reply
	Message string `xml:"message"`
}

// GetAccountBalance was generated 2019-03-20 19:35:05.
type GetAccountBalance struct {
	XMLName xml.Name               `xml:"namesilo"`
	Request Request                `xml:"request"`
	Reply   GetAccountBalanceReply `xml:"reply"`
}

// GetAccountBalanceReply A reply representation.
type GetAccountBalanceReply struct {
	Reply
	Balance string `xml:"balance"`
}

// GetDomainInfo was generated 2019-03-20 19:35:05.
type GetDomainInfo struct {
	XMLName xml.Name           `xml:"namesilo"`
	Request Request            `xml:"request"`
	Reply   GetDomainInfoReply `xml:"reply"`
}

// GetDomainInfoReply A reply representation.
type GetDomainInfoReply struct {
	Reply
	Created                   string       `xml:"created"`
	Expires                   string       `xml:"expires"`
	Status                    string       `xml:"status"`
	Locked                    string       `xml:"locked"`
	Private                   string       `xml:"private"`
	AutoRenew                 string       `xml:"auto_renew"`
	TrafficType               string       `xml:"traffic_type"`
	EmailVerificationRequired string       `xml:"email_verification_required"`
	Portfolio                 string       `xml:"portfolio"`
	ForwardURL                string       `xml:"forward_url"`
	ForwardType               string       `xml:"forward_type"`
	Nameservers               []Nameserver `xml:"nameservers>nameserver"`
	ContactIDs                ContactIDs   `xml:"contact_ids"`
}

// Nameserver A Nameserver representation.
type Nameserver struct {
	Name     string `xml:",chardata"`
	Position string `xml:"position,attr"`
}

// ContactIDs A Contact IDs representation.
type ContactIDs struct {
	Registrant     string `xml:"registrant"`
	Administrative string `xml:"administrative"`
	Technical      string `xml:"technical"`
	Billing        string `xml:"billing"`
}

// GetPrices was generated 2019-03-20 19:35:05.
type GetPrices struct {
	XMLName xml.Name       `xml:"namesilo"`
	Request Request        `xml:"request"`
	Reply   GetPricesReply `xml:"reply"`
}

// GetPricesReply A reply representation.
type GetPricesReply struct {
	Reply
	Com ComNet `xml:"com"`
	Net ComNet `xml:"net"`
}

// ComNet A Com/Net representation.
type ComNet struct {
	Registration string `xml:"registration"`
	Transfer     string `xml:"transfer"`
	Renew        string `xml:"renew"`
}

// ListDomains was generated 2019-03-20 19:35:05.
type ListDomains struct {
	XMLName xml.Name         `xml:"namesilo"`
	Request Request          `xml:"request"`
	Reply   ListDomainsReply `xml:"reply"`
}

// ListDomainsReply A reply representation.
type ListDomainsReply struct {
	Reply
	Domains struct {
		Domain []string `xml:"domain"`
	} `xml:"domains"`
}

// ListEmailForwards was generated 2019-03-20 19:35:05.
type ListEmailForwards struct {
	XMLName xml.Name               `xml:"namesilo"`
	Request Request                `xml:"request"`
	Reply   ListEmailForwardsReply `xml:"reply"`
}

// ListEmailForwardsReply A reply representation.
type ListEmailForwardsReply struct {
	Reply
	Addresses []Address `xml:"addresses"`
}

// Address An Address representation.
type Address struct {
	Email      string   `xml:"email"`
	ForwardsTo []string `xml:"forwards_to"`
}

// ListOrders was generated 2019-03-20 19:35:05.
type ListOrders struct {
	XMLName xml.Name        `xml:"namesilo"`
	Request Request         `xml:"request"`
	Reply   ListOrdersReply `xml:"reply"`
}

// ListOrdersReply A reply representation.
type ListOrdersReply struct {
	Reply
	Order []Order `xml:"order"`
}

// Order An Order representation.
type Order struct {
	OrderNumber string `xml:"order_number"`
	OrderDate   string `xml:"order_date"`
	Method      string `xml:"method"`
	Total       string `xml:"total"`
}

// ListRegisteredNameServers was generated 2019-03-20 19:35:05.
type ListRegisteredNameServers struct {
	XMLName xml.Name                       `xml:"namesilo"`
	Request Request                        `xml:"request"`
	Reply   ListRegisteredNameServersReply `xml:"reply"`
}

// ListRegisteredNameServersReply A reply representation.
type ListRegisteredNameServersReply struct {
	Reply
	Hosts []Host `xml:"hosts"`
}

// Host A Host representation.
type Host struct {
	Host string   `xml:"host"`
	IP   []string `xml:"ip"`
}

// MarketplaceActiveSalesOverview was generated 2019-03-20 19:35:05.
type MarketplaceActiveSalesOverview struct {
	XMLName xml.Name                            `xml:"namesilo"`
	Request Request                             `xml:"request"`
	Reply   MarketplaceActiveSalesOverviewReply `xml:"reply"`
}

// MarketplaceActiveSalesOverviewReply A reply representation.
type MarketplaceActiveSalesOverviewReply struct {
	Reply
	SaleDetails []SaleDetail `xml:"sale_details"`
}

// SaleDetail A Sale Detail representation.
type SaleDetail struct {
	Domain           string `xml:"domain"`
	Status           string `xml:"status"`
	Reserve          string `xml:"reserve"`
	BuyNow           string `xml:"buy_now"`
	Portfolio        string `xml:"portfolio"`
	SaleType         string `xml:"sale_type"`
	PayPlanOffered   string `xml:"pay_plan_offered"`
	EndDate          string `xml:"end_date"`
	AutoExtendDays   string `xml:"auto_extend_days"`
	TimeRemaining    string `xml:"time_remaining"`
	Private          string `xml:"private"`
	ActiveBidOrOffer string `xml:"active_bid_or_offer"`
}

// MarketplaceAddOrModifySale was generated 2019-03-20 19:35:05.
type MarketplaceAddOrModifySale struct {
	XMLName xml.Name                        `xml:"namesilo"`
	Request Request                         `xml:"request"`
	Reply   MarketplaceAddOrModifySaleReply `xml:"reply"`
}

// MarketplaceAddOrModifySaleReply A reply representation.
type MarketplaceAddOrModifySaleReply struct {
	Reply
	Message string `xml:"message"`
}

// MarketplaceLandingPageUpdate was generated 2019-03-20 19:35:05.
type MarketplaceLandingPageUpdate struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// ModifyRegisteredNameServer was generated 2019-03-20 19:35:05.
type ModifyRegisteredNameServer struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// OrderDetails was generated 2019-03-20 19:35:05.
type OrderDetails struct {
	XMLName xml.Name          `xml:"namesilo"`
	Request Request           `xml:"request"`
	Reply   OrderDetailsReply `xml:"reply"`
}

// OrderDetailsReply A reply representation.
type OrderDetailsReply struct {
	Reply
	OrderDate    string        `xml:"order_date"`
	Method       string        `xml:"method"`
	Total        string        `xml:"total"`
	OrderDetails []OrderDetail `xml:"order_details"`
}

// OrderDetail An Order Detail representation.
type OrderDetail struct {
	Description    string `xml:"description"`
	YearsQty       string `xml:"years_qty"`
	Price          string `xml:"price"`
	Subtotal       string `xml:"subtotal"`
	Status         string `xml:"status"`
	CreditedDate   string `xml:"credited_date,omitempty"`
	CreditedAmount string `xml:"credited_amount,omitempty"`
}

// PortfolioAdd was generated 2019-03-20 19:35:05.
type PortfolioAdd struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// PortfolioDelete was generated 2019-03-20 19:35:05.
type PortfolioDelete struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// PortfolioDomainAssociate was generated 2019-03-20 19:35:05.
type PortfolioDomainAssociate struct {
	XMLName xml.Name                      `xml:"namesilo"`
	Request Request                       `xml:"request"`
	Reply   PortfolioDomainAssociateReply `xml:"reply"`
}

// PortfolioDomainAssociateReply A reply representation.
type PortfolioDomainAssociateReply struct {
	Reply
	Message string `xml:"message"`
}

// PortfolioList was generated 2019-03-20 19:35:05.
type PortfolioList struct {
	XMLName xml.Name           `xml:"namesilo"`
	Request Request            `xml:"request"`
	Reply   PortfolioListReply `xml:"reply"`
}

// PortfolioListReply A reply representation.
type PortfolioListReply struct {
	Reply
	Portfolios Portfolios `xml:"portfolios"`
}

// Portfolios A Portfolios representation.
type Portfolios struct {
	Name string `xml:"name"`
	Code string `xml:"code"`
}

// RegisterDomainDrop was generated 2019-03-20 19:35:05.
type RegisterDomainDrop struct {
	XMLName xml.Name                `xml:"namesilo"`
	Request Request                 `xml:"request"`
	Reply   RegisterDomainDropReply `xml:"reply"`
}

// RegisterDomainDropReply A reply representation.
type RegisterDomainDropReply struct {
	Reply
	Message     string `xml:"message"`
	Domain      string `xml:"domain"`
	OrderAmount string `xml:"order_amount"`
}

// RegisterDomain was generated 2019-03-20 19:35:05.
type RegisterDomain struct {
	XMLName xml.Name            `xml:"namesilo"`
	Request Request             `xml:"request"`
	Reply   RegisterDomainReply `xml:"reply"`
}

// RegisterDomainReply A reply representation.
type RegisterDomainReply struct {
	Reply
	Message     string `xml:"message"`
	Domain      string `xml:"domain"`
	OrderAmount string `xml:"order_amount"`
}

// RegistrantVerificationStatus was generated 2019-03-20 19:35:05.
type RegistrantVerificationStatus struct {
	XMLName xml.Name                          `xml:"namesilo"`
	Request Request                           `xml:"request"`
	Reply   RegistrantVerificationStatusReply `xml:"reply"`
}

// RegistrantVerificationStatusReply A reply representation.
type RegistrantVerificationStatusReply struct {
	Reply
	Emails []RegistrantEmail `xml:"email"`
}

// RegistrantEmail A email representation.
type RegistrantEmail struct {
	EmailAddress string `xml:"email_address"`
	Domains      string `xml:"domains"`
	Verified     string `xml:"verified"`
}

// RemoveAutoRenewal was generated 2019-03-20 19:35:05.
type RemoveAutoRenewal struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// RemovePrivacy was generated 2019-03-20 19:35:05.
type RemovePrivacy struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// RenewDomain was generated 2019-03-20 19:35:05.
type RenewDomain struct {
	XMLName xml.Name         `xml:"namesilo"`
	Request Request          `xml:"request"`
	Reply   RenewDomainReply `xml:"reply"`
}

// RenewDomainReply A reply representation.
type RenewDomainReply struct {
	Reply
	Message     string `xml:"message"`
	Domain      string `xml:"domain"`
	OrderAmount string `xml:"order_amount"`
}

// RetrieveAuthCode was generated 2019-03-20 19:35:05.
type RetrieveAuthCode struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// TransferDomain was generated 2019-03-20 19:35:05.
type TransferDomain struct {
	XMLName xml.Name            `xml:"namesilo"`
	Request Request             `xml:"request"`
	Reply   TransferDomainReply `xml:"reply"`
}

// TransferDomainReply A reply representation.
type TransferDomainReply struct {
	Reply
	Message     string `xml:"message"`
	Domain      string `xml:"domain"`
	OrderAmount string `xml:"order_amount"`
}

// TransferUpdateChangeEPPCode was generated 2019-03-20 19:35:05.
type TransferUpdateChangeEPPCode struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// TransferUpdateResendAdminEmail was generated 2019-03-20 19:35:05.
type TransferUpdateResendAdminEmail struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}

// TransferUpdateResubmitToRegistry was generated 2019-03-20 19:35:05.
type TransferUpdateResubmitToRegistry struct {
	XMLName xml.Name `xml:"namesilo"`
	Request Request  `xml:"request"`
	Reply   Reply    `xml:"reply"`
}
