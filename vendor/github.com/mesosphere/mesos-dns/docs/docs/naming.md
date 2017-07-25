---
title: Service Naming
---

# Service Naming

Mesos-DNS defines a DNS domain for Mesos tasks (default `.mesos`, see [instructions on configuration](configuration-parameters.html)). Running tasks can be discovered by looking up A and, optionally, SRV records within the Mesos domain. 

## A Records

An A record associates a hostname to an IP address.
For task `task` launched by framework `framework`, Mesos-DNS generates an A record for hostname `task.framework.domain` that provides one of the following:

- the IP address of the task's network container (provided by a Mesos containerizer); or
- the IP address of the specific slave running the task.

For example, other Mesos tasks can discover the IP address for service `search` launched by the `marathon` framework with a lookup for `search.marathon.mesos`:

``` console
$ dig search.marathon.mesos

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> search.marathon.mesos
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 24471
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 1, ADDITIONAL: 0

;; QUESTION SECTION:
;search.marathon.mesos.			IN	A

;; ANSWER SECTION:
search.marathon.mesos.		60	IN	A	10.9.87.94
```

If the following conditions are true...

- the Mesos-DNS IP-source configuration prioritizes container IPs; and
- the Mesos containerizer that launches the task provides a container IP `10.0.4.1` for the task `search.marathon.mesos`

...then the lookup would give:

``` console
$ dig search.marathon.mesos

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> search.marathon.mesos
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 24471
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 1, ADDITIONAL: 0

;; QUESTION SECTION:
;search.marathon.mesos.         IN  A

;; ANSWER SECTION:
search.marathon.mesos.      60  IN  A   10.0.4.1
```

In addition to the `task.framework.domain` semantics above Mesos-DNS always generates an A record `task.framework.slave.domain` that references the IP address(es) of the slave(s) upon which the task is running.
For example, a query of the A records for `search.marathon.slave.mesos` would yield the IP address of each slave running one or more instances of the `search` application on the `marathon` framework.

*Note*: Container IPs must be provided by the executor of a task in one of the following task status labels:

- `Docker.NetworkSettings.IPAddress`
- `MesosContainerizer.NetworkSettings.IPAddress`.

In general support for these will not be available before Mesos 0.24.
 
## SRV Records

An SRV record associates a service name to a hostname and an IP port.
For task `task` launched by framework `framework`, Mesos-DNS generates an SRV record for service name `_task._protocol.framework.domain`, where `protocol` is `udp` or `tcp`.
For example, other Mesos tasks can discover service `search` launched by the `marathon` framework with a lookup for lookup `_search._tcp.marathon.mesos`:

```console
$ dig _search._tcp.marathon.mesos SRV

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> _search._tcp.marathon.mesos SRV
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 33793
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;_search._tcp.marathon.mesos.	IN SRV

;; ANSWER SECTION:
_search._tcp.marathon.mesos.	60 IN SRV 0 0 31302 10.254.132.41.
``` 

Mesos-DNS supports the use of a task's DiscoveryInfo for SRV record generation.
If no DiscoveryInfo is available then Mesos-DNS will fall back to those "ports" resources allocated for the task.
The following table illustrates the rules that govern SRV generation:

|Service   	|CT-IP Avail  	|DI Avail   	|Target Host   	|Target Port   	|A (Target Resolution)	  |
|---		|---		|---		|---		|---		|---			  |
|_{task}._{proto}.framework.domain |no  | no  	|{task}.framework.slave.domain | host-port | slave-ip	  |
|				   |yes | no  	|{task}.framework.slave.domain | host-port | slave-ip	  |
|				   |no  | yes  	|{task}.framework.domain       | di-port   | slave-ip	  |
|				   |yes | yes  	|{task}.framework.domain       | di-port   | container-ip |
|_{task}._{proto}.framework.slave.domain |n/a | n/a |{task}.framework.slave.domain | host-port | slave-ip |

## Other Records

Mesos-DNS generates a few special records:
- for the leading master: A record (`leader.domain`) and SRV records (`_leader._tcp.domain` and `_leader._udp.domain`); and
- for all framework schedulers: A records (`{framework}.domain`) and SRV records (`_framework._tcp.{framework}.domain`)
- for every known Mesos master: A records (`master.domain`) and SRV records (`_master._tcp.domain` and `_master._udp.domain`); and
- for every known Mesos slave: A records (`slave.domain`) and SRV records (`_slave._tcp.domain`).

Note that, if you configure Mesos-DNS to detect the leading master through Zookeeper, then this is the only master it knows about.
If you configure Mesos-DNS using the `masters` field, it will generate master records for every master in the list.
Also note that there is inherent delay between the election of a new master and the update of leader/master records in Mesos-DNS. 

Mesos-DNS generates A records for itself that list all the IP addresses that Mesos-DNS is listening to. The name for Mesos-DNS can be selected using the `SOAMname` [configuration parameter](configuration-parameters.html). The default name is `ns1.mesos`.

In addition to A and SRV records for Mesos tasks, Mesos-DNS supports requests for SOA and NS records for the Mesos domain. DNS requests for records of other types in the Mesos domain will return `NXDOMAIN`. Mesos-DNS does not support PTR records needed for reverse lookups. 

## Notes

If a framework launches multiple tasks with the same name, the DNS lookup will return multiple records, one per task. Mesos-DNS randomly shuffles the order of records to provide rudimentary load balancing between these tasks. 

Mesos-DNS follows [RFC 952](https://tools.ietf.org/html/rfc952) for name formatting. All fields used to construct hostnames for A records and service names for SRV records must be up to 24 characters and drawn from the alphabet (A-Z), digits (0-9) and minus sign (-). No distinction is made between upper and lower case. If the task name does not comply with these constraints, Mesos-DNS will trim it, remove all invalid characters, and replace period (.) with sign (-) for task names. For framework names, we allow period (.) but all other constraints apply.  For example, a task named `apiserver.myservice` launch by framework `marathon.prod`, will have A records associated with the name `apiserver-myservice.marathon.prod.mesos` and SRV records associated with name `_apiserver-myservice._tcp.marathon.prod.mesos`. 

Some frameworks register with longer, less friendly names. For example, earlier versions of marathon may register with names like `marathon-0.7.5`, which will lead to names like `search.marathon-0.7.5.mesos`. Make sure your framework registers with the desired name. For instance, you can launch marathon with ` --framework_name marathon` to get the framework registered as `marathon`.  



