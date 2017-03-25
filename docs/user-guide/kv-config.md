
# Key-value store configuration

Both [static global configuration](/user-guide/kv-config/#static-configuration-in-key-value-store) and [dynamic](/user-guide/kv-config/#dynamic-configuration-in-key-value-store) configuration can be sorted in a Key-value store.

This section explains how to launch Træfɪk using a configuration loaded from a Key-value store.

Træfɪk supports several Key-value stores:

- [Consul](https://consul.io)
- [etcd](https://coreos.com/etcd/)
- [ZooKeeper](https://zookeeper.apache.org/) 
- [boltdb](https://github.com/boltdb/bolt)

# Static configuration in Key-value store

We will see the steps to set it up with an easy example. 
Note that we could do the same with any other Key-value Store.

## docker-compose file for Consul

The Træfɪk global configuration will be getted from a [Consul](https://consul.io) store. 

First we have to launch Consul in a container. 
The [docker-compose file](https://docs.docker.com/compose/compose-file/) allows us to launch Consul and four instances of the trivial app [emilevauge/whoamI](https://github.com/emilevauge/whoamI) : 

```yml
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
  image: emilevauge/whoami
  
whoami2:
  image: emilevauge/whoami
  
whoami3:
  image: emilevauge/whoami
  
whoami4:
  image: emilevauge/whoami
```

## Upload the configuration in the Key-value store

We should now fill the store with the Træfɪk global configuration, as we do with a [TOML file configuration](/toml).
To do that, we can send the Key-value pairs via [curl commands](https://www.consul.io/intro/getting-started/kv.html) or via the [Web UI](https://www.consul.io/intro/getting-started/ui.html).

Fortunately, Træfɪk allows automation of this process using the `storeconfig` subcommand.
Please refer to the [store Træfɪk configuration](/user-guide/kv-config/#store-configuration-in-key-value-store) section to get documentation on it.

Here is the toml configuration we would like to store in the Key-value Store  :

```toml
logLevel = "DEBUG"

defaultEntryPoints = ["http", "https"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = """-----BEGIN CERTIFICATE-----
                      <cert file content>
                      -----END CERTIFICATE-----"""
      KeyFile = """-----BEGIN CERTIFICATE-----
                      <key file content>
                      -----END CERTIFICATE-----"""


[consul]
  endpoint = "127.0.0.1:8500"
  watch = true
  prefix = "traefik"
  
[web]
  address = ":8081"
```

And there, the same global configuration in the Key-value Store (using `prefix = "traefik"`): 

| Key                                                       | Value                                                         |
|-----------------------------------------------------------|---------------------------------------------------------------|
| `/traefik/loglevel`                                       | `DEBUG`                                                       |
| `/traefik/defaultentrypoints/0`                           | `http`                                                        |
| `/traefik/defaultentrypoints/1`                           | `https`                                                       |
| `/traefik/entrypoints/http/address`                       | `:80`                                                         |
| `/traefik/entrypoints/https/address`                      | `:443`                                                        |
| `/traefik/entrypoints/https/tls/certificates/0/certfile`  | `integration/fixtures/https/snitest.com.cert`                 |
| `/traefik/entrypoints/https/tls/certificates/0/keyfile`   | `integration/fixtures/https/snitest.com.key`                  |
| `/traefik/entrypoints/https/tls/certificates/1/certfile`  | `--BEGIN CERTIFICATE--<cert file content>--END CERTIFICATE--` |
| `/traefik/entrypoints/https/tls/certificates/1/keyfile`   | `--BEGIN CERTIFICATE--<key file content>--END CERTIFICATE--`  |
| `/traefik/consul/endpoint`                                | `127.0.0.1:8500`                                              |
| `/traefik/consul/watch`                                   | `true`                                                        |
| `/traefik/consul/prefix`                                  | `traefik`                                                     |
| `/traefik/web/address`                                    | `:8081`                                                       |

In case you are setting key values manually,:
 - Remember to specify the indexes (`0`,`1`, `2`, ... ) under prefixes `/traefik/defaultentrypoints/` and `/traefik/entrypoints/https/tls/certificates/` in order to match the global configuration structure.
 - Be careful to give the correct IP address and port on the key `/traefik/consul/endpoint`.

Note that we can either give path to certificate file or directly the file content itself.

## Launch Træfɪk

We will now launch Træfɪk in a container.
We use CLI flags to setup the connection between Træfɪk and Consul.
All the rest of the global configuration is stored in Consul.

Here is the [docker-compose file](https://docs.docker.com/compose/compose-file/) :

```yml
traefik:
  image: traefik
  command: --consul --consul.endpoint=127.0.0.1:8500
  ports:
    - "80:80"
    - "8080:8080"
```

NB : Be careful to give the correct IP address and port in the flag `--consul.endpoint`.

## TLS support

So far, only [Consul](https://consul.io) and [etcd](https://coreos.com/etcd/) support TLS connections. 
To set it up, we should enable [consul security](https://www.consul.io/docs/internals/security.html) (or [etcd security](https://coreos.com/etcd/docs/latest/security.html)).

Then, we have to provide CA, Cert and Key to Træfɪk using `consul` flags :

- `--consul.tls`
- `--consul.tls.ca=path/to/the/file`
- `--consul.tls.cert=path/to/the/file`
- `--consul.tls.key=path/to/the/file` 

Or etcd flags :

- `--etcd.tls`
- `--etcd.tls.ca=path/to/the/file`
- `--etcd.tls.cert=path/to/the/file`
- `--etcd.tls.key=path/to/the/file`

Note that we can either give directly directly the file content itself (instead of the path to certificate) in a TOML file configuration.

Remember the command `traefik --help` to display the updated list of flags.

# Dynamic configuration in Key-value store
Following our example, we will provide backends/frontends rules to Træfɪk.

Note that this section is independent of the way Træfɪk got its static configuration. 
It means that the static configuration can either come from the same Key-value store or from any other sources.

## Key-value storage structure
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
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"
  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
    rule = "Path:/test"
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

| Key                                                | Value              |
|----------------------------------------------------|--------------------|
| `/traefik/frontends/frontend2/backend`             | `backend1`         |
| `/traefik/frontends/frontend2/passHostHeader`      | `true`             |
| `/traefik/frontends/frontend2/priority`            | `10`               |
| `/traefik/frontends/frontend2/entrypoints`         | `http,https`       |
| `/traefik/frontends/frontend2/routes/test_2/rule`  | `PathPrefix:/test` |

## Atomic configuration changes

Træfɪk can watch the backends/frontends configuration changes and generate its configuration automatically. 

Note that only backends/frontends rules are dynamic, the rest of the Træfɪk configuration stay static. 

The [Etcd](https://github.com/coreos/etcd/issues/860) and [Consul](https://github.com/hashicorp/consul/issues/886) backends do not support updating multiple keys atomically. As a result, it may be possible for Træfɪk to read an intermediate configuration state despite judicious use of the `--providersThrottleDuration` flag. To solve this problem, Træfɪk supports a special key called `/traefik/alias`. If set, Træfɪk use the value as an alternative key prefix.

Given the key structure below, Træfɪk will use the `http://172.17.0.2:80` as its only backend (frontend keys have been omitted for brevity).

| Key                                                                     | Value                       |
|-------------------------------------------------------------------------|-----------------------------|
| `/traefik/alias`                                                        | `/traefik_configurations/1` |
| `/traefik_configurations/1/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/1/backends/backend1/servers/server1/weight`    | `10`                        |

When an atomic configuration change is required, you may write a new configuration at an alternative prefix. Here, although the `/traefik_configurations/2/...` keys have been set, the old configuration is still active because the `/traefik/alias` key still points to `/traefik_configurations/1`:

| Key                                                                     | Value                       |
|-------------------------------------------------------------------------|-----------------------------|
| `/traefik/alias`                                                        | `/traefik_configurations/1` |
| `/traefik_configurations/1/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/1/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik_configurations/2/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server1/weight`    | `5`                        |
| `/traefik_configurations/2/backends/backend1/servers/server2/url`       | `http://172.17.0.3:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server2/weight`    | `5`                        |

Once the `/traefik/alias` key is updated, the new `/traefik_configurations/2` configuration becomes active atomically. Here, we have a 50% balance between the `http://172.17.0.3:80` and the `http://172.17.0.4:80` hosts while no traffic is sent to the `172.17.0.2:80` host:

| Key                                                                     | Value                       |
|-------------------------------------------------------------------------|-----------------------------|
| `/traefik/alias`                                                        | `/traefik_configurations/2` |
| `/traefik_configurations/1/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik_configurations/1/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik_configurations/2/backends/backend1/servers/server1/url`       | `http://172.17.0.3:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server1/weight`    | `5`                        |
| `/traefik_configurations/2/backends/backend1/servers/server2/url`       | `http://172.17.0.4:80`      |
| `/traefik_configurations/2/backends/backend1/servers/server2/weight`    | `5`                        |

Note that Træfɪk *will not watch for key changes in the `/traefik_configurations` prefix*. It will only watch for changes in the `/traefik/alias`. 
Further, if the `/traefik/alias` key is set, all other configuration with `/traefik/backends` or `/traefik/frontends` prefix are ignored.

# Store configuration in Key-value store

Don't forget to [setup the connection between Træfɪk and Key-value store](/user-guide/kv-config/#launch-trfk).
The static Træfɪk configuration in a key-value store can be automatically created and updated, using the [`storeconfig` subcommand](/basics/#commands).

```bash
$ traefik storeconfig [flags] ...
```
This command is here only to automate the [process which upload the configuration into the Key-value store](/user-guide/kv-config/#upload-the-configuration-in-the-key-value-store).
Træfɪk will not start but the [static configuration](/basics/#static-trfk-configuration) will be uploaded into the Key-value store.
If you configured ACME (Let's Encrypt), your registration account and your certificates will also be uploaded.

To upload your ACME certificates to the KV store, get your traefik TOML file and add the new `storage` option in the `acme` section:

```
[acme]
email = "test@traefik.io"
storage = "traefik/acme/account" # the key where to store your certificates in the KV store
storageFile = "acme.json" # your old certificates store
```

Call `traefik storeconfig` to upload your config in the KV store.
Then remove the line `storageFile = "acme.json"` from your TOML config file.
That's it!


