package govultr

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ServerService is the interface to interact with the server endpoints on the Vultr API
// Link: https://www.vultr.com/api/#server
type ServerService interface {
	ChangeApp(ctx context.Context, instanceID, appID string) error
	ListApps(ctx context.Context, instanceID string) ([]Application, error)
	AppInfo(ctx context.Context, instanceID string) (*AppInfo, error)
	EnableBackup(ctx context.Context, instanceID string) error
	DisableBackup(ctx context.Context, instanceID string) error
	GetBackupSchedule(ctx context.Context, instanceID string) (*BackupSchedule, error)
	SetBackupSchedule(ctx context.Context, instanceID string, backup *BackupSchedule) error
	RestoreBackup(ctx context.Context, instanceID, backupID string) error
	RestoreSnapshot(ctx context.Context, instanceID, snapshotID string) error
	SetLabel(ctx context.Context, instanceID, label string) error
	SetTag(ctx context.Context, instanceID, tag string) error
	Neighbors(ctx context.Context, instanceID string) ([]int, error)
	EnablePrivateNetwork(ctx context.Context, instanceID, networkID string) error
	DisablePrivateNetwork(ctx context.Context, instanceID, networkID string) error
	ListPrivateNetworks(ctx context.Context, instanceID string) ([]PrivateNetwork, error)
	ListUpgradePlan(ctx context.Context, instanceID string) ([]int, error)
	UpgradePlan(ctx context.Context, instanceID, vpsPlanID string) error
	ListOS(ctx context.Context, instanceID string) ([]OS, error)
	ChangeOS(ctx context.Context, instanceID, osID string) error
	IsoAttach(ctx context.Context, instanceID, isoID string) error
	IsoDetach(ctx context.Context, instanceID string) error
	IsoStatus(ctx context.Context, instanceID string) (*ServerIso, error)
	SetFirewallGroup(ctx context.Context, instanceID, firewallGroupID string) error
	GetUserData(ctx context.Context, instanceID string) (*UserData, error)
	SetUserData(ctx context.Context, instanceID, userData string) error
	IPV4Info(ctx context.Context, instanceID string, public bool) ([]IPV4, error)
	IPV6Info(ctx context.Context, instanceID string) ([]IPV6, error)
	AddIPV4(ctx context.Context, instanceID string) error
	DestroyIPV4(ctx context.Context, instanceID, ip string) error
	EnableIPV6(ctx context.Context, instanceID string) error
	Bandwidth(ctx context.Context, instanceID string) ([]map[string]string, error)
	ListReverseIPV6(ctx context.Context, instanceID string) ([]ReverseIPV6, error)
	SetDefaultReverseIPV4(ctx context.Context, instanceID, ip string) error
	DeleteReverseIPV6(ctx context.Context, instanceID, ip string) error
	SetReverseIPV4(ctx context.Context, instanceID, ipv4, entry string) error
	SetReverseIPV6(ctx context.Context, instanceID, ipv6, entry string) error
	Start(ctx context.Context, instanceID string) error
	Halt(ctx context.Context, instanceID string) error
	Reboot(ctx context.Context, instanceID string) error
	Reinstall(ctx context.Context, instanceID string) error
	Delete(ctx context.Context, instanceID string) error
	Create(ctx context.Context, regionID, vpsPlanID, osID int, options *ServerOptions) (*Server, error)
	List(ctx context.Context) ([]Server, error)
	ListByLabel(ctx context.Context, label string) ([]Server, error)
	ListByMainIP(ctx context.Context, mainIP string) ([]Server, error)
	ListByTag(ctx context.Context, tag string) ([]Server, error)
	GetServer(ctx context.Context, instanceID string) (*Server, error)
}

// ServerServiceHandler handles interaction with the server methods for the Vultr API
type ServerServiceHandler struct {
	client *Client
}

// AppInfo represents information about the application on your VPS
type AppInfo struct {
	AppInfo string `json:"app_info"`
}

// BackupSchedule represents a schedule of a backup that runs on a VPS
type BackupSchedule struct {
	Enabled  bool   `json:"enabled"`
	CronType string `json:"cron_type"`
	NextRun  string `json:"next_scheduled_time_utc"`
	Hour     int    `json:"hour"`
	Dow      int    `json:"dow"`
	Dom      int    `json:"dom"`
}

// PrivateNetwork represents a private network attached to a VPS
type PrivateNetwork struct {
	NetworkID  string `json:"NETWORKID"`
	MacAddress string `json:"mac_address"`
	IPAddress  string `json:"ip_address"`
}

// ServerIso represents a iso attached to a VPS
type ServerIso struct {
	State string `json:"state"`
	IsoID string `json:"ISOID"`
}

// UserData represents the user data you can give a VPS
type UserData struct {
	UserData string `json:"userdata"`
}

// IPV4 represents IPV4 information for a VPS
type IPV4 struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
	Type    string `json:"type"`
	Reverse string `json:"reverse"`
}

// IPV6 represents IPV6 information for a VPS
type IPV6 struct {
	IP          string `json:"ip"`
	Network     string `json:"network"`
	NetworkSize string `json:"network_size"`
	Type        string `json:"type"`
}

// ReverseIPV6 represents IPV6 reverse DNS entries
type ReverseIPV6 struct {
	IP      string `json:"ip"`
	Reverse string `json:"reverse"`
}

// Server represents a VPS
type Server struct {
	InstanceID       string      `json:"SUBID"`
	Os               string      `json:"os"`
	RAM              string      `json:"ram"`
	Disk             string      `json:"disk"`
	MainIP           string      `json:"main_ip"`
	VPSCpus          string      `json:"vcpu_count"`
	Location         string      `json:"location"`
	RegionID         string      `json:"DCID"`
	DefaultPassword  string      `json:"default_password"`
	Created          string      `json:"date_created"`
	PendingCharges   string      `json:"pending_charges"`
	Status           string      `json:"status"`
	Cost             string      `json:"cost_per_month"`
	CurrentBandwidth float64     `json:"current_bandwidth_gb"`
	AllowedBandwidth string      `json:"allowed_bandwidth_gb"`
	NetmaskV4        string      `json:"netmask_v4"`
	GatewayV4        string      `json:"gateway_v4"`
	PowerStatus      string      `json:"power_status"`
	ServerState      string      `json:"server_state"`
	PlanID           string      `json:"VPSPLANID"`
	V6Networks       []V6Network `json:"v6_networks"`
	Label            string      `json:"label"`
	InternalIP       string      `json:"internal_ip"`
	KVMUrl           string      `json:"kvm_url"`
	AutoBackups      string      `json:"auto_backups"`
	Tag              string      `json:"tag"`
	OsID             string      `json:"OSID"`
	AppID            string      `json:"APPID"`
	FirewallGroupID  string      `json:"FIREWALLGROUPID"`
}

// V6Network represents an IPV6 network on a VPS
type V6Network struct {
	Network     string `json:"v6_network"`
	MainIP      string `json:"v6_main_ip"`
	NetworkSize string `json:"v6_network_size"`
}

// ServerOptions are all optional fields that can be used during vps creation
type ServerOptions struct {
	IPXEChain            string
	IsoID                int
	SnapshotID           string
	ScriptID             string
	EnableIPV6           bool
	EnablePrivateNetwork bool
	NetworkID            []string
	Label                string
	SSHKeyIDs            []string
	AutoBackups          bool
	AppID                string
	UserData             string
	NotifyActivate       bool
	DDOSProtection       bool
	ReservedIPV4         string
	Hostname             string
	Tag                  string
	FirewallGroupID      string
}

// ChangeApp changes the VPS to a different application.
func (s *ServerServiceHandler) ChangeApp(ctx context.Context, instanceID, appID string) error {

	uri := "/v1/server/app_change"

	values := url.Values{
		"SUBID": {instanceID},
		"APPID": {appID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// ListApps retrieves a list of applications to which a virtual machine can be changed.
func (s *ServerServiceHandler) ListApps(ctx context.Context, instanceID string) ([]Application, error) {

	uri := "/v1/server/app_change_list"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var appMap map[string]Application
	err = s.client.DoWithContext(ctx, req, &appMap)

	if err != nil {
		return nil, err
	}

	var appList []Application
	for _, a := range appMap {
		appList = append(appList, a)
	}

	return appList, nil
}

// AppInfo retrieves the application information for a given VPS ID
func (s *ServerServiceHandler) AppInfo(ctx context.Context, instanceID string) (*AppInfo, error) {

	uri := "/v1/server/get_app_info"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	appInfo := new(AppInfo)

	err = s.client.DoWithContext(ctx, req, appInfo)

	if err != nil {
		return nil, err
	}

	return appInfo, nil
}

// EnableBackup enables automatic backups on a given VPS
func (s *ServerServiceHandler) EnableBackup(ctx context.Context, instanceID string) error {

	uri := "/v1/server/backup_enable"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// DisableBackup disable automatic backups on a given VPS
func (s *ServerServiceHandler) DisableBackup(ctx context.Context, instanceID string) error {

	uri := "/v1/server/backup_disable"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// GetBackupSchedule retrieves the backup schedule for a given vps - all time values are in UTC
func (s *ServerServiceHandler) GetBackupSchedule(ctx context.Context, instanceID string) (*BackupSchedule, error) {

	uri := "/v1/server/backup_get_schedule"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	backup := new(BackupSchedule)
	err = s.client.DoWithContext(ctx, req, backup)

	if err != nil {
		return nil, err
	}

	return backup, nil
}

// SetBackupSchedule sets the backup schedule for a given vps - all time values are in UTC
func (s *ServerServiceHandler) SetBackupSchedule(ctx context.Context, instanceID string, backup *BackupSchedule) error {

	uri := "/v1/server/backup_set_schedule"

	values := url.Values{
		"SUBID":     {instanceID},
		"cron_type": {backup.CronType},
		"hour":      {strconv.Itoa(backup.Hour)},
		"dow":       {strconv.Itoa(backup.Dow)},
		"dom":       {strconv.Itoa(backup.Dom)},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// RestoreBackup will restore the specified backup to the given VPS
func (s *ServerServiceHandler) RestoreBackup(ctx context.Context, instanceID, backupID string) error {

	uri := "/v1/server/restore_backup"

	values := url.Values{
		"SUBID":    {instanceID},
		"BACKUPID": {backupID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// RestoreSnapshot will restore the specified snapshot to the given VPS
func (s *ServerServiceHandler) RestoreSnapshot(ctx context.Context, instanceID, snapshotID string) error {

	uri := "/v1/server/restore_snapshot"

	values := url.Values{
		"SUBID":      {instanceID},
		"SNAPSHOTID": {snapshotID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetLabel will set a label for a given VPS
func (s *ServerServiceHandler) SetLabel(ctx context.Context, instanceID, label string) error {

	uri := "/v1/server/label_set"

	values := url.Values{
		"SUBID": {instanceID},
		"label": {label},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetTag will set a tag for a given VPS
func (s *ServerServiceHandler) SetTag(ctx context.Context, instanceID, tag string) error {

	uri := "/v1/server/tag_set"

	values := url.Values{
		"SUBID": {instanceID},
		"tag":   {tag},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Neighbors will determine what other vps are hosted on the same physical host as a given vps.
func (s *ServerServiceHandler) Neighbors(ctx context.Context, instanceID string) ([]int, error) {

	uri := "/v1/server/neighbors"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var neighbors []int
	err = s.client.DoWithContext(ctx, req, &neighbors)

	if err != nil {
		return nil, err
	}

	return neighbors, nil
}

// EnablePrivateNetwork enables private networking on a server.
// The server will be automatically rebooted to complete the request.
// No action occurs if private networking was already enabled
func (s *ServerServiceHandler) EnablePrivateNetwork(ctx context.Context, instanceID, networkID string) error {

	uri := "/v1/server/private_network_enable"

	values := url.Values{
		"SUBID":     {instanceID},
		"NETWORKID": {networkID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// DisablePrivateNetwork removes a private network from a server.
// The server will be automatically rebooted to complete the request.
func (s *ServerServiceHandler) DisablePrivateNetwork(ctx context.Context, instanceID, networkID string) error {

	uri := "/v1/server/private_network_disable"

	values := url.Values{
		"SUBID":     {instanceID},
		"NETWORKID": {networkID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// ListPrivateNetworks will list private networks attached to a vps
func (s *ServerServiceHandler) ListPrivateNetworks(ctx context.Context, instanceID string) ([]PrivateNetwork, error) {

	uri := "/v1/server/private_networks"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var networkMap map[string]PrivateNetwork
	err = s.client.DoWithContext(ctx, req, &networkMap)

	if err != nil {
		return nil, err
	}

	var privateNetworks []PrivateNetwork
	for _, p := range networkMap {
		privateNetworks = append(privateNetworks, p)
	}

	return privateNetworks, nil
}

// ListUpgradePlan Retrieve a list of the planIDs for which the vps can be upgraded.
// An empty response array means that there are currently no upgrades available
func (s *ServerServiceHandler) ListUpgradePlan(ctx context.Context, instanceID string) ([]int, error) {

	uri := "/v1/server/upgrade_plan_list"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var plans []int
	err = s.client.DoWithContext(ctx, req, &plans)

	if err != nil {
		return nil, err
	}

	return plans, nil
}

// UpgradePlan will upgrade the plan of a virtual machine.
// The vps will be rebooted upon a successful upgrade.
func (s *ServerServiceHandler) UpgradePlan(ctx context.Context, instanceID, vpsPlanID string) error {

	uri := "/v1/server/upgrade_plan"

	values := url.Values{
		"SUBID":     {instanceID},
		"VPSPLANID": {vpsPlanID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// ListOS retrieves a list of operating systems to which the VPS can be changed to.
func (s *ServerServiceHandler) ListOS(ctx context.Context, instanceID string) ([]OS, error) {

	uri := "/v1/server/os_change_list"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var osMap map[string]OS
	err = s.client.DoWithContext(ctx, req, &osMap)

	if err != nil {
		return nil, err
	}

	var os []OS
	for _, o := range osMap {
		os = append(os, o)
	}

	return os, nil
}

// ChangeOS changes the VPS to a different operating system.
// All data will be permanently lost.
func (s *ServerServiceHandler) ChangeOS(ctx context.Context, instanceID, osID string) error {

	uri := "/v1/server/os_change"

	values := url.Values{
		"SUBID": {instanceID},
		"OSID":  {osID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// IsoAttach will attach an ISO to the given VPS and reboot it
func (s *ServerServiceHandler) IsoAttach(ctx context.Context, instanceID, isoID string) error {

	uri := "/v1/server/iso_attach"

	values := url.Values{
		"SUBID": {instanceID},
		"ISOID": {isoID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// IsoDetach will detach the currently mounted ISO and reboot the server.
func (s *ServerServiceHandler) IsoDetach(ctx context.Context, instanceID string) error {

	uri := "/v1/server/iso_detach"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// IsoStatus retrieves the current ISO state for a given VPS.
// The returned state may be one of: ready | isomounting | isomounted.
func (s *ServerServiceHandler) IsoStatus(ctx context.Context, instanceID string) (*ServerIso, error) {

	uri := "/v1/server/iso_status"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	serverIso := new(ServerIso)
	err = s.client.DoWithContext(ctx, req, serverIso)

	if err != nil {
		return nil, err
	}

	return serverIso, nil
}

// SetFirewallGroup will set, change, or remove the firewall group currently applied to a vps.
//  A value of "0" means "no firewall group"
func (s *ServerServiceHandler) SetFirewallGroup(ctx context.Context, instanceID, firewallGroupID string) error {

	uri := "/v1/server/firewall_group_set"

	values := url.Values{
		"SUBID":           {instanceID},
		"FIREWALLGROUPID": {firewallGroupID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetUserData sets the user-data for this subscription.
// User-data is a generic data store, which some provisioning tools and cloud operating systems use as a configuration file.
// It is generally consumed only once after an instance has been launched, but individual needs may vary.
func (s *ServerServiceHandler) SetUserData(ctx context.Context, instanceID, userData string) error {

	uri := "/v1/server/set_user_data"

	encodedUserData := base64.StdEncoding.EncodeToString([]byte(userData))

	values := url.Values{
		"SUBID":    {instanceID},
		"userdata": {encodedUserData},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// GetUserData retrieves the (base64 encoded) user-data for this VPS
func (s *ServerServiceHandler) GetUserData(ctx context.Context, instanceID string) (*UserData, error) {

	uri := "/v1/server/get_user_data"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	userData := new(UserData)
	err = s.client.DoWithContext(ctx, req, userData)

	if err != nil {
		return nil, err
	}

	return userData, nil
}

// IPV4Info will list the IPv4 information of a virtual machine.
// Public if set to 'true', includes information about the public network adapter (such as MAC address) with the "main_ip" entry.
func (s *ServerServiceHandler) IPV4Info(ctx context.Context, instanceID string, public bool) ([]IPV4, error) {

	uri := "/v1/server/list_ipv4"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)

	if public == true {
		q.Add("public_network", instanceID)
	}

	req.URL.RawQuery = q.Encode()

	var ipMap map[string][]IPV4
	err = s.client.DoWithContext(ctx, req, &ipMap)

	if err != nil {
		return nil, err
	}

	var ipv4 []IPV4
	for _, i := range ipMap {
		ipv4 = i
	}

	return ipv4, nil
}

// IPV6Info will list the IPv6 information of a virtual machine.
// If the virtual machine does not have IPv6 enabled, then an empty array is returned.
func (s *ServerServiceHandler) IPV6Info(ctx context.Context, instanceID string) ([]IPV6, error) {
	uri := "/v1/server/list_ipv6"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var ipMap map[string][]IPV6
	err = s.client.DoWithContext(ctx, req, &ipMap)

	if err != nil {
		return nil, err
	}

	var ipv6 []IPV6
	for _, i := range ipMap {
		ipv6 = i
	}

	return ipv6, nil
}

// AddIPV4 will add a new IPv4 address to a server.
func (s *ServerServiceHandler) AddIPV4(ctx context.Context, instanceID string) error {

	uri := "/v1/server/create_ipv4"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// DestroyIPV4 removes a secondary IPv4 address from a server.
// Your server will be hard-restarted. We suggest halting the machine gracefully before removing IPs.
func (s *ServerServiceHandler) DestroyIPV4(ctx context.Context, instanceID, ip string) error {

	uri := "/v1/server/destroy_ipv4"

	values := url.Values{
		"SUBID": {instanceID},
		"ip":    {ip},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// EnableIPV6 enables IPv6 networking on a server by assigning an IPv6 subnet to it.
func (s *ServerServiceHandler) EnableIPV6(ctx context.Context, instanceID string) error {

	uri := "/v1/server/ipv6_enable"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Bandwidth will get the bandwidth used by a VPS
func (s *ServerServiceHandler) Bandwidth(ctx context.Context, instanceID string) ([]map[string]string, error) {

	uri := "/v1/server/bandwidth"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var bandwidthMap map[string][][]string
	err = s.client.DoWithContext(ctx, req, &bandwidthMap)

	if err != nil {
		return nil, err
	}

	var bandwidth []map[string]string

	for _, b := range bandwidthMap["incoming_bytes"] {
		inMap := make(map[string]string)
		inMap["date"] = b[0]
		inMap["incoming"] = b[1]
		bandwidth = append(bandwidth, inMap)
	}

	for _, b := range bandwidthMap["outgoing_bytes"] {
		for i := range bandwidth {
			if bandwidth[i]["date"] == b[0] {
				bandwidth[i]["outgoing"] = b[1]
				break
			}
		}
	}

	return bandwidth, nil
}

// ListReverseIPV6 List the IPv6 reverse DNS entries of a virtual machine.
// Reverse DNS entries are only available for virtual machines in the "active" state.
// If the virtual machine does not have IPv6 enabled, then an empty array is returned.
func (s *ServerServiceHandler) ListReverseIPV6(ctx context.Context, instanceID string) ([]ReverseIPV6, error) {

	uri := "/v1/server/reverse_list_ipv6"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	var reverseMap map[string][]ReverseIPV6
	err = s.client.DoWithContext(ctx, req, &reverseMap)

	if err != nil {
		return nil, err
	}

	var reverseIPV6 []ReverseIPV6
	for _, r := range reverseMap {

		if len(r) == 0 {
			break
		}

		for _, i := range r {
			reverseIPV6 = append(reverseIPV6, i)
		}
	}

	return reverseIPV6, nil
}

// SetDefaultReverseIPV4 will set a reverse DNS entry for an IPv4 address of a virtual machine to the original setting.
// Upon success, DNS changes may take 6-12 hours to become active.
func (s *ServerServiceHandler) SetDefaultReverseIPV4(ctx context.Context, instanceID, ip string) error {

	uri := "/v1/server/reverse_default_ipv4"

	values := url.Values{
		"SUBID": {instanceID},
		"ip":    {ip},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// DeleteReverseIPV6 Remove a reverse DNS entry for an IPv6 address of a VPS.
// Upon success, DNS changes may take 6-12 hours to become active.
func (s *ServerServiceHandler) DeleteReverseIPV6(ctx context.Context, instanceID, ip string) error {

	uri := "/v1/server/reverse_delete_ipv6"

	values := url.Values{
		"SUBID": {instanceID},
		"ip":    {ip},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetReverseIPV4 will set a reverse DNS entry for an IPv4 address of a virtual machine.
// Upon success, DNS changes may take 6-12 hours to become active.
func (s *ServerServiceHandler) SetReverseIPV4(ctx context.Context, instanceID, ipv4, entry string) error {

	uri := "/v1/server/reverse_set_ipv4"

	values := url.Values{
		"SUBID": {instanceID},
		"ip":    {ipv4},
		"entry": {entry},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetReverseIPV6 will set a reverse DNS entry for an IPv4 address of a virtual machine.
// Upon success, DNS changes may take 6-12 hours to become active.
func (s *ServerServiceHandler) SetReverseIPV6(ctx context.Context, instanceID, ipv6, entry string) error {
	uri := "/v1/server/reverse_set_ipv6"

	values := url.Values{
		"SUBID": {instanceID},
		"ip":    {ipv6},
		"entry": {entry},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Start will start a vps. If the machine is already running, it will be restarted.
func (s *ServerServiceHandler) Start(ctx context.Context, instanceID string) error {
	uri := "/v1/server/start"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Halt will halt a virtual machine. This is a hard power off
func (s *ServerServiceHandler) Halt(ctx context.Context, instanceID string) error {

	uri := "/v1/server/halt"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Reboot will reboot a VPS. This is a hard reboot
func (s *ServerServiceHandler) Reboot(ctx context.Context, instanceID string) error {

	uri := "/v1/server/reboot"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Reinstall will reinstall the operating system on a VPS.
func (s *ServerServiceHandler) Reinstall(ctx context.Context, instanceID string) error {
	uri := "/v1/server/reinstall"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Delete a VPS. All data will be permanently lost, and the IP address will be released
func (s *ServerServiceHandler) Delete(ctx context.Context, instanceID string) error {

	uri := "/v1/server/destroy"

	values := url.Values{
		"SUBID": {instanceID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Create will create a new VPS
// In order to create a server using a snapshot, use OSID 164 and specify a SNAPSHOTID.
// Similarly, to create a server using an ISO use OSID 159 and specify an ISOID.
func (s *ServerServiceHandler) Create(ctx context.Context, regionID, vpsPlanID, osID int, options *ServerOptions) (*Server, error) {

	uri := "/v1/server/create"

	values := url.Values{
		"DCID":      {strconv.Itoa(regionID)},
		"VPSPLANID": {strconv.Itoa(vpsPlanID)},
		"OSID":      {strconv.Itoa(osID)},
	}

	if options != nil {
		if options.IPXEChain != "" {
			values.Add("ipxe_chain_url", options.IPXEChain)
		}

		if options.IsoID != 0 {
			values.Add("ISOID", strconv.Itoa(options.IsoID))
		}

		if options.SnapshotID != "" {
			values.Add("SNAPSHOTID", options.SnapshotID)
		}

		if options.ScriptID != "" {
			values.Add("SCRIPTID", options.ScriptID)
		}

		if options.EnableIPV6 == true {
			values.Add("enable_ipv6", "yes")
		}

		// Use either EnabledPrivateNetwork or NetworkIDs, not both
		if options.EnablePrivateNetwork == true {
			values.Add("enable_private_network", "yes")
		} else {
			if options.NetworkID != nil && len(options.NetworkID) != 0 {
				for _, n := range options.NetworkID {
					values.Add("NETWORKID[]", n)
				}
			}
		}

		if options.Label != "" {
			values.Add("label", options.Label)
		}

		if options.SSHKeyIDs != nil && len(options.SSHKeyIDs) != 0 {
			values.Add("SSHKEYID", strings.Join(options.SSHKeyIDs, ","))
		}

		if options.AutoBackups == true {
			values.Add("auto_backups", "yes")
		}

		if options.AppID != "" {
			values.Add("APPID", options.AppID)
		}

		if options.UserData != "" {
			values.Add("userdata", base64.StdEncoding.EncodeToString([]byte(options.UserData)))
		}

		if options.NotifyActivate == true {
			values.Add("notify_activate", "yes")
		} else if options.NotifyActivate == false {
			values.Add("notify_activate", "no")
		}

		if options.DDOSProtection == true {
			values.Add("ddos_protection", "yes")
		}

		if options.ReservedIPV4 != "" {
			values.Add("reserved_ip_v4", options.ReservedIPV4)
		}

		if options.Hostname != "" {
			values.Add("hostname", options.Hostname)
		}

		if options.Tag != "" {
			values.Add("tag", options.Tag)
		}

		if options.FirewallGroupID != "" {
			values.Add("FIREWALLGROUPID", options.FirewallGroupID)
		}
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	server := new(Server)
	err = s.client.DoWithContext(ctx, req, server)

	if err != nil {
		return nil, err
	}

	return server, nil
}

// List lists all VPS on the current account. This includes both pending and active servers.
func (s *ServerServiceHandler) List(ctx context.Context) ([]Server, error) {
	return s.list(ctx, "", "")
}

// ListByLabel lists all VPS that match the given label on the current account. This includes both pending and active servers.
func (s *ServerServiceHandler) ListByLabel(ctx context.Context, label string) ([]Server, error) {
	return s.list(ctx, "label", label)
}

// ListByMainIP lists all VPS that match the given IP address on the current account. This includes both pending and active servers.
func (s *ServerServiceHandler) ListByMainIP(ctx context.Context, mainIP string) ([]Server, error) {
	return s.list(ctx, "main_ip", mainIP)
}

// ListByTag lists all VPS that match the given tag on the current account. This includes both pending and active servers.
func (s *ServerServiceHandler) ListByTag(ctx context.Context, tag string) ([]Server, error) {
	return s.list(ctx, "tag", tag)
}

// list is used to consolidate the optional params to get a VPS
func (s *ServerServiceHandler) list(ctx context.Context, key, value string) ([]Server, error) {

	uri := "/v1/server/list"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	if key != "" {
		q := req.URL.Query()
		q.Add(key, value)
		req.URL.RawQuery = q.Encode()
	}

	var serverMap map[string]Server
	err = s.client.DoWithContext(ctx, req, &serverMap)

	if err != nil {
		return nil, err
	}

	var servers []Server
	for _, s := range serverMap {
		servers = append(servers, s)
	}

	return servers, nil
}

// GetServer will get the server with the given instanceID
func (s *ServerServiceHandler) GetServer(ctx context.Context, instanceID string) (*Server, error) {

	uri := "/v1/server/list"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", instanceID)
	req.URL.RawQuery = q.Encode()

	server := new(Server)
	err = s.client.DoWithContext(ctx, req, server)

	if err != nil {
		return nil, err
	}

	return server, nil

}
