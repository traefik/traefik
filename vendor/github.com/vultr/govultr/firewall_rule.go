package govultr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

// FireWallRuleService is the interface to interact with the firewall rule endpoints on the Vultr API
// Link: https://www.vultr.com/api/#firewall
type FireWallRuleService interface {
	Create(ctx context.Context, groupID, protocol, port, network, notes string) (*FirewallRule, error)
	Delete(ctx context.Context, groupID, ruleID string) error
	ListByIPType(ctx context.Context, groupID, ipType string) ([]FirewallRule, error)
	List(ctx context.Context, groupID string) ([]FirewallRule, error)
}

// FireWallRuleServiceHandler handles interaction with the firewall rule methods for the Vultr API
type FireWallRuleServiceHandler struct {
	client *Client
}

// FirewallRule represents a Vultr firewall rule
type FirewallRule struct {
	RuleNumber int        `json:"rulenumber"`
	Action     string     `json:"action"`
	Protocol   string     `json:"protocol"`
	Port       string     `json:"port"`
	Network    *net.IPNet `json:"network"`
	Notes      string     `json:"notes"`
}

// UnmarshalJSON implements a custom unmarshaler on FirewallRule
// This is done to help reduce data inconsistency with V1 of the Vultr API
// It also merges the subnet & subnet_mask into a single type of *net.IPNet
func (f *FirewallRule) UnmarshalJSON(data []byte) (err error) {
	if f == nil {
		*f = FirewallRule{}
	}

	// Pull out all of the data that was given to us and put it into a map
	var fields map[string]interface{}
	err = json.Unmarshal(data, &fields)

	if err != nil {
		return err
	}

	// Unmarshal RuleNumber
	value := fmt.Sprintf("%v", fields["rulenumber"])
	number, _ := strconv.Atoi(value)
	f.RuleNumber = number

	// Unmarshal all other strings

	action := fmt.Sprintf("%v", fields["action"])
	if action == "<nil>" {
		action = ""
	}
	f.Action = action

	protocol := fmt.Sprintf("%v", fields["protocol"])
	if protocol == "<nil>" {
		protocol = ""
	}
	f.Protocol = protocol

	port := fmt.Sprintf("%v", fields["port"])
	if port == "<nil>" {
		port = ""
	}
	f.Port = port

	notes := fmt.Sprintf("%v", fields["notes"])
	if notes == "<nil>" {
		notes = ""
	}
	f.Notes = notes

	// Unmarshal subnet_size & subnet and convert to *net.IP
	value = fmt.Sprintf("%v", fields["subnet_size"])
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}
	subnetSize, _ := strconv.Atoi(value)

	subnet := fmt.Sprintf("%v", fields["subnet"])
	if subnet == "<nil>" {
		subnet = ""
	}

	if len(subnet) > 0 {
		_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", subnet, subnetSize))

		if err != nil {
			return errors.New("an issue has occurred while parsing subnet")
		}

		f.Network = ipNet
	}

	return
}

// Create will create a rule in a firewall group.
func (f *FireWallRuleServiceHandler) Create(ctx context.Context, groupID, protocol, port, cdirBlock, notes string) (*FirewallRule, error) {

	uri := "/v1/firewall/rule_create"

	ip, ipNet, err := net.ParseCIDR(cdirBlock)

	if err != nil {
		return nil, err
	}

	values := url.Values{
		"FIREWALLGROUPID": {groupID},
		"direction":       {"in"},
		"protocol":        {protocol},
		"subnet":          {ip.String()},
	}

	// mask
	mask, _ := ipNet.Mask.Size()
	values.Add("subnet_size", strconv.Itoa(mask))

	// ip Type
	if ipNet.IP.To4() != nil {
		values.Add("ip_type", "v4")
	} else {
		values.Add("ip_type", "v6")
	}

	// Optional params
	if port != "" {
		values.Add("port", port)
	}

	if notes != "" {
		values.Add("notes", notes)
	}

	req, err := f.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	firewallRule := new(FirewallRule)
	err = f.client.DoWithContext(ctx, req, firewallRule)

	if err != nil {
		return nil, err
	}

	return firewallRule, nil
}

// Delete will delete a firewall rule on your Vultr account
func (f *FireWallRuleServiceHandler) Delete(ctx context.Context, groupID, ruleID string) error {

	uri := "/v1/firewall/rule_delete"

	values := url.Values{
		"FIREWALLGROUPID": {groupID},
		"rulenumber":      {ruleID},
	}

	req, err := f.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = f.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List will list the current firewall rules in a firewall group.
// ipType values that can be passed in are "v4", "v6"
func (f *FireWallRuleServiceHandler) ListByIPType(ctx context.Context, groupID, ipType string) ([]FirewallRule, error) {

	uri := "/v1/firewall/rule_list"

	req, err := f.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("FIREWALLGROUPID", groupID)
	q.Add("direction", "in")
	q.Add("ip_type", ipType)
	req.URL.RawQuery = q.Encode()
	var firewallRuleMap map[string]FirewallRule

	err = f.client.DoWithContext(ctx, req, &firewallRuleMap)

	if err != nil {
		return nil, err
	}

	var firewallRule []FirewallRule
	for _, f := range firewallRuleMap {
		firewallRule = append(firewallRule, f)
	}

	return firewallRule, nil
}

// List will return both ipv4 an ipv6 firewall rules that are defined within a firewall group
func (f *FireWallRuleServiceHandler) List(ctx context.Context, groupID string) ([]FirewallRule, error) {
	uri := "/v1/firewall/rule_list"

	req, err := f.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("FIREWALLGROUPID", groupID)
	q.Add("direction", "in")
	q.Add("ip_type", "v4")

	req.URL.RawQuery = q.Encode()

	var firewallRuleMap map[string]FirewallRule

	// V4 call
	err = f.client.DoWithContext(ctx, req, &firewallRuleMap)

	if err != nil {
		return nil, err
	}

	// V6 call
	q.Del("ip_type")
	q.Add("ip_type", "v6")
	req.URL.RawQuery = q.Encode()

	err = f.client.DoWithContext(ctx, req, &firewallRuleMap)

	if err != nil {
		return nil, err
	}

	var firewallRule []FirewallRule
	for _, f := range firewallRuleMap {
		firewallRule = append(firewallRule, f)
	}

	return firewallRule, nil
}
