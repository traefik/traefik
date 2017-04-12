package lib

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// IP on Vultr
type IP struct {
	ID         string `json:"SUBID,string"`
	RegionID   int    `json:"DCID,string"`
	IPType     string `json:"ip_type"`
	Subnet     string `json:"subnet"`
	SubnetSize int    `json:"subnet_size"`
	Label      string `json:"label"`
	AttachedTo string `json:"attached_SUBID,string"`
}

type ips []IP

func (s ips) Len() int      { return len(s) }
func (s ips) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ips) Less(i, j int) bool {
	// sort order: label, iptype, subnet
	if strings.ToLower(s[i].Label) < strings.ToLower(s[j].Label) {
		return true
	} else if strings.ToLower(s[i].Label) > strings.ToLower(s[j].Label) {
		return false
	}
	if s[i].IPType < s[j].IPType {
		return true
	} else if s[i].IPType > s[j].IPType {
		return false
	}
	return s[i].Subnet < s[j].Subnet
}

// UnmarshalJSON implements json.Unmarshaller on IP.
// This is needed because the Vultr API is inconsistent in it's JSON responses.
// Some fields can change type, from JSON number to JSON string and vice-versa.
func (i *IP) UnmarshalJSON(data []byte) (err error) {
	if i == nil {
		*i = IP{}
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	value := fmt.Sprintf("%v", fields["SUBID"])
	if len(value) == 0 || value == "<nil>" || value == "0" {
		i.ID = ""
	} else {
		id, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		i.ID = strconv.FormatFloat(id, 'f', -1, 64)
	}

	value = fmt.Sprintf("%v", fields["DCID"])
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}
	region, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	i.RegionID = int(region)

	value = fmt.Sprintf("%v", fields["attached_SUBID"])
	if len(value) == 0 || value == "<nil>" || value == "0" || value == "false" {
		i.AttachedTo = ""
	} else {
		attached, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		i.AttachedTo = strconv.FormatFloat(attached, 'f', -1, 64)
	}

	value = fmt.Sprintf("%v", fields["subnet_size"])
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}
	size, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	i.SubnetSize = int(size)

	i.IPType = fmt.Sprintf("%v", fields["ip_type"])
	i.Subnet = fmt.Sprintf("%v", fields["subnet"])
	i.Label = fmt.Sprintf("%v", fields["label"])

	return
}

// ListReservedIP returns a list of all available reserved IPs on Vultr account
func (c *Client) ListReservedIP() ([]IP, error) {
	var ipMap map[string]IP

	err := c.get(`reservedip/list`, &ipMap)
	if err != nil {
		return nil, err
	}

	ipList := make([]IP, 0)
	for _, ip := range ipMap {
		ipList = append(ipList, ip)
	}
	sort.Sort(ips(ipList))
	return ipList, nil
}

// GetReservedIP returns reserved IP with given ID
func (c *Client) GetReservedIP(id string) (IP, error) {
	var ipMap map[string]IP

	err := c.get(`reservedip/list`, &ipMap)
	if err != nil {
		return IP{}, err
	}
	if ip, ok := ipMap[id]; ok {
		return ip, nil
	}
	return IP{}, fmt.Errorf("IP with ID %v not found", id)
}

// CreateReservedIP creates a new reserved IP on Vultr account
func (c *Client) CreateReservedIP(regionID int, ipType string, label string) (string, error) {
	values := url.Values{
		"DCID":    {fmt.Sprintf("%v", regionID)},
		"ip_type": {ipType},
	}
	if len(label) > 0 {
		values.Add("label", label)
	}

	result := IP{}
	err := c.post(`reservedip/create`, values, &result)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

// DestroyReservedIP deletes an existing reserved IP
func (c *Client) DestroyReservedIP(id string) error {
	values := url.Values{
		"SUBID": {id},
	}
	return c.post(`reservedip/destroy`, values, nil)
}

// AttachReservedIP attaches a reserved IP to a virtual machine
func (c *Client) AttachReservedIP(ip string, serverID string) error {
	values := url.Values{
		"ip_address":   {ip},
		"attach_SUBID": {serverID},
	}
	return c.post(`reservedip/attach`, values, nil)
}

// DetachReservedIP detaches a reserved IP from an existing virtual machine
func (c *Client) DetachReservedIP(serverID string, ip string) error {
	values := url.Values{
		"ip_address":   {ip},
		"detach_SUBID": {serverID},
	}
	return c.post(`reservedip/detach`, values, nil)
}

// ConvertReservedIP converts an existing virtual machines IP to a reserved IP
func (c *Client) ConvertReservedIP(serverID string, ip string) (string, error) {
	values := url.Values{
		"SUBID":      {serverID},
		"ip_address": {ip},
	}

	result := IP{}
	err := c.post(`reservedip/convert`, values, &result)
	if err != nil {
		return "", err
	}
	return result.ID, err
}
