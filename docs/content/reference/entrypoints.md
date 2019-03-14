# EntryPoints - Reference

Every Options for EntryPoints
{: .subtitle} 

## TOML

```toml
# ...

[entrypoints]
  [entrypoints.web]
    address = ":80"

    [entrypoints.web.proxyProtocol]
      insecure = true
      trustedIPs = ["10.10.10.1", "10.10.10.2"]

    [entrypoints.web.forwardedHeaders]
      trustedIPs = ["10.10.10.1", "10.10.10.2"]
      insecure = false

  [entrypoints.web-secure]
    # ...
```

## CLI

```ini
Name:foo
Address::80
ProxyProtocol.TrustedIPs:192.168.0.1
ProxyProtocol.Insecure:true
ForwardedHeaders.TrustedIPs:10.0.0.3/24,20.0.0.3/24
```
