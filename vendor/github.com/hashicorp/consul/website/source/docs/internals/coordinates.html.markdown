---
layout: "docs"
page_title: "Network Coordinates"
sidebar_current: "docs-internals-coordinates"
description: |-
Serf uses a network tomography system to compute network coordinates for nodes in the cluster. These coordinates are useful for easily calculating the estimated network round trip time between any two nodes in the cluster. This page documents the details of this system. The core of the network tomography system us based on Vivaldi: A Decentralized Network Coordinate System, with several improvements based on several follow-on papers.
---

# Network Coordinates

Consul uses a [network tomography](https://en.wikipedia.org/wiki/Network_tomography)
system to compute network coordinates for nodes in the cluster. These coordinates
allow the network round trip time to be estimated between any two nodes using a
very simple calculation. This allows for many useful applications, such as finding
the service node nearest a requesting node, or failing over to services in the next
closest datacenter.

All of this is provided through the use of the [Serf library](https://www.serf.io/).
Serf's network tomography is based on ["Vivaldi: A Decentralized Network Coordinate System"](http://www.cs.ucsb.edu/~ravenben/classes/276/papers/vivaldi-sigcomm04.pdf),
with some enhancements based on other research. There are more details about
[Serf's network coordinates here](https://www.serf.io/docs/internals/coordinates.html).

~> **Advanced Topic!** This page covers the technical details of
the internals of Consul. You don't need to know these details to effectively
operate and use Consul. These details are documented here for those who wish
to learn about them without having to go spelunking through the source code.

## Network Coordinates in Consul

Network coordinates manifest in several ways inside Consul:

* The [`consul rtt`](/docs/commands/rtt.html) command can be used to query for the
  network round trip time between any two nodes.

* The [Catalog endpoints](/api/catalog.html) and
  [Health endpoints](/api/health.html) can sort the results of queries based
  on the network round trip time from a given node using a "?near=" parameter.

* [Prepared queries](/api/query.html) can automatically fail over services 
  to other Consul datacenters based on network round trip times.

* The [Coordinate endpoint](/api/coordinate.html) exposes raw network
  coordinates for use in other applications.

Consul uses Serf to manage two different gossip pools, one for the LAN with members
of a given datacenter, and one for the WAN which is made up of just the Consul servers
in all datacenters. It's important to note that **network coordinates are not compatible
between these two pools**. LAN coordinates only make sense in calculations with other
LAN coordinates, and WAN coordinates only make sense with other WAN coordinates.

## Working with Coordinates

Computing the estimated network round trip time between any two nodes is simple
once you have their coordinates. Here's a sample coordinate, as returned from the
[Coordinate endpoint](/api/coordinate.html).

```
    "Coord": {
        "Adjustment": 0.1,
        "Error": 1.5,
        "Height": 0.02,
        "Vec": [0.34,0.68,0.003,0.01,0.05,0.1,0.34,0.06]
    }
```

All values are floating point numbers in units of seconds, except for the error
term which isn't used for distance calculations.

Here's a complete example in Go showing how to compute the distance between two
coordinates:

```
import (
    "github.com/hashicorp/serf/coordinate"
    "math"
    "time"
)

func dist(a *coordinate.Coordinate, b *coordinate.Coordinate) time.Duration {
    // Coordinates will always have the same dimensionality, so this is
    // just a sanity check.
    if len(a.Vec) != len(b.Vec) {
        panic("dimensions aren't compatible")
    }

    // Calculate the Euclidean distance plus the heights.
    sumsq := 0.0
    for i := 0; i < len(a.Vec); i++ {
        diff := a.Vec[i] - b.Vec[i]
        sumsq += diff * diff
    }
    rtt := math.Sqrt(sumsq) + a.Height + b.Height

    // Apply the adjustment components, guarding against negatives.
    adjusted := rtt + a.Adjustment + b.Adjustment
    if adjusted > 0.0 {
        rtt = adjusted
    }

    // Go's times are natively nanoseconds, so we convert from seconds.
    const secondsToNanoseconds = 1.0e9
    return time.Duration(rtt * secondsToNanoseconds)
}
```
