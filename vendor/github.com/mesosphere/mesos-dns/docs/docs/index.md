---
title: Installing and running Mesos-DNS
---

## Installing and running Mesos-DNS

Stable binaries for the project are published via the GitHub release
channel: https://github.com/mesosphere/mesos-dns/releases.

To run Mesos-DNS, you first need to install the `mesos-dns` binary somewhere on a selected server. The server can be the same machine as one of the Mesos masters, one of the slaves, or a dedicated machine on the same network. Next, follow [these instructions](configuration-parameters.html) to create a configuration file for your cluster. You can launch Mesos-DNS with: 

```
sudo mesos-dns -config=config.json & 
```

### Mesos-DNS with Docker
If you choose to use Mesos-DNS with Docker, with a version of Mesos after 0.25, be aware that there are some caveats. By default the Docker executor publishes the IP of the Docker container into the NetworkInfo field. Unfortunately, unless you're running some kind of SDN solution, bridged, or host networking with Docker, this can prove to make the containers unreachable.

The default configuration that Mesos-DNS ships with in config.json.sample omits `netinfo` from the sources. The default options if you omit this field from the configuration includes `netinfo`. If you have trouble with Docker, ensure you check the IPSources field to omit netinfo.

### Slave Setup

To allow Mesos tasks to use Mesos-DNS as the primary DNS server, you must edit the file `/etc/resolv.conf` in every slave and add a new nameserver. For instance, if `mesos-dns` runs on the server with IP address `10.181.64.13`, you should add the line `nameserver 10.181.64.13` at the ***beginning*** of `/etc/resolv.conf` on every slave node. This can be achieve by running:

```
sudo sed -i '1s/^/nameserver 10.181.64.13\n /' /etc/resolv.conf
```

If multiple instances of Mesos-DNS are launched, add a nameserver line for each one at the beginning of `/etc/resolv.conf`. The order of these entries determines the order that the slave will use to contact Mesos-DNS instances. You can set `options rotate` to instruct select between the listed nameservers in a round-robin manner for load balancing.  

All other nameserver settings in `/etc/resolv.conf` should remain unchanged. The `/etc/resolv.conf` file in the masters should only change if the master machines are also used as slaves. 

You can also use Mesos-DNS to serve just a *forward lookup zone* from your primary DNS server (see [this tutorial](tutorial-forward.html)). In this case, you do not need to make any changes to the slaves in the cluster.
