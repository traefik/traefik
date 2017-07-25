* `-http-addr=<addr>` - Address of the Consul agent with the port. This can be
  an IP address or DNS address, but it must include the port. This can also be
  specified via the `CONSUL_HTTP_ADDR` environment variable. In Consul 0.8 and
  later, the default value is http://127.0.0.1:8500, and https can optionally
  be used instead. The scheme can also be set to HTTPS by setting the
  environment variable `CONSUL_HTTP_SSL=true`.

* `-token=<value>` - ACL token to use in the request. This can also be specified
  via the `CONSUL_HTTP_TOKEN` environment variable. If unspecified, the query
  will default to the token of the Consul agent at the HTTP address.
