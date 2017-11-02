package lib

import (
	"fmt"
	"net/url"
	"sort"
)

// IPv4 information of a virtual machine
type IPv4 struct {
	IP         string `json:"ip"`
	Netmask    string `json:"netmask"`
	Gateway    string `json:"gateway"`
	Type       string `json:"type"`
	ReverseDNS string `json:"reverse"`
}

type ipv4s []IPv4

func (s ipv4s) Len() int      { return len(s) }
func (s ipv4s) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ipv4s) Less(i, j int) bool {
	// sort order: type, ip
	if s[i].Type < s[j].Type {
		return true
	} else if s[i].Type > s[j].Type {
		return false
	}
	return s[i].IP < s[j].IP
}

// IPv6 information of a virtual machine
type IPv6 struct {
	IP          string `json:"ip"`
	Network     string `json:"network"`
	NetworkSize string `json:"network_size"`
	Type        string `json:"type"`
}

type ipv6s []IPv6

func (s ipv6s) Len() int      { return len(s) }
func (s ipv6s) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ipv6s) Less(i, j int) bool {
	// sort order: type, ip
	if s[i].Type < s[j].Type {
		return true
	} else if s[i].Type > s[j].Type {
		return false
	}
	return s[i].IP < s[j].IP
}

// ReverseDNSIPv6 information of a virtual machine
type ReverseDNSIPv6 struct {
	IP         string `json:"ip"`
	ReverseDNS string `json:"reverse"`
}

type reverseDNSIPv6s []ReverseDNSIPv6

func (s reverseDNSIPv6s) Len() int           { return len(s) }
func (s reverseDNSIPv6s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s reverseDNSIPv6s) Less(i, j int) bool { return s[i].IP < s[j].IP }

// ListIPv4 lists the IPv4 information of a virtual machine
func (c *Client) ListIPv4(id string) (list []IPv4, err error) {
	var ipMap map[string][]IPv4
	if err := c.get(`server/list_ipv4?SUBID=`+id, &ipMap); err != nil {
		return nil, err
	}

	for _, iplist := range ipMap {
		for _, ip := range iplist {
			list = append(list, ip)
		}
	}
	sort.Sort(ipv4s(list))
	return list, nil
}

// CreateIPv4 creates an IPv4 address and attaches it to a virtual machine
func (c *Client) CreateIPv4(id string, reboot bool) error {
	values := url.Values{
		"SUBID":  {id},
		"reboot": {fmt.Sprintf("%t", reboot)},
	}

	if err := c.post(`server/create_ipv4`, values, nil); err != nil {
		return err
	}
	return nil
}

// DeleteIPv4 deletes an IPv4 address and detaches it from a virtual machine
func (c *Client) DeleteIPv4(id, ip string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
	}

	if err := c.post(`server/destroy_ipv4`, values, nil); err != nil {
		return err
	}
	return nil
}

// ListIPv6 lists the IPv4 information of a virtual machine
func (c *Client) ListIPv6(id string) (list []IPv6, err error) {
	var ipMap map[string][]IPv6
	if err := c.get(`server/list_ipv6?SUBID=`+id, &ipMap); err != nil {
		return nil, err
	}

	for _, iplist := range ipMap {
		for _, ip := range iplist {
			list = append(list, ip)
		}
	}
	sort.Sort(ipv6s(list))
	return list, nil
}

// ListIPv6ReverseDNS lists the IPv6 reverse DNS entries of a virtual machine
func (c *Client) ListIPv6ReverseDNS(id string) (list []ReverseDNSIPv6, err error) {
	var ipMap map[string][]ReverseDNSIPv6
	if err := c.get(`server/reverse_list_ipv6?SUBID=`+id, &ipMap); err != nil {
		return nil, err
	}

	for _, iplist := range ipMap {
		for _, ip := range iplist {
			list = append(list, ip)
		}
	}
	sort.Sort(reverseDNSIPv6s(list))
	return list, nil
}

// DeleteIPv6ReverseDNS removes a reverse DNS entry for an IPv6 address of a virtual machine
func (c *Client) DeleteIPv6ReverseDNS(id string, ip string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
	}

	if err := c.post(`server/reverse_delete_ipv6`, values, nil); err != nil {
		return err
	}
	return nil
}

// SetIPv6ReverseDNS sets a reverse DNS entry for an IPv6 address of a virtual machine
func (c *Client) SetIPv6ReverseDNS(id, ip, entry string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
		"entry": {entry},
	}

	if err := c.post(`server/reverse_set_ipv6`, values, nil); err != nil {
		return err
	}
	return nil
}

// DefaultIPv4ReverseDNS sets a reverse DNS entry for an IPv4 address of a virtual machine to the original setting
func (c *Client) DefaultIPv4ReverseDNS(id, ip string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
	}

	if err := c.post(`server/reverse_default_ipv4`, values, nil); err != nil {
		return err
	}
	return nil
}

// SetIPv4ReverseDNS sets a reverse DNS entry for an IPv4 address of a virtual machine
func (c *Client) SetIPv4ReverseDNS(id, ip, entry string) error {
	values := url.Values{
		"SUBID": {id},
		"ip":    {ip},
		"entry": {entry},
	}

	if err := c.post(`server/reverse_set_ipv4`, values, nil); err != nil {
		return err
	}
	return nil
}
