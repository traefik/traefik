package agent

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/circonus"
	"github.com/armon/go-metrics/datadog"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/consul/logger"
	"github.com/hashicorp/consul/watch"
	"github.com/hashicorp/go-checkpoint"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/logutils"
	"github.com/hashicorp/scada-client/scada"
	"github.com/mitchellh/cli"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

// gracefulTimeout controls how long we wait before forcefully terminating
var gracefulTimeout = 5 * time.Second

// validDatacenter is used to validate a datacenter
var validDatacenter = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

// Command is a Command implementation that runs a Consul agent.
// The command will not end unless a shutdown message is sent on the
// ShutdownCh. If two messages are sent on the ShutdownCh it will forcibly
// exit.
type Command struct {
	base.Command
	Revision          string
	Version           string
	VersionPrerelease string
	HumanVersion      string
	ShutdownCh        <-chan struct{}
	configReloadCh    chan chan error
	args              []string
	logFilter         *logutils.LevelFilter
	logOutput         io.Writer
	agent             *Agent
	httpServers       []*HTTPServer
	dnsServer         *DNSServer
	scadaProvider     *scada.Provider
	scadaHttp         *HTTPServer
}

// readConfig is responsible for setup of our configuration using
// the command line and any file configs
func (c *Command) readConfig() *Config {
	var cmdConfig Config
	var configFiles []string
	var retryInterval string
	var retryIntervalWan string
	var dnsRecursors []string
	var dev bool
	var dcDeprecated string
	var nodeMeta []string

	f := c.Command.NewFlagSet(c)

	f.Var((*AppendSliceValue)(&configFiles), "config-file",
		"Path to a JSON file to read configuration from. This can be specified multiple times.")
	f.Var((*AppendSliceValue)(&configFiles), "config-dir",
		"Path to a directory to read configuration files from. This will read every file ending "+
			"in '.json' as configuration in this directory in alphabetical order. This can be "+
			"specified multiple times.")
	f.Var((*AppendSliceValue)(&dnsRecursors), "recursor",
		"Address of an upstream DNS server. Can be specified multiple times.")
	f.Var((*AppendSliceValue)(&nodeMeta), "node-meta",
		"An arbitrary metadata key/value pair for this node, of the format `key:value`. Can be specified multiple times.")
	f.BoolVar(&dev, "dev", false, "Starts the agent in development mode.")

	f.StringVar(&cmdConfig.LogLevel, "log-level", "", "Log level of the agent.")
	f.StringVar(&cmdConfig.NodeName, "node", "", "Name of this node. Must be unique in the cluster.")
	f.StringVar((*string)(&cmdConfig.NodeID), "node-id", "",
		"A unique ID for this node across space and time. Defaults to a randomly-generated ID"+
			" that persists in the data-dir.")
	f.StringVar(&dcDeprecated, "dc", "", "Datacenter of the agent (deprecated: use 'datacenter' instead).")
	f.StringVar(&cmdConfig.Datacenter, "datacenter", "", "Datacenter of the agent.")
	f.StringVar(&cmdConfig.DataDir, "data-dir", "", "Path to a data directory to store agent state.")
	f.BoolVar(&cmdConfig.EnableUi, "ui", false, "Enables the built-in static web UI server.")
	f.StringVar(&cmdConfig.UiDir, "ui-dir", "", "Path to directory containing the web UI resources.")
	f.StringVar(&cmdConfig.PidFile, "pid-file", "", "Path to file to store agent PID.")
	f.StringVar(&cmdConfig.EncryptKey, "encrypt", "", "Provides the gossip encryption key.")

	f.BoolVar(&cmdConfig.Server, "server", false, "Switches agent to server mode.")
	f.BoolVar(&cmdConfig.NonVotingServer, "non-voting-server", false,
		"(Enterprise-only) This flag is used to make the server not participate in the Raft quorum, "+
			"and have it only receive the data replication stream. This can be used to add read scalability "+
			"to a cluster in cases where a high volume of reads to servers are needed.")
	f.BoolVar(&cmdConfig.Bootstrap, "bootstrap", false, "Sets server to bootstrap mode.")
	f.IntVar(&cmdConfig.BootstrapExpect, "bootstrap-expect", 0, "Sets server to expect bootstrap mode.")
	f.StringVar(&cmdConfig.Domain, "domain", "", "Domain to use for DNS interface.")

	f.StringVar(&cmdConfig.ClientAddr, "client", "",
		"Sets the address to bind for client access. This includes RPC, DNS, HTTP and HTTPS (if configured).")
	f.StringVar(&cmdConfig.BindAddr, "bind", "", "Sets the bind address for cluster communication.")
	f.StringVar(&cmdConfig.SerfWanBindAddr, "serf-wan-bind", "", "Address to bind Serf WAN listeners to.")
	f.StringVar(&cmdConfig.SerfLanBindAddr, "serf-lan-bind", "", "Address to bind Serf LAN listeners to.")
	f.IntVar(&cmdConfig.Ports.HTTP, "http-port", 0, "Sets the HTTP API port to listen on.")
	f.IntVar(&cmdConfig.Ports.DNS, "dns-port", 0, "DNS port to use.")
	f.StringVar(&cmdConfig.AdvertiseAddr, "advertise", "", "Sets the advertise address to use.")
	f.StringVar(&cmdConfig.AdvertiseAddrWan, "advertise-wan", "",
		"Sets address to advertise on WAN instead of -advertise address.")

	f.StringVar(&cmdConfig.AtlasInfrastructure, "atlas", "",
		"(deprecated) Sets the Atlas infrastructure name, enables SCADA.")
	f.StringVar(&cmdConfig.AtlasToken, "atlas-token", "",
		"(deprecated) Provides the Atlas API token.")
	f.BoolVar(&cmdConfig.AtlasJoin, "atlas-join", false,
		"(deprecated) Enables auto-joining the Atlas cluster.")
	f.StringVar(&cmdConfig.AtlasEndpoint, "atlas-endpoint", "",
		"(deprecated) The address of the endpoint for Atlas integration.")

	f.IntVar(&cmdConfig.Protocol, "protocol", -1,
		"Sets the protocol version. Defaults to latest.")
	f.IntVar(&cmdConfig.RaftProtocol, "raft-protocol", -1,
		"Sets the Raft protocol version. Defaults to latest.")

	f.BoolVar(&cmdConfig.EnableSyslog, "syslog", false,
		"Enables logging to syslog.")
	f.BoolVar(&cmdConfig.RejoinAfterLeave, "rejoin", false,
		"Ignores a previous leave and attempts to rejoin the cluster.")
	f.Var((*AppendSliceValue)(&cmdConfig.StartJoin), "join",
		"Address of an agent to join at start time. Can be specified multiple times.")
	f.Var((*AppendSliceValue)(&cmdConfig.StartJoinWan), "join-wan",
		"Address of an agent to join -wan at start time. Can be specified multiple times.")
	f.Var((*AppendSliceValue)(&cmdConfig.RetryJoin), "retry-join",
		"Address of an agent to join at start time with retries enabled. Can be specified multiple times.")
	f.IntVar(&cmdConfig.RetryMaxAttempts, "retry-max", 0,
		"Maximum number of join attempts. Defaults to 0, which will retry indefinitely.")
	f.StringVar(&retryInterval, "retry-interval", "",
		"Time to wait between join attempts.")
	f.StringVar(&cmdConfig.RetryJoinEC2.Region, "retry-join-ec2-region", "",
		"EC2 Region to discover servers in.")
	f.StringVar(&cmdConfig.RetryJoinEC2.TagKey, "retry-join-ec2-tag-key", "",
		"EC2 tag key to filter on for server discovery.")
	f.StringVar(&cmdConfig.RetryJoinEC2.TagValue, "retry-join-ec2-tag-value", "",
		"EC2 tag value to filter on for server discovery.")
	f.StringVar(&cmdConfig.RetryJoinGCE.ProjectName, "retry-join-gce-project-name", "",
		"Google Compute Engine project to discover servers in.")
	f.StringVar(&cmdConfig.RetryJoinGCE.ZonePattern, "retry-join-gce-zone-pattern", "",
		"Google Compute Engine region or zone to discover servers in (regex pattern).")
	f.StringVar(&cmdConfig.RetryJoinGCE.TagValue, "retry-join-gce-tag-value", "",
		"Google Compute Engine tag value to filter on for server discovery.")
	f.StringVar(&cmdConfig.RetryJoinGCE.CredentialsFile, "retry-join-gce-credentials-file", "",
		"Path to credentials JSON file to use with Google Compute Engine.")
	f.Var((*AppendSliceValue)(&cmdConfig.RetryJoinWan), "retry-join-wan",
		"Address of an agent to join -wan at start time with retries enabled. "+
			"Can be specified multiple times.")
	f.IntVar(&cmdConfig.RetryMaxAttemptsWan, "retry-max-wan", 0,
		"Maximum number of join -wan attempts. Defaults to 0, which will retry indefinitely.")
	f.StringVar(&retryIntervalWan, "retry-interval-wan", "",
		"Time to wait between join -wan attempts.")

	if err := c.Command.Parse(c.args); err != nil {
		return nil
	}

	if retryInterval != "" {
		dur, err := time.ParseDuration(retryInterval)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
			return nil
		}
		cmdConfig.RetryInterval = dur
	}

	if retryIntervalWan != "" {
		dur, err := time.ParseDuration(retryIntervalWan)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
			return nil
		}
		cmdConfig.RetryIntervalWan = dur
	}

	if len(nodeMeta) > 0 {
		cmdConfig.Meta = make(map[string]string)
		for _, entry := range nodeMeta {
			key, value := parseMetaPair(entry)
			cmdConfig.Meta[key] = value
		}
	}

	var config *Config
	if dev {
		config = DevConfig()
	} else {
		config = DefaultConfig()
	}

	if len(configFiles) > 0 {
		fileConfig, err := ReadConfigPaths(configFiles)
		if err != nil {
			c.Ui.Error(err.Error())
			return nil
		}

		config = MergeConfig(config, fileConfig)
	}

	cmdConfig.DNSRecursors = append(cmdConfig.DNSRecursors, dnsRecursors...)

	config = MergeConfig(config, &cmdConfig)

	if config.NodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error determining node name: %s", err))
			return nil
		}
		config.NodeName = hostname
	}
	config.NodeName = strings.TrimSpace(config.NodeName)
	if config.NodeName == "" {
		c.Ui.Error("Node name can not be empty")
		return nil
	}

	// Make sure LeaveOnTerm and SkipLeaveOnInt are set to the right
	// defaults based on the agent's mode (client or server).
	if config.LeaveOnTerm == nil {
		config.LeaveOnTerm = Bool(!config.Server)
	}
	if config.SkipLeaveOnInt == nil {
		config.SkipLeaveOnInt = Bool(config.Server)
	}

	// Ensure we have a data directory if we are not in dev mode.
	if !dev {
		if config.DataDir == "" {
			c.Ui.Error("Must specify data directory using -data-dir")
			return nil
		}

		if finfo, err := os.Stat(config.DataDir); err != nil {
			if !os.IsNotExist(err) {
				c.Ui.Error(fmt.Sprintf("Error getting data-dir: %s", err))
				return nil
			}
		} else if !finfo.IsDir() {
			c.Ui.Error(fmt.Sprintf("The data-dir specified at %q is not a directory", config.DataDir))
			return nil
		}
	}

	// Ensure all endpoints are unique
	if err := config.verifyUniqueListeners(); err != nil {
		c.Ui.Error(fmt.Sprintf("All listening endpoints must be unique: %s", err))
		return nil
	}

	// Check the data dir for signs of an un-migrated Consul 0.5.x or older
	// server. Consul refuses to start if this is present to protect a server
	// with existing data from starting on a fresh data set.
	if config.Server {
		mdbPath := filepath.Join(config.DataDir, "mdb")
		if _, err := os.Stat(mdbPath); !os.IsNotExist(err) {
			if os.IsPermission(err) {
				c.Ui.Error(fmt.Sprintf("CRITICAL: Permission denied for data folder at %q!", mdbPath))
				c.Ui.Error("Consul will refuse to boot without access to this directory.")
				c.Ui.Error("Please correct permissions and try starting again.")
				return nil
			}
			c.Ui.Error(fmt.Sprintf("CRITICAL: Deprecated data folder found at %q!", mdbPath))
			c.Ui.Error("Consul will refuse to boot with this directory present.")
			c.Ui.Error("See https://www.consul.io/docs/upgrade-specific.html for more information.")
			return nil
		}
	}

	// Verify DNS settings
	if config.DNSConfig.UDPAnswerLimit < 1 {
		c.Ui.Error(fmt.Sprintf("dns_config.udp_answer_limit %d too low, must always be greater than zero", config.DNSConfig.UDPAnswerLimit))
	}

	if config.EncryptKey != "" {
		if _, err := config.EncryptBytes(); err != nil {
			c.Ui.Error(fmt.Sprintf("Invalid encryption key: %s", err))
			return nil
		}
		keyfileLAN := filepath.Join(config.DataDir, serfLANKeyring)
		if _, err := os.Stat(keyfileLAN); err == nil {
			c.Ui.Error("WARNING: LAN keyring exists but -encrypt given, using keyring")
		}
		if config.Server {
			keyfileWAN := filepath.Join(config.DataDir, serfWANKeyring)
			if _, err := os.Stat(keyfileWAN); err == nil {
				c.Ui.Error("WARNING: WAN keyring exists but -encrypt given, using keyring")
			}
		}
	}

	// Output a warning if the 'dc' flag has been used.
	if dcDeprecated != "" {
		c.Ui.Error("WARNING: the 'dc' flag has been deprecated. Use 'datacenter' instead")

		// Making sure that we don't break previous versions.
		config.Datacenter = dcDeprecated
	}

	// Ensure the datacenter is always lowercased. The DNS endpoints automatically
	// lowercase all queries, and internally we expect DC1 and dc1 to be the same.
	config.Datacenter = strings.ToLower(config.Datacenter)

	// Verify datacenter is valid
	if !validDatacenter.MatchString(config.Datacenter) {
		c.Ui.Error("Datacenter must be alpha-numeric with underscores and hypens only")
		return nil
	}

	// If 'acl_datacenter' is set, ensure it is lowercased.
	if config.ACLDatacenter != "" {
		config.ACLDatacenter = strings.ToLower(config.ACLDatacenter)

		// Verify 'acl_datacenter' is valid
		if !validDatacenter.MatchString(config.ACLDatacenter) {
			c.Ui.Error("ACL datacenter must be alpha-numeric with underscores and hypens only")
			return nil
		}
	}

	// Only allow bootstrap mode when acting as a server
	if config.Bootstrap && !config.Server {
		c.Ui.Error("Bootstrap mode cannot be enabled when server mode is not enabled")
		return nil
	}

	// Expect can only work when acting as a server
	if config.BootstrapExpect != 0 && !config.Server {
		c.Ui.Error("Expect mode cannot be enabled when server mode is not enabled")
		return nil
	}

	// Expect can only work when dev mode is off
	if config.BootstrapExpect > 0 && config.DevMode {
		c.Ui.Error("Expect mode cannot be enabled when dev mode is enabled")
		return nil
	}

	// Expect & Bootstrap are mutually exclusive
	if config.BootstrapExpect != 0 && config.Bootstrap {
		c.Ui.Error("Bootstrap cannot be provided with an expected server count")
		return nil
	}

	// Compile all the watches
	for _, params := range config.Watches {
		// Parse the watches, excluding the handler
		wp, err := watch.ParseExempt(params, []string{"handler"})
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to parse watch (%#v): %v", params, err))
			return nil
		}

		// Get the handler
		if err := verifyWatchHandler(wp.Exempt["handler"]); err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to setup watch handler (%#v): %v", params, err))
			return nil
		}

		// Store the watch plan
		config.WatchPlans = append(config.WatchPlans, wp)
	}

	// Warn if we are in expect mode
	if config.BootstrapExpect == 1 {
		c.Ui.Error("WARNING: BootstrapExpect Mode is specified as 1; this is the same as Bootstrap mode.")
		config.BootstrapExpect = 0
		config.Bootstrap = true
	} else if config.BootstrapExpect > 0 {
		c.Ui.Error(fmt.Sprintf("WARNING: Expect Mode enabled, expecting %d servers", config.BootstrapExpect))
	}

	// Warn if we are in bootstrap mode
	if config.Bootstrap {
		c.Ui.Error("WARNING: Bootstrap mode enabled! Do not enable unless necessary")
	}

	// Need both tag key and value for EC2 discovery
	if config.RetryJoinEC2.TagKey != "" || config.RetryJoinEC2.TagValue != "" {
		if config.RetryJoinEC2.TagKey == "" || config.RetryJoinEC2.TagValue == "" {
			c.Ui.Error("tag key and value are both required for EC2 retry-join")
			return nil
		}
	}

	// EC2 and GCE discovery are mutually exclusive
	if config.RetryJoinEC2.TagKey != "" && config.RetryJoinEC2.TagValue != "" && config.RetryJoinGCE.TagValue != "" {
		c.Ui.Error("EC2 and GCE discovery are mutually exclusive. Please provide one or the other.")
		return nil
	}

	// Verify the node metadata entries are valid
	if err := structs.ValidateMetadata(config.Meta); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse node metadata: %v", err))
		return nil
	}

	// Set the version info
	config.Revision = c.Revision
	config.Version = c.Version
	config.VersionPrerelease = c.VersionPrerelease

	return config
}

// verifyUniqueListeners checks to see if an address was used more than once in
// the config
func (config *Config) verifyUniqueListeners() error {
	listeners := []struct {
		host  string
		port  int
		descr string
	}{
		{config.Addresses.DNS, config.Ports.DNS, "DNS"},
		{config.Addresses.HTTP, config.Ports.HTTP, "HTTP"},
		{config.Addresses.HTTPS, config.Ports.HTTPS, "HTTPS"},
		{config.AdvertiseAddr, config.Ports.Server, "Server RPC"},
		{config.AdvertiseAddr, config.Ports.SerfLan, "Serf LAN"},
		{config.AdvertiseAddr, config.Ports.SerfWan, "Serf WAN"},
	}

	type key struct {
		host string
		port int
	}
	m := make(map[key]string, len(listeners))

	for _, l := range listeners {
		if l.host == "" {
			l.host = "0.0.0.0"
		} else if strings.HasPrefix(l.host, "unix") {
			// Don't compare ports on unix sockets
			l.port = 0
		}
		if l.host == "0.0.0.0" && l.port <= 0 {
			continue
		}

		k := key{l.host, l.port}
		v, ok := m[k]
		if ok {
			return fmt.Errorf("%s address already configured for %s", l.descr, v)
		}
		m[k] = l.descr
	}
	return nil
}

// discoverEc2Hosts searches an AWS region, returning a list of instance ips
// where EC2TagKey = EC2TagValue
func (c *Config) discoverEc2Hosts(logger *log.Logger) ([]string, error) {
	config := c.RetryJoinEC2

	ec2meta := ec2metadata.New(session.New())
	if config.Region == "" {
		logger.Printf("[INFO] agent: No EC2 region provided, querying instance metadata endpoint...")
		identity, err := ec2meta.GetInstanceIdentityDocument()
		if err != nil {
			return nil, err
		}
		config.Region = identity.Region
	}

	awsConfig := &aws.Config{
		Region: &config.Region,
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     config.AccessKeyID,
						SecretAccessKey: config.SecretAccessKey,
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				defaults.RemoteCredProvider(*(defaults.Config()), defaults.Handlers()),
			}),
	}

	svc := ec2.New(session.New(), awsConfig)

	resp, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + config.TagKey),
				Values: []*string{
					aws.String(config.TagValue),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	var servers []string
	for i := range resp.Reservations {
		for _, instance := range resp.Reservations[i].Instances {
			// Terminated instances don't have the PrivateIpAddress field
			if instance.PrivateIpAddress != nil {
				servers = append(servers, *instance.PrivateIpAddress)
			}
		}
	}

	return servers, nil
}

// discoverGCEHosts searches a Google Compute Engine region, returning a list
// of instance ips that match the tags given in GCETags.
func (c *Config) discoverGCEHosts(logger *log.Logger) ([]string, error) {
	config := c.RetryJoinGCE
	ctx := oauth2.NoContext
	var client *http.Client
	var err error

	logger.Printf("[INFO] agent: Initializing GCE client")
	if config.CredentialsFile != "" {
		logger.Printf("[INFO] agent: Loading credentials from %s", config.CredentialsFile)
		key, err := ioutil.ReadFile(config.CredentialsFile)
		if err != nil {
			return nil, err
		}
		jwtConfig, err := google.JWTConfigFromJSON(key, compute.ComputeScope)
		if err != nil {
			return nil, err
		}
		client = jwtConfig.Client(ctx)
	} else {
		logger.Printf("[INFO] agent: Using default credential chain")
		client, err = google.DefaultClient(ctx, compute.ComputeScope)
		if err != nil {
			return nil, err
		}
	}

	computeService, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	if config.ProjectName == "" {
		logger.Printf("[INFO] agent: No GCE project provided, will discover from metadata.")
		config.ProjectName, err = gceProjectIDFromMetadata(logger)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Printf("[INFO] agent: Using pre-defined GCE project name: %s", config.ProjectName)
	}

	zones, err := gceDiscoverZones(logger, ctx, computeService, config.ProjectName, config.ZonePattern)
	if err != nil {
		return nil, err
	}

	logger.Printf("[INFO] agent: Discovering GCE hosts with tag %s in zones: %s", config.TagValue, strings.Join(zones, ", "))

	var servers []string
	for _, zone := range zones {
		addresses, err := gceInstancesAddressesForZone(logger, ctx, computeService, config.ProjectName, zone, config.TagValue)
		if err != nil {
			return nil, err
		}
		if len(addresses) > 0 {
			logger.Printf("[INFO] agent: Discovered %d instances in %s/%s: %v", len(addresses), config.ProjectName, zone, addresses)
		}
		servers = append(servers, addresses...)
	}

	return servers, nil
}

// gceProjectIDFromMetadata queries the metadata service on GCE to get the
// project ID (name) of an instance.
func gceProjectIDFromMetadata(logger *log.Logger) (string, error) {
	logger.Printf("[INFO] agent: Attempting to discover GCE project from metadata.")
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	project, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	logger.Printf("[INFO] agent: GCE project discovered as: %s", project)
	return string(project), nil
}

// gceDiscoverZones discovers a list of zones from a supplied zone pattern, or
// all of the zones available to a project.
func gceDiscoverZones(logger *log.Logger, ctx context.Context, computeService *compute.Service, project, pattern string) ([]string, error) {
	var zones []string

	if pattern != "" {
		logger.Printf("[INFO] agent: Discovering zones for project %s matching pattern: %s", project, pattern)
	} else {
		logger.Printf("[INFO] agent: Discovering all zones available to project: %s", project)
	}

	call := computeService.Zones.List(project)
	if pattern != "" {
		call = call.Filter(fmt.Sprintf("name eq %s", pattern))
	}

	if err := call.Pages(ctx, func(page *compute.ZoneList) error {
		for _, v := range page.Items {
			zones = append(zones, v.Name)
		}
		return nil
	}); err != nil {
		return zones, err
	}

	logger.Printf("[INFO] agent: Discovered GCE zones: %s", strings.Join(zones, ", "))
	return zones, nil
}

// gceInstancesAddressesForZone locates all instances within a specific project
// and zone, matching the supplied tag. Only the private IP addresses are
// returned, but ID is also logged.
func gceInstancesAddressesForZone(logger *log.Logger, ctx context.Context, computeService *compute.Service, project, zone, tag string) ([]string, error) {
	var addresses []string
	call := computeService.Instances.List(project, zone)
	if err := call.Pages(ctx, func(page *compute.InstanceList) error {
		for _, v := range page.Items {
			for _, t := range v.Tags.Items {
				if t == tag && len(v.NetworkInterfaces) > 0 && v.NetworkInterfaces[0].NetworkIP != "" {
					addresses = append(addresses, v.NetworkInterfaces[0].NetworkIP)
				}
			}
		}
		return nil
	}); err != nil {
		return addresses, err
	}

	return addresses, nil
}

// setupAgent is used to start the agent and various interfaces
func (c *Command) setupAgent(config *Config, logOutput io.Writer, logWriter *logger.LogWriter) error {
	c.Ui.Output("Starting Consul agent...")
	agent, err := Create(config, logOutput, logWriter, c.configReloadCh)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error starting agent: %s", err))
		return err
	}
	c.agent = agent

	// Enable the SCADA integration
	if err := c.setupScadaConn(config); err != nil {
		agent.Shutdown()
		c.Ui.Error(fmt.Sprintf("Error starting SCADA connection: %s", err))
		return err
	}

	if config.Ports.HTTP > 0 || config.Ports.HTTPS > 0 {
		servers, err := NewHTTPServers(agent, config, logOutput)
		if err != nil {
			agent.Shutdown()
			c.Ui.Error(fmt.Sprintf("Error starting http servers: %s", err))
			return err
		}
		c.httpServers = servers
	}

	if config.Ports.DNS > 0 {
		dnsAddr, err := config.ClientListener(config.Addresses.DNS, config.Ports.DNS)
		if err != nil {
			agent.Shutdown()
			c.Ui.Error(fmt.Sprintf("Invalid DNS bind address: %s", err))
			return err
		}

		server, err := NewDNSServer(agent, &config.DNSConfig, logOutput,
			config.Domain, dnsAddr.String(), config.DNSRecursors)
		if err != nil {
			agent.Shutdown()
			c.Ui.Error(fmt.Sprintf("Error starting dns server: %s", err))
			return err
		}
		c.dnsServer = server
	}

	// Setup update checking
	if !config.DisableUpdateCheck {
		version := config.Version
		if config.VersionPrerelease != "" {
			version += fmt.Sprintf("-%s", config.VersionPrerelease)
		}
		updateParams := &checkpoint.CheckParams{
			Product: "consul",
			Version: version,
		}
		if !config.DisableAnonymousSignature {
			updateParams.SignatureFile = filepath.Join(config.DataDir, "checkpoint-signature")
		}

		// Schedule a periodic check with expected interval of 24 hours
		checkpoint.CheckInterval(updateParams, 24*time.Hour, c.checkpointResults)

		// Do an immediate check within the next 30 seconds
		go func() {
			time.Sleep(lib.RandomStagger(30 * time.Second))
			c.checkpointResults(checkpoint.Check(updateParams))
		}()
	}
	return nil
}

// checkpointResults is used to handler periodic results from our update checker
func (c *Command) checkpointResults(results *checkpoint.CheckResponse, err error) {
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to check for updates: %v", err))
		return
	}
	if results.Outdated {
		c.Ui.Error(fmt.Sprintf("Newer Consul version available: %s (currently running: %s)", results.CurrentVersion, c.Version))
	}
	for _, alert := range results.Alerts {
		switch alert.Level {
		case "info":
			c.Ui.Info(fmt.Sprintf("Bulletin [%s]: %s (%s)", alert.Level, alert.Message, alert.URL))
		default:
			c.Ui.Error(fmt.Sprintf("Bulletin [%s]: %s (%s)", alert.Level, alert.Message, alert.URL))
		}
	}
}

// startupJoin is invoked to handle any joins specified to take place at start time
func (c *Command) startupJoin(config *Config) error {
	if len(config.StartJoin) == 0 {
		return nil
	}

	c.Ui.Output("Joining cluster...")
	n, err := c.agent.JoinLAN(config.StartJoin)
	if err != nil {
		return err
	}

	c.Ui.Info(fmt.Sprintf("Join completed. Synced with %d initial agents", n))
	return nil
}

// startupJoinWan is invoked to handle any joins -wan specified to take place at start time
func (c *Command) startupJoinWan(config *Config) error {
	if len(config.StartJoinWan) == 0 {
		return nil
	}

	c.Ui.Output("Joining -wan cluster...")
	n, err := c.agent.JoinWAN(config.StartJoinWan)
	if err != nil {
		return err
	}

	c.Ui.Info(fmt.Sprintf("Join -wan completed. Synced with %d initial agents", n))
	return nil
}

// retryJoin is used to handle retrying a join until it succeeds or all
// retries are exhausted.
func (c *Command) retryJoin(config *Config, errCh chan<- struct{}) {
	ec2Enabled := config.RetryJoinEC2.TagKey != "" && config.RetryJoinEC2.TagValue != ""

	if len(config.RetryJoin) == 0 && !ec2Enabled && config.RetryJoinGCE.TagValue == "" {
		return
	}

	logger := c.agent.logger
	logger.Printf("[INFO] agent: Joining cluster...")

	attempt := 0
	for {
		var servers []string
		var err error
		switch {
		case ec2Enabled:
			servers, err = config.discoverEc2Hosts(logger)
			if err != nil {
				logger.Printf("[ERROR] agent: Unable to query EC2 instances: %s", err)
			}
			logger.Printf("[INFO] agent: Discovered %d servers from EC2", len(servers))
		case config.RetryJoinGCE.TagValue != "":
			servers, err = config.discoverGCEHosts(logger)
			if err != nil {
				logger.Printf("[ERROR] agent: Unable to query GCE insances: %s", err)
			}
			logger.Printf("[INFO] agent: Discovered %d servers from GCE", len(servers))
		}

		servers = append(servers, config.RetryJoin...)
		if len(servers) == 0 {
			err = fmt.Errorf("No servers to join")
		} else {
			n, err := c.agent.JoinLAN(servers)
			if err == nil {
				logger.Printf("[INFO] agent: Join completed. Synced with %d initial agents", n)
				return
			}
		}

		attempt++
		if config.RetryMaxAttempts > 0 && attempt > config.RetryMaxAttempts {
			logger.Printf("[ERROR] agent: max join retry exhausted, exiting")
			close(errCh)
			return
		}

		logger.Printf("[WARN] agent: Join failed: %v, retrying in %v", err,
			config.RetryInterval)
		time.Sleep(config.RetryInterval)
	}
}

// retryJoinWan is used to handle retrying a join -wan until it succeeds or all
// retries are exhausted.
func (c *Command) retryJoinWan(config *Config, errCh chan<- struct{}) {
	if len(config.RetryJoinWan) == 0 {
		return
	}

	logger := c.agent.logger
	logger.Printf("[INFO] agent: Joining WAN cluster...")

	attempt := 0
	for {
		n, err := c.agent.JoinWAN(config.RetryJoinWan)
		if err == nil {
			logger.Printf("[INFO] agent: Join -wan completed. Synced with %d initial agents", n)
			return
		}

		attempt++
		if config.RetryMaxAttemptsWan > 0 && attempt > config.RetryMaxAttemptsWan {
			logger.Printf("[ERROR] agent: max join -wan retry exhausted, exiting")
			close(errCh)
			return
		}

		logger.Printf("[WARN] agent: Join -wan failed: %v, retrying in %v", err,
			config.RetryIntervalWan)
		time.Sleep(config.RetryIntervalWan)
	}
}

// gossipEncrypted determines if the consul instance is using symmetric
// encryption keys to protect gossip protocol messages.
func (c *Command) gossipEncrypted() bool {
	if c.agent.config.EncryptKey != "" {
		return true
	}

	server := c.agent.server
	if server != nil {
		return server.KeyManagerLAN() != nil || server.KeyManagerWAN() != nil
	}

	client := c.agent.client
	return client != nil && client.KeyManagerLAN() != nil
}

func (c *Command) Run(args []string) int {
	c.Ui = &cli.PrefixedUi{
		OutputPrefix: "==> ",
		InfoPrefix:   "    ",
		ErrorPrefix:  "==> ",
		Ui:           c.Ui,
	}

	// Parse our configs
	c.args = args
	config := c.readConfig()
	if config == nil {
		return 1
	}

	// Setup the log outputs
	logConfig := &logger.Config{
		LogLevel:       config.LogLevel,
		EnableSyslog:   config.EnableSyslog,
		SyslogFacility: config.SyslogFacility,
	}
	logFilter, logGate, logWriter, logOutput, ok := logger.Setup(logConfig, c.Ui)
	if !ok {
		return 1
	}
	c.logFilter = logFilter
	c.logOutput = logOutput

	// Setup the channel for triggering config reloads
	c.configReloadCh = make(chan chan error)

	/* Setup telemetry
	Aggregate on 10 second intervals for 1 minute. Expose the
	metrics over stderr when there is a SIGUSR1 received.
	*/
	inm := metrics.NewInmemSink(10*time.Second, time.Minute)
	metrics.DefaultInmemSignal(inm)
	metricsConf := metrics.DefaultConfig(config.Telemetry.StatsitePrefix)
	metricsConf.EnableHostname = !config.Telemetry.DisableHostname

	// Configure the statsite sink
	var fanout metrics.FanoutSink
	if config.Telemetry.StatsiteAddr != "" {
		sink, err := metrics.NewStatsiteSink(config.Telemetry.StatsiteAddr)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to start statsite sink. Got: %s", err))
			return 1
		}
		fanout = append(fanout, sink)
	}

	// Configure the statsd sink
	if config.Telemetry.StatsdAddr != "" {
		sink, err := metrics.NewStatsdSink(config.Telemetry.StatsdAddr)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to start statsd sink. Got: %s", err))
			return 1
		}
		fanout = append(fanout, sink)
	}

	// Configure the DogStatsd sink
	if config.Telemetry.DogStatsdAddr != "" {
		var tags []string

		if config.Telemetry.DogStatsdTags != nil {
			tags = config.Telemetry.DogStatsdTags
		}

		sink, err := datadog.NewDogStatsdSink(config.Telemetry.DogStatsdAddr, metricsConf.HostName)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to start DogStatsd sink. Got: %s", err))
			return 1
		}
		sink.SetTags(tags)
		fanout = append(fanout, sink)
	}

	if config.Telemetry.CirconusAPIToken != "" || config.Telemetry.CirconusCheckSubmissionURL != "" {
		cfg := &circonus.Config{}
		cfg.Interval = config.Telemetry.CirconusSubmissionInterval
		cfg.CheckManager.API.TokenKey = config.Telemetry.CirconusAPIToken
		cfg.CheckManager.API.TokenApp = config.Telemetry.CirconusAPIApp
		cfg.CheckManager.API.URL = config.Telemetry.CirconusAPIURL
		cfg.CheckManager.Check.SubmissionURL = config.Telemetry.CirconusCheckSubmissionURL
		cfg.CheckManager.Check.ID = config.Telemetry.CirconusCheckID
		cfg.CheckManager.Check.ForceMetricActivation = config.Telemetry.CirconusCheckForceMetricActivation
		cfg.CheckManager.Check.InstanceID = config.Telemetry.CirconusCheckInstanceID
		cfg.CheckManager.Check.SearchTag = config.Telemetry.CirconusCheckSearchTag
		cfg.CheckManager.Check.DisplayName = config.Telemetry.CirconusCheckDisplayName
		cfg.CheckManager.Check.Tags = config.Telemetry.CirconusCheckTags
		cfg.CheckManager.Broker.ID = config.Telemetry.CirconusBrokerID
		cfg.CheckManager.Broker.SelectTag = config.Telemetry.CirconusBrokerSelectTag

		if cfg.CheckManager.Check.DisplayName == "" {
			cfg.CheckManager.Check.DisplayName = "Consul"
		}

		if cfg.CheckManager.API.TokenApp == "" {
			cfg.CheckManager.API.TokenApp = "consul"
		}

		if cfg.CheckManager.Check.SearchTag == "" {
			cfg.CheckManager.Check.SearchTag = "service:consul"
		}

		sink, err := circonus.NewCirconusSink(cfg)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to start Circonus sink. Got: %s", err))
			return 1
		}
		sink.Start()
		fanout = append(fanout, sink)
	}

	// Initialize the global sink
	if len(fanout) > 0 {
		fanout = append(fanout, inm)
		metrics.NewGlobal(metricsConf, fanout)
	} else {
		metricsConf.EnableHostname = false
		metrics.NewGlobal(metricsConf, inm)
	}

	// Create the agent
	if err := c.setupAgent(config, logOutput, logWriter); err != nil {
		return 1
	}
	defer c.agent.Shutdown()
	if c.dnsServer != nil {
		defer c.dnsServer.Shutdown()
	}
	for _, server := range c.httpServers {
		defer server.Shutdown()
	}

	// Check and shut down the SCADA listeners at the end
	defer func() {
		if c.scadaHttp != nil {
			c.scadaHttp.Shutdown()
		}
		if c.scadaProvider != nil {
			c.scadaProvider.Shutdown()
		}
	}()

	// Join startup nodes if specified
	if err := c.startupJoin(config); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// Join startup nodes if specified
	if err := c.startupJoinWan(config); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// Get the new client http listener addr
	var httpAddr net.Addr
	var err error
	if config.Ports.HTTP != -1 {
		httpAddr, err = config.ClientListener(config.Addresses.HTTP, config.Ports.HTTP)
	} else if config.Ports.HTTPS != -1 {
		httpAddr, err = config.ClientListener(config.Addresses.HTTPS, config.Ports.HTTPS)
	} else if len(config.WatchPlans) > 0 {
		c.Ui.Error("Error: cannot use watches if both HTTP and HTTPS are disabled")
		return 1
	}
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to determine HTTP address: %v", err))
	}

	// Register the watches
	for _, wp := range config.WatchPlans {
		go func(wp *watch.WatchPlan) {
			wp.Handler = makeWatchHandler(logOutput, wp.Exempt["handler"])
			wp.LogOutput = c.logOutput
			addr := httpAddr.String()
			// If it's a unix socket, prefix with unix:// so the client initializes correctly
			if httpAddr.Network() == "unix" {
				addr = "unix://" + addr
			}
			if err := wp.Run(addr); err != nil {
				c.Ui.Error(fmt.Sprintf("Error running watch: %v", err))
			}
		}(wp)
	}

	// Figure out if gossip is encrypted
	var gossipEncrypted bool
	if config.Server {
		gossipEncrypted = c.agent.server.Encrypted()
	} else {
		gossipEncrypted = c.agent.client.Encrypted()
	}

	// Determine the Atlas cluster
	atlas := "<disabled>"
	if config.AtlasInfrastructure != "" {
		atlas = fmt.Sprintf("(Infrastructure: '%s' Join: %v)", config.AtlasInfrastructure, config.AtlasJoin)
	}

	// Let the agent know we've finished registration
	c.agent.StartSync()

	c.Ui.Output("Consul agent running!")
	c.Ui.Info(fmt.Sprintf("       Version: '%s'", c.HumanVersion))
	c.Ui.Info(fmt.Sprintf("       Node ID: '%s'", config.NodeID))
	c.Ui.Info(fmt.Sprintf("     Node name: '%s'", config.NodeName))
	c.Ui.Info(fmt.Sprintf("    Datacenter: '%s'", config.Datacenter))
	c.Ui.Info(fmt.Sprintf("        Server: %v (bootstrap: %v)", config.Server, config.Bootstrap))
	c.Ui.Info(fmt.Sprintf("   Client Addr: %v (HTTP: %d, HTTPS: %d, DNS: %d)", config.ClientAddr,
		config.Ports.HTTP, config.Ports.HTTPS, config.Ports.DNS))
	c.Ui.Info(fmt.Sprintf("  Cluster Addr: %v (LAN: %d, WAN: %d)", config.AdvertiseAddr,
		config.Ports.SerfLan, config.Ports.SerfWan))
	c.Ui.Info(fmt.Sprintf("Gossip encrypt: %v, RPC-TLS: %v, TLS-Incoming: %v",
		gossipEncrypted, config.VerifyOutgoing, config.VerifyIncoming))
	c.Ui.Info(fmt.Sprintf("         Atlas: %s", atlas))

	// Enable log streaming
	c.Ui.Info("")
	c.Ui.Output("Log data will now stream in as it occurs:\n")
	logGate.Flush()

	// Start retry join process
	errCh := make(chan struct{})
	go c.retryJoin(config, errCh)

	// Start retry -wan join process
	errWanCh := make(chan struct{})
	go c.retryJoinWan(config, errWanCh)

	// Wait for exit
	return c.handleSignals(config, errCh, errWanCh)
}

// handleSignals blocks until we get an exit-causing signal
func (c *Command) handleSignals(config *Config, retryJoin <-chan struct{}, retryJoinWan <-chan struct{}) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGPIPE)

	// Wait for a signal
WAIT:
	var sig os.Signal
	var reloadErrCh chan error
	select {
	case s := <-signalCh:
		sig = s
	case ch := <-c.configReloadCh:
		sig = syscall.SIGHUP
		reloadErrCh = ch
	case <-c.ShutdownCh:
		sig = os.Interrupt
	case <-retryJoin:
		return 1
	case <-retryJoinWan:
		return 1
	case <-c.agent.ShutdownCh():
		// Agent is already shutdown!
		return 0
	}
	c.Ui.Output(fmt.Sprintf("Caught signal: %v", sig))

	// Skip SIGPIPE signals
	if sig == syscall.SIGPIPE {
		goto WAIT
	}

	// Check if this is a SIGHUP
	if sig == syscall.SIGHUP {
		conf, err := c.handleReload(config)
		if conf != nil {
			config = conf
		}
		if err != nil {
			c.Ui.Error(err.Error())
		}
		// Send result back if reload was called via HTTP
		if reloadErrCh != nil {
			reloadErrCh <- err
		}
		goto WAIT
	}

	// Check if we should do a graceful leave
	graceful := false
	if sig == os.Interrupt && !(*config.SkipLeaveOnInt) {
		graceful = true
	} else if sig == syscall.SIGTERM && (*config.LeaveOnTerm) {
		graceful = true
	}

	// Bail fast if not doing a graceful leave
	if !graceful {
		return 1
	}

	// Attempt a graceful leave
	gracefulCh := make(chan struct{})
	c.Ui.Output("Gracefully shutting down agent...")
	go func() {
		if err := c.agent.Leave(); err != nil {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
			return
		}
		close(gracefulCh)
	}()

	// Wait for leave or another signal
	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

// handleReload is invoked when we should reload our configs, e.g. SIGHUP
func (c *Command) handleReload(config *Config) (*Config, error) {
	c.Ui.Output("Reloading configuration...")
	var errs error
	newConf := c.readConfig()
	if newConf == nil {
		errs = multierror.Append(errs, fmt.Errorf("Failed to reload configs"))
		return config, errs
	}

	// Change the log level
	minLevel := logutils.LogLevel(strings.ToUpper(newConf.LogLevel))
	if logger.ValidateLevelFilter(minLevel, c.logFilter) {
		c.logFilter.SetMinLevel(minLevel)
	} else {
		errs = multierror.Append(fmt.Errorf(
			"Invalid log level: %s. Valid log levels are: %v",
			minLevel, c.logFilter.Levels))

		// Keep the current log level
		newConf.LogLevel = config.LogLevel
	}

	// Bulk update the services and checks
	c.agent.PauseSync()
	defer c.agent.ResumeSync()

	// Snapshot the current state, and restore it afterwards
	snap := c.agent.snapshotCheckState()
	defer c.agent.restoreCheckState(snap)

	// First unload all checks, services, and metadata. This lets us begin the reload
	// with a clean slate.
	if err := c.agent.unloadServices(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Failed unloading services: %s", err))
		return nil, errs
	}
	if err := c.agent.unloadChecks(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Failed unloading checks: %s", err))
		return nil, errs
	}
	c.agent.unloadMetadata()

	// Reload service/check definitions and metadata.
	if err := c.agent.loadServices(newConf); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Failed reloading services: %s", err))
		return nil, errs
	}
	if err := c.agent.loadChecks(newConf); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Failed reloading checks: %s", err))
		return nil, errs
	}
	if err := c.agent.loadMetadata(newConf); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Failed reloading metadata: %s", err))
		return nil, errs
	}

	// Get the new client listener addr
	httpAddr, err := newConf.ClientListener(config.Addresses.HTTP, config.Ports.HTTP)
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Failed to determine HTTP address: %v", err))
	}

	// Deregister the old watches
	for _, wp := range config.WatchPlans {
		wp.Stop()
	}

	// Register the new watches
	for _, wp := range newConf.WatchPlans {
		go func(wp *watch.WatchPlan) {
			wp.Handler = makeWatchHandler(c.logOutput, wp.Exempt["handler"])
			wp.LogOutput = c.logOutput
			if err := wp.Run(httpAddr.String()); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Error running watch: %v", err))
			}
		}(wp)
	}

	// Reload SCADA client if we have a change
	if newConf.AtlasInfrastructure != config.AtlasInfrastructure ||
		newConf.AtlasToken != config.AtlasToken ||
		newConf.AtlasEndpoint != config.AtlasEndpoint {
		if err := c.setupScadaConn(newConf); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Failed reloading SCADA client: %s", err))
			return nil, errs
		}
	}

	return newConf, errs
}

// startScadaClient is used to start a new SCADA provider and listener,
// replacing any existing listeners.
func (c *Command) setupScadaConn(config *Config) error {
	// Shut down existing SCADA listeners
	if c.scadaProvider != nil {
		c.scadaProvider.Shutdown()
	}
	if c.scadaHttp != nil {
		c.scadaHttp.Shutdown()
	}

	// No-op if we don't have an infrastructure
	if config.AtlasInfrastructure == "" {
		return nil
	}

	c.Ui.Error("WARNING: The hosted version of Consul Enterprise will be deprecated " +
		"on March 7th, 2017. For details, see " +
		"https://atlas.hashicorp.com/help/consul/alternatives")

	scadaConfig := &scada.Config{
		Service:      "consul",
		Version:      fmt.Sprintf("%s%s", config.Version, config.VersionPrerelease),
		ResourceType: "infrastructures",
		Meta: map[string]string{
			"auto-join":  strconv.FormatBool(config.AtlasJoin),
			"datacenter": config.Datacenter,
			"server":     strconv.FormatBool(config.Server),
		},
		Atlas: scada.AtlasConfig{
			Endpoint:       config.AtlasEndpoint,
			Infrastructure: config.AtlasInfrastructure,
			Token:          config.AtlasToken,
		},
	}

	// Create the new provider and listener
	c.Ui.Output("Connecting to Atlas: " + config.AtlasInfrastructure)
	provider, list, err := scada.NewHTTPProvider(scadaConfig, c.logOutput)
	if err != nil {
		return err
	}
	c.scadaProvider = provider
	c.scadaHttp = newScadaHttp(c.agent, list)
	return nil
}

func (c *Command) Synopsis() string {
	return "Runs a Consul agent"
}

func (c *Command) Help() string {
	helpText := `
Usage: consul agent [options]

  Starts the Consul agent and runs until an interrupt is received. The
  agent represents a single node in a cluster.

 ` + c.Command.Help()

	return strings.TrimSpace(helpText)
}
