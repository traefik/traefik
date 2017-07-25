---
title: Mesos-DNS Performance Tuning
---

## Mesos-DNS Performance Tuning

Mesos-DNS is light-weight and can achieve high throughput with a small amount of resources. For example, our benchmarking on [Google Cloud Platform](https://cloud.google.com/) shows that, using a single core of an `n1-standard-2` instance, it can serve up to 8.5K queries/second for lookups in the Mesos cluster domain. For many use cases, this should be sufficiently performant. 

The following are suggestions for further performance tuning, focusing primarily on Linux systems. 

### Basic Tuning

#### GOMAXPROCS
In Mesos-DNS 0.5.0, we've moved to Go *1.5*, which automatically sets `GOMAXPROCS` to the number of CPUs that are detected by the Go runtime via `runtime.NumCPU`.

In previous versions, it may make sense to increase the number of cores used by Mesos-DNS by setting the `GOMAXPROCS` environment variable. For instance, use the following command line to launch Mesos-DNS on 8 cores:

````
GOMAXPROCS=8 mesos-dns
````

#### Scaling out
If a single Mesos-DNS server cannot meet the performance requirements in a very large cluster, you can bring up multiple Mesos-DNS servers and configure a subset of the slaves to use each Mesos-DNS server. All Mesos-DNS servers will serve the same settings derived from the Mesos master state. 


#### Fundamental Operating system limits
You should also increase the shell limits for the maximum number of file descriptors and processes. Use `ulimit -a` to check the current settings and, if needed, increase them by executing the following shell commands before launching Mesos-DNS. Mesos-DNS uses file descriptors for forwarding queries, making requests to the Mesos master, and keeps FDs open for accepting queries.

```
ulimit -n 65536
ulimit -p 16384

```
### Tuning External Lookups

When Mesos-DNS receives lookup requests for a name outside the Mesos domain, it forwards them to an external nameserver. Depending on the external nameserver used and the network conditions, external requests can have significant latency and can pose a performance bottleneck. 

Make sure you use a performant and reliable external nameserver. Its performance places an upper bound on Mesos-DNS for external requests. The [Google public DNS](https://developers.google.com/speed/public-dns/) is a good option but keep in mind that it enforces [request rate limits](https://developers.google.com/speed/public-dns/docs/security) for security reasons. You can also try other public nameservers from this [list](http://public-dns.tk/nameservers). 

Increasing the shell limits can also be important for the performance of external requests. 

### Tuning Mesos Slaves

You can improve the performance of Mesos slaves by increasing shell limits as discussed above. 

You can also install a DNS cache on each slave, such as [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) (see [instructions](http://www.g-loaded.eu/2010/09/18/caching-nameserver-using-dnsmasq/)). This will limit the number of repeated requests reaching Mesos-DNS. 

