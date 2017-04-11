package monitor

// Job wraps an NS1 /monitoring/jobs resource
type Job struct {
	ID string `json:"id,omitempty"`

	// The id of the notification list to send notifications to.
	NotifyListID string `json:"notify_list"`

	// Type of monitor to be run.
	// Available job types:
	//  - http: Do an HTTP request against a webserver
	//  - dns: Do a DNS lookup against a nameserver
	//  - tcp: Connect to a TCP port on a host
	//  - ping: Ping a host using ICMP packets
	Type string `json:"job_type"`

	// Configuration dictionary(key/vals depend on the jobs' type).
	Config Config `json:"config"`

	// The current status of the monitor.
	Status map[string]*Status `json:"status,omitempty"`

	// Rules for determining failure conditions.
	Rules []*Rule `json:"rules,omitempty"`

	// List of regions in which to run the monitor.
	// eg, ["dal", "sin", "sjc", "lga", "ams"]
	Regions []string `json:"regions"`

	// Indicates if the job is active or temporarily disabled.
	Active bool `json:"active"`

	// Frequency(in seconds), at which to run the monitor.
	Frequency int `json:"frequency"`

	// The policy for determining the monitor's global status based
	// on the status of the job in all regions.
	// Available policies:
	//  - quorum: Status change when majority status
	//  - all: Status change only when all regions are in agreement
	//  - one: Status change if any region changes
	Policy string `json:"policy"`

	// Controls behavior of how the job is assigned to monitoring regions.
	// Currently this must be fixed — indicating monitoring regions are explicitly chosen.
	RegionScope string `json:"region_scope"`

	// Freeform notes to be included in any notifications about this job,
	// e.g., instructions for operators who will receive the notifications.
	Notes string `json:"notes,omitempty"`

	// A free-form display name for the monitoring job.
	Name string `json:"name"`

	// Time(in seconds) between repeat notifications of a failed job.
	// Set to 0 to disable repeating notifications.
	NotifyRepeat int `json:"notify_repeat"`

	// If true, on any apparent state change, the job is quickly re-run after
	// one second to confirm the state change before notification.
	RapidRecheck bool `json:"rapid_recheck"`

	// Time(in seconds) after a failure to wait before sending a notification.
	NotifyDelay int `json:"notify_delay"`

	// If true, notifications are sent for any regional failure (and failback if desired),
	// in addition to global state notifications.
	NotifyRegional bool `json:"notify_regional"`

	// If true, a notification is sent when a job returns to an "up" state.
	NotifyFailback bool `json:"notify_failback"`
}

// Activate a monitoring job.
func (j *Job) Activate() {
	j.Active = true

}

// Deactivate a monitoring job.
func (j *Job) Deactivate() {
	j.Active = false
}

// Result wraps an element of a JobType's "results" attribute
type Result struct {
	Comparators []string `json:"comparators"`
	Metric      bool     `json:"metric"`
	Validator   string   `json:"validator"`
	ShortDesc   string   `json:"shortdesc"`
	Type        string   `json:"type"`
	Desc        string   `json:"desc"`
}

// Status wraps an value of a Job's "status" attribute
type Status struct {
	Since  int    `json:"since"`
	Status string `json:"status"`
}

// StatusLog wraps an NS1 /monitoring/history resource
type StatusLog struct {
	Job    string `json:"job"`
	Region string `json:"region"`
	Status string `json:"status"`
	Since  int    `json:"since"`
	Until  int    `json:"until"`
}

// Rule wraps an element of a Job's "rules" attribute
type Rule struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	Comparison string      `json:"comparison"`
}

// NewHTTPConfig constructs/returns a job configuration for HTTP type jobs.
// url is the URL to query. (Required)
// method is the HTTP method(valid methods are HEAD, GET, and POST).
// ua is the user agent text in the request header.
// auth is the authorization header to use in request.
// connTimeout is the timeout(in sec) to wait for query output.
func NewHTTPConfig(url, method, ua, auth string, connTimeout int) *Config {
	return &Config{
		"url":                url, // Required
		"method":             method,
		"user_agent":         ua,
		"auth":               auth,
		"connection_timeout": connTimeout,
	}

}

// NewDNSConfig constructs/returns a job configuration for DNS type jobs.
// host is the IP address or hostname of the nameserver to query. (Required)
// domain name to query. (Required)
// port is the dns port to query on host.
// t is the type of the DNS record type to query.
// respTimeout is the timeout(in ms) after sending query to wait for the output.
func NewDNSConfig(host, domain string, port int, t string, respTimeout int) *Config {
	return &Config{
		"host":             host,   // Required
		"domain":           domain, // Required
		"port":             port,
		"type":             t,
		"response_timeout": respTimeout,
	}
}

// NewTCPConfig constructs/returns a job configuration for TCP type jobs.
// host is the IP address or hostname to connect to. (Required)
// port is the tcp port to connect to on host. (Required)
// connTimeout is the timeout(in ms) before giving up on trying to connect.
// respTimeout is the timeout(in sec) after connecting to wait for output.
// send is the string to send to the host upon connecting.
// ssl determines whether to attempt negotiating an SSL connection.
func NewTCPConfig(host string, port, connTimeout, respTimeout int, send string, ssl bool) *Config {
	return &Config{
		"host":               host, // Required
		"port":               port, // Required
		"connection_timeout": connTimeout,
		"response_timeout":   respTimeout,
		"send":               send,
		"ssl":                ssl,
	}
}

// NewPINGConfig constructs/returns a job configuration for PING type jobs.
// host is the IP address or hostname to ping. (Required)
// timeout is the timeout(in ms) before marking the host as failed.
// count is the number of packets to send.
// interval is the minimum time(in ms) to wait between sending each packet.
func NewPINGConfig(host string, timeout, count, interval int) *Config {
	return &Config{
		"host":     host, // Required
		"timeout":  timeout,
		"count":    count,
		"interval": interval,
	}
}
