---
title: Mesos-DNS FAQ
---

##  Frequently Asked Questions & Troubleshooting

---

#### Mesos-DNS version

You can check the Mesos-DNS version by executing `mesos-dns -version`. 


---

#### SOA record customization

You can customize all fields in the SOA records for the Mesos domain. See the `SOA*` [configuration parameters](configuration-parameters.html).

---

#### Verbose and very verbose modes

If you start Mesos-DNS in verbose mode using the `-v=1` or `-v=2` arguments, it  prints a variety of messages that are useful for debugging and performance tuning. The `-v=2` option will periodically print every A or SRV record Mesos-DNS generates. In large clusters, this can overwhelm log management tools, such as [systemd-journald](http://www.freedesktop.org/software/systemd/man/systemd-journald.service.html), and they will drop messages. This is not a viable method to enumerate the state of a Mesos-DNS server.

---


#### Mesos-DNS fails to launch

Make sure that the port used for Mesos-DNS is available and not in use by another process. To use the recommended port `53`, you must start Mesos-DNS as root.

Alternatively, if you have an operating system and file system that supports [capabilities](http://manpages.ubuntu.com/manpages/hardy/man7/capabilities.7.html), you can run `sudo setcap 'cap_net_bind_service=+ep' mesos-dns`, which will then allow Mesos-DNS to bind to privileged ports, without requiring it to run as root.

---

#### Slaves cannot connect to Mesos-DNS

Make sure that port `53` is not blocked by a firewall rule on your cluster. For example, [Google Cloud Platform](https://cloud.google.com/) blocks port `53` by default. If you use the `zk` field, you should also check if the Zookeeper port is not blocked either. Finally, if you use the HTTP interface for Mesos-DNS, make sure that the `httpport` is open. 

Check the `/etc/resolv.conf` file. If multiple nameservers are listed and Mesos-DNS is not the first one, the slave will first connect to the other name servers. If `options rotate` is used and one of the listed nameservers is not Mesos-DNS, then you will get intermittent failures.

---

#### Mesos-DNS does not resolve names in the Mesos domain

Check the configuration file to make sure that Mesos-DNS is directed to the right Zookeeper or master(s) for the Mesos cluster (`masters`). 
 
---

#### Mesos-DNS does not resolve names outside of the Mesos domain

Check the configuration file to make sure that Mesos-DNS is configured with the IP address of  external DNS servers (`resolvers`).

---

#### Updating the configuration file

When you update the configuration file, you need to restart Mesos-DNS. No state is lost long-term on restart as Mesos-DNS is stateless and retrieves task state from the Mesos master(s). There is a short inconsistency window where records may be missing while the zone is being generated from the state.json.

---

### DNS names are not user-friendly

Some frameworks register with longer, less user-friendly names. For example, earlier versions of marathon may register with names like `marathon-0.7.5`, which will lead to names like `search.marathon-0.7.5.mesos`. Make sure your framework registers with the desired name. For instance, you can launch marathon with ` --framework_name marathon` to get the framework registered as `marathon`.  

---

### Mesos-DNS is not communicating with the Mesos Master using the hosts configured in the 'masters' field

If the `zk` field is defined, Mesos-DNS will ignore the `masters` field. It will contact Zookeeper to detect the leading Mesos master. If the `zk` field is not used, Mesos-DNS uses the `masters` field in the configuration file only for the initial requests to the Mesos master. The initial request for task state also return information about the current masters. This information is used for subsequent task state request. If you launch Mesos-DNS in verbose mode using `-v=2 `, there will be a period stdout message that identifies which master Mesos-DNS is contacting at the moment. 

