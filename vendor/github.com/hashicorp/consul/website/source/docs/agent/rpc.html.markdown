---
layout: "docs"
page_title: "RPC"
sidebar_current: "docs-agent-rpc"
description: |-
  The Consul agent provides a complete RPC mechanism that can be used to control the agent programmatically. This RPC mechanism is the same one used by the CLI but can be used by other applications to easily leverage the power of Consul without directly embedding.
---

# RPC Protocol

~> The RPC Protocol is deprecated and support was removed in Consul
   0.8. Please use the [HTTP API](/api/index.html), which has
   support for all features of the RPC Protocol.

The Consul agent provides a complete RPC mechanism that can
be used to control the agent programmatically. This RPC
mechanism is the same one used by the CLI but can be
used by other applications to easily leverage the power
of Consul without directly embedding.

It is important to note that the RPC protocol does not support
all the same operations as the [HTTP API](/api/index.html).

## Implementation Details

The RPC protocol is implemented using [MsgPack](http://msgpack.org/)
over TCP. This choice was driven by the fact that all operating
systems support TCP, and MsgPack provides a fast serialization format
that is broadly available across languages.

All RPC requests have a request header, and some requests have
a request body. The request header looks like:

```javascript
{
  "Command": "Handshake",
  "Seq": 0
}
```

All responses have a response header, and some may contain
a response body. The response header looks like:

```javascript
{
  "Seq": 0,
  "Error": ""
}
```

The `Command` in the request is used to specify what command the server should
run, and the `Seq` is used to track the request. Responses are
tagged with the same `Seq` as the request. This allows for some
concurrency on the server side as requests are not purely FIFO.
Thus, the `Seq` value should not be re-used between commands.
All responses may be accompanied by an error.

Possible commands include:

* handshake - Initializes the connection and sets the version
* force-leave - Removes a failed node from the cluster
* join - Requests Consul join another node
* members-lan - Returns the list of LAN members
* members-wan - Returns the list of WAN members
* monitor - Starts streaming logs over the connection
* stop - Stops streaming logs
* leave - Instructs the Consul agent to perform a graceful leave and shutdown
* stats - Provides various debugging statistics
* reload - Triggers a configuration reload

Each command is documented below along with any request or
response body that is applicable.

### handshake

This command is used to initialize an RPC connection. As it informs
the server which version the client is using, handshake MUST be the
first command sent.

The request header must be followed by a handshake body, like:

```javascript
{
  "Version": 1
}
```

The body specifies the IPC version being used; however, only version
1 is currently supported. This is to ensure backwards compatibility
in the future.

There is no special response body, but the client should wait for the
response and check for an error.

### force-leave

This command is used to remove failed nodes from a cluster. It takes
the following body:

```javascript
{
  "Node": "failed-node-name"
}
```

There is no special response body.

### join

This command is used to join an existing cluster using one or more known nodes.
It takes the following body:

```javascript
{
  "Existing": [
    "192.168.0.1:6000",
    "192.168.0.2:6000"
  ],
  "WAN": false
}
```

The `Existing` nodes are each contacted, and `WAN` controls if we are adding a
WAN member or LAN member. LAN members are expected to be in the same datacenter
and should be accessible at relatively low latencies. WAN members are expected to
be operating in different datacenters with relatively high access latencies. It is
important that only agents running in "server" mode are able to join nodes over the
WAN.

The response contains both a header and body. The body looks like:

```javascript
{
  "Num": 2
}
```

'Num' indicates the number of nodes successfully joined.

### members-lan

This command is used to return all the known LAN members and associated
information. All agents will respond to this command.

There is no request body, but the response looks like:

```javascript
{
  "Members": [
    {
      "Name": "TestNode"
      "Addr": [127, 0, 0, 1],
      "Port": 5000,
      "Tags": {
        "role": "test"
      },
      "Status": "alive",
      "ProtocolMin": 0,
      "ProtocolMax": 3,
      "ProtocolCur": 2,
      "DelegateMin": 0,
      "DelegateMax": 1,
      "DelegateCur": 1,
    },
  ...
  ]
}
```

### members-wan

This command is used to return all the known WAN members and associated
information. Only agents in server mode will respond to this command.

There is no request body, and the response is the same as `members-lan`

### monitor

The monitor command subscribes the channel to log messages from the Agent.

The request looks like:

```javascript
{
  "LogLevel": "DEBUG"
}
```

This subscribes the client to all messages of at least DEBUG level.

The server will respond with a standard response header indicating if the monitor
was successful. If so, any future logs will be sent and tagged with
the same `Seq` as in the `monitor` request.

Assume we issued the previous monitor command with `"Seq": 50`. We may start
getting messages like:

```javascript
{
  "Seq": 50,
  "Error": ""
}

{
  "Log": "2013/12/03 13:06:53 [INFO] agent: Received event: member-join"
}
```

It is important to realize that these messages are sent asynchronously
and not in response to any command. If a client is streaming
commands, there may be logs streamed while a client is waiting for a
response to a command. This is why the `Seq` must be used to pair requests
with their corresponding responses.

The client can only be subscribed to at most a single monitor instance.
To stop streaming, the `stop` command is used.

### stop

This command stops a monitor.

The request looks like:

```javascript
{
  "Stop": 50
}
```

This unsubscribes the client from the monitor with `Seq` value of 50.

There is no response body.

### leave

This command is used to trigger a graceful leave and shutdown.
There is no request body or response body.

### stats

This command provides debug information. There is no request body, and the
response body looks like:

```javascript
{
  "agent": {
    "check_monitors": 0,
    ...
  },
  "consul: {
    "server": "true",
    ...
  },
  ...
}
```

### reload

This command is used to trigger a reload of configurations.
There is no request body or response body.
