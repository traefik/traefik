# EntryPoints - Reference

Every Options for EntryPoints
{: .subtitle} 

## TOML

```toml
defaultEntryPoints = ["http", "https"]

# ...
# ...

[entryPoints]
  [entryPoints.http]
    address = ":80"
    [entryPoints.http.compress]
    
    [entryPoints.http.clientIPStrategy]
      depth = 5
      excludedIPs = ["127.0.0.1/32", "192.168.1.7"]

    [entryPoints.http.whitelist]
      sourceRange = ["10.42.0.0/16", "152.89.1.33/32", "afed:be44::/16"]
      [entryPoints.http.whitelist.IPStrategy]
        depth = 5
        excludedIPs = ["127.0.0.1/32", "192.168.1.7"]
          
    [entryPoints.http.tls]
      minVersion = "VersionTLS12"
      cipherSuites = [
        "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
        "TLS_RSA_WITH_AES_256_GCM_SHA384"
       ]
      [[entryPoints.http.tls.certificates]]
        certFile = "path/to/my.cert"
        keyFile = "path/to/my.key"
      [[entryPoints.http.tls.certificates]]
        certFile = "path/to/other.cert"
        keyFile = "path/to/other.key"
      # ...
      [entryPoints.http.tls.clientCA]
        files = ["path/to/ca1.crt", "path/to/ca2.crt"]
        optional = false

    [entryPoints.http.redirect]
      entryPoint = "https"
      regex = "^http://localhost/(.*)"
      replacement = "http://mydomain/$1"
      permanent = true

    [entryPoints.http.auth]
      headerField = "X-WebAuth-User"
      [entryPoints.http.auth.basic]
        removeHeader = true
        realm = "Your realm"
        users = [
          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
          "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
        ]
        usersFile = "/path/to/.htpasswd"
      [entryPoints.http.auth.digest]
        removeHeader = true
        users = [
          "test:traefik:a2688e031edb4be6a3797f3882655c05",
          "test2:traefik:518845800f9e2bfb1f1f740ec24f074e",
        ]
        usersFile = "/path/to/.htdigest"
      [entryPoints.http.auth.forward]
        address = "https://authserver.com/auth"
        trustForwardHeader = true
        authResponseHeaders = ["X-Auth-User"]
        [entryPoints.http.auth.forward.tls]
          ca = "path/to/local.crt"
          caOptional = true
          cert = "path/to/foo.cert"
          key = "path/to/foo.key"
          insecureSkipVerify = true

    [entryPoints.http.proxyProtocol]
      insecure = true
      trustedIPs = ["10.10.10.1", "10.10.10.2"]

    [entryPoints.http.forwardedHeaders]
      trustedIPs = ["10.10.10.1", "10.10.10.2"]
      insecure = false

  [entryPoints.https]
    # ...
```

## CLI

```ini
Name:foo
Address::80
TLS:/my/path/foo.cert,/my/path/foo.key;/my/path/goo.cert,/my/path/goo.key;/my/path/hoo.cert,/my/path/hoo.key
TLS
TLS.MinVersion:VersionTLS11
TLS.CipherSuites:TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA384
TLS.SniStrict:true
TLS.DefaultCertificate.Cert:path/to/foo.cert
TLS.DefaultCertificate.Key:path/to/foo.key
CA:car
CA.Optional:true
Redirect.EntryPoint:https
Redirect.Regex:http://localhost/(.*)
Redirect.Replacement:http://mydomain/$1
Redirect.Permanent:true
Compress:true
WhiteList.SourceRange:10.42.0.0/16,152.89.1.33/32,afed:be44::/16
WhiteList.IPStrategy.depth:3
WhiteList.IPStrategy.ExcludedIPs:10.0.0.3/24,20.0.0.3/24
ProxyProtocol.TrustedIPs:192.168.0.1
ProxyProtocol.Insecure:true
ForwardedHeaders.TrustedIPs:10.0.0.3/24,20.0.0.3/24
Auth.Basic.Users:test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0
Auth.Basic.Removeheader:true
Auth.Basic.Realm:traefik
Auth.Digest.Users:test:traefik:a2688e031edb4be6a3797f3882655c05,test2:traefik:518845800f9e2bfb1f1f740ec24f074e
Auth.Digest.Removeheader:true
Auth.HeaderField:X-WebAuth-User
Auth.Forward.Address:https://authserver.com/auth
Auth.Forward.AuthResponseHeaders:X-Auth,X-Test,X-Secret
Auth.Forward.TrustForwardHeader:true
Auth.Forward.TLS.CA:path/to/local.crt
Auth.Forward.TLS.CAOptional:true
Auth.Forward.TLS.Cert:path/to/foo.cert
Auth.Forward.TLS.Key:path/to/foo.key
Auth.Forward.TLS.InsecureSkipVerify:true
```
