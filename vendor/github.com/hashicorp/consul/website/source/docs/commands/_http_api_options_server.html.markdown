* `-datacenter=<name>` -  Name of the datacenter to query. If unspecified, the
  query will default to the datacenter of the Consul agent at the HTTP address.

* `-stale` - Permit any Consul server (non-leader) to respond to this request.
  This allows for lower latency and higher throughput, but can result in stale
  data. This option has no effect on non-read operations. The default value is
  false.