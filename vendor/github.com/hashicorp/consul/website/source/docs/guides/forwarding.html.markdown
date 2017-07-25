---
layout: "docs"
page_title: "Forwarding"
sidebar_current: "docs-guides-forwarding"
description: |-
  By default, DNS is served from port 53. On most operating systems, this requires elevated privileges. Instead of running Consul with an administrative or root account, it is possible to instead forward appropriate queries to Consul, running on an unprivileged port, from another DNS server or port redirect.
---

# Forwarding DNS

By default, DNS is served from port 53. On most operating systems, this
requires elevated privileges. Instead of running Consul with an administrative
or root account, it is possible to instead forward appropriate queries to Consul,
running on an unprivileged port, from another DNS server or port redirect.

In this guide, we will demonstrate forwarding from
[BIND](https://www.isc.org/downloads/bind/) as well as
[dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) and
[iptables](http://www.netfilter.org/). For the sake of simplicity, BIND
and Consul are running on the same machine in this example. For iptables
the rules must be set on the same host as the Consul instance and relay
hosts should not be on the same host or the redirects will intercept the
traffic.

It is worth mentioning that, by default, Consul does not resolve DNS
records outside the `.consul.` zone unless the
[recursors](/docs/agent/options.html#recursors) configuration option
has been set. As an example of how this changes Consul's behavior,
suppose a Consul DNS reply includes a CNAME record pointing outside
the `.consul` TLD. The DNS reply will only include CNAME records by
default. By contrast, when `recursors` is set and the upstream resolver is
functioning correctly, Consul will try to resolve CNAMEs and include
any records (e.g. A, AAAA, PTR) for them in its DNS reply.

You can either do one of the following:

### BIND Setup

First, you have to disable DNSSEC so that Consul and BIND can communicate.
Here is an example of such a configuration:

```text
options {
  listen-on port 53 { 127.0.0.1; };
  listen-on-v6 port 53 { ::1; };
  directory       "/var/named";
  dump-file       "/var/named/data/cache_dump.db";
  statistics-file "/var/named/data/named_stats.txt";
  memstatistics-file "/var/named/data/named_mem_stats.txt";
  allow-query     { localhost; };
  recursion yes;

  dnssec-enable no;
  dnssec-validation no;

  /* Path to ISC DLV key */
  bindkeys-file "/etc/named.iscdlv.key";

  managed-keys-directory "/var/named/dynamic";
};

include "/etc/named/consul.conf";
```

### Zone File

Then we set up a zone for our Consul managed records in `consul.conf`:

```text
zone "consul" IN {
  type forward;
  forward only;
  forwarders { 127.0.0.1 port 8600; };
};
```

Here we assume Consul is running with default settings and is serving
DNS on port 8600.

### Dnsmasq Setup

Dnsmasq is typically configured via a `dnsmasq.conf` or a series of files in
the `/etc/dnsmasq.d` directory. In Dnsmasq's configuration file
(e.g. `/etc/dnsmasq.d/10-consul`), add the following:

```text
# Enable forward lookup of the 'consul' domain:
server=/consul/127.0.0.1#8600

# Uncomment and modify as appropriate to enable reverse DNS lookups for
# common netblocks found in RFC 1918, 5735, and 6598:
#rev-server=0.0.0.0/8,127.0.0.1#8600
#rev-server=10.0.0.0/8,127.0.0.1#8600
#rev-server=100.64.0.0/10,127.0.0.1#8600
#rev-server=127.0.0.1/8,127.0.0.1#8600
#rev-server=169.254.0.0/16,127.0.0.1#8600
#rev-server=172.16.0.0/12,127.0.0.1#8600
#rev-server=192.168.0.0/16,127.0.0.1#8600
#rev-server=224.0.0.0/4,127.0.0.1#8600
#rev-server=240.0.0.0/4,127.0.0.1#8600
```

Once that configuration is created, restart the `dnsmasq` service.

Additional useful settings in `dnsmasq` to consider include (see
[`dnsmasq(8)`](http://www.thekelleys.org.uk/dnsmasq/docs/dnsmasq-man.html)
for additional details):

```
# Accept DNS queries only from hosts whose address is on a local subnet.
#local-service

# Don't poll /etc/resolv.conf for changes.
#no-poll

# Don't read /etc/resolv.conf. Get upstream servers only from the command
# line or the dnsmasq configuration file (see the "server" directive below).
#no-resolv

# Specify IP address(es) of other DNS servers for queries not handled
# directly by consul. There is normally one 'server' entry set for every
# 'nameserver' parameter found in '/etc/resolv.conf'. See dnsmasq(8)'s
# 'server' configuration option for details.
#server=1.2.3.4
#server=208.67.222.222
#server=8.8.8.8

# Set the size of dnsmasq's cache. The default is 150 names. Setting the
# cache size to zero disables caching.
#cache-size=65536
```

### iptables Setup

On Linux systems that support it, incoming requests and requests to
the local host can use `iptables` to forward ports on the same machine
without a secondary service. Since Consul, by default, only resolves
the `.consul` TLD, it is especially important to use the `recursors`
option if you wish the `iptables` setup to resolve for other domains.
The recursors should not include the local host as the redirects would
just intercept the requests.

The iptables method is suited for situations where an external DNS
service is already running in your infrastructure and is used as the
recursor or if you want to use an existing DNS server as your query
endpoint and forward requests for the consul domain to the Consul
server. In both of those cases you may want to query the Consul server
but not need the overhead of a separate service on the Consul host.

```
[root@localhost ~]# iptables -t nat -A PREROUTING -p udp -m udp --dport 53 -j REDIRECT --to-ports 8600
[root@localhost ~]# iptables -t nat -A PREROUTING -p tcp -m tcp --dport 53 -j REDIRECT --to-ports 8600
[root@localhost ~]# iptables -t nat -A OUTPUT -d localhost -p udp -m udp --dport 53 -j REDIRECT --to-ports 8600
[root@localhost ~]# iptables -t nat -A OUTPUT -d localhost -p tcp -m tcp --dport 53 -j REDIRECT --to-ports 8600
```

### Testing

First, perform a DNS query against Consul directly to be sure that the record exists:

```text
[root@localhost ~]# dig @localhost -p 8600 primary.redis.service.dc-1.consul. A

; <<>> DiG 9.8.2rc1-RedHat-9.8.2-0.23.rc1.32.amzn1 <<>> @localhost primary.redis.service.dc-1.consul. A
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 11536
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;primary.redis.service.dc-1.consul. IN A

;; ANSWER SECTION:
primary.redis.service.dc-1.consul. 0 IN A 172.31.3.234

;; Query time: 4 msec
;; SERVER: 127.0.0.1#53(127.0.0.1)
;; WHEN: Wed Apr  9 17:36:12 2014
;; MSG SIZE  rcvd: 76
```

Then run the same query against your BIND instance and make sure you get a
valid result:

```text
[root@localhost ~]# dig @localhost -p 53 primary.redis.service.dc-1.consul. A

; <<>> DiG 9.8.2rc1-RedHat-9.8.2-0.23.rc1.32.amzn1 <<>> @localhost primary.redis.service.dc-1.consul. A
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 11536
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;primary.redis.service.dc-1.consul. IN A

;; ANSWER SECTION:
primary.redis.service.dc-1.consul. 0 IN A 172.31.3.234

;; Query time: 4 msec
;; SERVER: 127.0.0.1#53(127.0.0.1)
;; WHEN: Wed Apr  9 17:36:12 2014
;; MSG SIZE  rcvd: 76
```

If desired, verify reverse DNS using the same methodology:

```text
[root@localhost ~]# dig @127.0.0.1 -p 8600 133.139.16.172.in-addr.arpa. PTR

; <<>> DiG 9.10.3-P3 <<>> @127.0.0.1 -p 8600 133.139.16.172.in-addr.arpa. PTR
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 3713
;; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;133.139.16.172.in-addr.arpa.   IN  PTR

;; ANSWER SECTION:
133.139.16.172.in-addr.arpa. 0  IN  PTR consul1.node.dc1.consul.

;; Query time: 3 msec
;; SERVER: 127.0.0.1#8600(127.0.0.1)
;; WHEN: Sun Jan 31 04:25:39 UTC 2016
;; MSG SIZE  rcvd: 109
[root@localhost ~]# dig @127.0.0.1 +short -x 172.16.139.133
consul1.node.dc1.consul.
```

### Troubleshooting

If you don't get an answer from your DNS server (e.g. BIND, Dnsmasq) but you
do get an answer from Consul, your best bet is to turn on your DNS server's
query log to see what's happening.

For BIND:

```text
[root@localhost ~]# rndc querylog
[root@localhost ~]# tail -f /var/log/messages
```

The log may show errors like this:

```text
error (no valid RRSIG) resolving
error (no valid DS) resolving
```

This indicates that DNSSEC is not disabled properly.

If you see errors about network connections, verify that there are no firewall
or routing problems between the servers running BIND and Consul.

For Dnsmasq, see the `log-queries` configuration option and the `USR1`
signal.
