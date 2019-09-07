package namesilo

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// AddAccountFunds Execute operation addAccountFunds.
func (c *Client) AddAccountFunds(params *AddAccountFundsParams) (*AddAccountFunds, error) {
	resp, err := c.get("addAccountFunds", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &AddAccountFunds{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// AddAutoRenewal Execute operation addAutoRenewal.
func (c *Client) AddAutoRenewal(params *AddAutoRenewalParams) (*AddAutoRenewal, error) {
	resp, err := c.get("addAutoRenewal", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &AddAutoRenewal{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// AddPrivacy Execute operation addPrivacy.
func (c *Client) AddPrivacy(params *AddPrivacyParams) (*AddPrivacy, error) {
	resp, err := c.get("addPrivacy", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &AddPrivacy{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// AddRegisteredNameServer Execute operation addRegisteredNameServer.
func (c *Client) AddRegisteredNameServer(params *AddRegisteredNameServerParams) (*AddRegisteredNameServer, error) {
	resp, err := c.get("addRegisteredNameServer", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &AddRegisteredNameServer{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ChangeNameServers Execute operation changeNameServers.
func (c *Client) ChangeNameServers(params *ChangeNameServersParams) (*ChangeNameServers, error) {
	resp, err := c.get("changeNameServers", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ChangeNameServers{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// CheckRegisterAvailability Execute operation checkRegisterAvailability.
func (c *Client) CheckRegisterAvailability(params *CheckRegisterAvailabilityParams) (*CheckRegisterAvailability, error) {
	resp, err := c.get("checkRegisterAvailability", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &CheckRegisterAvailability{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// CheckTransferAvailability Execute operation checkTransferAvailability.
func (c *Client) CheckTransferAvailability(params *CheckTransferAvailabilityParams) (*CheckTransferAvailability, error) {
	resp, err := c.get("checkTransferAvailability", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &CheckTransferAvailability{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// CheckTransferStatus Execute operation checkTransferStatus.
func (c *Client) CheckTransferStatus(params *CheckTransferStatusParams) (*CheckTransferStatus, error) {
	resp, err := c.get("checkTransferStatus", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &CheckTransferStatus{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ConfigureEmailForward Execute operation configureEmailForward.
func (c *Client) ConfigureEmailForward(params *ConfigureEmailForwardParams) (*ConfigureEmailForward, error) {
	resp, err := c.get("configureEmailForward", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ConfigureEmailForward{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ContactAdd Execute operation contactAdd.
func (c *Client) ContactAdd(params *ContactAddParams) (*ContactAdd, error) {
	resp, err := c.get("contactAdd", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ContactAdd{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ContactDelete Execute operation contactDelete.
func (c *Client) ContactDelete(params *ContactDeleteParams) (*ContactDelete, error) {
	resp, err := c.get("contactDelete", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ContactDelete{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ContactDomainAssociate Execute operation contactDomainAssociate.
func (c *Client) ContactDomainAssociate(params *ContactDomainAssociateParams) (*ContactDomainAssociate, error) {
	resp, err := c.get("contactDomainAssociate", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ContactDomainAssociate{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ContactList Execute operation contactList.
func (c *Client) ContactList(params *ContactListParams) (*ContactList, error) {
	resp, err := c.get("contactList", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ContactList{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ContactUpdate Execute operation contactUpdate.
func (c *Client) ContactUpdate(params *ContactUpdateParams) (*ContactUpdate, error) {
	resp, err := c.get("contactUpdate", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ContactUpdate{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DeleteEmailForward Execute operation deleteEmailForward.
func (c *Client) DeleteEmailForward(params *DeleteEmailForwardParams) (*DeleteEmailForward, error) {
	resp, err := c.get("deleteEmailForward", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DeleteEmailForward{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DeleteRegisteredNameServer Execute operation deleteRegisteredNameServer.
func (c *Client) DeleteRegisteredNameServer(params *DeleteRegisteredNameServerParams) (*DeleteRegisteredNameServer, error) {
	resp, err := c.get("deleteRegisteredNameServer", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DeleteRegisteredNameServer{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DnsAddRecord Execute operation dnsAddRecord.
func (c *Client) DnsAddRecord(params *DnsAddRecordParams) (*DnsAddRecord, error) {
	resp, err := c.get("dnsAddRecord", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DnsAddRecord{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DnsDeleteRecord Execute operation dnsDeleteRecord.
func (c *Client) DnsDeleteRecord(params *DnsDeleteRecordParams) (*DnsDeleteRecord, error) {
	resp, err := c.get("dnsDeleteRecord", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DnsDeleteRecord{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DnsListRecords Execute operation dnsListRecords.
func (c *Client) DnsListRecords(params *DnsListRecordsParams) (*DnsListRecords, error) {
	resp, err := c.get("dnsListRecords", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DnsListRecords{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DnsSecAddRecord Execute operation dnsSecAddRecord.
func (c *Client) DnsSecAddRecord(params *DnsSecAddRecordParams) (*DnsSecAddRecord, error) {
	resp, err := c.get("dnsSecAddRecord", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DnsSecAddRecord{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DnsSecDeleteRecord Execute operation dnsSecDeleteRecord.
func (c *Client) DnsSecDeleteRecord(params *DnsSecDeleteRecordParams) (*DnsSecDeleteRecord, error) {
	resp, err := c.get("dnsSecDeleteRecord", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DnsSecDeleteRecord{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DnsSecListRecords Execute operation dnsSecListRecords.
func (c *Client) DnsSecListRecords(params *DnsSecListRecordsParams) (*DnsSecListRecords, error) {
	resp, err := c.get("dnsSecListRecords", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DnsSecListRecords{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DnsUpdateRecord Execute operation dnsUpdateRecord.
func (c *Client) DnsUpdateRecord(params *DnsUpdateRecordParams) (*DnsUpdateRecord, error) {
	resp, err := c.get("dnsUpdateRecord", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DnsUpdateRecord{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DomainForward Execute operation domainForward.
func (c *Client) DomainForward(params *DomainForwardParams) (*DomainForward, error) {
	resp, err := c.get("domainForward", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DomainForward{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DomainForwardSubDomain Execute operation domainForwardSubDomain.
func (c *Client) DomainForwardSubDomain(params *DomainForwardSubDomainParams) (*DomainForwardSubDomain, error) {
	resp, err := c.get("domainForwardSubDomain", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DomainForwardSubDomain{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DomainForwardSubDomainDelete Execute operation domainForwardSubDomainDelete.
func (c *Client) DomainForwardSubDomainDelete(params *DomainForwardSubDomainDeleteParams) (*DomainForwardSubDomainDelete, error) {
	resp, err := c.get("domainForwardSubDomainDelete", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DomainForwardSubDomainDelete{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DomainLock Execute operation domainLock.
func (c *Client) DomainLock(params *DomainLockParams) (*DomainLock, error) {
	resp, err := c.get("domainLock", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DomainLock{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// DomainUnlock Execute operation domainUnlock.
func (c *Client) DomainUnlock(params *DomainUnlockParams) (*DomainUnlock, error) {
	resp, err := c.get("domainUnlock", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &DomainUnlock{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// EmailVerification Execute operation emailVerification.
func (c *Client) EmailVerification(params *EmailVerificationParams) (*EmailVerification, error) {
	resp, err := c.get("emailVerification", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &EmailVerification{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// GetAccountBalance Execute operation getAccountBalance.
func (c *Client) GetAccountBalance(params *GetAccountBalanceParams) (*GetAccountBalance, error) {
	resp, err := c.get("getAccountBalance", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &GetAccountBalance{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// GetDomainInfo Execute operation getDomainInfo.
func (c *Client) GetDomainInfo(params *GetDomainInfoParams) (*GetDomainInfo, error) {
	resp, err := c.get("getDomainInfo", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &GetDomainInfo{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// GetPrices Execute operation getPrices.
func (c *Client) GetPrices(params *GetPricesParams) (*GetPrices, error) {
	resp, err := c.get("getPrices", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &GetPrices{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ListDomains Execute operation listDomains.
func (c *Client) ListDomains(params *ListDomainsParams) (*ListDomains, error) {
	resp, err := c.get("listDomains", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ListDomains{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ListEmailForwards Execute operation listEmailForwards.
func (c *Client) ListEmailForwards(params *ListEmailForwardsParams) (*ListEmailForwards, error) {
	resp, err := c.get("listEmailForwards", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ListEmailForwards{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ListOrders Execute operation listOrders.
func (c *Client) ListOrders(params *ListOrdersParams) (*ListOrders, error) {
	resp, err := c.get("listOrders", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ListOrders{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ListRegisteredNameServers Execute operation listRegisteredNameServers.
func (c *Client) ListRegisteredNameServers(params *ListRegisteredNameServersParams) (*ListRegisteredNameServers, error) {
	resp, err := c.get("listRegisteredNameServers", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ListRegisteredNameServers{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// MarketplaceActiveSalesOverview Execute operation marketplaceActiveSalesOverview.
func (c *Client) MarketplaceActiveSalesOverview(params *MarketplaceActiveSalesOverviewParams) (*MarketplaceActiveSalesOverview, error) {
	resp, err := c.get("marketplaceActiveSalesOverview", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &MarketplaceActiveSalesOverview{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// MarketplaceAddOrModifySale Execute operation marketplaceAddOrModifySale.
func (c *Client) MarketplaceAddOrModifySale(params *MarketplaceAddOrModifySaleParams) (*MarketplaceAddOrModifySale, error) {
	resp, err := c.get("marketplaceAddOrModifySale", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &MarketplaceAddOrModifySale{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// MarketplaceLandingPageUpdate Execute operation marketplaceLandingPageUpdate.
func (c *Client) MarketplaceLandingPageUpdate(params *MarketplaceLandingPageUpdateParams) (*MarketplaceLandingPageUpdate, error) {
	resp, err := c.get("marketplaceLandingPageUpdate", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &MarketplaceLandingPageUpdate{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// ModifyRegisteredNameServer Execute operation modifyRegisteredNameServer.
func (c *Client) ModifyRegisteredNameServer(params *ModifyRegisteredNameServerParams) (*ModifyRegisteredNameServer, error) {
	resp, err := c.get("modifyRegisteredNameServer", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &ModifyRegisteredNameServer{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// OrderDetails Execute operation orderDetails.
func (c *Client) OrderDetails(params *OrderDetailsParams) (*OrderDetails, error) {
	resp, err := c.get("orderDetails", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &OrderDetails{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// PortfolioAdd Execute operation portfolioAdd.
func (c *Client) PortfolioAdd(params *PortfolioAddParams) (*PortfolioAdd, error) {
	resp, err := c.get("portfolioAdd", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &PortfolioAdd{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// PortfolioDelete Execute operation portfolioDelete.
func (c *Client) PortfolioDelete(params *PortfolioDeleteParams) (*PortfolioDelete, error) {
	resp, err := c.get("portfolioDelete", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &PortfolioDelete{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// PortfolioDomainAssociate Execute operation portfolioDomainAssociate.
func (c *Client) PortfolioDomainAssociate(params *PortfolioDomainAssociateParams) (*PortfolioDomainAssociate, error) {
	resp, err := c.get("portfolioDomainAssociate", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &PortfolioDomainAssociate{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// PortfolioList Execute operation portfolioList.
func (c *Client) PortfolioList(params *PortfolioListParams) (*PortfolioList, error) {
	resp, err := c.get("portfolioList", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &PortfolioList{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// RegisterDomain Execute operation registerDomain.
func (c *Client) RegisterDomain(params *RegisterDomainParams) (*RegisterDomain, error) {
	resp, err := c.get("registerDomain", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &RegisterDomain{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// RegisterDomainDrop Execute operation registerDomainDrop.
func (c *Client) RegisterDomainDrop(params *RegisterDomainDropParams) (*RegisterDomainDrop, error) {
	resp, err := c.get("registerDomainDrop", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &RegisterDomainDrop{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// RegistrantVerificationStatus Execute operation registrantVerificationStatus.
func (c *Client) RegistrantVerificationStatus(params *RegistrantVerificationStatusParams) (*RegistrantVerificationStatus, error) {
	resp, err := c.get("registrantVerificationStatus", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &RegistrantVerificationStatus{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// RemoveAutoRenewal Execute operation removeAutoRenewal.
func (c *Client) RemoveAutoRenewal(params *RemoveAutoRenewalParams) (*RemoveAutoRenewal, error) {
	resp, err := c.get("removeAutoRenewal", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &RemoveAutoRenewal{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// RemovePrivacy Execute operation removePrivacy.
func (c *Client) RemovePrivacy(params *RemovePrivacyParams) (*RemovePrivacy, error) {
	resp, err := c.get("removePrivacy", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &RemovePrivacy{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// RenewDomain Execute operation renewDomain.
func (c *Client) RenewDomain(params *RenewDomainParams) (*RenewDomain, error) {
	resp, err := c.get("renewDomain", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &RenewDomain{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// RetrieveAuthCode Execute operation retrieveAuthCode.
func (c *Client) RetrieveAuthCode(params *RetrieveAuthCodeParams) (*RetrieveAuthCode, error) {
	resp, err := c.get("retrieveAuthCode", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &RetrieveAuthCode{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// TransferDomain Execute operation transferDomain.
func (c *Client) TransferDomain(params *TransferDomainParams) (*TransferDomain, error) {
	resp, err := c.get("transferDomain", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &TransferDomain{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// TransferUpdateChangeEPPCode Execute operation transferUpdateChangeEPPCode.
func (c *Client) TransferUpdateChangeEPPCode(params *TransferUpdateChangeEPPCodeParams) (*TransferUpdateChangeEPPCode, error) {
	resp, err := c.get("transferUpdateChangeEPPCode", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &TransferUpdateChangeEPPCode{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// TransferUpdateResendAdminEmail Execute operation transferUpdateResendAdminEmail.
func (c *Client) TransferUpdateResendAdminEmail(params *TransferUpdateResendAdminEmailParams) (*TransferUpdateResendAdminEmail, error) {
	resp, err := c.get("transferUpdateResendAdminEmail", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &TransferUpdateResendAdminEmail{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}

// TransferUpdateResubmitToRegistry Execute operation transferUpdateResubmitToRegistry.
func (c *Client) TransferUpdateResubmitToRegistry(params *TransferUpdateResubmitToRegistryParams) (*TransferUpdateResubmitToRegistry, error) {
	resp, err := c.get("transferUpdateResubmitToRegistry", params)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %v", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	op := &TransferUpdateResubmitToRegistry{}
	err = xml.Unmarshal(bytes, op)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %v: %s", err, bytes)
	}

	switch op.Reply.Code {
	case SuccessfulAPIOperation:
		// Successful API operation
		return op, nil
	case SuccessfulRegistration:
		// Successful registration, but not all provided hosts were valid resulting in our nameservers being used
		return op, nil
	case SuccessfulOrder:
		// Successful order, but there was an error with the contact information provided so your account default contact profile was used
		return op, nil
	default:
		// error
		return op, fmt.Errorf("code: %s, details: %s", op.Reply.Code, op.Reply.Detail)
	}
}
