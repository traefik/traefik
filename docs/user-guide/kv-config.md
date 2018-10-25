# Key-value store configuration

Both [static global configuration](/user-guide/kv-config/#static-configuration-in-key-value-store) and [dynamic](/user-guide/kv-config/#dynamic-configuration-in-key-value-store) configuration can be stored in a Key-value store.

This section explains how to launch Traefik using a configuration loaded from a Key-value store.

Traefik supports several Key-value stores:

- [Consul](https://consul.io)
- [etcd](https://coreos.com/etcd/)
- [ZooKeeper](https://zookeeper.apache.org/)
- [boltdb](https://github.com/boltdb/bolt)

## Static configuration in Key-value store

We will see the steps to set it up with an easy example.

!!! note
    We could do the same with any other Key-value Store.

### docker-compose file for Consul

The Traefik global configuration will be retrieved from a [Consul](https://consul.io) store.

First we have to launch Consul in a container.

The [docker-compose file](https://docs.docker.com/compose/compose-file/) allows us to launch Consul and four instances of the trivial app [containous/whoami](https://github.com/containous/whoami) :

```yaml
consul:
  image: progrium/consul
  command: -server -bootstrap -log-level debug -ui-dir /ui
  ports:
    - "8400:8400"
    - "8500:8500"
    - "8600:53/udp"
  expose:
    - "8300"
    - "8301"
    - "8301/udp"
    - "8302"
    - "8302/udp"

whoami1:
  image: containous/whoami

whoami2:
  image: containous/whoami

whoami3:
  image: containous/whoami

whoami4:
  image: containous/whoami
```

### Upload the configuration in the Key-value store

We should now fill the store with the Traefik global configuration.  
To do that, we can send the Key-value pairs via [curl commands](https://www.consul.io/intro/getting-started/kv.html) or via the [Web UI](https://www.consul.io/intro/getting-started/ui.html).

Fortunately, Traefik allows automation of this process using the `storeconfig` subcommand.  
Please refer to the [store Traefik configuration](/user-guide/kv-config/#store-configuration-in-key-value-store) section to get documentation on it.

Here is the toml configuration we would like to store in the Key-value Store  :

```toml
logLevel = "DEBUG"

defaultEntryPoints = ["http", "https"]

[entryPoints]
  [entryPoints.api]
    address = ":8081"
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"

    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.com.cert"
      keyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      certFile = """-----BEGIN CERTIFICATE-----
                      <cert file content>
                      -----END CERTIFICATE-----"""
      keyFile = """-----BEGIN PRIVATE KEY-----
                      <key file content>
                      -----END PRIVATE KEY-----"""
    [entryPoints.other-https]
    address = ":4443"
      [entryPoints.other-https.tls]

[consul]
  endpoint = "127.0.0.1:8500"
  watch = true
  prefix = "traefik"

[api]
  entrypoint = "api"
```

And there, the same global configuration in the Key-value Store (using `prefix = "traefik"`):

| Key                                                       | Value                                                         |
|-----------------------------------------------------------|---------------------------------------------------------------|
| `/traefik/loglevel`                                       | `DEBUG`                                                       |
| `/traefik/defaultentrypoints/0`                           | `http`                                                        |
| `/traefik/defaultentrypoints/1`                           | `https`                                                       |
| `/traefik/entrypoints/api/address`                        | `:8081`                                                       |
| `/traefik/entrypoints/http/address`                       | `:80`                                                         |
| `/traefik/entrypoints/https/address`                      | `:443`                                                        |
| `/traefik/entrypoints/https/tls/certificates/0/certfile`  | `integration/fixtures/https/snitest.com.cert`                 |
| `/traefik/entrypoints/https/tls/certificates/0/keyfile`   | `integration/fixtures/https/snitest.com.key`                  |
| `/traefik/entrypoints/https/tls/certificates/1/certfile`  | `--BEGIN CERTIFICATE--<cert file content>--END CERTIFICATE--` |
| `/traefik/entrypoints/https/tls/certificates/1/keyfile`   | `--BEGIN CERTIFICATE--<key file content>--END CERTIFICATE--`  |
| `/traefik/entrypoints/other-https/address`                | `:4443`                                                       |
| `/traefik/consul/endpoint`                                | `127.0.0.1:8500`                                              |
| `/traefik/consul/watch`                                   | `true`                                                        |
| `/traefik/consul/prefix`                                  | `traefik`                                                     |
| `/traefik/api/entrypoint`                                 | `api`                                                         |

In case you are setting key values manually:

- Remember to specify the indexes (`0`,`1`, `2`, ... ) under prefixes `/traefik/defaultentrypoints/` and `/traefik/entrypoints/https/tls/certificates/` in order to match the global configuration structure.
- Be careful to give the correct IP address and port on the key `/traefik/consul/endpoint`.

Note that we can either give path to certificate file or directly the file content itself.

### Launch Traefik

We will now launch Traefik in a container.

We use CLI flags to setup the connection between Traefik and Consul.
All the rest of the global configuration is stored in Consul.

Here is the [docker-compose file](https://docs.docker.com/compose/compose-file/) :

```yaml
traefik:
  image: traefik
  command: --consul --consul.endpoint=127.0.0.1:8500
  ports:
    - "80:80"
    - "8080:8080"
```

!!! warning
    Be careful to give the correct IP address and port in the flag `--consul.endpoint`.

### Consul ACL Token support

To specify a Consul ACL token for Traefik, we have to set a System Environment variable named `CONSUL_HTTP_TOKEN` prior to starting Traefik.
This variable must be initialized with the ACL token value.

If Traefik is launched into a Docker container, the variable `CONSUL_HTTP_TOKEN` can be initialized with the `-e` Docker option : `-e "CONSUL_HTTP_TOKEN=[consul-acl-token-value]"`

If a Consul ACL is used to restrict Traefik read/write access, one of the following configurations is needed.

- HCL format :

```
    key "traefik" {
        policy = "write"
    },

    session "" {
        policy = "write"
    }
```

- JSON format :

```json
{
    "key": {
        "traefik": {
          "policy": "write"
        }
    },
    "session": {
        "": {
        "policy": "write"
        }
    }
}
```

### TLS support

To connect to a Consul endpoint using SSL, simply specify `https://` in the `consul.endpoint` property

- `--consul.endpoint=https://[consul-host]:[consul-ssl-port]`

### TLS support with client certificates

So far, only [Consul](https://consul.io) and [etcd](https://coreos.com/etcd/) support TLS connections with client certificates.

To set it up, we should enable [consul security](https://www.consul.io/docs/internals/security.html) (or [etcd security](https://coreos.com/etcd/docs/latest/security.html)).

Then, we have to provide CA, Cert and Key to Traefik using `consul` flags :

- `--consul.tls`
- `--consul.tls.ca=path/to/the/file`
- `--consul.tls.cert=path/to/the/file`
- `--consul.tls.key=path/to/the/file`

Or etcd flags :

- `--etcd.tls`
- `--etcd.tls.ca=path/to/the/file`
- `--etcd.tls.cert=path/to/the/file`
- `--etcd.tls.key=path/to/the/file`

!! note
    We can either give directly directly the file content itself (instead of the path to certificate) in a TOML file configuration.

Remember the command `traefik --help` to display the updated list of flags.

## Dynamic configuration in Key-value store

Following our example, we will provide backends/frontends  rules and HTTPS certificates to Traefik.

!!! note
    This section is independent of the way Traefik got its static configuration.
    It means that the static configuration can either come from the same Key-value store or from any other sources.

### Key-value storage structure

Here is the toml configuration we would like to store in the store :

```toml
[file]

# rules
[backends]
  [backends.backend1]
    [backends.backend1.circuitbreaker]
    expression = "NetworkErrorRatio() > 0.5"
    [backends.backend1.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1
  [backends.backend2]
    [backends.backend1.maxconn]
    amount = 10
    extractorfunc = "request.host"
    [backends.backend2.LoadBalancer]
    method = "drr"
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1
    [backends.backend2.servers.server2]
    url = "http://172.17.0.5:80"
    weight = 2

[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host:test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  priority = 10
      [frontends.frontend2.auth.basic]
      users = [
        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
        "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
      ]
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"
  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
  rule = "Path:/test"

[[tls]]
  [tls.certificate]
    certFile = "path/to/your.cert"
    keyFile = "path/to/your.key"

[[tls]]
  entryPoints = ["https","other-https"]
  [tls.certificate]
    certFile = """-----BEGIN CERTIFICATE-----
                      <cert file content>
                      -----END CERTIFICATE-----"""
    keyFile = """-----BEGIN CERTIFICATE-----
                      <key file content>
                      -----END CERTIFICATE-----"""
```

And there, the same dynamic configuration in a KV Store (using `prefix = "traefik"`):

- backend 1

| Key                                                    | Value                       |
|--------------------------------------------------------|-----------------------------|
| `/traefik/backends/backend1/circuitbreaker/expression` | `NetworkErrorRatio() > 0.5` |
| `/traefik/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik/backends/backend1/servers/server2/url`       | `http://172.17.0.3:80`      |
| `/traefik/backends/backend1/servers/server2/weight`    | `1`                         |
| `/traefik/backends/backend1/servers/server2/tags`      | `api,helloworld`            |

- backend 2

| Key                                                 | Value                  |
|-----------------------------------------------------|------------------------|
| `/traefik/backends/backend2/maxconn/amount`         | `10`                   |
| `/traefik/backends/backend2/maxconn/extractorfunc`  | `request.host`         |
| `/traefik/backends/backend2/loadbalancer/method`    | `drr`                  |
| `/traefik/backends/backend2/servers/server1/url`    | `http://172.17.0.4:80` |
| `/traefik/backends/backend2/servers/server1/weight` | `1`                    |
| `/traefik/backends/backend2/servers/server2/url`    | `http://172.17.0.5:80` |
| `/traefik/backends/backend2/servers/server2/weight` | `2`                    |
| `/traefik/backends/backend2/servers/server2/tags`   | `web`                  |

- frontend 1

| Key                                               | Value                 |
|---------------------------------------------------|-----------------------|
| `/traefik/frontends/frontend1/backend`            | `backend2`            |
| `/traefik/frontends/frontend1/routes/test_1/rule` | `Host:test.localhost` |

- frontend 2

| Key                                                | Value                                         |
|----------------------------------------------------|-----------------------------------------------|
| `/traefik/frontends/frontend2/backend`             | `backend1`                                    |
| `/traefik/frontends/frontend2/passhostheader`      | `true`                                        |
| `/traefik/frontends/frontend2/priority`            | `10`                                          |
| `/traefik/frontends/frontend2/auth/basic/users/0`  | `test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/`  |
| `/traefik/frontends/frontend2/auth/basic/users/1`  | `test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0` |
| `/traefik/frontends/frontend2/entrypoints`         | `http,https`                                  |
| `/traefik/frontends/frontend2/routes/test_2/rule`  | `PathPrefix:/test`                            |

- certificate 1

| Key                                   | Value              |
|---------------------------------------|--------------------|
| `/traefik/tls/1/certificate/certfile` | `path/to/your.cert`|
| `/traefik/tls/1/certificate/keyfile`  | `path/to/your.key` |

!!! note
    As `/traefik/tls/1/entrypoints` is not defined, the certificate will be attached to all `defaulEntryPoints` with a TLS configuration (in the example, the entryPoint `https`)

- certificate 2

| Key                                   | Value                 |
|---------------------------------------|-----------------------|
| `/traefik/tls/2/entrypoints`          | `https,other-https`   |
| `/traefik/tls/2/certificate/certfile` | `<cert file content>` |
| `/traefik/tls/2/certificate/keyfile`  | `<key file content>`  |

### Atomic configuration changes

Traefik can watch the backends/frontends configuration changes and generate its configuration automatically.

!!! note
    Only backends/frontends rules are dynamic, the rest of the Traefik configuration stay static.

The [Etcd](https://github.com/coreos/etcd/issues/860) and [Consul](https://github.com/hashicorp/consul/issues/886) backends do not support updating multiple keys atomically.  
As a result, it may be possible for Traefik to read an intermediate configuration state despite judicious use of the `--providersThrottleDuration` flag.  
To solve this problem, Traefik supports a special key called `/traefik/alias`.
If set, Traefik use the value as an alternative key prefix.

!!! note
    The field `useAPIV3` allows using Etcd V3 API which should support updating multiple keys atomically with Etcd.
    Etcd API V2 is deprecated and, in the future, Traefik will support API V3 by default.

Given the key structure below, Traefik will use the `http://172.17.0.2:80` as its only backend (frontend keys have been omitted for brevity).

| Key                                                                     | Value                       |
|-------------------------------------------------------------------------|-----------------------------|
| `/traefik/alias`                                                        | `/traefik_configurations/1` |
| `/traefik_configurations/1/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/1/backends/backend1/servers/server1/weight`    | `10`                        |

When an atomic configuration change is required, you may write a new configuration at an alternative prefix.

Here, although the `/traefik_configurations/2/...` keys have been set, the old configuration is still active because the `/traefik/alias` key still points to `/traefik_configurations/1`:

| Key                                                                     | Value                       |
|-------------------------------------------------------------------------|-----------------------------|
| `/traefik/alias`                                                        | `/traefik_configurations/1` |
| `/traefik_configurations/1/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/1/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik_configurations/2/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server1/weight`    | `5`                         |
| `/traefik_configurations/2/backends/backend1/servers/server2/url`       | `http://172.17.0.3:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server2/weight`    | `5`                         |

Once the `/traefik/alias` key is updated, the new `/traefik_configurations/2` configuration becomes active atomically.

Here, we have a 50% balance between the `http://172.17.0.3:80` and the `http://172.17.0.4:80` hosts while no traffic is sent to the `172.17.0.2:80` host:

| Key                                                                     | Value                       |
|-------------------------------------------------------------------------|-----------------------------|
| `/traefik/alias`                                                        | `/traefik_configurations/2` |
| `/traefik_configurations/1/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/1/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik_configurations/2/backends/backend1/servers/server1/url`       | `http://172.17.0.3:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server1/weight`    | `5`                         |
| `/traefik_configurations/2/backends/backend1/servers/server2/url`       | `http://172.17.0.4:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server2/weight`    | `5`                         |

!!! note
    Traefik *will not watch for key changes in the `/traefik_configurations` prefix*. It will only watch for changes in the `/traefik/alias`.  
    Further, if the `/traefik/alias` key is set, all other configuration with `/traefik/backends` or `/traefik/frontends` prefix are ignored.

## Store configuration in Key-value store

!!! note
    Don't forget to [setup the connection between Traefik and Key-value store](/user-guide/kv-config/#launch-traefik).

The static Traefik configuration in a key-value store can be automatically created and updated, using the [`storeconfig` subcommand](/basics/#commands).

```bash
traefik storeconfig [flags] ...
```
This command is here only to automate the [process which upload the configuration into the Key-value store](/user-guide/kv-config/#upload-the-configuration-in-the-key-value-store).
Traefik will not start but the [static configuration](/basics/#static-traefik-configuration) will be uploaded into the Key-value store.  

If you configured ACME (Let's Encrypt), your registration account and your certificates will also be uploaded.

If you configured a file provider `[file]`, all your dynamic configuration (backends, frontends...) will be uploaded to the Key-value store.

To upload your ACME certificates to the KV store, get your Traefik TOML file and add the new `storage` option in the `acme` section:

```toml
[acme]
email = "test@traefik.io"
storage = "traefik/acme/account" # the key where to store your certificates in the KV store
storageFile = "acme.json" # your old certificates store
```

Call `traefik storeconfig` to upload your config in the KV store.
Then remove the line `storageFile = "acme.json"` from your TOML config file.

That's it!

![GIF Magica](https://i.giphy.com/ujUdrdpX7Ok5W.gif)
