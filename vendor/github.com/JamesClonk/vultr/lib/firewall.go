package lib

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// FirewallGroup represents a firewall group on Vultr
type FirewallGroup struct {
	ID            string `json:"FIREWALLGROUPID"`
	Description   string `json:"description"`
	Created       string `json:"date_created"`
	Modified      string `json:"date_modified"`
	InstanceCount int    `json:"instance_count"`
	RuleCount     int    `json:"rule_count"`
	MaxRuleCount  int    `json:"max_rule_count"`
}

// FirewallRule represents a firewall rule on Vultr
type FirewallRule struct {
	RuleNumber int    `json:"rulenumber"`
	Action     string `json:"action"`
	Protocol   string `json:"protocol"`
	Port       string `json:"port"`
	Network    *net.IPNet
}

type firewallGroups []FirewallGroup

func (f firewallGroups) Len() int      { return len(f) }
func (f firewallGroups) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f firewallGroups) Less(i, j int) bool {
	// sort order: description
	return strings.ToLower(f[i].Description) < strings.ToLower(f[j].Description)
}

type firewallRules []FirewallRule

func (r firewallRules) Len() int      { return len(r) }
func (r firewallRules) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r firewallRules) Less(i, j int) bool {
	// sort order: rule number
	return r[i].RuleNumber < r[j].RuleNumber
}

// UnmarshalJSON implements json.Unmarshaller on FirewallRule.
// This is needed because the Vultr API is inconsistent in it's JSON responses.
// Some fields can change type, from JSON number to JSON string and vice-versa.
func (r *FirewallRule) UnmarshalJSON(data []byte) (err error) {
	if r == nil {
		*r = FirewallRule{}
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	value := fmt.Sprintf("%v", fields["rulenumber"])
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	r.RuleNumber = int(number)

	value = fmt.Sprintf("%v", fields["subnet_size"])
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}
	subnetSize, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	r.Action = fmt.Sprintf("%v", fields["action"])
	r.Protocol = fmt.Sprintf("%v", fields["protocol"])
	r.Port = fmt.Sprintf("%v", fields["port"])
	subnet := fmt.Sprintf("%v", fields["subnet"])

	if subnetSize > 0 && len(subnet) > 0 {
		_, r.Network, err = net.ParseCIDR(fmt.Sprintf("%s/%d", subnet, subnetSize))
		if err != nil {
			return fmt.Errorf("Failed to parse subnet from Vultr API")
		}
	} else {
		_, r.Network, _ = net.ParseCIDR("0.0.0.0/0")
	}

	return
}

// GetFirewallGroups returns a list of all available firewall groups on Vultr
func (c *Client) GetFirewallGroups() ([]FirewallGroup, error) {
	var groupMap map[string]FirewallGroup
	if err := c.get(`firewall/group_list`, &groupMap); err != nil {
		return nil, err
	}

	var groupList []FirewallGroup
	for _, g := range groupMap {
		groupList = append(groupList, g)
	}
	sort.Sort(firewallGroups(groupList))
	return groupList, nil
}

// GetFirewallGroup returns the firewall group with given ID
func (c *Client) GetFirewallGroup(id string) (FirewallGroup, error) {
	groups, err := c.GetFirewallGroups()
	if err != nil {
		return FirewallGroup{}, err
	}

	for _, g := range groups {
		if g.ID == id {
			return g, nil
		}
	}
	return FirewallGroup{}, fmt.Errorf("Firewall group with ID %v not found", id)
}

// CreateFirewallGroup creates a new firewall group in Vultr account
func (c *Client) CreateFirewallGroup(description string) (string, error) {
	values := url.Values{}

	// Optional description
	if len(description) > 0 {
		values.Add("description", description)
	}

	var result FirewallGroup
	err := c.post(`firewall/group_create`, values, &result)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

// DeleteFirewallGroup deletes an existing firewall group
func (c *Client) DeleteFirewallGroup(groupID string) error {
	values := url.Values{
		"FIREWALLGROUPID": {groupID},
	}

	if err := c.post(`firewall/group_delete`, values, nil); err != nil {
		return err
	}
	return nil
}

// SetFirewallGroupDescription sets the description of an existing firewall group
func (c *Client) SetFirewallGroupDescription(groupID, description string) error {
	values := url.Values{
		"FIREWALLGROUPID": {groupID},
		"description":     {description},
	}

	if err := c.post(`firewall/group_set_description`, values, nil); err != nil {
		return err
	}
	return nil
}

// GetFirewallRules returns a list of rules for the given firewall group
func (c *Client) GetFirewallRules(groupID string) ([]FirewallRule, error) {
	var ruleMap map[string]FirewallRule
	ipTypes := []string{"v4", "v6"}
	for _, ipType := range ipTypes {
		args := fmt.Sprintf("direction=in&FIREWALLGROUPID=%s&ip_type=%s",
			groupID, ipType)
		if err := c.get(`firewall/rule_list?`+args, &ruleMap); err != nil {
			return nil, err
		}
	}

	var ruleList []FirewallRule
	for _, r := range ruleMap {
		ruleList = append(ruleList, r)
	}
	sort.Sort(firewallRules(ruleList))
	return ruleList, nil
}

// CreateFirewallRule creates a new firewall rule in Vultr account.
// groupID is the ID of the firewall group to create the rule in
// protocol must be one of: "icmp", "tcp", "udp", "gre"
// port can be a port number or colon separated port range (TCP/UDP only)
func (c *Client) CreateFirewallRule(groupID, protocol, port string,
	network *net.IPNet) (int, error) {
	ip := network.IP.String()
	maskBits, _ := network.Mask.Size()
	if ip == "<nil>" {
		return 0, fmt.Errorf("Invalid network")
	}

	var ipType string
	if network.IP.To4() != nil {
		ipType = "v4"
	} else {
		ipType = "v6"
	}

	values := url.Values{
		"FIREWALLGROUPID": {groupID},
		// possible values: "in"
		"direction": {"in"},
		// possible values: "icmp", "tcp", "udp", "gre"
		"protocol": {protocol},
		// possible values: "v4", "v6"
		"ip_type": {ipType},
		// IP address representing a subnet
		"subnet": {ip},
		// IP prefix size in bits
		"subnet_size": {fmt.Sprintf("%v", maskBits)},
	}

	if len(port) > 0 {
		values.Add("port", port)
	}

	var result FirewallRule
	err := c.post(`firewall/rule_create`, values, &result)
	if err != nil {
		return 0, err
	}
	return result.RuleNumber, nil
}

// DeleteFirewallRule deletes an existing firewall rule
func (c *Client) DeleteFirewallRule(ruleNumber int, groupID string) error {
	values := url.Values{
		"FIREWALLGROUPID": {groupID},
		"rulenumber":      {fmt.Sprintf("%v", ruleNumber)},
	}

	if err := c.post(`firewall/rule_delete`, values, nil); err != nil {
		return err
	}
	return nil
}
