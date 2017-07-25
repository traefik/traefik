---
layout: "docs"
page_title: "Configuration"
sidebar_current: "docs-agent-config"
description: |-
  The agent has various configuration options that can be specified via the command-line or via configuration files. All of the configuration options are completely optional. Defaults are specified with their descriptions.
---

# Configuration

The agent has various configuration options that can be specified via
the command-line or via configuration files. All of the configuration
options are completely optional. Defaults are specified with their
descriptions.

Configuration precedence is evaluated in the following order:

1. Command line arguments
2. Environment Variables
3. Configuration files

When loading configuration, Consul loads the configuration from files
and directories in lexical order. For example, configuration file `basic_config.json`
will be processed before `extra_config.json`. Configuration specified later
will be merged into configuration specified earlier. In most cases,
"merge" means that the later version will override the earlier. In
some cases, such as event handlers, merging appends the handlers to the
existing configuration. The exact merging behavior is specified for each
option below.

Consul also supports reloading configuration when it receives the
SIGHUP signal. Not all changes are respected, but those that are
are documented below in the
[Reloadable Configuration](#reloadable-configuration) section. The
[reload command](/docs/commands/reload.html) can also be used to trigger a
configuration reload.

## <a name="commandline_options"></a>Command-line Options

The options below are all specified on the command-line.

* <a name="_advertise"></a><a href="#_advertise">`-advertise`</a> - The advertise
  address is used to change the address that we
  advertise to other nodes in the cluster. By default, the [`-bind`](#_bind) address is
  advertised. However, in some cases, there may be a routable address that cannot
  be bound. This flag enables gossiping a different address to support this.
  If this address is not routable, the node will be in a constant flapping state
  as other nodes will treat the non-routability as a failure.

* <a name="_advertise-wan"></a><a href="#_advertise-wan">`-advertise-wan`</a> - The
  advertise WAN address is used to change the address that we advertise to server nodes
  joining through the WAN. This can also be set on client agents when used in combination
  with the <a href="#translate_wan_addrs">`translate_wan_addrs`</a> configuration
  option. By default, the [`-advertise`](#_advertise) address is advertised. However, in some
  cases all members of all datacenters cannot be on the same physical or virtual network,
  especially on hybrid setups mixing cloud and private datacenters. This flag enables server
  nodes gossiping through the public network for the WAN while using private VLANs for gossiping
  to each other and their client agents, and it allows client agents to be reached at this
  address when being accessed from a remote datacenter if the remote datacenter is configured
  with <a href="#translate_wan_addrs">`translate_wan_addrs`</a>.

~> **Notice:** The hosted version of Consul Enterprise will be deprecated on
  March 7th, 2017. For details, see https://atlas.hashicorp.com/help/consul/alternatives

* <a name="_atlas"></a><a href="#_atlas">`-atlas`</a> - This flag
  enables [Atlas](https://atlas.hashicorp.com) integration.
  It is used to provide the Atlas infrastructure name and the SCADA connection. The format of
  this is `username/environment`. This enables Atlas features such as the Monitoring UI
  and node auto joining.

* <a name="_atlas_join"></a><a href="#_atlas_join">`-atlas-join`</a> - When set, enables auto-join
  via Atlas. Atlas will track the most
  recent members to join the infrastructure named by [`-atlas`](#_atlas) and automatically
  join them on start. For servers, the LAN and WAN pool are both joined.

* <a name="_atlas_token"></a><a href="#_atlas_token">`-atlas-token`</a> - Provides the Atlas
  API authentication token. This can also be provided
  using the `ATLAS_TOKEN` environment variable. Required for use with Atlas.

* <a name="_atlas_endpoint"></a><a href="#_atlas_endpoint">`-atlas-endpoint`</a> - The endpoint
  address used for Atlas integration. Used only if the `-atlas` and
  `-atlas-token` options are specified. This is optional, and defaults to the
  public Atlas endpoints. This can also be specified using the `SCADA_ENDPOINT`
  environment variable. The CLI option takes precedence, followed by the
  configuration file directive, and lastly, the environment variable.

* <a name="_bootstrap"></a><a href="#_bootstrap">`-bootstrap`</a> - This flag is used to control if a
  server is in "bootstrap" mode. It is important that
  no more than one server *per* datacenter be running in this mode. Technically, a server in bootstrap mode
  is allowed to self-elect as the Raft leader. It is important that only a single node is in this mode;
  otherwise, consistency cannot be guaranteed as multiple nodes are able to self-elect.
  It is not recommended to use this flag after a cluster has been bootstrapped.

* <a name="_bootstrap_expect"></a><a href="#_bootstrap_expect">`-bootstrap-expect`</a> - This flag
  provides the number of expected servers in the datacenter.
  Either this value should not be provided or the value must agree with other servers in
  the cluster. When provided, Consul waits until the specified number of servers are
  available and then bootstraps the cluster. This allows an initial leader to be elected
  automatically. This cannot be used in conjunction with the legacy [`-bootstrap`](#_bootstrap) flag.
  This flag implies server mode.

* <a name="_bind"></a><a href="#_bind">`-bind`</a> - The address that should be bound to
  for internal cluster communications.
  This is an IP address that should be reachable by all other nodes in the cluster.
  By default, this is "0.0.0.0", meaning Consul will bind to all addresses on
the local machine and will [advertise](/docs/agent/options.html#_advertise)
the first available private IPv4 address to the rest of the cluster. If there
are multiple private IPv4 addresses available, Consul will exit with an error
at startup. If you specify "[::]", Consul will
[advertise](/docs/agent/options.html#_advertise) the first available public
IPv6 address. If there are multiple public IPv6 addresses available, Consul
will exit with an error at startup.
  Consul uses both TCP and UDP and the same port for both. If you
  have any firewalls, be sure to allow both protocols.

* <a name="_serf_wan_bind"></a><a href="#_serf_wan_bind">`-serf-wan-bind`</a> - The address that should be bound to for Serf WAN gossip communications.
  By default, the value follows the same rules as [`-bind` command-line flag](#_bind), and if this is not specified, the `-bind` option is used. This
  is available in Consul 0.7.1 and later.

* <a name="_serf_lan_bind"></a><a href="#_serf_lan_bind">`-serf-lan-bind`</a> - The address that should be bound to for Serf LAN gossip communications.
  This is an IP address that should be reachable by all other LAN nodes in the cluster. By default, the value follows the same rules as
  [`-bind` command-line flag](#_bind), and if this is not specified, the `-bind` option is used. This is available in Consul 0.7.1 and later.

* <a name="_client"></a><a href="#_client">`-client`</a> - The address to which
  Consul will bind client interfaces, including the HTTP and DNS servers. By default,
  this is "127.0.0.1", allowing only loopback connections.

* <a name="_config_file"></a><a href="#_config_file">`-config-file`</a> - A configuration file
  to load. For more information on
  the format of this file, read the [Configuration Files](#configuration_files) section.
  This option can be specified multiple times to load multiple configuration
  files. If it is specified multiple times, configuration files loaded later
  will merge with configuration files loaded earlier. During a config merge,
  single-value keys (string, int, bool) will simply have their values replaced
  while list types will be appended together.

* <a name="_config_dir"></a><a href="#_config_dir">`-config-dir`</a> - A directory of
  configuration files to load. Consul will
  load all files in this directory with the suffix ".json". The load order
  is alphabetical, and the the same merge routine is used as with the
  [`config-file`](#_config_file) option above. This option can be specified multiple times
  to load multiple directories. Sub-directories of the config directory are not loaded.
  For more information on the format of the configuration files, see the
  [Configuration Files](#configuration_files) section.

* <a name="_data_dir"></a><a href="#_data_dir">`-data-dir`</a> - This flag provides
  a data directory for the agent to store state.
  This is required for all agents. The directory should be durable across reboots.
  This is especially critical for agents that are running in server mode as they
  must be able to persist cluster state. Additionally, the directory must support
  the use of filesystem locking, meaning some types of mounted folders (e.g. VirtualBox
  shared folders) may not be suitable.

* <a name="_dev"></a><a href="#_dev">`-dev`</a> - Enable development server
  mode. This is useful for quickly starting a Consul agent with all persistence
  options turned off, enabling an in-memory server which can be used for rapid
  prototyping or developing against the API. This mode is **not** intended for
  production use as it does not write any data to disk.

* <a name="_datacenter"></a><a href="#_datacenter">`-datacenter`</a> - This flag controls the datacenter in
  which the agent is running. If not provided,
  it defaults to "dc1". Consul has first-class support for multiple datacenters, but
  it relies on proper configuration. Nodes in the same datacenter should be on a single
  LAN.

* <a name="_dns_port"></a><a href="#_dns_port">`-dns-port`</a> - the DNS port to listen on.
  This overrides the default port 8600. This is available in Consul 0.7 and later.

* <a name="_domain"></a><a href="#_domain">`-domain`</a> - By default, Consul responds to DNS queries
  in the "consul." domain. This flag can be used to change that domain. All queries in this domain
  are assumed to be handled by Consul and will not be recursively resolved.

* <a name="_encrypt"></a><a href="#_encrypt">`-encrypt`</a> - Specifies the secret key to
  use for encryption of Consul
  network traffic. This key must be 16-bytes that are Base64-encoded. The
  easiest way to create an encryption key is to use
  [`consul keygen`](/docs/commands/keygen.html). All
  nodes within a cluster must share the same encryption key to communicate.
  The provided key is automatically persisted to the data directory and loaded
  automatically whenever the agent is restarted. This means that to encrypt
  Consul's gossip protocol, this option only needs to be provided once on each
  agent's initial startup sequence. If it is provided after Consul has been
  initialized with an encryption key, then the provided key is ignored and
  a warning will be displayed.

* <a name="_http_port"></a><a href="#_http_port">`-http-port`</a> - the HTTP API port to listen on.
  This overrides the default port 8500. This option is very useful when deploying Consul
  to an environment which communicates the HTTP port through the environment e.g. PaaS like CloudFoundry, allowing
  you to set the port directly via a Procfile.

* <a name="_join"></a><a href="#_join">`-join`</a> - Address of another agent
  to join upon starting up. This can be
  specified multiple times to specify multiple agents to join. If Consul is
  unable to join with any of the specified addresses, agent startup will
  fail. By default, the agent won't join any nodes when it starts up.

* <a name="_retry_join"></a><a href="#_retry_join">`-retry-join`</a> - Similar
  to [`-join`](#_join) but allows retrying a join if the first
  attempt fails. The list should contain IPv4 addresses with optional Serf
  LAN port number also specified or bracketed IPv6 addresses with optional
  port number — for example: `[::1]:8301`. This is useful for cases where we
  know the address will become available eventually.

* <a name="_retry_join_ec2_tag_key"></a><a href="#_retry_join_ec2_tag_key">`-retry-join-ec2-tag-key`
  </a> - The Amazon EC2 instance tag key to filter on. When used with
  [`-retry-join-ec2-tag-value`](#_retry_join_ec2_tag_value), Consul will attempt to join EC2
  instances with the given tag key and value on startup.
  </br></br>For AWS authentication the following methods are supported, in order:
  - Static credentials (from the config file)
  - Environment variables (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`)
  - Shared credentials file (`~/.aws/credentials` or the path specified by `AWS_SHARED_CREDENTIALS_FILE`)
  - ECS task role metadata (container-specific).
  - EC2 instance role metadata.
  
  The only required IAM permission is `ec2:DescribeInstances`, and it is recommended you make a dedicated
  key used only for auto-joining.

* <a name="_retry_join_ec2_tag_value"></a><a href="#_retry_join_ec2_tag_value">`-retry-join-ec2-tag-value`
  </a> - The Amazon EC2 instance tag value to filter on.

* <a name="_retry_join_ec2_region"></a><a href="#_retry_join_ec2_region">`-retry-join-ec2-region`
  </a> - (Optional) The Amazon EC2 region to use. If not specified, Consul
   will use the local instance's [EC2 metadata endpoint](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html)
   to discover the region.

* <a name="_retry_join_gce_tag_value"></a><a href="#_retry_join_gce_tag_value">`-retry-join-gce-tag-value`
  </a> - A Google Compute Engine instance tag to filter on. Much like the
  `-retry-join-ec2-*` options, this gives Consul the option of doing server
  discovery on [Google Compute Engine](https://cloud.google.com/compute/) by
  searching the tags assigned to any particular instance.

* <a name="_retry_join_gce_project_name"></a><a href="#_retry_join_gce_project_name">`-retry-join-gce-project-name`
  </a> - The project to search in for the tag supplied by
  [`-retry-join-gce-tag-value`](#_retry_join_gce_tag_value). If this is run
  from within a GCE instance, the default is the project the instance is
  located in.

* <a name="_retry_join_gce_zone_pattern"></a><a href="#_retry_join_gce_zone_pattern">`-retry-join-gce-zone-pattern`
  </a> - A regular expression that indicates the zones the tag should be
  searched in. For example, while `us-west1-a` would only search in
  `us-west1-a`, `us-west1-.*` would search in `us-west1-a` and `us-west1-b`.
  The default is to search globally.

* <a name="_retry_join_gce_credentials_file"></a><a href="#_retry_join_gce_credentials_file">`-retry-join-gce-credentials-file`
  </a> - The path to the JSON credentials file of the [GCE Service
  Account](https://cloud.google.com/compute/docs/access/service-accounts) that
  will be used to search for instances. Note that this can also reside in the
  following locations:
   - A path supplied by the `GOOGLE_APPLICATION_CREDENTIALS` environment
     variable
   - The `%APPDATA%/gcloud/application_default_credentials.json` file (Windows)
     or `$HOME/.config/gcloud/application_default_credentials.json` (Linux and
     other systems)
   - If none of these exist and discovery is being run from a GCE instance, the
     instance's configured service account will be used.

* <a name="_retry_interval"></a><a href="#_retry_interval">`-retry-interval`</a> - Time
  to wait between join attempts. Defaults to 30s.

* <a name="_retry_max"></a><a href="#_retry_max">`-retry-max`</a> - The maximum number
  of [`-join`](#_join) attempts to be made before exiting
  with return code 1. By default, this is set to 0 which is interpreted as infinite
  retries.

* <a name="_join_wan"></a><a href="#_join_wan">`-join-wan`</a> - Address of another
  wan agent to join upon starting up. This can be
  specified multiple times to specify multiple WAN agents to join. If Consul is
  unable to join with any of the specified addresses, agent startup will
  fail. By default, the agent won't [`-join-wan`](#_join_wan) any nodes when it starts up.

* <a name="_retry_join_wan"></a><a href="#_retry_join_wan">`-retry-join-wan`</a> - Similar
  to [`retry-join`](#_retry_join) but allows retrying a wan join if the first attempt fails.
  This is useful for cases where we know the address will become
  available eventually.

* <a name="_retry_interval_wan"></a><a href="#_retry_interval_wan">`-retry-interval-wan`</a> - Time
  to wait between [`-join-wan`](#_join_wan) attempts.
  Defaults to 30s.

* <a name="_retry_max_wan"></a><a href="#_retry_max_wan">`-retry-max-wan`</a> - The maximum
  number of [`-join-wan`](#_join_wan) attempts to be made before exiting with return code 1.
  By default, this is set to 0 which is interpreted as infinite retries.

* <a name="_log_level"></a><a href="#_log_level">`-log-level`</a> - The level of logging to
  show after the Consul agent has started. This defaults to "info". The available log levels are
  "trace", "debug", "info", "warn", and "err". You can always connect to an
  agent via [`consul monitor`](/docs/commands/monitor.html) and use any log level. Also, the
  log level can be changed during a config reload.

* <a name="_node"></a><a href="#_node">`-node`</a> - The name of this node in the cluster.
  This must be unique within the cluster. By default this is the hostname of the machine.

* <a name="_node_id"></a><a href="#_node_id">`-node-id`</a> - Available in Consul 0.7.3 and later, this
  is a unique identifier for this node across all time, even if the name of the node or address
  changes. This must be in the form of a hex string, 36 characters long, such as
  `adf4238a-882b-9ddc-4a9d-5b6758e4159e`. If this isn't supplied, which is the most common case, then
  the agent will generate an identifier at startup and persist it in the <a href="#_data_dir">data directory</a>
  so that it will remain the same across agent restarts. This is currently only exposed via
  <a href="/api/agent.html#agent_self">/v1/agent/self</a>,
  <a href="/api/catalog.html">/v1/catalog</a>, and
  <a href="/api/health.html">/v1/health</a> endpoints, but future versions of
  Consul will use this to better manage cluster changes, especially for Consul servers.

* <a name="_node_meta"></a><a href="#_node_meta">`-node-meta`</a> - Available in Consul 0.7.3 and later,
  this specifies an arbitrary metadata key/value pair to associate with the node, of the form `key:value`.
  This can be specified multiple times. Node metadata pairs have the following restrictions:
  - A maximum of 64 key/value pairs can be registered per node.
  - Metadata keys must be between 1 and 128 characters (inclusive) in length
  - Metadata keys must contain only alphanumeric, `-`, and `_` characters.
  - Metadata keys must not begin with the `consul-` prefix; that is reserved for internal use by Consul.
  - Metadata values must be between 0 and 512 (inclusive) characters in length.

* <a name="_pid_file"></a><a href="#_pid_file">`-pid-file`</a> - This flag provides the file
  path for the agent to store its PID. This is useful for sending signals (for example, `SIGINT`
  to close the agent or `SIGHUP` to update check definite

* <a name="_protocol"></a><a href="#_protocol">`-protocol`</a> - The Consul protocol version to
  use. This defaults to the latest version. This should be set only when [upgrading](/docs/upgrading.html).
  You can view the protocol versions supported by Consul by running `consul -v`.

* <a name="_raft_protocol"></a><a href="#_raft_protocol">`-raft-protocol`</a> - This controls the internal
  version of the Raft consensus protocol used for server communications. This defaults to 2 but must
  be set to 3 in order to gain access to Autopilot features, with the exception of
  [`cleanup_dead_servers`](#cleanup_dead_servers).

* <a name="_recursor"></a><a href="#_recursor">`-recursor`</a> - Specifies the address of an upstream DNS
  server. This option may be provided multiple times, and is functionally
  equivalent to the [`recursors` configuration option](#recursors).

* <a name="_rejoin"></a><a href="#_rejoin">`-rejoin`</a> - When provided, Consul will ignore a
  previous leave and attempt to rejoin the cluster when starting. By default, Consul treats leave
  as a permanent intent and does not attempt to join the cluster again when starting. This flag
  allows the previous state to be used to rejoin the cluster.

* <a name="_server"></a><a href="#_server">`-server`</a> - This flag is used to control if an
  agent is in server or client mode. When provided,
  an agent will act as a Consul server. Each Consul cluster must have at least one server and ideally
  no more than 5 per datacenter. All servers participate in the Raft consensus algorithm to ensure that
  transactions occur in a consistent, linearizable manner. Transactions modify cluster state, which
  is maintained on all server nodes to ensure availability in the case of node failure. Server nodes also
  participate in a WAN gossip pool with server nodes in other datacenters. Servers act as gateways
  to other datacenters and forward traffic as appropriate.

* <a name="_non_voting_server"></a><a href="#_non_voting_server">`-non-voting-server`</a> - (Enterprise-only)
  This flag is used to make the server not participate in the Raft quorum, and have it only receive the data
  replication stream. This can be used to add read scalability to a cluster in cases where a high volume of
  reads to servers are needed.

* <a name="_syslog"></a><a href="#_syslog">`-syslog`</a> - This flag enables logging to syslog. This
  is only supported on Linux and OSX. It will result in an error if provided on Windows.

* <a name="_ui"></a><a href="#_ui">`-ui`</a> - Enables the built-in web UI
  server and the required HTTP routes. This eliminates the need to maintain the
  Consul web UI files separately from the binary.

* <a name="_ui_dir"></a><a href="#_ui_dir">`-ui-dir`</a> - This flag provides the directory containing
  the Web UI resources for Consul. This will automatically enable the Web UI. The directory must be
  readable to the agent. Starting with Consul version 0.7.0 and later, the Web UI assets are included in the binary so this flag is no longer necessary; specifying only the `-ui` flag is enough to enable the Web UI.

## <a name="configuration_files"></a>Configuration Files

In addition to the command-line options, configuration can be put into
files. This may be easier in certain situations, for example when Consul is
being configured using a configuration management system.

The configuration files are JSON formatted, making them easily readable
and editable by both humans and computers. The configuration is formatted
as a single JSON object with configuration within it.

Configuration files are used for more than just setting up the agent,
they are also used to provide check and service definitions. These are used
to announce the availability of system servers to the rest of the cluster.
They are documented separately under [check configuration](/docs/agent/checks.html) and
[service configuration](/docs/agent/services.html) respectively. The service and check
definitions support being updated during a reload.

#### Example Configuration File

```javascript
{
  "datacenter": "east-aws",
  "data_dir": "/opt/consul",
  "log_level": "INFO",
  "node_name": "foobar",
  "server": true,
  "watches": [
    {
        "type": "checks",
        "handler": "/usr/bin/health-check-handler.sh"
    }
  ],
  "telemetry": {
     "statsite_address": "127.0.0.1:2180"
  }
}
```

#### Example Configuration File, with TLS

```javascript
{
  "datacenter": "east-aws",
  "data_dir": "/opt/consul",
  "log_level": "INFO",
  "node_name": "foobar",
  "server": true,
  "addresses": {
    "https": "0.0.0.0"
  },
  "ports": {
    "https": 8080
  },
  "key_file": "/etc/pki/tls/private/my.key",
  "cert_file": "/etc/pki/tls/certs/my.crt",
  "ca_file": "/etc/pki/tls/certs/ca-bundle.crt"
}
```

See, especially, the use of the `ports` setting:

```javascript
"ports": {
  "https": 8080
}
```

Consul will not enable TLS for the HTTP API unless the `https` port has been assigned a port number `> 0`.

#### Configuration Key Reference

* <a name="acl_datacenter"></a><a href="#acl_datacenter">`acl_datacenter`</a> - This designates
  the datacenter which is authoritative for ACL information. It must be provided to enable ACLs.
  All servers and datacenters must agree on the ACL datacenter. Setting it on the servers is all
  you need for cluster-level enforcement, but for the APIs to forward properly from the clients,
  it must be set on them too. In Consul 0.8 and later, this also enables agent-level enforcement
  of ACLs. Please see the [ACL Guide](/docs/guides/acl.html) for more details.

* <a name="acl_default_policy"></a><a href="#acl_default_policy">`acl_default_policy`</a> - Either
  "allow" or "deny"; defaults to "allow". The default policy controls the behavior of a token when
  there is no matching rule. In "allow" mode, ACLs are a blacklist: any operation not specifically
  prohibited is allowed. In "deny" mode, ACLs are a whitelist: any operation not
  specifically allowed is blocked. *Note*: this will not take effect until you've set `acl_datacenter` 
  to enable ACL support.

* <a name="acl_down_policy"></a><a href="#acl_down_policy">`acl_down_policy`</a> - Either
  "allow", "deny" or "extend-cache"; "extend-cache" is the default. In the case that the
  policy for a token cannot be read from the [`acl_datacenter`](#acl_datacenter) or leader
  node, the down policy is applied. In "allow" mode, all actions are permitted, "deny" restricts
  all operations, and "extend-cache" allows any cached ACLs to be used, ignoring their TTL
  values. If a non-cached ACL is used, "extend-cache" acts like "deny".

* <a name="acl_agent_master_token"></a><a href="#acl_agent_master_token">`acl_agent_master_token`</a> -
  Used to access <a href="/api/agent.html">agent endpoints</a> that require agent read
  or write privileges even if Consul servers aren't present to validate any tokens. This should only
  be used by operators during outages, regular ACL tokens should normally be used by applications.
  This was added in Consul 0.7.2 and is only used when <a href="#acl_enforce_version_8">`acl_enforce_version_8`</a>
  is set to true.

* <a name="acl_agent_token"></a><a href="#acl_agent_token">`acl_agent_token`</a> - Used for clients
  and servers to perform internal operations to the service catalog. If this isn't specified, then
  the <a href="#acl_token">`acl_token`</a> will be used. This was added in Consul 0.7.2.
  <br><br>
  This token must at least have write access to the node name it will register as in order to set any
  of the node-level information in the catalog such as metadata, or the node's tagged addresses.

* <a name="acl_enforce_version_8"></a><a href="#acl_enforce_version_8">`acl_enforce_version_8`</a> -
  Used for clients and servers to determine if enforcement should occur for new ACL policies being
  previewed before Consul 0.8. Added in Consul 0.7.2, this defaults to false in versions of
  Consul prior to 0.8, and defaults to true in Consul 0.8 and later. This helps ease the
  transition to the new ACL features by allowing policies to be in place before enforcement begins.
  Please see the [ACL Guide](/docs/guides/acl.html#version_8_acls) for more details.

* <a name="acl_master_token"></a><a href="#acl_master_token">`acl_master_token`</a> - Only used
  for servers in the [`acl_datacenter`](#acl_datacenter). This token will be created with management-level
  permissions if it does not exist. It allows operators to bootstrap the ACL system
  with a token ID that is well-known.
  <br><br>
  The `acl_master_token` is only installed when a server acquires cluster leadership. If
  you would like to install or change the `acl_master_token`, set the new value for `acl_master_token`
  in the configuration for all servers. Once this is done, restart the current leader to force a
  leader election. If the `acl_master_token` is not supplied, then the servers do not create a master
  token. When you provide a value, it can be any string value. Using a UUID would ensure that it looks
  the same as the other tokens, but isn't strictly necessary.

* <a name="acl_replication_token"></a><a href="#acl_replication_token">`acl_replication_token`</a> -
  Only used for servers outside the [`acl_datacenter`](#acl_datacenter) running Consul 0.7 or later.
  When provided, this will enable [ACL replication](/docs/guides/acl.html#replication) using this
  token to retrieve and replicate the ACLs to the non-authoritative local datacenter.
  <br><br>
  If there's a partition or other outage affecting the authoritative datacenter, and the
  [`acl_down_policy`](/docs/agent/options.html#acl_down_policy) is set to "extend-cache", tokens not
  in the cache can be resolved during the outage using the replicated set of ACLs. Please see the
  [ACL Guide](/docs/guides/acl.html#replication) replication section for more details.

* <a name="acl_token"></a><a href="#acl_token">`acl_token`</a> - When provided, the agent will use this
  token when making requests to the Consul servers. Clients can override this token on a per-request
  basis by providing the "?token" query parameter. When not provided, the empty token, which maps to
  the 'anonymous' ACL policy, is used.

* <a name="acl_ttl"></a><a href="#acl_ttl">`acl_ttl`</a> - Used to control Time-To-Live caching of ACLs.
  By default, this is 30 seconds. This setting has a major performance impact: reducing it will cause
  more frequent refreshes while increasing it reduces the number of caches. However, because the caches
  are not actively invalidated, ACL policy may be stale up to the TTL value.

* <a name="addresses"></a><a href="#addresses">`addresses`</a> - This is a nested object that allows
  setting bind addresses.
  <br><br>
  `http` supports binding to a Unix domain socket. A socket can be
  specified in the form `unix:///path/to/socket`. A new domain socket will be
  created at the given path. If the specified file path already exists, Consul
  will attempt to clear the file and create the domain socket in its place. The
  permissions of the socket file are tunable via the [`unix_sockets` config construct](#unix_sockets).
  <br><br>
  When running Consul agent commands against Unix socket interfaces, use the
  `-http-addr` argument to specify the path to the socket. You can also place
  the desired values in the `CONSUL_HTTP_ADDR` environment variable.
  <br><br>
  For TCP addresses, the variable values should be an IP address with the port. For
  example: `10.0.0.1:8500` and not `10.0.0.1`. However, ports are set separately in the
  <a href="#ports">`ports`</a> structure when defining them in a configuration file.
  <br><br>
  The following keys are valid:
  * `dns` - The DNS server. Defaults to `client_addr`
  * `http` - The HTTP API. Defaults to `client_addr`
  * `https` - The HTTPS API. Defaults to `client_addr`
* <a name="advertise_addr"></a><a href="#advertise_addr">`advertise_addr`</a> Equivalent to
  the [`-advertise` command-line flag](#_advertise).

* <a name="serf_wan_bind"></a><a href="#serf_wan_bind">`serf_wan_bind`</a> Equivalent to
  the [`-serf-wan-bind` command-line flag](#_serf_wan_bind).

* <a name="serf_lan_bind"></a><a href="#serf_lan_bind">`serf_lan_bind`</a> Equivalent to
  the [`-serf-lan-bind` command-line flag](#_serf_lan_bind).

* <a name="advertise_addrs"></a><a href="#advertise_addrs">`advertise_addrs`</a> Allows to set
  the advertised addresses for SerfLan, SerfWan and RPC together with the port. This gives
  you more control than <a href="#_advertise">`-advertise`</a> or <a href="#_advertise-wan">`-advertise-wan`</a>
  while it serves the same purpose. These settings might override <a href="#_advertise">`-advertise`</a> or
  <a href="#_advertise-wan">`-advertise-wan`</a>
  <br><br>
  This is a nested setting that allows the following keys:
  * `serf_lan` - The SerfLan address. Accepts values in the form of "host:port" like "10.23.31.101:8301".
  * `serf_wan` - The SerfWan address. Accepts values in the form of "host:port" like "10.23.31.101:8302".
  * `rpc` - The server RPC address. Accepts values in the form of "host:port" like "10.23.31.101:8300".

* <a name="advertise_addr_wan"></a><a href="#advertise_addr_wan">`advertise_addr_wan`</a> Equivalent to
  the [`-advertise-wan` command-line flag](#_advertise-wan).

* <a name="atlas_acl_token"></a><a href="#atlas_acl_token">`atlas_acl_token`</a> When provided,
  any requests made by Atlas will use this ACL token unless explicitly overridden. When not provided
  the [`acl_token`](#acl_token) is used. This can be set to 'anonymous' to reduce permission below
  that of [`acl_token`](#acl_token).

* <a name="atlas_infrastructure"></a><a href="#atlas_infrastructure">`atlas_infrastructure`</a>
  Equivalent to the [`-atlas` command-line flag](#_atlas).

* <a name="atlas_join"></a><a href="#atlas_join">`atlas_join`</a> Equivalent to the
  [`-atlas-join` command-line flag](#_atlas_join).

* <a name="atlas_token"></a><a href="#atlas_token">`atlas_token`</a> Equivalent to the
  [`-atlas-token` command-line flag](#_atlas_token).

* <a name="atlas_endpoint"></a><a href="#atlas_endpoint">`atlas_endpoint`</a> Equivalent to the
  [`-atlas-endpoint` command-line flag](#_atlas_endpoint).

* <a name="autopilot"></a><a href="#autopilot">`autopilot`</a> Added in Consul 0.8, this object
  allows a number of sub-keys to be set which can configure operator-friendly settings for Consul servers.
  For more information about Autopilot, see the [Autopilot Guide](/docs/guides/autopilot.html).
  <br><br>
  The following sub-keys are available:

  * <a name="cleanup_dead_servers"></a><a href="#cleanup_dead_servers">`cleanup_dead_servers`</a> - This controls
  the automatic removal of dead server nodes periodically and whenever a new server is added to the cluster.
  Defaults to `true`.

  * <a name="last_contact_threshold"></a><a href="#last_contact_threshold">`last_contact_threshold`</a> - Controls
  the maximum amount of time a server can go without contact from the leader before being considered unhealthy.
  Must be a duration value such as `10s`. Defaults to `200ms`.

  * <a name="max_trailing_threshold"></a><a href="#max_trailing_threshold">`max_trailing_threshold`</a> - Controls
  the maximum number of log entries that a server can trail the leader by before being considered unhealthy. Defaults
  to 250.

  * <a name="server_stabilization_time"></a><a href="#server_stabilization_time">`server_stabilization_time`</a> -
  Controls the minimum amount of time a server must be stable in the 'healthy' state before being added to the
  cluster. Only takes effect if all servers are running Raft protocol version 3 or higher. Must be a duration value
  such as `30s`. Defaults to `10s`.

  * <a name="redundancy_zone_tag"></a><a href="#redundancy_zone_tag">`redundancy_zone_tag`</a> - (Enterprise-only)
  This controls the [`-node-meta`](#_node_meta) key to use when Autopilot is separating servers into zones for
  redundancy. Only one server in each zone can be a voting member at one time. If left blank (the default), this
  feature will be disabled.

  * <a name="disable_upgrade_migration"></a><a href="#disable_upgrade_migration">`disable_upgrade_migration`</a> - (Enterprise-only)
  If set to `true`, this setting will disable Autopilot's upgrade migration strategy in Consul Enterprise of waiting
  until enough newer-versioned servers have been added to the cluster before promoting any of them to voters. Defaults
  to `false`.

* <a name="bootstrap"></a><a href="#bootstrap">`bootstrap`</a> Equivalent to the
  [`-bootstrap` command-line flag](#_bootstrap).

* <a name="bootstrap_expect"></a><a href="#bootstrap_expect">`bootstrap_expect`</a> Equivalent
  to the [`-bootstrap-expect` command-line flag](#_bootstrap_expect).

* <a name="bind_addr"></a><a href="#bind_addr">`bind_addr`</a> Equivalent to the
  [`-bind` command-line flag](#_bind).

* <a name="ca_file"></a><a href="#ca_file">`ca_file`</a> This provides a file path to a PEM-encoded
  certificate authority. The certificate authority is used to check the authenticity of client and
  server connections with the appropriate [`verify_incoming`](#verify_incoming) or
  [`verify_outgoing`](#verify_outgoing) flags.

* <a name="cert_file"></a><a href="#cert_file">`cert_file`</a> This provides a file path to a
  PEM-encoded certificate. The certificate is provided to clients or servers to verify the agent's
  authenticity. It must be provided along with [`key_file`](#key_file).

* <a name="check_update_interval"></a><a href="#check_update_interval">`check_update_interval`</a>
  This interval controls how often check output from
  checks in a steady state is synchronized with the server. By default, this is
  set to 5 minutes ("5m"). Many checks which are in a steady state produce
  slightly different output per run (timestamps, etc) which cause constant writes.
  This configuration allows deferring the sync of check output for a given interval to
  reduce write pressure. If a check ever changes state, the new state and associated
  output is synchronized immediately. To disable this behavior, set the value to "0s".

* <a name="client_addr"></a><a href="#client_addr">`client_addr`</a> Equivalent to the
  [`-client` command-line flag](#_client).

* <a name="datacenter"></a><a href="#datacenter">`datacenter`</a> Equivalent to the
  [`-datacenter` command-line flag](#_datacenter).

* <a name="data_dir"></a><a href="#data_dir">`data_dir`</a> Equivalent to the
  [`-data-dir` command-line flag](#_data_dir).

* <a name="disable_anonymous_signature"></a><a href="#disable_anonymous_signature">
  `disable_anonymous_signature`</a> Disables providing an anonymous signature for de-duplication
  with the update check. See [`disable_update_check`](#disable_update_check).

* <a name="disable_remote_exec"></a><a href="#disable_remote_exec">`disable_remote_exec`</a>
  Disables support for remote execution. When set to true, the agent will ignore any incoming
  remote exec requests. In versions of Consul prior to 0.8, this defaulted to false. In Consul
  0.8 the default was changed to true, to make remote exec opt-in instead of opt-out.

* <a name="disable_update_check"></a><a href="#disable_update_check">`disable_update_check`</a>
  Disables automatic checking for security bulletins and new version releases.

* <a name="dns_config"></a><a href="#dns_config">`dns_config`</a> This object allows a number
  of sub-keys to be set which can tune how DNS queries are serviced. See this guide on
  [DNS caching](/docs/guides/dns-cache.html) for more detail.
  <br><br>
  The following sub-keys are available:

  * <a name="allow_stale"></a><a href="#allow_stale">`allow_stale`</a> - Enables a stale query
  for DNS information. This allows any Consul server, rather than only the leader, to service
  the request. The advantage of this is you get linear read scalability with Consul servers.
  In versions of Consul prior to 0.7, this defaulted to false, meaning all requests are serviced
  by the leader, providing stronger consistency but less throughput and higher latency. In Consul
  0.7 and later, this defaults to true for better utilization of available servers.

  * <a name="max_stale"></a><a href="#max_stale">`max_stale`</a> - When [`allow_stale`](#allow_stale)
  is specified, this is used to limit how stale results are allowed to be. If a Consul server is
  behind the leader by more than `max_stale`, the query will be re-evaluated on the leader to get
  more up-to-date results. Prior to Consul 0.7.1 this defaulted to 5 seconds; in Consul 0.7.1
  and later this defaults to 10 years ("87600h") which effectively allows DNS queries to be answered
  by any server, no matter how stale. In practice, servers are usually only milliseconds behind the
  leader, so this lets Consul continue serving requests in long outage scenarios where no leader can
  be elected.

  * <a name="node_ttl"></a><a href="#node_ttl">`node_ttl`</a> - By default, this is "0s", so all
  node lookups are served with a 0 TTL value. DNS caching for node lookups can be enabled by
  setting this value. This should be specified with the "s" suffix for second or "m" for minute.

  * <a name="service_ttl"></a><a href="#service_ttl">`service_ttl`</a> - This is a sub-object
  which allows for setting a TTL on service lookups with a per-service policy. The "*" wildcard
  service can be used when there is no specific policy available for a service. By default, all
  services are served with a 0 TTL value. DNS caching for service lookups can be enabled by
  setting this value.

  * <a name="enable_truncate"></a><a href="#enable_truncate">`enable_truncate`</a> - If set to
  true, a UDP DNS query that would return more than 3 records, or more than would fit into a valid
  UDP response, will set the truncated flag, indicating to clients that they should re-query
  using TCP to get the full set of records.

  * <a name="only_passing"></a><a href="#only_passing">`only_passing`</a> - If set to true, any
  nodes whose health checks are warning or critical will be excluded from DNS results. If false,
  the default, only nodes whose healthchecks are failing as critical will be excluded. For
  service lookups, the health checks of the node itself, as well as the service-specific checks
  are considered. For example, if a node has a health check that is critical then all services on
  that node will be excluded because they are also considered critical.

  * <a name="recursor_timeout"></a><a href="#recursor_timeout">`recursor_timeout`</a> - Timeout used
  by Consul when recursively querying an upstream DNS server. See <a href="#recursors">`recursors`</a>
  for more details. Default is 2s. This is available in Consul 0.7 and later.

  * <a name="disable_compression"></a><a href="#disable_compression">`disable_compression`</a> - If
  set to true, DNS responses will not be compressed. Compression was added and enabled by default
  in Consul 0.7.

  * <a name="udp_answer_limit"></a><a
  href="#udp_answer_limit">`udp_answer_limit`</a> - Limit the number of
  resource records contained in the answer section of a UDP-based DNS
  response. When answering a question, Consul will use the complete list of
  matching hosts, shuffle the list randomly, and then limit the number of
  answers to `udp_answer_limit` (default `3`). In environments where
  [RFC 3484 Section 6](https://tools.ietf.org/html/rfc3484#section-6) Rule 9
  is implemented and enforced (i.e. DNS answers are always sorted and
  therefore never random), clients may need to set this value to `1` to
  preserve the expected randomized distribution behavior (note:
  [RFC 3484](https://tools.ietf.org/html/rfc3484) has been obsoleted by
  [RFC 6724](https://tools.ietf.org/html/rfc6724) and as a result it should
  be increasingly uncommon to need to change this value with modern
  resolvers).

* <a name="domain"></a><a href="#domain">`domain`</a> Equivalent to the
  [`-domain` command-line flag](#_domain).

* <a name="enable_debug"></a><a href="#enable_debug">`enable_debug`</a> When set, enables some
  additional debugging features. Currently, this is only used to set the runtime profiling HTTP endpoints.

* <a name="enable_syslog"></a><a href="#enable_syslog">`enable_syslog`</a> Equivalent to
  the [`-syslog` command-line flag](#_syslog).

* <a name="encrypt"></a><a href="#encrypt">`encrypt`</a> Equivalent to the
  [`-encrypt` command-line flag](#_encrypt).

* <a name="key_file"></a><a href="#key_file">`key_file`</a> This provides a the file path to a
  PEM-encoded private key. The key is used with the certificate to verify the agent's authenticity.
  This must be provided along with [`cert_file`](#cert_file).

* <a name="http_api_response_headers"></a><a href="#http_api_response_headers">`http_api_response_headers`</a>
  This object allows adding headers to the HTTP API
  responses. For example, the following config can be used to enable
  [CORS](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) on
  the HTTP API endpoints:

    ```javascript
      {
        "http_api_response_headers": {
            "Access-Control-Allow-Origin": "*"
        }
      }
    ```

* <a name="leave_on_terminate"></a><a href="#leave_on_terminate">`leave_on_terminate`</a> If
  enabled, when the agent receives a TERM signal, it will send a `Leave` message to the rest
  of the cluster and gracefully leave. The default behavior for this feature varies based on
  whether or not the agent is running as a client or a server (prior to Consul 0.7 the default
  value was unconditionally set to `false`). On agents in client-mode, this defaults to `true`
  and for agents in server-mode, this defaults to `false`.

* <a name="log_level"></a><a href="#log_level">`log_level`</a> Equivalent to the
  [`-log-level` command-line flag](#_log_level).

* <a name="node_id"></a><a href="#node_id">`node_id`</a> Equivalent to the
  [`-node-id` command-line flag](#_node_id).

* <a name="node_name"></a><a href="#node_name">`node_name`</a> Equivalent to the
  [`-node` command-line flag](#_node).

* <a name="node_meta"></a><a href="#node_meta">`node_meta`</a> Available in Consul 0.7.3 and later,
  This object allows associating arbitrary metadata key/value pairs with the local node, which can
  then be used for filtering results from certain catalog endpoints. See the
  [`-node-meta` command-line flag](#_node_meta) for more information.

    ```javascript
      {
        "node_meta": {
            "instance_type": "t2.medium"
        }
      }
    ```

* <a name="performance"></a><a href="#performance">`performance`</a> Available in Consul 0.7 and
  later, this is a nested object that allows tuning the performance of different subsystems in
  Consul. See the [Server Performance](/docs/guides/performance.html) guide for more details. The
  following parameters are available:

  * <a name="raft_multiplier"></a><a href="#raft_multiplier">`raft_multiplier`</a> - An integer
    multiplier used by Consul servers to scale key Raft timing parameters. Omitting this value
    or setting it to 0 uses default timing described below. Lower values are used to tighten
    timing and increase sensitivity while higher values relax timings and reduce sensitivity.
    Tuning this affects the time it takes Consul to detect leader failures and to perform
    leader elections, at the expense of requiring more network and CPU resources for better
    performance.<br><br>By default, Consul will use a lower-performance timing that's suitable
    for [minimal Consul servers](/docs/guides/performance.html#minumum), currently equivalent
    to setting this to a value of 5 (this default may be changed in future versions of Consul,
    depending if the target minimum server profile changes). Setting this to a value of 1 will
    configure Raft to its highest-performance mode, equivalent to the default timing of Consul
    prior to 0.7, and is recommended for [production Consul servers](/docs/guides/performance.html#production).
    See the note on [last contact](/docs/guides/performance.html#last-contact) timing for more
    details on tuning this parameter. The maximum allowed value is 10.

* <a name="ports"></a><a href="#ports">`ports`</a> This is a nested object that allows setting
  the bind ports for the following keys:
    * <a name="dns_port"></a><a href="#dns_port">`dns`</a> - The DNS server, -1 to disable. Default 8600.
    * <a name="http_port"></a><a href="#http_port">`http`</a> - The HTTP API, -1 to disable. Default 8500.
    * <a name="https_port"></a><a href="#https_port">`https`</a> - The HTTPS API, -1 to disable. Default -1 (disabled).
    * <a name="rpc_port"></a><a href="#rpc_port">`rpc`</a> - The CLI RPC endpoint. Default 8400. This is deprecated
      in Consul 0.8 and later.
    * <a name="serf_lan_port"></a><a href="#serf_lan_port">`serf_lan`</a> - The Serf LAN port. Default 8301.
    * <a name="serf_wan_port"></a><a href="#serf_wan_port">`serf_wan`</a> - The Serf WAN port. Default 8302.
    * <a name="server_rpc_port"></a><a href="#server_rpc_port">`server`</a> - Server RPC address. Default 8300.

* <a name="protocol"></a><a href="#protocol">`protocol`</a> Equivalent to the
  [`-protocol` command-line flag](#_protocol).

* <a name="raft_protocol"></a><a href="#raft_protocol">`raft_protocol`</a> Equivalent to the
  [`-raft-protocol` command-line flag](#_raft_protocol).

* <a name="reap"></a><a href="#reap">`reap`</a> This controls Consul's automatic reaping of child processes,
  which is useful if Consul is running as PID 1 in a Docker container. If this isn't specified, then Consul will
  automatically reap child processes if it detects it is running as PID 1. If this is set to true or false, then
  it controls reaping regardless of Consul's PID (forces reaping on or off, respectively). This option was removed
  in Consul 0.7.1. For later versions of Consul, you will need to reap processes using a wrapper, please see the
  [Consul Docker image entry point script](https://github.com/hashicorp/docker-consul/blob/master/0.X/docker-entrypoint.sh)
  for an example.

* <a name="reconnect_timeout"></a><a href="#reconnect_timeout">`reconnect_timeout`</a> This controls
  how long it takes for a failed node to be completely removed from the cluster. This defaults to
  72 hours and it is recommended that this is set to at least double the maximum expected recoverable
  outage time for a node or network partition. WARNING: Setting this time too low could cause Consul
  servers to be removed from quorum during an extended node failure or partition, which could complicate
  recovery of the cluster. The value is a time with a unit suffix, which can be "s", "m", "h" for seconds,
  minutes, or hours. The value must be >= 8 hours.

* <a name="reconnect_timeout_wan"></a><a href="#reconnect_timeout_wan">`reconnect_timeout_wan`</a> This
  is the WAN equivalent of the <a href="#reconnect_timeout">`reconnect_timeout`</a> parameter, which
  controls how long it takes for a failed server to be completely removed from the WAN pool. This also
  defaults to 72 hours, and must be >= 8 hours.

* <a name="recursor"></a><a href="#recursor">`recursor`</a> Provides a single recursor address.
  This has been deprecated, and the value is appended to the [`recursors`](#recursors) list for
  backwards compatibility.

* <a name="recursors"></a><a href="#recursors">`recursors`</a> This flag provides addresses of
  upstream DNS servers that are used to recursively resolve queries if they are not inside the service
  domain for Consul. For example, a node can use Consul directly as a DNS server, and if the record is
  outside of the "consul." domain, the query will be resolved upstream.

* <a name="rejoin_after_leave"></a><a href="#rejoin_after_leave">`rejoin_after_leave`</a> Equivalent
  to the [`-rejoin` command-line flag](#_rejoin).

* <a name="retry_join"></a><a href="#retry_join">`retry_join`</a> Equivalent to the
  [`-retry-join` command-line flag](#_retry_join). Takes a list
  of addresses to attempt joining every [`retry_interval`](#_retry_interval) until at least one
  join works. The list should contain IPv4 addresses with optional Serf LAN port number also specified or bracketed IPv6 addresses with optional port number — for example: `[::1]:8301`.

* <a name="retry_join_ec2"></a><a href="#retry_join_ec2">`retry_join_ec2`</a> - This is a nested object
  that allows the setting of EC2-related [`-retry-join`](#_retry_join) options.
  <br><br>
  The following keys are valid:
  * `region` - The AWS region. Equivalent to the
    [`-retry-join-ec2-region` command-line flag](#_retry_join_ec2_region).
  * `tag_key` - The EC2 instance tag key to filter on. Equivalent to the</br>
    [`-retry-join-ec2-tag-key` command-line flag](#_retry_join_ec2_tag_key).
  * `tag_value` - The EC2 instance tag value to filter on. Equivalent to the</br>
    [`-retry-join-ec2-tag-value` command-line flag](#_retry_join_ec2_tag_value).
  * `access_key_id` - The AWS access key ID to use for authentication.
  * `secret_access_key` - The AWS secret access key to use for authentication.

* <a name="retry_join_gce"></a><a href="#retry_join_gce">`retry_join_gce`</a> - This is a nested object
  that allows the setting of GCE-related [`-retry-join`](#_retry_join) options.
  <br><br>
  The following keys are valid:
  * `project_name` - The GCE project name. Equivalent to the<br>
    [`-retry-join-gce-project-name` command-line
    flag](#_retry_join_gce_project_name).
  * `zone_pattern` - The regular expression indicating the zones to search in.
    Equivalent to the <br>
    [`-retry-join-gce-zone-pattern` command-line
    flag](#_retry_join_gce_zone_pattern).
  * `tag_value` - The GCE instance tag value to filter on. Equivalent to the <br>
    [`-retry-join-gce-tag-value` command-line
    flag](#_retry_join_gce_tag_value).
  * `credentials_file` - The path to the GCE service account credentials file.
    Equivalent to the <br>
    [`-retry-join-gce-credentials-file` command-line
    flag](#_retry_join_gce_credentials_file).

* <a name="retry_interval"></a><a href="#retry_interval">`retry_interval`</a> Equivalent to the
  [`-retry-interval` command-line flag](#_retry_interval).

* <a name="retry_join_wan"></a><a href="#retry_join_wan">`retry_join_wan`</a> Equivalent to the
  [`-retry-join-wan` command-line flag](#_retry_join_wan). Takes a list
  of addresses to attempt joining to WAN every [`retry_interval_wan`](#_retry_interval_wan) until at least one
  join works.

* <a name="retry_interval_wan"></a><a href="#retry_interval_wan">`retry_interval_wan`</a> Equivalent to the
  [`-retry-interval-wan` command-line flag](#_retry_interval_wan).

* <a name="server"></a><a href="#server">`server`</a> Equivalent to the
  [`-server` command-line flag](#_server).

* <a name="non_voting_server"></a><a href="#non_voting_server">`non_voting_server`</a> - Equivalent to the
  [`-non-voting-server` command-line flag](#_non_voting_server).

* <a name="server_name"></a><a href="#server_name">`server_name`</a> When provided, this overrides
  the [`node_name`](#_node) for the TLS certificate. It can be used to ensure that the certificate
  name matches the hostname we declare.

* <a name="session_ttl_min"></a><a href="#session_ttl_min">`session_ttl_min`</a>
  The minimum allowed session TTL. This ensures sessions are not created with
  TTL's shorter than the specified limit. It is recommended to keep this limit
  at or above the default to encourage clients to send infrequent heartbeats.
  Defaults to 10s.

* <a name="skip_leave_on_interrupt"></a><a
  href="#skip_leave_on_interrupt">`skip_leave_on_interrupt`</a> This is
  similar to [`leave_on_terminate`](#leave_on_terminate) but only affects
  interrupt handling. When Consul receives an interrupt signal (such as
  hitting Control-C in a terminal), Consul will gracefully leave the cluster.
  Setting this to `true` disables that behavior. The default behavior for
  this feature varies based on whether or not the agent is running as a
  client or a server (prior to Consul 0.7 the default value was
  unconditionally set to `false`). On agents in client-mode, this defaults
  to `false` and for agents in server-mode, this defaults to `true`
  (i.e. Ctrl-C on a server will keep the server in the cluster and therefore
  quorum, and Ctrl-C on a client will gracefully leave).

* <a name="start_join"></a><a href="#start_join">`start_join`</a> An array of strings specifying addresses
  of nodes to [`-join`](#_join) upon startup.

* <a name="start_join_wan"></a><a href="#start_join_wan">`start_join_wan`</a> An array of strings specifying
  addresses of WAN nodes to [`-join-wan`](#_join_wan) upon startup.

* <a name="telemetry"></a><a href="#telemetry">`telemetry`</a> This is a nested object that configures where Consul
  sends its runtime telemetry, and contains the following keys:

  * <a name="telemetry-statsd_address"></a><a href="#telemetry-statsd_address">`statsd_address`</a> This provides the
    address of a statsd instance in the format `host:port`. If provided, Consul will send various telemetry information to that instance for
    aggregation. This can be used to capture runtime information. This sends UDP packets only and can be used with
    statsd or statsite.

  * <a name="telemetry-statsite_address"></a><a href="#telemetry-statsite_address">`statsite_address`</a> This provides
    the address of a statsite instance in the format `host:port`. If provided, Consul will stream various telemetry information to that instance
    for aggregation. This can be used to capture runtime information. This streams via TCP and can only be used with
    statsite.

  * <a name="telemetry-statsite_prefix"></a><a href="#telemetry-statsite_prefix">`statsite_prefix`</a>
    The prefix used while writing all telemetry data to statsite. By default, this is set to "consul".

  * <a name="telemetry-dogstatsd_addr"></a><a href="#telemetry-dogstatsd_addr">`dogstatsd_addr`</a> This provides the
    address of a DogStatsD instance in the format `host:port`. DogStatsD is a protocol-compatible flavor of
    statsd, with the added ability to decorate metrics with tags and event information. If provided, Consul will
    send various telemetry information to that instance for aggregation. This can be used to capture runtime
    information.

  * <a name="telemetry-dogstatsd_tags"></a><a href="#telemetry-dogstatsd_tags">`dogstatsd_tags`</a> This provides a list of global tags
    that will be added to all telemetry packets sent to DogStatsD. It is a list of strings, where each string
    looks like "my_tag_name:my_tag_value".

  * <a name="telemetry-disable_hostname"></a><a href="#telemetry-disable_hostname">`disable_hostname`</a>
    This controls whether or not to prepend runtime telemetry with the machine's hostname, defaults to false.

  * <a name="telemetry-circonus_api_token"></a><a href="#telemetry-circonus_api_token">`circonus_api_token`</a>
    A valid API Token used to create/manage check. If provided, metric management is enabled.

  * <a name="telemetry-circonus_api_app"></a><a href="#telemetry-circonus_api_app">`circonus_api_app`</a>
    A valid app name associated with the API token. By default, this is set to "consul".

  * <a name="telemetry-circonus_api_url"></a><a href="#telemetry-circonus_api_url">`circonus_api_url`</a>
    The base URL to use for contacting the Circonus API. By default, this is set to "https://api.circonus.com/v2".

  * <a name="telemetry-circonus_submission_interval"></a><a href="#telemetry-circonus_submission_interval">`circonus_submission_interval`</a>
    The interval at which metrics are submitted to Circonus. By default, this is set to "10s" (ten seconds).

  * <a name="telemetry-circonus_submission_url"></a><a href="#telemetry-circonus_submission_url">`circonus_submission_url`</a>
    The `check.config.submission_url` field, of a Check API object, from a previously created HTTPTRAP check.

  * <a name="telemetry-circonus_check_id"></a><a href="#telemetry-circonus_check_id">`circonus_check_id`</a>
    The Check ID (not **check bundle**) from a previously created HTTPTRAP check. The numeric portion of the `check._cid` field in the Check API object.

  * <a name="telemetry-circonus_check_force_metric_activation"></a><a href="#telemetry-circonus_check_force_metric_activation">`circonus_check_force_metric_activation`</a>
    Force activation of metrics which already exist and are not currently active. If check management is enabled, the default behavior is to add new metrics as they are encoutered. If the metric already exists in the check, it will **not** be activated. This setting overrides that behavior. By default, this is set to false.

  * <a name="telemetry-circonus_check_instance_id"></a><a href="#telemetry-circonus_check_instance_id">`circonus_check_instance_id`</a>
    Uniquely identifies the metrics coming from this *instance*. It can be used to maintain metric continuity with transient or ephemeral instances as they move around within an infrastructure. By default, this is set to hostname:application name (e.g. "host123:consul").

  * <a name="telemetry-circonus_check_search_tag"></a><a href="#telemetry-circonus_check_search_tag">`circonus_check_search_tag`</a>
    A special tag which, when coupled with the instance id, helps to narrow down the search results when neither a Submission URL or Check ID is provided. By default, this is set to service:application name (e.g. "service:consul").

  * <a name="telemetry-circonus_check_display_name"</a><a href="#telemetry-circonus_check_display_name">`circonus_check_display_name`</a>
    Specifies a name to give a check when it is created. This name is displayed in the Circonus UI Checks list. Available in Consul 0.7.2 and later.

  * <a name="telemetry-circonus_check_tags"</a><a href="#telemetry-circonus_check_tags">`circonus_check_tags`</a>
    Comma separated list of additional tags to add to a check when it is created. Available in Consul 0.7.2 and later.

  * <a name="telemetry-circonus_broker_id"></a><a href="#telemetry-circonus_broker_id">`circonus_broker_id`</a>
    The ID of a specific Circonus Broker to use when creating a new check. The numeric portion of `broker._cid` field in a Broker API object. If metric management is enabled and neither a Submission URL nor Check ID is provided, an attempt will be made to search for an existing check using Instance ID and Search Tag. If one is not found, a new HTTPTRAP check will be created. By default, this is not used and a random Enterprise Broker is selected, or the default Circonus Public Broker.

  * <a name="telemetry-circonus_broker_select_tag"></a><a href="#telemetry-circonus_broker_select_tag">`circonus_broker_select_tag`</a>
    A special tag which will be used to select a Circonus Broker when a Broker ID is not provided. The best use of this is to as a hint for which broker should be used based on *where* this particular instance is running (e.g. a specific geo location or datacenter, dc:sfo). By default, this is left blank and not used.

* <a name="statsd_addr"></a><a href="#statsd_addr">`statsd_addr`</a> Deprecated, see
  the <a href="#telemetry">telemetry</a> structure

* <a name="statsite_addr"></a><a href="#statsite_addr">`statsite_addr`</a> Deprecated, see
  the <a href="#telemetry">telemetry</a> structure

* <a name="statsite_prefix"></a><a href="#statsite_prefix">`statsite_prefix`</a> Deprecated, see
  the <a href="#telemetry">telemetry</a> structure

* <a name="dogstatsd_addr"></a><a href="#dogstatsd_addr">`dogstatsd_addr`</a> Deprecated, see
  the <a href="#telemetry">telemetry</a> structure

* <a name="dogstatsd_tags"></a><a href="#dogstatsd_tags">`dogstatsd_tags`</a> Deprecated, see
  the <a href="#telemetry">telemetry</a> structure

* <a name="syslog_facility"></a><a href="#syslog_facility">`syslog_facility`</a> When
  [`enable_syslog`](#enable_syslog) is provided, this controls to which
  facility messages are sent. By default, `LOCAL0` will be used.

* <a name="tls_min_version"></a><a href="#tls_min_version">`tls_min_version`</a> Added in Consul
  0.7.4, this specifies the minimum supported version of TLS. Accepted values are "tls10", "tls11"
  or "tls12". This defaults to "tls10". WARNING: TLS 1.1 and lower are generally considered less
  secure; avoid using these if possible. This will be changed to default to "tls12" in Consul 0.8.0.

* <a name="translate_wan_addrs"</a><a href="#translate_wan_addrs">`translate_wan_addrs`</a> If
  set to true, Consul will prefer a node's configured <a href="#_advertise-wan">WAN address</a>
  when servicing DNS and HTTP requests for a node in a remote datacenter. This allows the node to
  be reached within its own datacenter using its local address, and reached from other datacenters
  using its WAN address, which is useful in hybrid setups with mixed networks. This is disabled by
  default.
  <br>
  <br>
  Starting in Consul 0.7 and later, node addresses in responses to HTTP requests will also prefer a
  node's configured <a href="#_advertise-wan">WAN address</a> when querying for a node in a remote
  datacenter. An [`X-Consul-Translate-Addresses`](/api/index.html#translate_header) header
  will be present on all responses when translation is enabled to help clients know that the addresses
  may be translated. The `TaggedAddresses` field in responses also have a `lan` address for clients that
  need knowledge of that address, regardless of translation.
  <br>
  <br>The following endpoints translate addresses:
  <br>
  * [`/v1/catalog/nodes`](/api/catalog.html#catalog_nodes)
  * [`/v1/catalog/node/<node>`](/api/catalog.html#catalog_node)
  * [`/v1/catalog/service/<service>`](/api/catalog.html#catalog_service)
  * [`/v1/health/service/<service>`](/api/health.html#health_service)
  * [`/v1/query/<query or name>/execute`](/api/query.html#execute)

* <a name="ui"></a><a href="#ui">`ui`</a> - Equivalent to the [`-ui`](#_ui)
  command-line flag.

* <a name="ui_dir"></a><a href="#ui_dir">`ui_dir`</a> - Equivalent to the
  [`-ui-dir`](#_ui_dir) command-line flag. This configuration key is not required as of Consul version 0.7.0 and later.

* <a name="unix_sockets"></a><a href="#unix_sockets">`unix_sockets`</a> - This
  allows tuning the ownership and permissions of the
  Unix domain socket files created by Consul. Domain sockets are only used if
  the HTTP address is configured with the `unix://` prefix.
  <br>
  <br>
  It is important to note that this option may have different effects on
  different operating systems. Linux generally observes socket file permissions
  while many BSD variants ignore permissions on the socket file itself. It is
  important to test this feature on your specific distribution. This feature is
  currently not functional on Windows hosts.
  <br>
  <br>
  The following options are valid within this construct and apply globally to all
  sockets created by Consul:
  <br>
  * `user` - The name or ID of the user who will own the socket file.
  * `group` - The group ID ownership of the socket file. This option
    currently only supports numeric IDs.
  * `mode` - The permission bits to set on the file.

* <a name="verify_incoming"></a><a href="#verify_incoming">`verify_incoming`</a> - If
  set to true, Consul requires that all incoming
  connections make use of TLS and that the client provides a certificate signed
  by the Certificate Authority from the [`ca_file`](#ca_file). By default, this is false, and
  Consul will not enforce the use of TLS or verify a client's authenticity. This
  applies to both server RPC and to the HTTPS API. To enable the HTTPS API, you
  must define an HTTPS port via the [`ports`](#ports) configuration. By default, HTTPS
  is disabled.

* <a name="verify_outgoing"></a><a href="#verify_outgoing">`verify_outgoing`</a> - If set to
  true, Consul requires that all outgoing connections
  make use of TLS and that the server provides a certificate that is signed by
  the Certificate Authority from the [`ca_file`](#ca_file). By default, this is false, and Consul
  will not make use of TLS for outgoing connections. This applies to clients and servers
  as both will make outgoing connections.

* <a name="verify_server_hostname"></a><a href="#verify_server_hostname">`verify_server_hostname`</a> - If set to
  true, Consul verifies for all outgoing connections that the TLS certificate presented by the servers
  matches "server.&lt;datacenter&gt;.&lt;domain&gt;" hostname. This implies `verify_outgoing`.
  By default, this is false, and Consul does not verify the hostname of the certificate, only
  that it is signed by a trusted CA. This setting is important to prevent a compromised
  client from being restarted as a server, and thus being able to perform a MITM attack
  or to be added as a Raft peer. This is new in 0.5.1.

* <a name="watches"></a><a href="#watches">`watches`</a> - Watches is a list of watch
  specifications which allow an external process to be automatically invoked when a
  particular data view is updated. See the
   [watch documentation](/docs/agent/watches.html) for more detail. Watches can be
   modified when the configuration is reloaded.

## <a id="ports"></a>Ports Used

Consul requires up to 5 different ports to work properly, some on
TCP, UDP, or both protocols. Below we document the requirements for each
port.

* Server RPC (Default 8300). This is used by servers to handle incoming
  requests from other agents. TCP only.

* Serf LAN (Default 8301). This is used to handle gossip in the LAN.
  Required by all agents. TCP and UDP.

* Serf WAN (Default 8302). This is used by servers to gossip over the
  WAN to other servers. TCP and UDP.

* CLI RPC (Default 8400). This is used by all agents to handle RPC
  from the CLI, but is deprecated in Consul 0.8 and later. TCP only.
  In Consul 0.8 all CLI commands were changed to use the HTTP API and
  the RPC interface was completely removed.

* HTTP API (Default 8500). This is used by clients to talk to the HTTP
  API. TCP only.

* DNS Interface (Default 8600). Used to resolve DNS queries. TCP and UDP.

Consul will also make an outgoing connection to HashiCorp's servers for
Atlas-related features and to check for the availability of newer versions
of Consul. This will be a TLS-secured TCP connection to `scada.hashicorp.com:7223`.

## <a id="reloadable-configuration"></a>Reloadable Configuration

Reloading configuration does not reload all configuration items. The
items which are reloaded include:

* Log level
* Checks
* Services
* Watches
* HTTP Client Address
* <a href="#node_meta">Node Metadata</a>
* Atlas Token
* Atlas Infrastructure
* Atlas Endpoint
