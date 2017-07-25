## 0.8.1 (UNRELEASED)

FEATURES:

IMPROVEMENTS:

BUG FIXES:

## 0.8.0 (April 5, 2017)

BREAKING CHANGES:

* **Command-Line Interface RPC Deprecation:** The RPC client interface has been removed. All CLI commands that used RPC and the `-rpc-addr` flag to communicate with Consul have been converted to use the HTTP API and the appropriate flags for it, and the `rpc` field has been removed from the port and address binding configs. You will need to remove these fields from your config files and update any scripts that passed a custom `-rpc-addr` to the following commands: `force-leave`, `info`,  `join`, `keyring`, `leave`, `members`, `monitor`, `reload`

* **Version 8 ACLs Are Now Opt-Out:** The [`acl_enforce_version_8`](https://www.consul.io/docs/agent/options.html#acl_enforce_version_8) configuration now defaults to `true` to enable [full version 8 ACL support](https://www.consul.io/docs/internals/acl.html#version_8_acls) by default. If you are upgrading an existing cluster with ACLs enabled, you will need to set this to `false` during the upgrade on **both Consul agents and Consul servers**. Version 8 ACLs were also changed so that [`acl_datacenter`](https://www.consul.io/docs/agent/options.html#acl_datacenter) must be set on agents in order to enable the agent-side enforcement of ACLs. This makes for a smoother experience in clusters where ACLs aren't enabled at all, but where the agents would have to wait to contact a Consul server before learning that. [GH-2844]

* **Remote Exec Is Now Opt-In:** The default for [`disable_remote_exec`](https://www.consul.io/docs/agent/options.html#disable_remote_exec) was changed to "true", so now operators need to opt-in to having agents support running commands remotely via [`consul exec`](/docs/commands/exec.html). [GH-2854]

* **Raft Protocol Compatibility:** When upgrading to Consul 0.8.0 from a version lower than 0.7.0, users will need to
set the [`-raft-protocol`](https://www.consul.io/docs/agent/options.html#_raft_protocol) option to 1 in order to maintain backwards compatibility with the old servers during the upgrade. See [Upgrading Specific Versions](https://www.consul.io/docs/upgrade-specific.html) guide for more details.

FEATURES:

* **Autopilot:** A set of features has been added to allow for automatic operator-friendly management of Consul servers. For more information about Autopilot, see the [Autopilot Guide](https://www.consul.io/docs/guides/autopilot.html).
  - **Dead Server Cleanup:** Dead servers will periodically be cleaned up and removed from the Raft peer set, to prevent them from interfering with the quorum size and leader elections.
  - **Server Health Checking:** An internal health check has been added to track the stability of servers. The thresholds of this health check are tunable as part of the [Autopilot configuration](https://www.consul.io/docs/agent/options.html#autopilot) and the status can be viewed through the [`/v1/operator/autopilot/health`](https://www.consul.io/docs/agent/http/operator.html#autopilot-health) HTTP endpoint.
  - **New Server Stabilization:** When a new server is added to the cluster, there will be a waiting period where it must be healthy and stable for a certain amount of time before being promoted to a full, voting member. This threshold can be configured using the new [`server_stabilization_time`](https://www.consul.io/docs/agent/options.html#server_stabilization_time) setting.
  - **Advanced Redundancy:** (Consul Enterprise) A new [`-non-voting-server`](https://www.consul.io/docs/agent/options.html#_non_voting_server) option flag has been added for Consul servers to configure a server that does not participate in the Raft quorum. This can be used to add read scalability to a cluster in cases where a high volume of reads to servers are needed, but non-voting servers can be lost without causing an outage. There's also a new [`redundancy_zone_tag`](https://www.consul.io/docs/agent/options.html#redundancy_zone_tag) configuration that allows Autopilot to manage separating servers into zones for redundancy. Only one server in each zone can be a voting member at one time. This helps when Consul servers are managed with automatic replacement with a system like a resource scheduler or auto-scaling group. Extra non-voting servers in each zone will be available as hot standbys (that help with read-scaling) that can be quickly promoted into service when the voting server in a zone fails.
  - **Upgrade Orchestration:** (Consul Enterprise) Autopilot will automatically orchestrate an upgrade strategy for Consul servers where it will initially add newer versions of Consul servers as non-voters, wait for a full set of newer versioned servers to be added, and then gradually swap into service as voters and swap out older versioned servers to non-voters. This allows operators to safely bring up new servers, wait for the upgrade to be complete, and then terminate the old servers.
* **Network Areas:** (Consul Enterprise) A new capability has been added which allows operators to define network areas that join together two Consul datacenters. Unlike Consul's WAN feature, network areas use just the server RPC port for communication, and pairwise relationships can be made between arbitrary datacenters, so not all servers need to be fully connected. This allows for complex topologies among Consul datacenters like hub/spoke and more general trees. See the [Network Areas Guide](https://www.consul.io/docs/guides/areas.html) for more details.
* **WAN Soft Fail:** Request routing between servers in the WAN is now more robust by treating Serf failures as advisory but not final. This means that if there are issues between some subset of the servers in the WAN, Consul will still be able to route RPC requests as long as RPCs are actually still working. Prior to WAN Soft Fail, any datacenters having connectivity problems on the WAN would mean that all DCs might potentially stop sending RPCs to those datacenters. [GH-2801]
* **WAN Join Flooding:** A new routine was added that looks for Consul servers in the LAN and makes sure that they are joined into the WAN as well. This catches up up newly-added servers onto the WAN as soon as they join the LAN, keeping them in sync automatically. [GH-2801]
* **Validate command:** To provide consistency across our products, the `configtest` command has been deprecated and replaced with the `validate` command (to match Nomad and Terraform). The `configtest` command will be removed in Consul 0.9. [GH-2732]

IMPROVEMENTS:

* agent: Fixed a missing case where gossip would stop flowing to dead nodes for a short while. [GH-2722]
* agent: Changed agent to seed Go's random number generator. [GH-2722]
* agent: Serf snapshots no longer have the executable bit set on the file. [GH-2722]
* agent: Consul is now built with Go 1.8. [GH-2752]
* agent: Updated aws-sdk-go version (used for EC2 auto join) for Go 1.8 compatibility. [GH-2755]
* agent: User-supplied node IDs are now normalized to lower-case. [GH-2798]
* agent: Added checks to enforce uniqueness of agent node IDs at cluster join time and when registering with the catalog. [GH-2832]
* cli: Standardized handling of CLI options for connecting to the Consul agent. This makes sure that the same set of flags and environment variables works in all CLI commands (see https://www.consul.io/docs/commands/index.html#environment-variables). [GH-2717]
* cli: Updated go-cleanhttp library for better HTTP connection handling between CLI commands and the Consul agent (tunes reuse settings). [GH-2735]
* cli: The `operator raft` subcommand has had its two modes split into the `list-peers` and `remove-peer` subcommands. The old flags for these will continue to work for backwards compatibility, but will be removed in Consul 0.9.
* cli: Added an `-id` flag to the `operator raft remove-peer` command to allow removing a peer by ID. [GH-2847]
* dns: Allows the `.service` tag to be optional in RFC 2782 lookups. [GH-2690]
* server: Changed the internal `EnsureRegistration` RPC endpoint to prevent registering checks that aren't associated with the top-level node being registered. [GH-2846]

BUG FIXES:

* agent: Fixed an issue with `consul watch` not working when http was listening on a unix socket. [GH-2385]
* agent: Fixed an issue where checks and services could not sync deregister operations back to the catalog when version 8 ACL support is enabled. [GH-2818]
* agent: Fixed an issue where agents could use the ACL token registered with a service when registering checks for the same service that were registered with a different ACL token. [GH-2829]
* cli: Fixed `consul kv` commands not reading the `CONSUL_HTTP_TOKEN` environment variable. [GH-2566]
* cli: Fixed an issue where prefixing an address with a protocol (such as 'http://' or 'https://') in `-http-addr` or `CONSUL_HTTP_ADDR` would give an error.
* cli: Fixed an issue where error messages would get printed to stdout instead of stderr. [GH-2548]
* server: Fixed an issue with version 8 ACLs where servers couldn't deregister nodes from the catalog during reconciliation. [GH-2792] This fix was generalized and applied to registering nodes as well. [GH-2826]
* server: Fixed an issue where servers could temporarily roll back changes to a node's metadata or tagged addresses when making updates to the node's health checks. [GH-2826]
* server: Fixed an issue where the service name `consul` was not subject to service ACL policies with version 8 ACLs enabled. [GH-2816]

## 0.7.5 (February 15, 2017)

BUG FIXES:

* server: Fixed a rare but serious issue where Consul servers could panic when performing a large delete operation followed by a specific sequence of other updates to related parts of the state store (affects KV, sessions, prepared queries, and the catalog). [GH-2724]

## 0.7.4 (February 6, 2017)

IMPROVEMENTS:

* agent: Integrated gopsutil library to use built in host UUID as node ID, if available, instead of a randomly generated UUID. This makes it easier for other applications on the same host to generate the same node ID without coordinating with Consul. [GH-2697]
* agent: Added a configuration option, `tls_min_version`, for setting the minimum allowed TLS version used for the HTTP API and RPC. [GH-2699]
* agent: Added a `relay-factor` option to keyring operations to allow nodes to relay their response through N randomly-chosen other nodes in the cluster. [GH-2704]
* build: Consul is now built with Go 1.7.5. [GH-2682]
* dns: Add ability to lookup Consul agents by either their Node ID or Node Name through the node interface (e.g. DNS `(node-id|node-name).node.consul`). [GH-2702]

BUG FIXES:

* dns: Fixed an issue where SRV lookups for services on a node registered with non-IP addresses were missing the CNAME record in the additional section of the response. [GH-2695]

## 0.7.3 (January 26, 2017)

FEATURES:

* **KV Import/Export CLI:** `consul kv export` and `consul kv import` can be used to move parts of the KV tree between disconnected consul clusters, using JSON as the intermediate representation. [GH-2633]
* **Node Metadata:** Support for assigning user-defined metadata key/value pairs to nodes has been added. This can be viewed when looking up node info, and can be used to filter the results of various catalog and health endpoints. For more information, see the [Catalog](https://www.consul.io/docs/agent/http/catalog.html), [Health](https://www.consul.io/docs/agent/http/health.html), and [Prepared Query](https://www.consul.io/docs/agent/http/query.html) endpoint documentation, as well as the [Node Meta](https://www.consul.io/docs/agent/options.html#_node_meta) section of the agent configuration. [GH-2654]
* **Node Identifiers:** Consul agents can now be configured with a unique identifier, or they will generate one at startup that will persist across agent restarts. This identifier is designed to represent a node across all time, even if the name or address of the node changes. Identifiers are currently only exposed in node-related endpoints, but they will be used in future versions of Consul to help manage Consul servers and the Raft quorum in a more robust manner, as the quorum is currently tracked via addresses, which can change. [GH-2661]
* **Improved Blocking Queries:** Consul's [blocking query](https://www.consul.io/api/index.html#blocking-queries) implementation was improved to provide a much more fine-grained mechanism for detecting changes. For example, in previous versions of Consul blocking to wait on a change to a specific service would result in a wake up if any service changed. Now, wake ups are scoped to the specific service being watched, if possible. This support has been added to all endpoints that support blocking queries, nothing new is required to take advantage of this feature. [GH-2671]
* **GCE auto-discovery:** New `-retry-join-gce` configuration options added to allow bootstrapping by automatically discovering Google Cloud instances with a given tag at startup. [GH-2570]

IMPROVEMENTS:

* build: Consul is now built with Go 1.7.4. [GH-2676]
* cli: `consul kv get` now has a `-base64` flag to base 64 encode the value. [GH-2631]
* cli: `consul kv put` now has a `-base64` flag for setting values which are base 64 encoded. [GH-2632]
* ui: Added a notice that JS is required when viewing the web UI with JS disabled. [GH-2636]

BUG FIXES:

* agent: Redacted the AWS access key and secret key ID from the /v1/agent/self output so they are not disclosed. [GH-2677]
* agent: Fixed a rare startup panic due to a Raft/Serf race condition. [GH-1899]
* cli: Fixed a panic when an empty quoted argument was given to `consul kv put`. [GH-2635]
* tests: Fixed a race condition with check mock's map usage. [GH-2578]

## 0.7.2 (December 19, 2016)

FEATURES:

* **Keyring API:** A new `/v1/operator/keyring` HTTP endpoint was added that allows for performing operations such as list, install, use, and remove on the encryption keys in the gossip keyring. See the [Keyring Endpoint](https://www.consul.io/docs/agent/http/operator.html#keyring) for more details. [GH-2509]
* **Monitor API:** A new `/v1/agent/monitor` HTTP endpoint was added to allow for viewing streaming log output from the agent, similar to the `consul monitor` command. See the [Monitor Endpoint](https://www.consul.io/docs/agent/http/agent.html#agent_monitor) for more details. [GH-2511]
* **Reload API:** A new `/v1/agent/reload` HTTP endpoint was added for triggering a reload of the agent's configuration. See the [Reload Endpoint](https://www.consul.io/docs/agent/http/agent.html#agent_reload) for more details. [GH-2516]
* **Leave API:** A new `/v1/agent/leave` HTTP endpoint was added for causing an agent to gracefully shutdown and leave the cluster (previously, only `force-leave` was present in the HTTP API). See the [Leave Endpoint](https://www.consul.io/docs/agent/http/agent.html#agent_leave) for more details. [GH-2516]
* **Bind Address Templates (beta):** Consul agents now allow [go-sockaddr/template](https://godoc.org/github.com/hashicorp/go-sockaddr/template) syntax to be used for any bind address configuration (`advertise_addr`, `bind_addr`, `client_addr`, and others). This allows for easy creation of immutable images for Consul that can fetch their own address based on an interface name, network CIDR, address family from an actual RFC number, and many other possible schemes. This feature is in beta and we may tweak the template syntax before final release, but we encourage the community to try this and provide feedback. [GH-2563]
* **Complete ACL Coverage (beta):** Consul 0.8 will feature complete ACL coverage for all of Consul. To ease the transition to the new policies, a beta version of complete ACL support was added to help with testing and migration to the new features. Please see the [ACLs Internals Guide](https://www.consul.io/docs/internals/acl.html#version_8_acls) for more details. [GH-2594, GH-2592, GH-2590]

IMPROVEMENTS:

* agent: Defaults to `?pretty` JSON for HTTP API requests when in `-dev` mode. [GH-2518]
* agent: Updated Circonus metrics library and added new Circonus configration options for Consul for customizing check display name and tags. [GH-2555]
* agent: Added a checksum to UDP gossip messages to guard against packet corruption. [GH-2574]
* agent: Check whether a snapshot needs to be taken more often (every 5 seconds instead of 2 minutes) to keep the raft file smaller and to avoid doing huge truncations when writing lots of entries very quickly. [GH-2591]
* agent: Allow gossiping to suspected/recently dead nodes. [GH-2593]
* agent: Changed the gossip suspicion timeout to grow smoothly as the number of nodes grows. [GH-2593]
* agent: Added a deprecation notice for Atlas features to the CLI and docs. [GH-2597]
* agent: Give a better error message when the given data-dir is not a directory. [GH-2529]

BUG FIXES:

* agent: Fixed a panic when SIGPIPE signal was received. [GH-2404]
* api: Added missing Raft index fields to `CatalogService` structure. [GH-2366]
* api: Added missing notes field to `AgentServiceCheck` structure. [GH-2336]
* api: Changed type of `AgentServiceCheck.TLSSkipVerify` from `string` to `bool`. [GH-2530]
* api: Added new `HealthChecks.AggregatedStatus()` method that makes it easy get an overall health status from a list of checks. [GH-2544]
* api: Changed type of `KVTxnOp.Verb` from `string` to `KVOp`. [GH-2531]
* cli: Fixed an issue with the `consul kv put` command where a negative value would be interpreted as an argument to read from standard input. [GH-2526]
* ui: Fixed an issue where extra commas would be shown around service tags. [GH-2340]
* ui: Customized Bootstrap config to avoid missing font file references. [GH-2485]
* ui: Removed "Deregister" button as removing nodes from the catalog isn't a common operation and leads to lots of user confusion. [GH-2541]

## 0.7.1 (November 10, 2016)

BREAKING CHANGES:

* Child process reaping support has been removed, along with the `reap` configuration option. Reaping is also done via [dumb-init](https://github.com/Yelp/dumb-init) in the [Consul Docker image](https://github.com/hashicorp/docker-consul), so removing it from Consul itself simplifies the code and eases future maintainence for Consul. If you are running Consul as PID 1 in a container you will need to arrange for a wrapper process to reap child processes. [GH-1988]
* The default for `max_stale` has been increased to a near-indefinite threshold (10 years) to allow DNS queries to continue to be served in the event of a long outage with no leader. A new telemetry counter has also been added at `consul.dns.stale_queries` to track when agents serve DNS queries that are over a certain staleness (>5 seconds). [GH-2481]
* The api package's `PreparedQuery.Delete()` method now takes `WriteOptions` instead of `QueryOptions`. [GH-2417]

FEATURES:

* **Key/Value Store Command Line Interface:** New `consul kv` commands were added for easy access to all basic key/value store operations. [GH-2360]
* **Snapshot/Restore:** A new /v1/snapshot HTTP endpoint and corresponding set of `consul snapshot` commands were added for easy point-in-time snapshots for disaster recovery. Snapshots include all state managed by Consul's Raft [consensus protocol](/docs/internals/consensus.html), including Key/Value Entries, Service Catalog, Prepared Queries, Sessions, and ACLs. Snapshots can be restored on the fly into a completely fresh cluster. [GH-2396]
* **AWS auto-discovery:** New `-retry-join-ec2` configuration options added to allow bootstrapping by automatically discovering AWS instances with a given tag key/value at startup. [GH-2459]

IMPROVEMENTS:

* api: All session options can now be set when using `api.Lock()`. [GH-2372]
* agent: Added the ability to bind Serf WAN and LAN to different interfaces than the general bind address. [GH-2007]
* agent: Added a new `tls_skip_verify` configuration option for HTTP checks. [GH-1984]
* build: Consul is now built with Go 1.7.3. [GH-2281]

BUG FIXES:

* agent: Fixed a Go race issue with log buffering at startup. [GH-2262]
* agent: Fixed a panic during anti-entropy sync for services and checks. [GH-2125]
* agent: Fixed an issue on Windows where "wsarecv" errors were logged when CLI commands accessed the RPC interface. [GH-2356]
* agent: Syslog initialization will now retry on errors for up to 60 seconds to avoid a race condition at system startup. [GH-1610]
* agent: Fixed a panic when both -dev and -bootstrap-expect flags were provided. [GH-2464]
* agent: Added a retry with backoff when a session fails to invalidate after expiring. [GH-2435]
* agent: Fixed an issue where Consul would fail to start because of leftover malformed check/service state files. [GH-1221]
* agent: Fixed agent crashes on macOS Sierra by upgrading Go. [GH-2407, GH-2281]
* agent: Log a warning instead of success when attempting to deregister a nonexistent service. [GH-2492]
* api: Trim leading slashes from keys/prefixes when querying KV endpoints to avoid a bug with redirects in Go 1.7 (golang/go#4800). [GH-2403]
* dns: Fixed external services that pointed to consul addresses (CNAME records) not resolving to A-records. [GH-1228]
* dns: Fixed an issue with SRV lookups where the service address was different from the node's. [GH-832]
* dns: Fixed an issue where truncated records from a recursor query were improperly reported as errors. [GH-2384]
* server: Fixed the port numbers in the sample JSON inside peers.info. [GH-2391]
* server: Squashes ACL datacenter name to lower case and checks for proper formatting at startup. [GH-2059, GH-1778, GH-2478]
* ui: Fixed an XSS issue with the display of sessions and ACLs in the web UI. [GH-2456]

## 0.7.0 (September 14, 2016)

BREAKING CHANGES:

* The default behavior of `leave_on_terminate` and `skip_leave_on_interrupt` are now dependent on whether or not the agent is acting as a server or client. When Consul is started as a server the defaults for these are `false` and `true`, respectively, which means that you have to explicitly configure a server to leave the cluster. When Consul is started as a client the defaults are the opposite, which means by default, clients will leave the cluster if shutdown or interrupted. [GH-1909] [GH-2320]
* The `allow_stale` configuration for DNS queries to the Consul agent now defaults to `true`, allowing for better utilization of available Consul servers and higher throughput at the expense of weaker consistency. This is almost always an acceptable tradeoff for DNS queries, but this can be reconfigured to use the old default behavior if desired. [GH-2315]
* Output from HTTP checks is truncated to 4k when stored on the servers, similar to script check output. [GH-1952]
* Consul's Go API client will now send ACL tokens using HTTP headers instead of query parameters, requiring Consul 0.6.0 or later. [GH-2233]
* Removed support for protocol version 1, so Consul 0.7 is no longer compatible with Consul versions prior to 0.3. [GH-2259]
* The Raft peers information in `consul info` has changed format and includes information about the suffrage of a server, which will be used in future versions of Consul. [GH-2222]
* New [`translate_wan_addrs`](https://www.consul.io/docs/agent/options.html#translate_wan_addrs) behavior from [GH-2118] translates addresses in HTTP responses and could break clients that are expecting local addresses. A new `X-Consul-Translate-Addresses` header was added to allow clients to detect if translation is enabled for HTTP responses, and a "lan" tag was added to `TaggedAddresses` for clients that need the local address regardless of translation. [GH-2280]
* The behavior of the `peers.json` file is different in this version of Consul. This file won't normally be present and is used only during outage recovery. Be sure to read the updated [Outage Recovery Guide](https://www.consul.io/docs/guides/outage.html) for details. [GH-2222]
* Consul's default Raft timing is now set to work more reliably on lower-performance servers, which allows small clusters to use lower cost compute at the expense of reduced performance for failed leader detection and leader elections. You will need to configure Consul to get the same performance as before. See the new [Server Performance](https://www.consul.io/docs/guides/performance.html) guide for more details. [GH-2303]

FEATURES:

* **Transactional Key/Value API:** A new `/v1/txn` API was added that allows for atomic updates to and fetches from multiple entries in the key/value store inside of an atomic transaction. This includes conditional updates based on obtaining locks, and all other key/value store operations. See the [Key/Value Store Endpoint](https://www.consul.io/docs/agent/http/kv.html#txn) for more details. [GH-2028]
* **Native ACL Replication:** Added a built-in full replication capability for ACLs. Non-ACL datacenters can now replicate the complete ACL set locally to their state store and fall back to that if there's an outage. Additionally, this provides a good way to make a backup ACL datacenter, or to migrate the ACL datacenter to a different one. See the [ACL Internals Guide](https://www.consul.io/docs/internals/acl.html#replication) for more details. [GH-2237]
* **Server Connection Rebalancing:** Consul agents will now periodically reconnect to available Consul servers in order to redistribute their RPC query load. Consul clients will, by default, attempt to establish a new connection every 120s to 180s unless the size of the cluster is sufficiently large. The rate at which agents begin to query new servers is proportional to the size of the Consul cluster (servers should never receive more than 64 new connections per second per Consul server as a result of rebalancing). Clusters in stable environments who use `allow_stale` should see a more even distribution of query load across all of their Consul servers. [GH-1743]
* **Raft Updates and Consul Operator Interface:** This version of Consul upgrades to "stage one" of the v2 HashiCorp Raft library. This version offers improved handling of cluster membership changes and recovery after a loss of quorum. This version also provides a foundation for new features that will appear in future Consul versions once the remainder of the v2 library is complete. [GH-2222] <br> Consul's default Raft timing is now set to work more reliably on lower-performance servers, which allows small clusters to use lower cost compute at the expense of reduced performance for failed leader detection and leader elections. You will need to configure Consul to get the same performance as before. See the new [Server Performance](https://www.consul.io/docs/guides/performance.html) guide for more details. [GH-2303] <br> Servers will now abort bootstrapping if they detect an existing cluster with configured Raft peers. This will help prevent safe but spurious leader elections when introducing new nodes with `bootstrap_expect` enabled into an existing cluster. [GH-2319] <br> Added new `consul operator` command, HTTP endpoint, and associated ACL to allow Consul operators to view and update the Raft configuration. This allows a stale server to be removed from the Raft peers without requiring downtime and peers.json recovery file use. See the new [Consul Operator Command](https://www.consul.io/docs/commands/operator.html) and the [Consul Operator Endpoint](https://www.consul.io/docs/agent/http/operator.html) for details, as well as the updated [Outage Recovery Guide](https://www.consul.io/docs/guides/outage.html). [GH-2312]
* **Serf Lifeguard Updates:** Implemented a new set of feedback controls for the gossip layer that help prevent degraded nodes that can't meet the soft real-time requirements from erroneously causing `serfHealth` flapping in other, healthy nodes. This feature tunes itself automatically and requires no configuration. [GH-2101]
* **Prepared Query Near Parameter:** Prepared queries support baking in a new `Near` sorting parameter. This allows results to be sorted by network round trip time based on a static node, or based on the round trip time from the Consul agent where the request originated. This can be used to find a co-located service instance is one is available, with a transparent fallback to the next best alternate instance otherwise. [GH-2137]
* **Automatic Service Deregistration:** Added a new `deregister_critical_service_after` timeout field for health checks which will cause the service associated with that check to get deregistered if the check is critical for longer than the timeout. This is useful for cleanup of health checks registered natively by applications, or in other situations where services may not always be cleanly shutdown. [GH-679]
* **WAN Address Translation Everywhere:** Extended the [`translate_wan_addrs`](https://www.consul.io/docs/agent/options.html#translate_wan_addrs) config option to also translate node addresses in HTTP responses, making it easy to use this feature from non-DNS clients. [GH-2118]
* **RPC Retries:** Consul will now retry RPC calls that result in "no leader" errors for up to 5 seconds. This allows agents to ride out leader elections with a delayed response vs. an error. [GH-2175]
* **Circonus Telemetry Support:** Added support for Circonus as a telemetry destination. [GH-2193]

IMPROVEMENTS:

* agent: Reap time for failed nodes is now configurable via new `reconnect_timeout` and `reconnect_timeout_wan` config options ([use with caution](https://www.consul.io/docs/agent/options.html#reconnect_timeout)). [GH-1935]
* agent: Joins based on a DNS lookup will use TCP and attempt to join with the full list of returned addresses. [GH-2101]
* agent: Consul will now refuse to start with a helpful message if the same UNIX socket is used for more than one listening endpoint. [GH-1910]
* agent: Removed an obsolete warning message when Consul starts on Windows. [GH-1920]
* agent: Defaults bind address to 127.0.0.1 when running in `-dev` mode. [GH-1878]
* agent: Added version information to the log when Consul starts up. [GH-1404]
* agent: Added timing metrics for HTTP requests in the form of `consul.http.<verb>.<path>`. [GH-2256]
* build: Updated all vendored dependencies. [GH-2258]
* build: Consul releases are now built with Go 1.6.3. [GH-2260]
* checks: Script checks now support an optional `timeout` parameter. [GH-1762]
* checks: HTTP health checks limit saved output to 4K to avoid performance issues. [GH-1952]
* cli: Added a `-stale` mode for watchers to allow them to pull data from any Consul server, not just the leader. [GH-2045] [GH-917]
* dns: Consul agents can now limit the number of UDP answers returned via the DNS interface. The default number of UDP answers is `3`, however by adjusting the `dns_config.udp_answer_limit` configuration parameter, it is now possible to limit the results down to `1`. This tunable provides environments where RFC3484 section 6, rule 9 is enforced with an important workaround in order to preserve the desired behavior of randomized DNS results. Most modern environments will not need to adjust this setting as this RFC was made obsolete by RFC 6724\. See the [agent options](https://www.consul.io/docs/agent/options.html#udp_answer_limit) documentation for additional details for when this should be used. [GH-1712]
* dns: Consul now compresses all DNS responses by default. This prevents issues when recursing records that were originally compressed, where Consul would sometimes generate an invalid, uncompressed response that was too large. [GH-2266]
* dns: Added a new `recursor_timeout` configuration option to set the timeout for Consul's internal DNS client that's used for recursing queries to upstream DNS servers. [GH-2321]
* dns: Added a new `-dns-port` command line option so this can be set without a config file. [GH-2263]
* ui: Added a new network tomography visualization to the UI. [GH-2046]

BUG FIXES:

* agent: Fixed an issue where a health check's output never updates if the check status doesn't change after the Consul agent starts. [GH-1934]
* agent: External services can now be registered with ACL tokens. [GH-1738]
* agent: Fixed an issue where large events affecting many nodes could cause infinite intent rebroadcasts, leading to many log messages about intent queue overflows. [GH-1062]
* agent: Gossip encryption keys are now validated before being made persistent in the keyring, avoiding delayed feedback at runtime. [GH-1299]
* dns: Fixed an issue where DNS requests for SRV records could be incorrectly trimmed, resulting in an ADDITIONAL section that was out of sync with the ANSWER. [GH-1931]
* dns: Fixed two issues where DNS requests for SRV records on a prepared query that failed over would report the wrong domain and fail to translate addresses. [GH-2218] [GH-2220]
* server: Fixed a deadlock related to sorting the list of available datacenters by round trip time. [GH-2130]
* server: Fixed an issue with the state store's immutable radix tree that would prevent it from using cached modified objects during transactions, leading to extra copies and increased memory / GC pressure. [GH-2106]
* server: Upgraded Bolt DB to v1.2.1 to fix an issue on Windows where Consul would sometimes fail to start due to open user-mapped sections. [GH-2203]

OTHER CHANGES:

* build: Switched from Godep to govendor. [GH-2252]

## 0.6.4 (March 16, 2016)

BACKWARDS INCOMPATIBILITIES:

* Added a new `query` ACL type to manage prepared query names, and stopped capturing
  ACL tokens by default when prepared queries are created. This won't affect existing
  queries and how they are executed, but this will affect how they are managed. Now
  management of prepared queries can be delegated within an organization. If you use
  prepared queries, you'll need to read the
  [Consul 0.6.4 upgrade instructions](https://www.consul.io/docs/upgrade-specific.html)
  before upgrading to this version of Consul. [GH-1748]
* Consul's Go API client now pools connections by default, and requires you to manually
  opt-out of this behavior. Previously, idle connections were supported and their
  lifetime was managed by a finalizer, but this wasn't reliable in certain situations.
  If you reuse an API client object during the lifetime of your application, then there's
  nothing to do. If you have short-lived API client objects, you may need to configure them
  using the new `api.DefaultNonPooledConfig()` method to avoid leaking idle connections. [GH-1825]
* Consul's Go API client's `agent.UpdateTTL()` function was updated in a way that will
  only work with Consul 0.6.4 and later. The `agent.PassTTL()`, `agent.WarnTTL()`, and
  `agent.FailTTL()` functions were not affected and will continue work with older
  versions of Consul. [GH-1794]

FEATURES:

* Added new template prepared queries which allow you to define a prefix (possibly even
  an empty prefix) to apply prepared query features like datacenter failover to multiple
  services with a single query definition. This makes it easy to apply a common policy to
  multiple services without having to manage many prepared queries. See
  [Prepared Query Templates](https://www.consul.io/docs/agent/http/query.html#templates)
  for more details. [GH-1764]
* Added a new ability to translate address lookups when doing queries of nodes in
  remote datacenters via DNS using a new `translate_wan_addrs` configuration
  option. This allows the node to be reached within its own datacenter using its
  local address, and reached from other datacenters using its WAN address, which is
  useful in hybrid setups with mixed networks. [GH-1698]

IMPROVEMENTS:

* Added a new `disable_hostname` configuration option to control whether Consul's
  runtime telemetry gets prepended with the host name. All of the telemetry
  configuration has also been moved to a `telemetry` nested structure, but the old
  format is currently still supported. [GH-1284]
* Consul's Go dependencies are now vendored using Godep. [GH-1714]
* Added support for `EnableTagOverride` for the catalog in the Go API client. [GH-1726]
* Consul now ships built from Go 1.6. [GH-1735]
* Added a new `/v1/agent/check/update/<check id>` API for updating TTL checks which
  makes it easier to send large check output as part of a PUT body and not a query
  parameter. [GH-1785].
* Added a default set of `Accept` headers for HTTP checks. [GH-1819]
* Added support for RHEL7/Systemd in Terraform example. [GH-1629]

BUG FIXES:

* Updated the internal web UI (`-ui` option) to latest released build, fixing
  an ACL-related issue and the broken settings icon. [GH-1619]
* Fixed an issue where blocking KV reads could miss updates and return stale data
  when another key whose name is a prefix of the watched key was updated. [GH-1632]
* Fixed the redirect from `/` to `/ui` when the internal web UI (`-ui` option) is
  enabled. [GH-1713]
* Updated memberlist to pull in a fix for leaking goroutines when performing TCP
  fallback pings. This affected users with frequent UDP connectivity problems. [GH-1802]
* Added a fix to trim UDP DNS responses so they don't exceed 512 bytes. [GH-1813]
* Updated go-dockerclient to fix Docker health checks with Docker 1.10. [GH-1706]
* Removed fixed height display of nodes and services in UI, leading to broken displays
  when a node has a lot of services. [GH-2055]

## 0.6.3 (January 15, 2016)

BUG FIXES:

* Fixed an issue when running Consul as PID 1 in a Docker container where
  it could consume CPU and show spurious failures for health checks, watch
  handlers, and `consul exec` commands [GH-1592]

## 0.6.2 (January 13, 2016)

SECURITY:

* Build against Go 1.5.3 to mitigate a security vulnerability introduced
  in Go 1.5. For more information, please see https://groups.google.com/forum/#!topic/golang-dev/MEATuOi_ei4

This is a security-only release; other than the version number and building
against Go 1.5.3, there are no changes from 0.6.1.

## 0.6.1 (January 6, 2016)

BACKWARDS INCOMPATIBILITIES:

* The new `-monitor-retry` option to `consul lock` defaults to 3. This
  will cause the lock monitor to retry up to 3 times, waiting 1s between
  each attempt if it gets a 500 error from the Consul servers. For the
  vast majority of use cases this is desirable to prevent the lock from
  being given up during a brief period of Consul unavailability. If you
  want to get the previous default behavior you will need to set the
  `-monitor-retry=0` option.

IMPROVEMENTS:

* Consul is now built with Go 1.5.2
* Added source IP address and port information to RPC-related log error
  messages and HTTP access logs [GH-1513] [GH-1448]
* API clients configured for insecure SSL now use an HTTP transport that's
  set up the same way as the Go default transport [GH-1526]
* Added new per-host telemetry on DNS requests [GH-1537]
* Added support for reaping child processes which is useful when running
  Consul as PID 1 in Docker containers [GH-1539]
* Added new `-ui` command line and `ui` config option that enables a built-in
  Consul web UI, making deployment much simpler [GH-1543]
* Added new `-dev` command line option that creates a completely in-memory
  standalone Consul server for development
* Added a Solaris build, now that dependencies have been updated to support
  it [GH-1568]
* Added new `-try` option to `consul lock` to allow it to timeout with an error
  if it doesn't acquire the lock [GH-1567]
* Added a new `-monitor-retry` option to `consul lock` to help ride out brief
  periods of Consul unavailabily without causing the lock to be given up [GH-1567]

BUG FIXES:

* Fixed broken settings icon in web UI [GH-1469]
* Fixed a web UI bug where the supplied token wasn't being passed into
  the internal endpoint, breaking some pages when multiple datacenters
  were present [GH-1071]

## 0.6.0 (December 3, 2015)

BACKWARDS INCOMPATIBILITIES:

* A KV lock acquisition operation will now allow the lock holder to
  update the key's contents without giving up the lock by doing another
  PUT with `?acquire=<session>` and providing the same session that
  is holding the lock. Previously, this operation would fail.

FEATURES:

* Service ACLs now apply to service discovery [GH-1024]
* Added event ACLs to guard firing user events [GH-1046]
* Added keyring ACLs for gossip encryption keyring operations [GH-1090]
* Added a new TCP check type that does a connect as a check [GH-1130]
* Added new "tag override" feature that lets catalog updates to a
  service's tags flow down to agents [GH-1187]
* Ported in-memory database from LMDB to an immutable radix tree to improve
  read throughput, reduce garbage collection pressure, and make Consul 100%
  pure Go [GH-1291]
* Added support for sending telemetry to DogStatsD [GH-1293]
* Added new network tomography subsystem that estimates the network
  round trip times between nodes and exposes that in raw APIs, as well
  as in existing APIs (find the service node nearest node X); also
  includes a new `consul rtt` command to query interactively [GH-1331]
* Consul now builds under Go 1.5.1 by default [GH-1345]
* Added built-in support for running health checks inside Docker containers
  [GH-1343]
* Added prepared queries which support service health queries with rich
  features such as filters for multiple tags and failover to remote datacenters
  based on network coordinates; these are available via HTTP as well as the
  DNS interface [GH-1389]

BUG FIXES:

* Fixed expired certificates in unit tests [GH-979]
* Allow services with `/` characters in the UI [GH-988]
* Added SOA/NXDOMAIN records to negative DNS responses per RFC2308 [GH-995]
  [GH-1142] [GH-1195] [GH-1217]
* Token hiding in HTTP logs bug fixed [GH-1020]
* RFC6598 addresses are accepted as private IPs [GH-1050]
* Fixed reverse DNS lookups to recursor [GH-1137]
* Removes the trailing `/` added by the `consul lock` command [GH-1145]
* Fixed bad lock handler execution during shutdown [GH-1080] [GH-1158] [GH-1214]
* Added missing support for AAAA queries for nodes [GH-1222]
* Tokens passed from the CLI or API work for maint mode [GH-1230]
* Fixed service deregister/reregister flaps that could happen during
  `consul reload` [GH-1235]
* Fixed the Go API client to properly distinguish between expired sessions
  and sessions that don't exist [GH-1041]
* Fixed the KV section of the UI to work on Safari [GH-1321]
* Cleaned up JavaScript for built-in UI with bug fixes [GH-1338]

IMPROVEMENTS:

* Added sorting of `consul members` command output [GH-969]
* Updated AWS templates for RHEL6, CentOS6 [GH-992] [GH-1002]
* Advertised gossip/rpc addresses can now be configured [GH-1004]
* Failed lock acquisition handling now responds based on type of failure
  [GH-1006]
* Agents now remember check state across restarts [GH-1009]
* Always run ACL tests by default in API tests [GH-1030]
* Consul now refuses to start if there are multiple private IPs [GH-1099]
* Improved efficiency of servers managing incoming connections from agents
  [GH-1170]
* Added logging of the DNS client addresses in error messages [GH-1166]
* Added `-http-port` option to change the HTTP API port number [GH-1167]
* Atlas integration options are reload-able via SIGHUP [GH-1199]
* Atlas endpoint is a configurable option and CLI arg [GH-1201]
* Added `-pass-stdin` option to `consul lock` command [GH-1200]
* Enables the `/v1/internal/ui/*` endpoints, even if `-ui-dir` isn't set
  [GH-1215]
* Added HTTP method to Consul's log output for better debugging [GH-1270]
* Lock holders can `?acquire=<session>` a key again with the same session
  that holds the lock to update a key's contents without releasing the
  lock [GH-1291]
* Improved an O(n^2) algorithm in the agent's catalog sync code [GH-1296]
* Switched to net-rpc-msgpackrpc to reduce RPC overhead [GH-1307]
* Removed all uses of the http package's default client and transport in
  Consul to avoid conflicts with other packages [GH-1310] [GH-1327]
* Added new `X-Consul-Token` HTTP header option to avoid passing tokens
  in the query string [GH-1318]
* Increased session TTL max to 24 hours (use with caution, see note added
  to the Session HTTP endpoint documentation) [GH-1412]
* Added support to the API client for retrying lock monitoring when Consul
  is unavailable, helping prevent false indications of lost locks (eg. apps
  like Vault can avoid failing over when a Consul leader election occurs)
  [GH-1457]
* Added reap of receive buffer space for idle streams in the connection
  pool [GH-1452]

MISC:

* Lots of docs fixes
* Lots of Vagrantfile cleanup
* Data migrator utility removed to eliminate cgo dependency [GH-1309]

UPGRADE NOTES:

* Consul will refuse to start if the data directory contains an "mdb" folder.
  This folder was used in versions of Consul up to 0.5.1. Consul version 0.5.2
  included a baked-in utility to automatically upgrade the data format, but
  this has been removed in Consul 0.6 to eliminate the dependency on cgo.
* New service read, event firing, and keyring ACLs may require special steps to
  perform during an upgrade if ACLs are enabled and set to deny by default.
* Consul will refuse to start if there are multiple private IPs available, so
  if this is the case you will need to configure Consul's advertise or bind
  addresses before upgrading.

See https://www.consul.io/docs/upgrade-specific.html for detailed upgrade
instructions.

## 0.5.2 (May 18, 2015)

FEATURES:

* Include datacenter in the `members` output
* HTTP Health Check sets user agent "Consul Health Check" [GH-951]

BUG FIXES:

* Fixed memory leak caused by blocking query [GH-939]

MISC:

* Remove unused constant [GH-941]

## 0.5.1 (May 13, 2015)

FEATURES:

 * Ability to configure minimum session TTL. [GH-821]
 * Ability to set the initial state of a health check when registering [GH-859]
 * New `configtest` sub-command to verify config validity [GH-904]
 * ACL enforcement is prefix based for service names [GH-905]
 * ACLs support upsert for simpler restore and external generation [GH-909]
 * ACL tokens can be provided per-service during registration [GH-891]
 * Support for distinct LAN and WAN advertise addresses [GH-816]
 * Migrating Raft log from LMDB to BoltDB [GH-857]
 * `session_ttl_min` is now configurable to reduce the minimum TTL [GH-821]
 * Adding `verify_server_hostname` to protect against server forging [GH-927]

BUG FIXES:

 * Datacenter is lowercased, fixes DNS lookups [GH-761]
 * Deregister all checks when service is deregistered [GH-918]
 * Fixing issues with updates of persisted services [GH-910]
 * Chained CNAME resolution fixes [GH-862]
 * Tokens are filtered out of log messages [GH-860]
 * Fixing anti-entropy issue if servers rollback Raft log [GH-850]
 * Datacenter name is case insensitive for DNS lookups
 * Queries for invalid datacenters do not leak sockets [GH-807]

IMPROVEMENTS:

 * HTTP health checks more reliable, avoid KeepAlives [GH-824]
 * Improved protection against a passive cluster merge
 * SIGTERM is properly handled for graceful shutdown [GH-827]
 * Better staggering of deferred updates to checks [GH-884]
 * Configurable stats prefix [GH-902]
 * Raft uses BoltDB as the backend store. [GH-857]
 * API RenewPeriodic more resilient to transient errors [GH-912]

## 0.5.0 (February 19, 2015)

FEATURES:

 * Key rotation support for gossip layer. This allows the `encrypt` key
   to be changed globally.  See "keyring" command. [GH-336]
 * Options to join the WAN pool on start (`start_join_wan`, `retry_join_wan`) [GH-477]
 * Optional HTTPS interface [GH-478]
 * Ephemeral keys via "delete" session behavior. This allows keys to be deleted when
   a session is invalidated instead of having the lock released. Adds new "Behavior"
   field to Session which is configurable. [GH-487]
 * Reverse DNS lookups via PTR for IPv4 and IPv6 [GH-475]
 * API added checks and services are persisted. This means services and
   checks will survive a crash or restart. [GH-497]
 * ACLs can now protect service registration. Users in blacklist mode should
   allow registrations before upgrading to prevent a service disruption. [GH-506] [GH-465]
 * Sessions support a heartbeat failure detector via use of TTLs. This adds a new
   "TTL" field to Sessions and a `/v1/session/renew` endpoint. Heartbeats act like a
   failure detector (health check), but are managed by the servers. [GH-524] [GH-172]
 * Support for service specific IP addresses. This allows the service to advertise an
   address that is different from the agent. [GH-229] [GH-570]
 * Support KV Delete with Check-And-Set  [GH-589]
 * Merge `armon/consul-api` into `api` as official Go client.
 * Support for distributed locks and semaphores in API client [GH-594] [GH-600]
 * Support for native HTTP health checks [GH-592]
 * Support for node and service maintenance modes [GH-606]
 * Added new "consul maint" command to easily toggle maintenance modes [GH-625]
 * Added new "consul lock" command for simple highly-available deployments.
   This lets Consul manage the leader election and easily handle N+1 deployments
   without the applications being Consul aware. [GH-619]
 * Multiple checks can be associated with a service [GH-591] [GH-230]

BUG FIXES:

 * Fixed X-Consul-Index calculation for KV ListKeys
 * Fixed errors under extremely high read parallelism
 * Fixed issue causing event watches to not fire reliably [GH-479]
 * Fixed non-monotonic X-Consul-Index with key deletion [GH-577] [GH-195]
 * Fixed use of default instead of custom TLD in some DNS responses [GH-582]
 * Fixed memory leaks in API client when an error response is returned [GH-608]
 * Fixed issues with graceful leave in single-node bootstrap cluster [GH-621]
 * Fixed issue preventing node reaping [GH-371]
 * Fixed gossip stability at very large scale
 * Fixed string of rpc error: rpc error: ... no known leader. [GH-611]
 * Fixed panic in `exec` during cancellation
 * Fixed health check state reset caused by SIGHUP [GH-693]
 * Fixed bug in UI when multiple datacenters exist.

IMPROVEMENTS:

 * Support "consul exec" in foreign datacenter [GH-584]
 * Improved K/V blocking query performance [GH-578]
 * CLI respects CONSUL_RPC_ADDR environment variable to load parameter [GH-542]
 * Added support for multiple DNS recursors [GH-448]
 * Added support for defining multiple services per configuration file [GH-433]
 * Added support for defining multiple checks per configuration file [GH-433]
 * Allow mixing of service and check definitions in a configuration file [GH-433]
 * Allow notes for checks in service definition file [GH-449]
 * Random stagger for agent checks to prevent thundering herd [GH-546]
 * More useful metrics are sent to statsd/statsite
 * Added configuration to set custom HTTP headers (CORS) [GH-558]
 * Reject invalid configurations to simplify validation [GH-576]
 * Guard against accidental cluster mixing [GH-580] [GH-260]
 * Added option to filter DNS results on warning [GH-595]
 * Improve write throughput with raft log caching [GH-604]
 * Added ability to bind RPC and HTTP listeners to UNIX sockets [GH-587] [GH-612]
 * K/V HTTP endpoint returns 400 on conflicting flags [GH-634] [GH-432]

MISC:

 * UI confirms before deleting key sub-tree [GH-520]
 * More useful output in "consul version" [GH-480]
 * Many documentation improvements
 * Reduce log messages when quorum member is logs [GH-566]

UPGRADE NOTES:

 * If `acl_default_policy` is "deny", ensure tokens are updated to enable
   service registration to avoid a service disruption. The new ACL policy
   can be submitted with 0.4 before upgrading to 0.5 where it will be
   enforced.

 * Servers running 0.5.X cannot be mixed with older servers. (Any client
   version is fine). There is a 15 minute upgrade window where mixed
   versions are allowed before older servers will panic due to an unsupported
   internal command. This is due to the new KV tombstones which are internal
   to servers.

## 0.4.1 (October 20, 2014)

FEATURES:

 * Adding flags for `-retry-join` to attempt a join with
   configurable retry behavior. [GH-395]

BUG FIXES:

 * Fixed ACL token in UI
 * Fixed ACL reloading in UI [GH-323]
 * Fixed long session names in UI [GH-353]
 * Fixed exit code from remote exec [GH-346]
 * Fixing only a single watch being run by an agent [GH-337]
 * Fixing potential race in connection multiplexing
 * Fixing issue with Session ID and ACL ID generation. [GH-391]
 * Fixing multiple headers for /v1/event/list endpoint [GH-361]
 * Fixing graceful leave of leader causing invalid Raft peers [GH-360]
 * Fixing bug with closing TLS connection on error
 * Fixing issue with node reaping [GH-371]
 * Fixing aggressive deadlock time [GH-389]
 * Fixing syslog filter level [GH-272]
 * Serf snapshot compaction works on Windows [GH-332]
 * Raft snapshots work on Windows [GH-265]
 * Consul service entry clean by clients now possible
 * Fixing improper deserialization

IMPROVEMENTS:

 * Use "critical" health state instead of "unknown" [GH-341]
 * Consul service can be targeted for exec [GH-344]
 * Provide debug logging for session invalidation [GH-390]
 * Added "Deregister" button to UI [GH-364]
 * Added `enable_truncate` DNS configuration flag [GH-376]
 * Reduce mmap() size on 32bit systems [GH-265]
 * Temporary state is cleaned after an abort [GH-338] [GH-178]

MISC:

 * Health state "unknown" being deprecated

## 0.4.0 (September 5, 2014)

FEATURES:

 * Fine-grained ACL system to restrict access to KV store. Clients
   use tokens which can be restricted to (read, write, deny) permissions
   using longest-prefix matches.

 * Watch mechanisms added to invoke a handler when data changes in consul.
   Used with the `consul watch` command, or by specifying `watches` in
   an agent configuration.

 * Event system added to support custom user events. Events are fired using
   the `consul event` command. They are handled using a standard watch.

 * Remote execution using `consul exec`. This allows for command execution on remote
   instances mediated through Consul.

 * RFC-2782 style DNS lookups supported

 * UI improvements, including support for ACLs.

IMPROVEMENTS:

  * DNS case-insensitivity [GH-189]
  * Support for HTTP `?pretty` parameter to pretty format JSON output.
  * Use $SHELL when invoking handlers. [GH-237]
  * Agent takes the `-encrypt` CLI Flag [GH-245]
  * New `statsd_add` config for Statsd support. [GH-247]
  * New `addresses` config for providing an override to `client_addr` for
    DNS, HTTP, or RPC endpoints. [GH-301] [GH-253]
  * Support [Checkpoint](http://checkpoint.hashicorp.com) for security bulletins
    and update announcements.

BUG FIXES:

  * Fixed race condition in `-bootstrap-expect` [GH-254]
  * Require PUT to /v1/session/destroy [GH-285]
  * Fixed registration race condition [GH-300] [GH-279]

UPGRADE NOTES:

  * ACL support should not be enabled until all server nodes are running
  Consul 0.4. Mixed server versions with ACL support enabled may result in
  panics.

## 0.3.1 (July 21, 2014)

FEATURES:

  * Improved bootstrapping process, thanks to @robxu9

BUG FIXES:

  * Fixed issue with service re-registration [GH-216]
  * Fixed handling of `-rejoin` flag
  * Restored 0.2 TLS behavior, thanks to @nelhage [GH-233]
  * Fix the statsite flags, thanks to @nelhage [GH-243]
  * Fixed filters on critical / non-passing checks [GH-241]
  * Fixed initial log compaction crash [GH-297]

IMPROVEMENTS:

  * UI Improvements
  * Improved handling of Serf snapshot data
  * Increase reliability of failure detector
  * More useful logging messages


## 0.3.0 (June 13, 2014)

FEATURES:

  * Better, faster, cleaner UI [GH-194] [GH-196]
  * Sessions, which  act as a binding layer between
  nodes, checks and KV data. [GH-162]
  * Key locking. KV data integrates with sessions to
  enable distributed locking. [GH-162]
  * DNS lookups can do stale reads and TTLs. [GH-200]
  * Added new /v1/agent/self endpoint [GH-173]
  * `reload` command can be used to trigger configuration
  reload from the CLI [GH-142]

IMPROVEMENTS:

  * `members` has a much cleaner output format [GH-143]
  * `info` includes build version information
  * Sorted results for datacneter list [GH-198]
  * Switch multiplexing to yamux
  * Allow multiple CA certs in ca_file [GH-174]
  * Enable logging to syslog. [GH-105]
  * Allow raw key value lookup [GH-150]
  * Log encryption enabled [GH-151]
  * Support `-rejoin` to rejoin a cluster after a previous leave. [GH-110]
  * Support the "any" wildcard for v1/health/state/ [GH-152]
  * Defer sync of health check output [GH-157]
  * Provide output for serfHealth check [GH-176]
  * Datacenter name is validated [GH-169]
  * Configurable syslog facilities [GH-170]
  * Pipelining replication of writes
  * Raft group commits
  * Increased stability of leader terms
  * Prevent previously left nodes from causing re-elections

BUG FIXES:

  * Fixed memory leak in in-memory stats system
  * Fixing race between RPC and Raft init [GH-160]
  * Server-local RPC is avoids network [GH-148]
  * Fixing builds for older OSX [GH-147]

MISC:

  * Fixed missing prefixes on some log messages
  * Removed the `-role` filter of `members` command
  * Lots of docs fixes

## 0.2.1 (May 20, 2014)

IMPROVEMENTS:

  * Improved the URL formatting for the key/value editor in the Web UI.
      Importantly, the editor now allows editing keys with dashes in the
      name. [GH-119]
  * The web UI now has cancel and delete folder actions in the key/value
      editor. [GH-124], [GH-122]
  * Add flag to agent to write pid to a file. [GH-106]
  * Time out commands if Raft exceeds command enqueue timeout
  * Adding support for the `-advertise` CLI flag. [GH-156]
  * Fixing potential name conflicts on the WAN gossip ring [GH-158]
  * /v1/catalog/services returns an empty slice instead of null. [GH-145]
  * `members` command returns exit code 2 if no results. [GH-116]

BUG FIXES:

  * Renaming "separator" to "separator". This is the correct spelling,
      but both spellings are respected for backwards compatibility. [GH-101]
  * Private IP is properly found on Windows clients.
  * Windows agents won't show "failed to decode" errors on every RPC
      request.
  * Fixed memory leak with RPC clients. [GH-149]
  * Serf name conflict resolution disabled. [GH-97]
  * Raft deadlock possibility fixed. [GH-141]

MISC:

  * Updating to latest version of LMDB
  * Reduced the limit of KV entries to 512KB. [GH-123].
  * Warn if any Raft log exceeds 1MB
  * Lots of docs fixes

## 0.2.0 (May 1, 2014)

FEATURES:

  * Adding Web UI for Consul. This is enabled by providing the `-ui-dir` flag
      with the path to the web directory. The UI is visited at the standard HTTP
      address (Defaults to http://127.0.0.1:8500/). There is a demo
  [available here](http://demo.consul.io).
  * Adding new read consistency modes. `?consistent` can be used for strongly
      consistent reads without caveats. `?stale` can be used for stale reads to
      allow for higher throughput and read scalability. [GH-68]
  * /v1/health/service/ endpoint can take an optional `?passing` flag
      to filter to only nodes with passing results. [GH-57]
  * The KV endpoint supports listing keys with the `?keys` query parameter,
      and limited up to a separator using `?separator=`.

IMPROVEMENTS:

  * Health check output goes into separate `Output` field instead
      of overriding `Notes`. [GH-59]
  * Adding a minimum check interval to prevent checks with extremely
      low intervals fork bombing. [GH-64]
  * Raft peer set cleared on leave. [GH-69]
  * Case insensitive parsing checks. [GH-78]
  * Increase limit of DB size and Raft log on 64bit systems. [GH-81]
  * Output of health checks limited to 4K. [GH-83]
  * More warnings if GOMAXPROCS == 1 [GH-87]
  * Added runtime information to `consul info`

BUG FIXES:

  * Fixed 404 on /v1/agent/service/deregister and
      /v1/agent/check/deregister. [GH-95]
  * Fixed JSON parsing for /v1/agent/check/register [GH-60]
  * DNS parser can handler period in a tag name. [GH-39]
  * "application/json" content-type is sent on HTTP requests. [GH-45]
  * Work around for LMDB delete issue. [GH-85]
  * Fixed tag gossip propagation for rapid restart. [GH-86]

MISC:

  * More conservative timing values for Raft
  * Provide a warning if attempting to commit a very large Raft entry
  * Improved timeliness of registration when server is in bootstrap mode. [GH-72]

## 0.1.0 (April 17, 2014)

  * Initial release
