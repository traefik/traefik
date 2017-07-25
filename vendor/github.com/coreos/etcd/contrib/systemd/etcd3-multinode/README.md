# etcd3 multi-node cluster

Here's how to deploy etcd cluster with systemd.

## Set up data directory

etcd needs data directory on host machine. Configure the data directory accessible to systemd as:

```
sudo mkdir -p /var/lib/etcd
sudo chown -R root:$(whoami) /var/lib/etcd
sudo chmod -R a+rw /var/lib/etcd
```

## Write systemd service file

In each machine, write etcd systemd service files:

```
cat > /tmp/my-etcd-1.service <<EOF
[Unit]
Description=etcd
Documentation=https://github.com/coreos/etcd
Conflicts=etcd.service
Conflicts=etcd2.service

[Service]
Type=notify
Restart=always
RestartSec=5s
LimitNOFILE=40000
TimeoutStartSec=0

ExecStart=etcd --name my-etcd-1 \
    --data-dir /var/lib/etcd \
    --listen-client-urls http://${IP_1}:2379 \
    --advertise-client-urls http://${IP_1}:2379 \
    --listen-peer-urls http://${IP_1}:2380 \
    --initial-advertise-peer-urls http://${IP_1}:2380 \
    --initial-cluster my-etcd-1=http://${IP_1}:2380,my-etcd-2=http://${IP_2}:2380,my-etcd-3=http://${IP_3}:2380 \
    --initial-cluster-token my-etcd-token \
    --initial-cluster-state new

[Install]
WantedBy=multi-user.target
EOF
sudo mv /tmp/my-etcd-1.service /etc/systemd/system/my-etcd-1.service
```

```
cat > /tmp/my-etcd-2.service <<EOF
[Unit]
Description=etcd
Documentation=https://github.com/coreos/etcd
Conflicts=etcd.service
Conflicts=etcd2.service

[Service]
Type=notify
Restart=always
RestartSec=5s
LimitNOFILE=40000
TimeoutStartSec=0

ExecStart=etcd --name my-etcd-2 \
    --data-dir /var/lib/etcd \
    --listen-client-urls http://${IP_2}:2379 \
    --advertise-client-urls http://${IP_2}:2379 \
    --listen-peer-urls http://${IP_2}:2380 \
    --initial-advertise-peer-urls http://${IP_2}:2380 \
    --initial-cluster my-etcd-1=http://${IP_1}:2380,my-etcd-2=http://${IP_2}:2380,my-etcd-3=http://${IP_3}:2380 \
    --initial-cluster-token my-etcd-token \
    --initial-cluster-state new

[Install]
WantedBy=multi-user.target
EOF
sudo mv /tmp/my-etcd-2.service /etc/systemd/system/my-etcd-2.service
```

```
cat > /tmp/my-etcd-3.service <<EOF
[Unit]
Description=etcd
Documentation=https://github.com/coreos/etcd
Conflicts=etcd.service
Conflicts=etcd2.service

[Service]
Type=notify
Restart=always
RestartSec=5s
LimitNOFILE=40000
TimeoutStartSec=0

ExecStart=etcd --name my-etcd-3 \
    --data-dir /var/lib/etcd \
    --listen-client-urls http://${IP_3}:2379 \
    --advertise-client-urls http://${IP_3}:2379 \
    --listen-peer-urls http://${IP_3}:2380 \
    --initial-advertise-peer-urls http://${IP_3}:2380 \
    --initial-cluster my-etcd-1=http://${IP_1}:2380,my-etcd-2=http://${IP_2}:2380,my-etcd-3=http://${IP_3}:2380 \
    --initial-cluster-token my-etcd-token \
    --initial-cluster-state new

[Install]
WantedBy=multi-user.target
EOF
sudo mv /tmp/my-etcd-3.service /etc/systemd/system/my-etcd-3.service
```

## Start the service

The service needs to be enabled first, in case of system reboot:

```
sudo systemctl daemon-reload
sudo systemctl enable my-etcd-1.service
sudo systemctl start my-etcd-1.service
```

```
sudo systemctl daemon-reload
sudo systemctl enable my-etcd-2.service
sudo systemctl start my-etcd-2.service
```

```
sudo systemctl daemon-reload
sudo systemctl enable my-etcd-3.service
sudo systemctl start my-etcd-3.service
```

## Check logs

systemd stores etcd server logs with journald:

```
sudo systemctl status my-etcd-1.service -l --no-pager
sudo journalctl -u my-etcd-1.service -l --no-pager|less
sudo journalctl -f -u my-etcd-1.service
```

```
sudo systemctl status my-etcd-2.service -l --no-pager
sudo journalctl -u my-etcd-2.service -l --no-pager|less
sudo journalctl -f -u my-etcd-2.service
```

```
sudo systemctl status my-etcd-3.service -l --no-pager
sudo journalctl -u my-etcd-3.service -l --no-pager|less
sudo journalctl -f -u my-etcd-3.service
```

## Stop etcd

To disable etcd process:

```
sudo systemctl stop my-etcd-1.service
sudo systemctl disable my-etcd-1.service
```

```
sudo systemctl stop my-etcd-2.service
sudo systemctl disable my-etcd-2.service
```

```
sudo systemctl stop my-etcd-3.service
sudo systemctl disable my-etcd-3.service
```
