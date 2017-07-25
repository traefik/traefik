---
title: Mesos-DNS and Bind
---

## Using Mesos-DNS as a Forward Lookup Zone with Bind

In many enterprise environments, it is difficult to change the settings for the first nameserver in every machine in the Mesos cluster and make it point to a Mesos-DNS server. You may also want to have machines outside of the Mesos cluster discover Mesos tasks using Mesos-DNS. To accomodate these cases, you can set up a *forward lookup zone* in your existing DNS server that forwards all requests in the Mesos domain to Mesos-DNS. 

This tutorial reviews how to set up a forward lookup zone for the Mesos domain with the [Bind9 DNS server](http://www.bind9.net). All DNS requests will first be received by the Bind9 server. Requests in the Mesos domain will be forwarded to Mesos-DNS. 
 
 
### Step 1: Bind9 Configuration

Assuming that Mesos-DNS will run on host `192.168.0.100` and uses port `8053`, you can configure configure a Bind9 server to forward Mesos requests to Mesos-DNS but adding the following information in the `/etc/bind/named.conf.local` file:

```
zone "mesos" {
  type forward;
  forward only;
  forwarders { 192.168.0.100 port 8053; };
}; 
```

Note that the zone name should match the domain name you selected for the Mesos cluster using the `domain` configuration parameter for Mesos-DNS. 

You will probably need to restart your Bind9 server using a command like `sudo service bind9 restart` or `sudo /etc/init.d/bind9 restart` depending on the system. Assuming that all other machines already point to the Bind9 server as their primary nameserver, no changes are needed in any other machines. 


### Step 2: Mesos-DNS Configuration

When you launch Mesos-DNS on host `192.168.0.100`, make sure that the following parameters are set in its configuration file:

``` 
  "externalon": false,
  "port": 8053,
```

Setting `externalon` to `false` instructs Mesos-DNS to refuse any requests outside of the Mesos domain. You can also skip setting the `resolvers` configuration parameter. 

Since we now use port `8053` instead of the standard but privileged port `53`, we can run Mesos-DNS as a non-root user. This is good for security. No DNS client will be affected by the port choice since all DNS requests are first send to the Bind9 server. 

### Notes 

If you are running multiple Mesos-DNS server for high availability, you can edit the Bind9 configuration to specify multiple fowarders for the Mesos zone. 

Since Mesos-DNS receives requests only through the Bind9 server(s), you can limit it or firewall it to decline any connections from any other machines within the enterprise network or across the Internet. This is a good measure for security. 





