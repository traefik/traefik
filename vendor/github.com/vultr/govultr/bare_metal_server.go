package govultr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// BareMetalServerService is the interface to interact with the bare metal endpoints on the Vultr API
// Link: https://www.vultr.com/api/#baremetal
type BareMetalServerService interface {
	AppInfo(ctx context.Context, serverID string) (*AppInfo, error)
	Bandwidth(ctx context.Context, serverID string) ([]map[string]string, error)
	ChangeApp(ctx context.Context, serverID, appID string) error
	ChangeOS(ctx context.Context, serverID, osID string) error
	Create(ctx context.Context, regionID, planID, osID string, options *BareMetalServerOptions) (*BareMetalServer, error)
	Delete(ctx context.Context, serverID string) error
	EnableIPV6(ctx context.Context, serverID string) error
	List(ctx context.Context) ([]BareMetalServer, error)
	ListByLabel(ctx context.Context, label string) ([]BareMetalServer, error)
	ListByMainIP(ctx context.Context, mainIP string) ([]BareMetalServer, error)
	ListByTag(ctx context.Context, tag string) ([]BareMetalServer, error)
	GetServer(ctx context.Context, serverID string) (*BareMetalServer, error)
	GetUserData(ctx context.Context, serverID string) (*UserData, error)
	Halt(ctx context.Context, serverID string) error
	IPV4Info(ctx context.Context, serverID string) ([]BareMetalServerIPV4, error)
	IPV6Info(ctx context.Context, serverID string) ([]BareMetalServerIPV6, error)
	ListApps(ctx context.Context, serverID string) ([]Application, error)
	ListOS(ctx context.Context, serverID string) ([]OS, error)
	Reboot(ctx context.Context, serverID string) error
	Reinstall(ctx context.Context, serverID string) error
	SetLabel(ctx context.Context, serverID, label string) error
	SetTag(ctx context.Context, serverID, tag string) error
	SetUserData(ctx context.Context, serverID, userData string) error
}

// BareMetalServerServiceHandler handles interaction with the bare metal methods for the Vultr API
type BareMetalServerServiceHandler struct {
	client *Client
}

// BareMetalServer represents a bare metal server on Vultr
type BareMetalServer struct {
	BareMetalServerID string      `json:"SUBID"`
	Os                string      `json:"os"`
	RAM               string      `json:"ram"`
	Disk              string      `json:"disk"`
	MainIP            string      `json:"main_ip"`
	CPUs              int         `json:"cpu_count"`
	Location          string      `json:"location"`
	RegionID          int         `json:"DCID"`
	DefaultPassword   string      `json:"default_password"`
	DateCreated       string      `json:"date_created"`
	Status            string      `json:"status"`
	NetmaskV4         string      `json:"netmask_v4"`
	GatewayV4         string      `json:"gateway_v4"`
	BareMetalPlanID   int         `json:"METALPLANID"`
	V6Networks        []V6Network `json:"v6_networks"`
	Label             string      `json:"label"`
	Tag               string      `json:"tag"`
	OsID              string      `json:"OSID"`
	AppID             string      `json:"APPID"`
}

// BareMetalServerOptions represents the optional parameters that can be set when creating a bare metal server
type BareMetalServerOptions struct {
	StartupScriptID string
	SnapshotID      string
	EnableIPV6      string
	Label           string
	SSHKeyIDs       []string
	AppID           string
	UserData        string
	NotifyActivate  string
	Hostname        string
	Tag             string
	ReservedIPV4    string
}

// BareMetalServerIPV4 represents IPV4 information for a bare metal server
type BareMetalServerIPV4 struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
	Type    string `json:"type"`
}

// BareMetalServerIPV6 represents IPV6 information for a bare metal server
type BareMetalServerIPV6 struct {
	IP          string `json:"ip"`
	Network     string `json:"network"`
	NetworkSize int    `json:"network_size"`
	Type        string `json:"type"`
}

// UnmarshalJSON implements a custom unmarshaler on BareMetalServer
// This is done to help reduce data inconsistency with V1 of the Vultr API
func (b *BareMetalServer) UnmarshalJSON(data []byte) error {
	if b == nil {
		*b = BareMetalServer{}
	}

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	cpu, err := b.unmarshalInt(fmt.Sprintf("%v", v["cpu_count"]))
	if err != nil {
		return err
	}
	b.CPUs = cpu

	region, err := b.unmarshalInt(fmt.Sprintf("%v", v["DCID"]))
	if err != nil {
		return err
	}
	b.RegionID = region

	plan, err := b.unmarshalInt(fmt.Sprintf("%v", v["METALPLANID"]))
	if err != nil {
		return err
	}
	b.BareMetalPlanID = plan

	b.BareMetalServerID = b.unmarshalStr(fmt.Sprintf("%v", v["SUBID"]))
	b.Os = b.unmarshalStr(fmt.Sprintf("%v", v["os"]))
	b.RAM = b.unmarshalStr(fmt.Sprintf("%v", v["ram"]))
	b.Label = b.unmarshalStr(fmt.Sprintf("%v", v["label"]))
	b.Disk = b.unmarshalStr(fmt.Sprintf("%v", v["disk"]))
	b.MainIP = b.unmarshalStr(fmt.Sprintf("%v", v["main_ip"]))
	b.Location = b.unmarshalStr(fmt.Sprintf("%v", v["location"]))
	b.DefaultPassword = b.unmarshalStr(fmt.Sprintf("%v", v["default_password"]))
	b.DateCreated = b.unmarshalStr(fmt.Sprintf("%v", v["date_created"]))
	b.Status = b.unmarshalStr(fmt.Sprintf("%v", v["status"]))
	b.NetmaskV4 = b.unmarshalStr(fmt.Sprintf("%v", v["netmask_v4"]))
	b.GatewayV4 = b.unmarshalStr(fmt.Sprintf("%v", v["gateway_v4"]))
	b.Tag = b.unmarshalStr(fmt.Sprintf("%v", v["tag"]))
	b.OsID = b.unmarshalStr(fmt.Sprintf("%v", v["OSID"]))
	b.AppID = b.unmarshalStr(fmt.Sprintf("%v", v["APPID"]))

	v6networks := make([]V6Network, 0)
	if networks, ok := v["v6_networks"].([]interface{}); ok {
		for _, network := range networks {
			if network, ok := network.(map[string]interface{}); ok {
				v6network := V6Network{
					Network:     fmt.Sprintf("%v", network["v6_network"]),
					MainIP:      fmt.Sprintf("%v", network["v6_main_ip"]),
					NetworkSize: fmt.Sprintf("%v", network["v6_network_size"]),
				}
				v6networks = append(v6networks, v6network)
			}
		}
		b.V6Networks = v6networks
	}

	return nil
}

func (b *BareMetalServer) unmarshalInt(value string) (int, error) {
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func (b *BareMetalServer) unmarshalStr(value string) string {
	if value == "<nil>" {
		value = ""
	}

	return value
}

// AppInfo retrieves the application information for a given server ID
func (b *BareMetalServerServiceHandler) AppInfo(ctx context.Context, serverID string) (*AppInfo, error) {
	uri := "/v1/baremetal/get_app_info"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)
	req.URL.RawQuery = q.Encode()

	appInfo := new(AppInfo)

	err = b.client.DoWithContext(ctx, req, appInfo)

	if err != nil {
		return nil, err
	}

	return appInfo, nil
}

// Bandwidth will get the bandwidth used by a bare metal server
func (b *BareMetalServerServiceHandler) Bandwidth(ctx context.Context, serverID string) ([]map[string]string, error) {
	uri := "/v1/baremetal/bandwidth"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)
	req.URL.RawQuery = q.Encode()

	var bandwidthMap map[string][][]interface{}
	err = b.client.DoWithContext(ctx, req, &bandwidthMap)

	if err != nil {
		return nil, err
	}

	var bandwidth []map[string]string

	for _, b := range bandwidthMap["incoming_bytes"] {
		inMap := make(map[string]string)
		inMap["date"] = fmt.Sprintf("%v", b[0])
		var bytes int64
		switch b[1].(type) {
		case float64:
			bytes = int64(b[1].(float64))
		case int64:
			bytes = b[1].(int64)
		}
		inMap["incoming"] = fmt.Sprintf("%v", bytes)
		bandwidth = append(bandwidth, inMap)
	}

	for _, b := range bandwidthMap["outgoing_bytes"] {
		for i := range bandwidth {
			if bandwidth[i]["date"] == b[0] {
				var bytes int64
				switch b[1].(type) {
				case float64:
					bytes = int64(b[1].(float64))
				case int64:
					bytes = b[1].(int64)
				}
				bandwidth[i]["outgoing"] = fmt.Sprintf("%v", bytes)
				break
			}
		}
	}

	return bandwidth, nil
}

// ChangeApp changes the bare metal server to a different application.
func (b *BareMetalServerServiceHandler) ChangeApp(ctx context.Context, serverID, appID string) error {
	uri := "/v1/baremetal/app_change"

	values := url.Values{
		"SUBID": {serverID},
		"APPID": {appID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// ChangeOS changes the bare metal server to a different operating system. All data will be permanently lost.
func (b *BareMetalServerServiceHandler) ChangeOS(ctx context.Context, serverID, osID string) error {
	uri := "/v1/baremetal/os_change"

	values := url.Values{
		"SUBID": {serverID},
		"OSID":  {osID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Create a new bare metal server.
func (b *BareMetalServerServiceHandler) Create(ctx context.Context, regionID, planID, osID string, options *BareMetalServerOptions) (*BareMetalServer, error) {
	uri := "/v1/baremetal/create"

	values := url.Values{
		"DCID":        {regionID},
		"METALPLANID": {planID},
		"OSID":        {osID},
	}

	if options != nil {
		if options.StartupScriptID != "" {
			values.Add("SCRIPTID", options.StartupScriptID)
		}
		if options.SnapshotID != "" {
			values.Add("SNAPSHOTID", options.SnapshotID)
		}
		if options.EnableIPV6 != "" {
			values.Add("enable_ipv6", options.EnableIPV6)
		}
		if options.Label != "" {
			values.Add("label", options.Label)
		}
		if options.SSHKeyIDs != nil && len(options.SSHKeyIDs) != 0 {
			values.Add("SSHKEYID", strings.Join(options.SSHKeyIDs, ","))
		}
		if options.AppID != "" {
			values.Add("APPID", options.AppID)
		}
		if options.UserData != "" {
			values.Add("userdata", base64.StdEncoding.EncodeToString([]byte(options.UserData)))
		}
		if options.NotifyActivate != "" {
			values.Add("notify_activate", options.NotifyActivate)
		}
		if options.Hostname != "" {
			values.Add("hostname", options.Hostname)
		}
		if options.Tag != "" {
			values.Add("tag", options.Tag)
		}
		if options.ReservedIPV4 != "" {
			values.Add("reserved_ip_v4", options.ReservedIPV4)
		}
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	bm := new(BareMetalServer)

	err = b.client.DoWithContext(ctx, req, bm)

	if err != nil {
		return nil, err
	}

	return bm, nil
}

// Delete a bare metal server.
// All data will be permanently lost, and the IP address will be released. There is no going back from this call.
func (b *BareMetalServerServiceHandler) Delete(ctx context.Context, serverID string) error {
	uri := "/v1/baremetal/destroy"

	values := url.Values{
		"SUBID": {serverID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// EnableIPV6 enables IPv6 networking on a bare metal server by assigning an IPv6 subnet to it.
// The server will not be rebooted when the subnet is assigned.
func (b *BareMetalServerServiceHandler) EnableIPV6(ctx context.Context, serverID string) error {
	uri := "/v1/baremetal/ipv6_enable"

	values := url.Values{
		"SUBID": {serverID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List lists all bare metal servers on the current account. This includes both pending and active servers.
func (b *BareMetalServerServiceHandler) List(ctx context.Context) ([]BareMetalServer, error) {
	return b.list(ctx, "", "")
}

// ListByLabel lists all bare metal servers that match the given label on the current account. This includes both pending and active servers.
func (b *BareMetalServerServiceHandler) ListByLabel(ctx context.Context, label string) ([]BareMetalServer, error) {
	return b.list(ctx, "label", label)
}

// ListByMainIP lists all bare metal servers that match the given IP address on the current account. This includes both pending and active servers.
func (b *BareMetalServerServiceHandler) ListByMainIP(ctx context.Context, mainIP string) ([]BareMetalServer, error) {
	return b.list(ctx, "main_ip", mainIP)
}

// ListByTag lists all bare metal servers that match the given tag on the current account. This includes both pending and active servers.
func (b *BareMetalServerServiceHandler) ListByTag(ctx context.Context, tag string) ([]BareMetalServer, error) {
	return b.list(ctx, "tag", tag)
}

func (b *BareMetalServerServiceHandler) list(ctx context.Context, key, value string) ([]BareMetalServer, error) {
	uri := "/v1/baremetal/list"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	if key != "" {
		q := req.URL.Query()
		q.Add(key, value)
		req.URL.RawQuery = q.Encode()
	}

	bmsMap := make(map[string]BareMetalServer)
	err = b.client.DoWithContext(ctx, req, &bmsMap)
	if err != nil {
		return nil, err
	}

	var bms []BareMetalServer
	for _, bm := range bmsMap {
		bms = append(bms, bm)
	}

	return bms, nil
}

// GetServer gets the server with the given ID
func (b *BareMetalServerServiceHandler) GetServer(ctx context.Context, serverID string) (*BareMetalServer, error) {
	uri := "/v1/baremetal/list"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)
	req.URL.RawQuery = q.Encode()

	bms := new(BareMetalServer)
	err = b.client.DoWithContext(ctx, req, bms)
	if err != nil {
		return nil, err
	}

	return bms, nil
}

// GetUserData retrieves the (base64 encoded) user-data for this bare metal server
func (b *BareMetalServerServiceHandler) GetUserData(ctx context.Context, serverID string) (*UserData, error) {
	uri := "/v1/baremetal/get_user_data"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)
	req.URL.RawQuery = q.Encode()

	userData := new(UserData)
	err = b.client.DoWithContext(ctx, req, userData)

	if err != nil {
		return nil, err
	}

	return userData, nil
}

// Halt a bare metal server.
// This is a hard power off, meaning that the power to the machine is severed.
// The data on the machine will not be modified, and you will still be billed for the machine.
func (b *BareMetalServerServiceHandler) Halt(ctx context.Context, serverID string) error {
	uri := "/v1/baremetal/halt"

	values := url.Values{
		"SUBID": {serverID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// IPV4Info will List the IPv4 information of a bare metal server.
// IP information is only available for bare metal servers in the "active" state.
func (b *BareMetalServerServiceHandler) IPV4Info(ctx context.Context, serverID string) ([]BareMetalServerIPV4, error) {
	uri := "/v1/baremetal/list_ipv4"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)

	req.URL.RawQuery = q.Encode()

	var ipMap map[string][]BareMetalServerIPV4
	err = b.client.DoWithContext(ctx, req, &ipMap)

	if err != nil {
		return nil, err
	}

	var ipv4 []BareMetalServerIPV4
	for _, i := range ipMap {
		ipv4 = i
	}

	return ipv4, nil
}

// IPV6Info ists the IPv6 information of a bare metal server.
// IP information is only available for bare metal servers in the "active" state.
// If the bare metal server does not have IPv6 enabled, then an empty array is returned.
func (b *BareMetalServerServiceHandler) IPV6Info(ctx context.Context, serverID string) ([]BareMetalServerIPV6, error) {
	uri := "/v1/baremetal/list_ipv6"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)
	req.URL.RawQuery = q.Encode()

	var ipMap map[string][]BareMetalServerIPV6
	err = b.client.DoWithContext(ctx, req, &ipMap)

	if err != nil {
		return nil, err
	}

	var ipv6 []BareMetalServerIPV6
	for _, i := range ipMap {
		ipv6 = i
	}

	return ipv6, nil
}

// ListApps retrieves a list of Vultr one-click applications to which a bare metal server can be changed.
// Always check against this list before trying to switch applications because it is not possible to switch between every application combination.
func (b *BareMetalServerServiceHandler) ListApps(ctx context.Context, serverID string) ([]Application, error) {
	uri := "/v1/baremetal/app_change_list"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)
	req.URL.RawQuery = q.Encode()

	var appMap map[string]Application
	err = b.client.DoWithContext(ctx, req, &appMap)

	if err != nil {
		return nil, err
	}

	var appList []Application
	for _, a := range appMap {
		appList = append(appList, a)
	}

	return appList, nil
}

// ListOS retrieves a list of operating systems to which a bare metal server can be changed.
// Always check against this list before trying to switch operating systems because it is not possible to switch between every operating system combination.
func (b *BareMetalServerServiceHandler) ListOS(ctx context.Context, serverID string) ([]OS, error) {
	uri := "/v1/baremetal/os_change_list"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", serverID)
	req.URL.RawQuery = q.Encode()

	var osMap map[string]OS
	err = b.client.DoWithContext(ctx, req, &osMap)

	if err != nil {
		return nil, err
	}

	var os []OS
	for _, o := range osMap {
		os = append(os, o)
	}

	return os, nil
}

// Reboot a bare metal server. This is a hard reboot, which means that the server is powered off, then back on.
func (b *BareMetalServerServiceHandler) Reboot(ctx context.Context, serverID string) error {
	uri := "/v1/baremetal/reboot"

	values := url.Values{
		"SUBID": {serverID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Reinstall the operating system on a bare metal server.
// All data will be permanently lost, but the IP address will remain the same. There is no going back from this call.
func (b *BareMetalServerServiceHandler) Reinstall(ctx context.Context, serverID string) error {
	uri := "/v1/baremetal/reinstall"

	values := url.Values{
		"SUBID": {serverID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetLabel sets the label of a bare metal server.
func (b *BareMetalServerServiceHandler) SetLabel(ctx context.Context, serverID, label string) error {
	uri := "/v1/baremetal/label_set"

	values := url.Values{
		"SUBID": {serverID},
		"label": {label},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetTag sets the tag of a bare metal server.
func (b *BareMetalServerServiceHandler) SetTag(ctx context.Context, serverID, tag string) error {
	uri := "/v1/baremetal/tag_set"

	values := url.Values{
		"SUBID": {serverID},
		"tag":   {tag},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetUserData sets the user-data for this server.
// User-data is a generic data store, which some provisioning tools and cloud operating systems use as a configuration file.
// It is generally consumed only once after an instance has been launched, but individual needs may vary.
func (b *BareMetalServerServiceHandler) SetUserData(ctx context.Context, serverID, userData string) error {
	uri := "/v1/baremetal/set_user_data"

	encodedUserData := base64.StdEncoding.EncodeToString([]byte(userData))

	values := url.Values{
		"SUBID":    {serverID},
		"userdata": {encodedUserData},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}
