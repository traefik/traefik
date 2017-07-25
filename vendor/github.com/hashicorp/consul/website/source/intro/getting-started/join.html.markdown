---
layout: "intro"
page_title: "Consul Cluster"
sidebar_current: "gettingstarted-join"
description: >
  When a Consul agent is started, it begins as an isolated cluster of its own.
  To learn about other cluster members, the agent must join one or more other
  nodes using a provided join address. In this step, we will set up a two-node
  cluster and join the nodes together.
---

# Consul Cluster

We've started our first agent and registered and queried a service on that
agent. This showed how easy it is to use Consul but didn't show how this could
be extended to a scalable, production-grade service discovery infrastructure.
In this step, we'll create our first real cluster with multiple members.

When a Consul agent is started, it begins without knowledge of any other node:
it is an isolated cluster of one. To learn about other cluster members, the
agent must _join_ an existing cluster. To join an existing cluster, it only
needs to know about a _single_ existing member. After it joins, the agent will
gossip with this member and quickly discover the other members in the cluster.
A Consul agent can join any other agent, not just agents in server mode.

## Starting the Agents

To simulate a more realistic cluster, we will start a two node cluster via
[Vagrant](https://www.vagrantup.com/). The Vagrantfile we will be using can
be found in the [demo section of the Consul repo]
(https://github.com/hashicorp/consul/tree/master/demo/vagrant-cluster).

We first boot our two nodes:

```text
$ vagrant up
```

Once the systems are available, we can ssh into them to begin configuration
of our cluster. We start by logging in to the first node:

```text
$ vagrant ssh n1
```

In our previous examples, we used the [`-dev`
flag](/docs/agent/options.html#_dev) to quickly set up a development server.
However, this is not sufficient for use in a clustered environment. We will
omit the `-dev` flag from here on, and instead specify our clustering flags as
outlined below.

Each node in a cluster must have a unique name. By default, Consul uses the
hostname of the machine, but we'll manually override it using the [`-node`
command-line option](/docs/agent/options.html#_node).

We will also specify a [`bind` address](/docs/agent/options.html#_bind):
this is the address that Consul listens on, and it *must* be accessible by
all other nodes in the cluster. While a `bind` address is not strictly
necessary, it's always best to provide one. Consul will by default attempt to
listen on all IPv4 interfaces on a system, but will fail to start with an
error if multiple private IPs are found. Since production servers often
have multiple interfaces, specifying a `bind` address assures that you will
never bind Consul to the wrong interface.

The first node will act as our sole server in this cluster, and we indicate
this with the [`server` switch](/docs/agent/options.html#_server).

The [`-bootstrap-expect` flag](/docs/agent/options.html#_bootstrap_expect)
hints to the Consul server the number of additional server nodes we are
expecting to join. The purpose of this flag is to delay the bootstrapping of
the replicated log until the expected number of servers has successfully joined.
You can read more about this in the [bootstrapping
guide](/docs/guides/bootstrapping.html).

Finally, we add the [`config-dir` flag](/docs/agent/options.html#_config_dir),
marking where service and check definitions can be found.

All together, these settings yield a
[`consul agent`](/docs/commands/agent.html) command like this:

```text
vagrant@n1:~$ consul agent -server -bootstrap-expect=1 \
	-data-dir=/tmp/consul -node=agent-one -bind=172.20.20.10 \
	-config-dir=/etc/consul.d
...
```

Now, in another terminal, we will connect to the second node:

```text
$ vagrant ssh n2
```

This time, we set the [`bind` address](/docs/agent/options.html#_bind)
address to match the IP of the second node as specified in the Vagrantfile
and the [`node` name](/docs/agent/options.html#_node) to be `agent-two`.
Since this node will not be a Consul server, we don't provide a
[`server` switch](/docs/agent/options.html#_server).

All together, these settings yield a
[`consul agent`](/docs/commands/agent.html) command like this:

```text
vagrant@n2:~$ consul agent -data-dir=/tmp/consul -node=agent-two \
	-bind=172.20.20.11 -config-dir=/etc/consul.d
...
```

At this point, you have two Consul agents running: one server and one client.
The two Consul agents still don't know anything about each other and are each
part of their own single-node clusters. You can verify this by running
[`consul members`](/docs/commands/members.html) against each agent and noting
that only one member is visible to each agent.

## Joining a Cluster

Now, we'll tell the first agent to join the second agent by running
the following commands in a new terminal:

```text
$ vagrant ssh n1
...
vagrant@n1:~$ consul join 172.20.20.11
Successfully joined cluster by contacting 1 nodes.
```

You should see some log output in each of the agent logs. If you read
carefully, you'll see that they received join information. If you
run [`consul members`](/docs/commands/members.html) against each agent,
you'll see that both agents now know about each other:

```text
vagrant@n2:~$ consul members
Node       Address            Status  Type    Build  Protocol
agent-two  172.20.20.11:8301  alive   client  0.5.0  2
agent-one  172.20.20.10:8301  alive   server  0.5.0  2
```

-> **Remember:** To join a cluster, a Consul agent only needs to
learn about <em>one existing member</em>. After joining the cluster, the
agents gossip with each other to propagate full membership information.

## Auto-joining a Cluster on Start
Ideally, whenever a new node is brought up in your datacenter, it should automatically join the Consul cluster without human intervention. Consul facilitates auto-join by enabling the auto-discovery of instances in AWS or Google Cloud with a given tag key/value. To use the integration, add the [`retry_join_ec2`](/docs/agent/options.html?#retry_join_ec2) or the [`retry_join_gce`](/docs/agent/options.html?#retry_join_gce) nested object to your Consul configuration file. This will allow a new node to join the cluster without any hardcoded configuration. Alternatively, you can join a cluster at startup using the [`-join` flag](/docs/agent/options.html#_join) or [`start_join` setting](/docs/agent/options.html#start_join) with hardcoded addresses of other known Consul agents.

## Querying Nodes

Just like querying services, Consul has an API for querying the
nodes themselves. You can do this via the DNS or HTTP API.

For the DNS API, the structure of the names is `NAME.node.consul` or
`NAME.node.DATACENTER.consul`. If the datacenter is omitted, Consul
will only search the local datacenter.

For example, from "agent-one", we can query for the address of the
node "agent-two":

```
vagrant@n1:~$ dig @127.0.0.1 -p 8600 agent-two.node.consul
...

;; QUESTION SECTION:
;agent-two.node.consul.	IN	A

;; ANSWER SECTION:
agent-two.node.consul.	0 IN	A	172.20.20.11
```

The ability to look up nodes in addition to services is incredibly
useful for system administration tasks. For example, knowing the address
of the node to SSH into is as easy as making the node a part of the
Consul cluster and querying it.

## Leaving a Cluster

To leave the cluster, you can either gracefully quit an agent (using
`Ctrl-C`) or force kill one of the agents. Gracefully leaving allows
the node to transition into the _left_ state; otherwise, other nodes
will detect it as having _failed_. The difference is covered
in more detail [here](/intro/getting-started/agent.html#stopping).

## Next Steps

We now have a multi-node Consul cluster up and running. Let's make
our services more robust by giving them [health checks](checks.html)!
