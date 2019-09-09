package govultr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ReservedIPService is the interface to interact with the reserved IP endpoints on the Vultr API
// Link: https://www.vultr.com/api/#reservedip
type ReservedIPService interface {
	Attach(ctx context.Context, ip, InstanceID string) error
	Convert(ctx context.Context, ip, InstanceID, label string) (*ReservedIP, error)
	Create(ctx context.Context, regionID int, ipType, label string) (*ReservedIP, error)
	Delete(ctx context.Context, ip string) error
	Detach(ctx context.Context, ip, InstanceID string) error
	List(ctx context.Context) ([]ReservedIP, error)
}

// ReservedIPServiceHandler handles interaction with the reserved IP methods for the Vultr API
type ReservedIPServiceHandler struct {
	client *Client
}

// ReservedIP represents an reserved IP on Vultr
type ReservedIP struct {
	ReservedIPID string `json:"SUBID"`
	RegionID     int    `json:"DCID"`
	IPType       string `json:"ip_type"`
	Subnet       string `json:"subnet"`
	SubnetSize   int    `json:"subnet_size"`
	Label        string `json:"label"`
	AttachedID   string `json:"attached_SUBID"`
}

// UnmarshalJSON implements json.Unmarshaller on ReservedIP to handle the inconsistent types returned from the Vultr API.
func (r *ReservedIP) UnmarshalJSON(data []byte) (err error) {
	if r == nil {
		*r = ReservedIP{}
	}

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	r.ReservedIPID, err = r.unmarshalStr(fmt.Sprintf("%v", v["SUBID"]))
	if err != nil {
		return err
	}

	r.AttachedID, err = r.unmarshalStr(fmt.Sprintf("%v", v["attached_SUBID"]))
	if err != nil {
		return err
	}

	r.RegionID, err = r.unmarshalInt(fmt.Sprintf("%v", v["DCID"]))
	if err != nil {
		return err
	}

	r.SubnetSize, err = r.unmarshalInt(fmt.Sprintf("%v", v["subnet_size"]))
	if err != nil {
		return err
	}

	if r.Subnet = fmt.Sprintf("%v", v["subnet"]); r.Subnet == "<nil>" {
		r.Subnet = ""
	}

	if r.IPType = fmt.Sprintf("%v", v["ip_type"]); r.IPType == "<nil>" {
		r.IPType = ""
	}

	if r.Label = fmt.Sprintf("%v", v["label"]); r.Label == "<nil>" {
		r.Label = ""
	}

	return nil
}

func (r *ReservedIP) unmarshalInt(value string) (int, error) {
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func (r *ReservedIP) unmarshalStr(value string) (string, error) {
	if len(value) == 0 || value == "<nil>" || value == "0" || value == "false" {
		return "", nil
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}

	return strconv.FormatFloat(f, 'f', -1, 64), nil
}

// Attach a reserved IP to an existing subscription
func (r *ReservedIPServiceHandler) Attach(ctx context.Context, ip, InstanceID string) error {
	uri := "/v1/reservedip/attach"

	values := url.Values{
		"ip_address":   {ip},
		"attach_SUBID": {InstanceID},
	}

	req, err := r.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = r.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Convert an existing IP on a subscription to a reserved IP.
func (r *ReservedIPServiceHandler) Convert(ctx context.Context, ip, InstanceID, label string) (*ReservedIP, error) {
	uri := "/v1/reservedip/convert"

	values := url.Values{
		"SUBID":      {InstanceID},
		"ip_address": {ip},
	}

	if label != "" {
		values.Add("label", label)
	}

	req, err := r.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	rip := new(ReservedIP)

	err = r.client.DoWithContext(ctx, req, rip)

	if err != nil {
		return nil, err
	}

	rip.Label = label

	return rip, nil
}

// Create adds the specified reserved IP to your Vultr account
func (r *ReservedIPServiceHandler) Create(ctx context.Context, regionID int, ipType, label string) (*ReservedIP, error) {

	uri := "/v1/reservedip/create"

	values := url.Values{
		"DCID":    {strconv.Itoa(regionID)},
		"ip_type": {ipType},
	}

	if label != "" {
		values.Add("label", label)
	}

	req, err := r.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	rip := new(ReservedIP)

	err = r.client.DoWithContext(ctx, req, rip)

	if err != nil {
		return nil, err
	}

	rip.RegionID = regionID
	rip.IPType = ipType
	rip.Label = label

	return rip, nil
}

// Delete removes the specified reserved IP from your Vultr account
func (r *ReservedIPServiceHandler) Delete(ctx context.Context, ip string) error {

	uri := "/v1/reservedip/destroy"

	values := url.Values{
		"ip_address": {ip},
	}

	req, err := r.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = r.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Detach a reserved IP from an existing subscription.
func (r *ReservedIPServiceHandler) Detach(ctx context.Context, ip, InstanceID string) error {
	uri := "/v1/reservedip/detach"

	values := url.Values{
		"ip_address":   {ip},
		"detach_SUBID": {InstanceID},
	}

	req, err := r.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = r.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List lists all the reserved IPs associated with your Vultr account
func (r *ReservedIPServiceHandler) List(ctx context.Context) ([]ReservedIP, error) {

	uri := "/v1/reservedip/list"

	req, err := r.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	ipMap := make(map[string]ReservedIP)
	err = r.client.DoWithContext(ctx, req, &ipMap)
	if err != nil {
		return nil, err
	}

	var ips []ReservedIP
	for _, ip := range ipMap {
		ips = append(ips, ip)
	}

	return ips, nil
}
