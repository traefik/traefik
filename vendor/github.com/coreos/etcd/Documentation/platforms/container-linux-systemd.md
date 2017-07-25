# Run etcd on Container Linux with systemd

The following guide shows how to run etcd with [systemd][systemd-docs] under [Container Linux][container-linux-docs].

## Provisioning an etcd cluster

Cluster bootstrapping in Container Linux is simplest with [Ignition][container-linux-ignition]; `coreos-metadata.service` dynamically fetches the machine's IP for discovery. Note that etcd's discovery service protocol is only meant for bootstrapping, and cannot be used with runtime reconfiguration or cluster monitoring.

The [Container Linux Config Transpiler][container-linux-ct] compiles etcd configuration files into Ignition configuration files:

```yaml container-linux-config:norender
etcd:
  version: 3.2.0
  name: s1
  data_dir: /var/lib/etcd
  advertise_client_urls:       http://{PUBLIC_IPV4}:2379
  initial_advertise_peer_urls: http://{PRIVATE_IPV4}:2380
  listen_client_urls:          http://0.0.0.0:2379
  listen_peer_urls:            http://{PRIVATE_IPV4}:2380
  discovery:                   https://discovery.etcd.io/<token>
```

`ct` would produce the following Ignition Config:

```
$ ct --platform=gce --in-file /tmp/ct-etcd.cnf
{"ignition":{"version":"2.0.0","config"...
```

```json ignition-config
{
  "ignition":{"version":"2.0.0","config":{}},
  "storage":{},
  "systemd":{
    "units":[{
      "name":"etcd-member.service",
      "enable":true,
      "dropins":[{
        "name":"20-clct-etcd-member.conf",
        "contents":"[Unit]\nRequires=coreos-metadata.service\nAfter=coreos-metadata.service\n\n[Service]\nEnvironmentFile=/run/metadata/coreos\nEnvironment=\"ETCD_IMAGE_TAG=v3.1.8\"\nExecStart=\nExecStart=/usr/lib/coreos/etcd-wrapper $ETCD_OPTS \\\n  --name=\"s1\" \\\n  --data-dir=\"/var/lib/etcd\" \\\n  --listen-peer-urls=\"http://${COREOS_GCE_IP_LOCAL_0}:2380\" \\\n  --listen-client-urls=\"http://0.0.0.0:2379\" \\\n  --initial-advertise-peer-urls=\"http://${COREOS_GCE_IP_LOCAL_0}:2380\" \\\n  --advertise-client-urls=\"http://${COREOS_GCE_IP_EXTERNAL_0}:2379\" \\\n  --discovery=\"https://discovery.etcd.io/\u003ctoken\u003e\""}]}]},
      "networkd":{},
      "passwd":{}}
```

To avoid accidental misconfiguration, the transpiler helpfully verifies etcd configurations when generating Ignition files:

```yaml container-linux-config:norender
etcd:
  version: 3.2.0
  name: s1
  data_dir_x: /var/lib/etcd
  advertise_client_urls:       http://{PUBLIC_IPV4}:2379
  initial_advertise_peer_urls: http://{PRIVATE_IPV4}:2380
  listen_client_urls:          http://0.0.0.0:2379
  listen_peer_urls:            http://{PRIVATE_IPV4}:2380
  discovery:                   https://discovery.etcd.io/<token>
```

```
$ ct --platform=gce --in-file /tmp/ct-etcd.cnf
warning at line 3, column 2
Config has unrecognized key: data_dir_x
```

See [Container Linux Provisioning][container-linux-provision] for more details.

## etcd 3.x service

[Container Linux][container-linux-docs] does not include etcd 3.x binaries by default. Different versions of etcd 3.x can be fetched via `etcd-member.service`.

Confirm unit file exists:

```
systemctl cat etcd-member.service
```

Check if the etcd service is running:

```
systemctl status etcd-member.service
```

Example systemd drop-in unit to override the default service settings:

```bash
cat > /tmp/20-cl-etcd-member.conf <<EOF
[Service]
Environment="ETCD_IMAGE_TAG=v3.2.0"
Environment="ETCD_DATA_DIR=/var/lib/etcd"
Environment="ETCD_SSL_DIR=/etc/ssl/certs"
Environment="ETCD_OPTS=--name s1 \
  --listen-client-urls https://10.240.0.1:2379 \
  --advertise-client-urls https://10.240.0.1:2379 \
  --listen-peer-urls https://10.240.0.1:2380 \
  --initial-advertise-peer-urls https://10.240.0.1:2380 \
  --initial-cluster s1=https://10.240.0.1:2380,s2=https://10.240.0.2:2380,s3=https://10.240.0.3:2380 \
  --initial-cluster-token mytoken \
  --initial-cluster-state new \
  --client-cert-auth \
  --trusted-ca-file /etc/ssl/certs/etcd-root-ca.pem \
  --cert-file /etc/ssl/certs/s1.pem \
  --key-file /etc/ssl/certs/s1-key.pem \
  --peer-client-cert-auth \
  --peer-trusted-ca-file /etc/ssl/certs/etcd-root-ca.pem \
  --peer-cert-file /etc/ssl/certs/s1.pem \
  --peer-key-file /etc/ssl/certs/s1-key.pem \
  --auto-compaction-retention 1"
EOF
mv /tmp/20-cl-etcd-member.conf /etc/systemd/system/etcd-member.service.d/20-cl-etcd-member.conf
```

Or use a Container Linux Config:

```yaml container-linux-config:norender
systemd:
  units:
    - name: etcd-member.service
      dropins:
        - name: conf1.conf
          contents: |
            [Service]
            Environment="ETCD_SSL_DIR=/etc/ssl/certs"

etcd:
  version: 3.2.0
  name: s1
  data_dir: /var/lib/etcd
  listen_client_urls:          https://0.0.0.0:2379
  advertise_client_urls:       https://{PUBLIC_IPV4}:2379
  listen_peer_urls:            https://{PRIVATE_IPV4}:2380
  initial_advertise_peer_urls: https://{PRIVATE_IPV4}:2380
  initial_cluster:             s1=https://{PRIVATE_IPV4}:2380,s2=https://10.240.0.2:2380,s3=https://10.240.0.3:2380
  initial_cluster_token:       mytoken
  initial_cluster_state:       new
  client_cert_auth:            true
  trusted_ca_file:             /etc/ssl/certs/etcd-root-ca.pem
  cert-file:                   /etc/ssl/certs/s1.pem
  key-file:                    /etc/ssl/certs/s1-key.pem
  peer-client-cert-auth:       true
  peer-trusted-ca-file:        /etc/ssl/certs/etcd-root-ca.pem
  peer-cert-file:              /etc/ssl/certs/s1.pem
  peer-key-file:               /etc/ssl/certs/s1-key.pem
  auto-compaction-retention:   1
```

```
$ ct --platform=gce --in-file /tmp/ct-etcd.cnf
{"ignition":{"version":"2.0.0","config"...
```

To see all runtime drop-in changes for system units:

```
systemd-delta --type=extended
```

To enable and start:

```
systemctl daemon-reload
systemctl enable --now etcd-member.service
```

To see the logs:

```
journalctl --unit etcd-member.service --lines 10
```

To stop and disable the service:

```
systemctl disable --now etcd-member.service
```

## etcd 2.x service

[Container Linux][container-linux-docs] includes a unit file `etcd2.service` for etcd 2.x, which will be removed in the near future. See [Container Linux FAQ][container-linux-faq] for more details.

Confirm unit file is installed:

```
systemctl cat etcd2.service
```

Check if the etcd service is running:

```
systemctl status etcd2.service
```

To stop and disable:

```
systemctl disable --now etcd2.service
```

[systemd-docs]: https://github.com/systemd/systemd
[container-linux-docs]: https://coreos.com/os/docs/latest
[container-linux-faq]: https://github.com/coreos/docs/blob/master/etcd/os-faq.md
[container-linux-provision]: https://github.com/coreos/docs/blob/master/os/provisioning.md
[container-linux-ignition]: https://github.com/coreos/docs/blob/master/ignition/what-is-ignition.md
[container-linux-ct]: https://github.com/coreos/container-linux-config-transpiler
