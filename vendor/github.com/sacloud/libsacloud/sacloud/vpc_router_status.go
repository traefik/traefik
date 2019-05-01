package sacloud

// VPCRouterStatus VPCルータのステータス情報
type VPCRouterStatus struct {
	FirewallReceiveLogs []string
	FirewallSendLogs    []string
	VPNLogs             []string
	SessionCount        int
	DHCPServerLeases    []struct {
		IPAddress  string
		MACAddress string
	}
	L2TPIPsecServerSessions []struct {
		User      string
		IPAddress string
		TimeSec   int
	}
	PPTPServerSessions []struct {
		User      string
		IPAddress string
		TimeSec   int
	}
	SiteToSiteIPsecVPNPeers []struct {
		Status string
		Peer   string
	}
}
