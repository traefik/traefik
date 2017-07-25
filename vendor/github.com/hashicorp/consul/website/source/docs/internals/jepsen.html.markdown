---
layout: "docs"
page_title: "Jepsen Testing"
sidebar_current: "docs-internals-jepsen"
description: |-
  Jepsen is a tool, written by Kyle Kingsbury, designed to test the partition tolerance of distributed systems. It creates network partitions while fuzzing the system with random operations. The results are analyzed to see if the system violates any of the consistency properties it claims to have.
---

# Jepsen Testing

[Jepsen](http://aphyr.com/posts/281-call-me-maybe-carly-rae-jepsen-and-the-perils-of-network-partitions)
is a tool, written by Kyle Kingsbury, designed to test the partition
tolerance of distributed systems. It creates network partitions while fuzzing
the system with random operations. The results are analyzed to see if the system
violates any of the consistency properties it claims to have.

As part of our Consul testing, we ran a Jepsen test to determine if
any consistency issues could be uncovered. In our testing, Consul
gracefully recovered from partitions without introducing any consistency
issues.

## Running the tests

At the moment, testing with Jepsen is rather complex as it requires
setting up multiple virtual machines, SSH keys, DNS configuration,
and a working Clojure environment. We hope to contribute our Consul
testing code upstream and to provide a Vagrant environment for Jepsen
testing soon.

## Output

Below is the output captured from Jepsen. We ran Jepsen multiple times,
and it passed each time. This output is only representative of a single
run.

<!--googleoff: all-->

```text
$ lein test :only jepsen.system.consul-test

lein test jepsen.system.consul-test
INFO  jepsen.os.debian - :n5 setting up debian
INFO  jepsen.os.debian - :n3 setting up debian
INFO  jepsen.os.debian - :n4 setting up debian
INFO  jepsen.os.debian - :n1 setting up debian
INFO  jepsen.os.debian - :n2 setting up debian
INFO  jepsen.os.debian - :n4 debian set up
INFO  jepsen.os.debian - :n5 debian set up
INFO  jepsen.os.debian - :n3 debian set up
INFO  jepsen.os.debian - :n1 debian set up
INFO  jepsen.os.debian - :n2 debian set up
INFO  jepsen.system.consul - :n1 consul nuked
INFO  jepsen.system.consul - :n4 consul nuked
INFO  jepsen.system.consul - :n5 consul nuked
INFO  jepsen.system.consul - :n3 consul nuked
INFO  jepsen.system.consul - :n2 consul nuked
INFO  jepsen.system.consul - Running nodes: {:n1 false, :n2 false, :n3 false, :n4 false, :n5 false}
INFO  jepsen.system.consul - :n2 consul nuked
INFO  jepsen.system.consul - :n3 consul nuked
INFO  jepsen.system.consul - :n4 consul nuked
INFO  jepsen.system.consul - :n5 consul nuked
INFO  jepsen.system.consul - :n1 consul nuked
INFO  jepsen.system.consul - :n1 starting consul
INFO  jepsen.system.consul - :n2 starting consul
INFO  jepsen.system.consul - :n4 starting consul
INFO  jepsen.system.consul - :n5 starting consul
INFO  jepsen.system.consul - :n3 starting consul
INFO  jepsen.system.consul - :n3 consul ready
INFO  jepsen.system.consul - :n2 consul ready
INFO  jepsen.system.consul - Running nodes: {:n1 true, :n2 true, :n3 true, :n4 true, :n5 true}
INFO  jepsen.system.consul - :n5 consul ready
INFO  jepsen.system.consul - :n1 consul ready
INFO  jepsen.system.consul - :n4 consul ready
INFO  jepsen.core - Worker 0 starting
INFO  jepsen.core - Worker 2 starting
INFO  jepsen.core - Worker 1 starting
INFO  jepsen.core - Worker 3 starting
INFO  jepsen.core - Worker 4 starting
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[4 4]
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 4	:invoke	:cas	[4 0]
INFO  jepsen.util - 2	:ok	:read	nil
INFO  jepsen.util - 4	:fail	:cas	[4 0]
INFO  jepsen.util - 1	:ok	:write	1
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 3	:fail	:cas	[4 4]
INFO  jepsen.util - 2	:invoke	:cas	[0 3]
INFO  jepsen.util - 2	:fail	:cas	[0 3]
INFO  jepsen.util - 4	:invoke	:cas	[4 4]
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 0	:invoke	:cas	[3 1]
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 4	:fail	:cas	[4 4]
INFO  jepsen.util - 0	:fail	:cas	[3 1]
INFO  jepsen.util - 1	:ok	:write	3
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[1 0]
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 0	:ok	:write	1
INFO  jepsen.util - 1	:fail	:cas	[1 0]
INFO  jepsen.util - 2	:invoke	:cas	[0 2]
INFO  jepsen.util - 2	:fail	:cas	[0 2]
INFO  jepsen.util - 4	:invoke	:cas	[1 2]
INFO  jepsen.util - 4	:fail	:cas	[1 2]
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:write	1
INFO  jepsen.util - 1	:ok	:read	1
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[2 4]
INFO  jepsen.util - 4	:fail	:cas	[2 4]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 1	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[4 2]
INFO  jepsen.util - 2	:fail	:cas	[4 2]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[2 4]
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 0	:fail	:cas	[2 4]
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:cas	[0 3]
INFO  jepsen.util - 2	:fail	:cas	[0 3]
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 3	:invoke	:cas	[0 2]
INFO  jepsen.util - 1	:invoke	:cas	[0 0]
INFO  jepsen.util - 0	:ok	:write	1
INFO  jepsen.util - 3	:fail	:cas	[0 2]
INFO  jepsen.util - 1	:fail	:cas	[0 0]
INFO  jepsen.util - 2	:invoke	:cas	[1 3]
INFO  jepsen.util - 2	:fail	:cas	[1 3]
INFO  jepsen.util - 4	:invoke	:cas	[1 0]
INFO  jepsen.util - 4	:fail	:cas	[1 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 2]
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 3	:fail	:cas	[2 2]
INFO  jepsen.util - 0	:ok	:read	1
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 4	:invoke	:cas	[1 2]
INFO  jepsen.util - 4	:fail	:cas	[1 2]
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 0	:invoke	:cas	[1 0]
INFO  jepsen.util - 1	:invoke	:cas	[0 1]
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 1	:fail	:cas	[0 1]
INFO  jepsen.util - 0	:fail	:cas	[1 0]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	0
INFO  jepsen.util - 3	:invoke	:cas	[3 3]
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 1	:invoke	:cas	[0 0]
INFO  jepsen.util - 3	:fail	:cas	[3 3]
INFO  jepsen.util - 1	:fail	:cas	[0 0]
INFO  jepsen.util - 0	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:cas	[0 0]
INFO  jepsen.util - 2	:fail	:cas	[0 0]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 3	:invoke	:cas	[3 0]
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 0	:invoke	:cas	[0 0]
INFO  jepsen.util - 3	:fail	:cas	[3 0]
INFO  jepsen.util - 0	:fail	:cas	[0 0]
INFO  jepsen.util - 1	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:cas	[0 0]
INFO  jepsen.util - 4	:fail	:cas	[0 0]
INFO  jepsen.util - 3	:invoke	:cas	[3 4]
INFO  jepsen.util - 0	:invoke	:cas	[4 3]
INFO  jepsen.util - 3	:fail	:cas	[3 4]
INFO  jepsen.util - 0	:fail	:cas	[4 3]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	0
INFO  jepsen.util - 2	:invoke	:cas	[1 1]
INFO  jepsen.util - 2	:fail	:cas	[1 1]
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 3	:invoke	:cas	[1 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 3	:fail	:cas	[1 0]
INFO  jepsen.util - 1	:invoke	:cas	[0 3]
INFO  jepsen.util - 1	:fail	:cas	[0 3]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[1 1]
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 1	:fail	:cas	[1 1]
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 0	:ok	:read	3
INFO  jepsen.util - 3	:ok	:write	4
INFO  jepsen.util - 1	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 0	:ok	:write	1
INFO  jepsen.util - 3	:ok	:read	1
INFO  jepsen.util - 1	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:cas	[0 3]
INFO  jepsen.util - 4	:fail	:cas	[0 3]
INFO  jepsen.util - 0	:invoke	:cas	[2 0]
INFO  jepsen.util - :nemesis	:info	:start	nil
INFO  jepsen.util - 3	:invoke	:cas	[0 1]
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 0	:fail	:cas	[2 0]
INFO  jepsen.util - 3	:fail	:cas	[0 1]
INFO  jepsen.util - 1	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:cas	[3 0]
INFO  jepsen.util - 2	:fail	:cas	[3 0]
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 1	:invoke	:cas	[1 0]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - :nemesis	:info	:start	"Cut off {:n5 #{:n3 :n1}, :n2 #{:n3 :n1}, :n4 #{:n3 :n1}, :n1 #{:n4 :n2 :n5}, :n3 #{:n4 :n2 :n5}}"
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:write	1
INFO  jepsen.util - 2	:fail	:write	1
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:ok	:write	1
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:fail	:write	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 0	:invoke	:write	2
INFO  jepsen.util - 0	:fail	:write	2
INFO  jepsen.util - 2	:invoke	:cas	[4 4]
INFO  jepsen.util - 2	:fail	:cas	[4 4]
INFO  jepsen.util - 4	:invoke	:cas	[3 0]
INFO  jepsen.util - 4	:fail	:cas	[3 0]
INFO  jepsen.util - 0	:invoke	:cas	[4 3]
INFO  jepsen.util - 0	:fail	:cas	[4 3]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:cas	[1 3]
INFO  jepsen.util - 4	:fail	:cas	[1 3]
INFO  jepsen.util - 0	:invoke	:cas	[3 0]
INFO  jepsen.util - 0	:fail	:cas	[3 0]
INFO  jepsen.util - 2	:invoke	:cas	[1 1]
INFO  jepsen.util - 2	:fail	:cas	[1 1]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 0	:invoke	:cas	[1 4]
INFO  jepsen.util - 0	:fail	:cas	[1 4]
INFO  jepsen.util - 2	:invoke	:cas	[2 2]
INFO  jepsen.util - 2	:fail	:cas	[2 2]
INFO  jepsen.util - 4	:invoke	:cas	[2 0]
INFO  jepsen.util - 4	:fail	:cas	[2 0]
INFO  jepsen.util - 0	:invoke	:cas	[0 4]
INFO  jepsen.util - 0	:fail	:cas	[0 4]
INFO  jepsen.util - 2	:invoke	:cas	[1 3]
INFO  jepsen.util - 2	:fail	:cas	[1 3]
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:cas	[2 3]
INFO  jepsen.util - 0	:fail	:cas	[2 3]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:cas	[4 0]
INFO  jepsen.util - 4	:fail	:cas	[4 0]
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:fail	:write	0
INFO  jepsen.util - 2	:invoke	:cas	[2 2]
INFO  jepsen.util - 2	:fail	:cas	[2 2]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:cas	[1 1]
INFO  jepsen.util - 0	:fail	:cas	[1 1]
INFO  jepsen.util - 2	:invoke	:cas	[4 0]
INFO  jepsen.util - 2	:fail	:cas	[4 0]
INFO  jepsen.util - 4	:invoke	:cas	[2 4]
INFO  jepsen.util - 4	:fail	:cas	[2 4]
INFO  jepsen.util - 0	:invoke	:cas	[0 1]
INFO  jepsen.util - 0	:fail	:cas	[0 1]
INFO  jepsen.util - 2	:invoke	:cas	[1 0]
INFO  jepsen.util - 2	:fail	:cas	[1 0]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:cas	[4 0]
INFO  jepsen.util - 0	:fail	:cas	[4 0]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:cas	[0 0]
INFO  jepsen.util - 0	:fail	:cas	[0 0]
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[0 0]
INFO  jepsen.util - 4	:fail	:cas	[0 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:write	2
INFO  jepsen.util - 0	:fail	:write	2
INFO  jepsen.util - 2	:invoke	:write	1
INFO  jepsen.util - 2	:fail	:write	1
INFO  jepsen.util - 4	:invoke	:cas	[0 1]
INFO  jepsen.util - 4	:fail	:cas	[0 1]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:fail	:write	3
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - :nemesis	:info	:stop	nil
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:cas	[2 3]
INFO  jepsen.util - 0	:fail	:cas	[2 3]
INFO  jepsen.util - :nemesis	:info	:stop	"fully connected"
INFO  jepsen.util - 2	:invoke	:cas	[2 1]
INFO  jepsen.util - 2	:fail	:cas	[2 1]
INFO  jepsen.util - 4	:invoke	:cas	[0 1]
INFO  jepsen.util - 4	:fail	:cas	[0 1]
INFO  jepsen.util - 0	:invoke	:cas	[1 1]
INFO  jepsen.util - 0	:fail	:cas	[1 1]
INFO  jepsen.util - 1	:fail	:cas	[1 0]
INFO  jepsen.util - 3	:fail	:write	0
INFO  jepsen.util - 2	:invoke	:cas	[2 1]
INFO  jepsen.util - 2	:fail	:cas	[2 1]
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:fail	:write	0
INFO  jepsen.util - 0	:invoke	:cas	[0 4]
INFO  jepsen.util - 0	:fail	:cas	[0 4]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:write	4
INFO  jepsen.util - 3	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[4 4]
INFO  jepsen.util - 2	:fail	:cas	[4 4]
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:fail	:write	1
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 3	:invoke	:cas	[0 0]
INFO  jepsen.util - 1	:ok	:write	1
INFO  jepsen.util - 3	:fail	:cas	[0 0]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 0	:invoke	:cas	[1 1]
INFO  jepsen.util - 0	:fail	:cas	[1 1]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[0 1]
INFO  jepsen.util - 1	:fail	:cas	[0 1]
INFO  jepsen.util - 3	:ok	:read	1
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[1 2]
INFO  jepsen.util - 4	:fail	:cas	[1 2]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 1	:ok	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[0 4]
INFO  jepsen.util - 2	:fail	:cas	[0 4]
INFO  jepsen.util - 4	:invoke	:cas	[2 4]
INFO  jepsen.util - 4	:fail	:cas	[2 4]
INFO  jepsen.util - 0	:invoke	:cas	[3 3]
INFO  jepsen.util - 0	:fail	:cas	[3 3]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[1 1]
INFO  jepsen.util - 0	:fail	:cas	[1 1]
INFO  jepsen.util - 1	:invoke	:cas	[2 3]
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:cas	[2 3]
INFO  jepsen.util - 3	:ok	:write	4
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:cas	[1 3]
INFO  jepsen.util - 0	:fail	:cas	[1 3]
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	1
INFO  jepsen.util - 3	:ok	:write	4
INFO  jepsen.util - 2	:invoke	:cas	[0 2]
INFO  jepsen.util - 2	:fail	:cas	[0 2]
INFO  jepsen.util - 4	:invoke	:cas	[2 4]
INFO  jepsen.util - 4	:fail	:cas	[2 4]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 1	:invoke	:write	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:write	2
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:cas	[4 4]
INFO  jepsen.util - 4	:fail	:cas	[4 4]
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 1	:invoke	:cas	[1 4]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:cas	[1 4]
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[3 4]
INFO  jepsen.util - 0	:fail	:cas	[3 4]
INFO  jepsen.util - 1	:invoke	:cas	[4 4]
INFO  jepsen.util - 3	:invoke	:cas	[0 4]
INFO  jepsen.util - 1	:fail	:cas	[4 4]
INFO  jepsen.util - 3	:fail	:cas	[0 4]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 1	:ok	:read	4
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:cas	[0 2]
INFO  jepsen.util - 2	:fail	:cas	[0 2]
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 1	:invoke	:cas	[3 2]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:cas	[3 2]
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 3	:invoke	:cas	[3 2]
INFO  jepsen.util - 3	:fail	:cas	[3 2]
INFO  jepsen.util - 1	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[2 3]
INFO  jepsen.util - 0	:fail	:cas	[2 3]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:cas	[2 2]
INFO  jepsen.util - 0	:fail	:cas	[2 2]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 1	:ok	:read	2
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:cas	[0 1]
INFO  jepsen.util - 0	:fail	:cas	[0 1]
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 3	:invoke	:cas	[2 4]
INFO  jepsen.util - 3	:fail	:cas	[2 4]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 4	:invoke	:cas	[1 0]
INFO  jepsen.util - 4	:fail	:cas	[1 0]
INFO  jepsen.util - 0	:invoke	:cas	[4 3]
INFO  jepsen.util - 0	:fail	:cas	[4 3]
INFO  jepsen.util - 1	:invoke	:cas	[4 4]
INFO  jepsen.util - 1	:fail	:cas	[4 4]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	0
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - :nemesis	:info	:start	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:cas	[4 2]
INFO  jepsen.util - 3	:fail	:cas	[4 2]
INFO  jepsen.util - :nemesis	:info	:start	"Cut off {:n1 #{:n3 :n2}, :n5 #{:n3 :n2}, :n4 #{:n3 :n2}, :n2 #{:n4 :n5 :n1}, :n3 #{:n4 :n5 :n1}}"
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:write	3
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 2	:invoke	:cas	[4 4]
INFO  jepsen.util - 2	:fail	:cas	[4 4]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[4 4]
INFO  jepsen.util - 3	:fail	:cas	[4 4]
INFO  jepsen.util - 2	:invoke	:cas	[3 1]
INFO  jepsen.util - 2	:fail	:cas	[3 1]
INFO  jepsen.util - 1	:invoke	:cas	[3 2]
INFO  jepsen.util - 1	:fail	:cas	[3 2]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[3 2]
INFO  jepsen.util - 1	:fail	:cas	[3 2]
INFO  jepsen.util - 3	:invoke	:cas	[0 2]
INFO  jepsen.util - 3	:fail	:cas	[0 2]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 1	:fail	:write	1
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 3	:ok	:write	4
INFO  jepsen.util - 2	:invoke	:cas	[4 0]
INFO  jepsen.util - 2	:fail	:cas	[4 0]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 2	:invoke	:cas	[0 1]
INFO  jepsen.util - 2	:fail	:cas	[0 1]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[3 3]
INFO  jepsen.util - 3	:fail	:cas	[3 3]
INFO  jepsen.util - 2	:invoke	:cas	[0 2]
INFO  jepsen.util - 2	:fail	:cas	[0 2]
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:fail	:write	0
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	2
INFO  jepsen.util - 1	:fail	:write	2
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:fail	:write	0
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:fail	:write	0
INFO  jepsen.util - 3	:invoke	:cas	[3 3]
INFO  jepsen.util - 3	:fail	:cas	[3 3]
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:fail	:write	0
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[2 3]
INFO  jepsen.util - 2	:fail	:cas	[2 3]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[4 1]
INFO  jepsen.util - 3	:fail	:cas	[4 1]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:fail	:write	0
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 2]
INFO  jepsen.util - 3	:fail	:cas	[2 2]
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:fail	:write	3
INFO  jepsen.util - 1	:invoke	:cas	[3 4]
INFO  jepsen.util - 1	:fail	:cas	[3 4]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:fail	:write	0
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:write	4
INFO  jepsen.util - 3	:invoke	:cas	[3 1]
INFO  jepsen.util - 3	:fail	:cas	[3 1]
INFO  jepsen.util - 2	:invoke	:cas	[0 0]
INFO  jepsen.util - 2	:fail	:cas	[0 0]
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:fail	:write	0
INFO  jepsen.util - 3	:invoke	:cas	[2 1]
INFO  jepsen.util - 3	:fail	:cas	[2 1]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 0]
INFO  jepsen.util - 3	:fail	:cas	[2 0]
INFO  jepsen.util - 2	:invoke	:cas	[4 3]
INFO  jepsen.util - 2	:fail	:cas	[4 3]
INFO  jepsen.util - :nemesis	:info	:stop	nil
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - :nemesis	:info	:stop	"fully connected"
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 1	:fail	:write	1
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[4 3]
INFO  jepsen.util - 0	:fail	:cas	[4 3]
INFO  jepsen.util - 1	:invoke	:cas	[3 3]
INFO  jepsen.util - 1	:fail	:cas	[3 3]
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 3	:fail	:write	4
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:fail	:write	0
INFO  jepsen.util - 3	:invoke	:cas	[4 1]
INFO  jepsen.util - 3	:fail	:cas	[4 1]
INFO  jepsen.util - 2	:invoke	:cas	[1 4]
INFO  jepsen.util - 2	:fail	:cas	[1 4]
INFO  jepsen.util - 0	:invoke	:cas	[4 1]
INFO  jepsen.util - 0	:fail	:cas	[4 1]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:fail	:write	1
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:fail	:write	2
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:fail	:write	0
INFO  jepsen.util - 4	:fail	:write	3
INFO  jepsen.util - 1	:invoke	:write	2
INFO  jepsen.util - 1	:fail	:write	2
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:fail	:write	1
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:fail	:write	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:write	4
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:fail	:write	0
INFO  jepsen.util - 0	:invoke	:cas	[2 1]
INFO  jepsen.util - 0	:fail	:cas	[2 1]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:write	4
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:fail	:write	2
INFO  jepsen.util - 2	:invoke	:cas	[3 2]
INFO  jepsen.util - 2	:fail	:cas	[3 2]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 1	:fail	:write	1
INFO  jepsen.util - 3	:invoke	:cas	[1 0]
INFO  jepsen.util - 3	:fail	:cas	[1 0]
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:fail	:write	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 1	:ok	:write	1
INFO  jepsen.util - 3	:invoke	:cas	[1 1]
INFO  jepsen.util - 3	:fail	:cas	[1 1]
INFO  jepsen.util - 2	:invoke	:write	1
INFO  jepsen.util - 2	:ok	:write	1
INFO  jepsen.util - 0	:invoke	:cas	[4 3]
INFO  jepsen.util - 0	:fail	:cas	[4 3]
INFO  jepsen.util - 4	:invoke	:cas	[1 0]
INFO  jepsen.util - 4	:fail	:cas	[1 0]
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 1	:ok	:write	3
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:write	1
INFO  jepsen.util - 2	:ok	:write	1
INFO  jepsen.util - 0	:invoke	:write	2
INFO  jepsen.util - 0	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:cas	[4 3]
INFO  jepsen.util - 4	:fail	:cas	[4 3]
INFO  jepsen.util - 1	:invoke	:cas	[2 0]
INFO  jepsen.util - 1	:fail	:cas	[2 0]
INFO  jepsen.util - 3	:invoke	:cas	[1 3]
INFO  jepsen.util - 3	:fail	:cas	[1 3]
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:cas	[4 2]
INFO  jepsen.util - 4	:fail	:cas	[4 2]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 2	:invoke	:write	1
INFO  jepsen.util - 2	:ok	:write	1
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 1	:invoke	:cas	[1 1]
INFO  jepsen.util - 1	:fail	:cas	[1 1]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	0
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:cas	[3 0]
INFO  jepsen.util - 1	:fail	:cas	[3 0]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[3 0]
INFO  jepsen.util - 2	:fail	:cas	[3 0]
INFO  jepsen.util - 0	:invoke	:cas	[4 3]
INFO  jepsen.util - 0	:fail	:cas	[4 3]
INFO  jepsen.util - 4	:invoke	:cas	[0 2]
INFO  jepsen.util - 4	:fail	:cas	[0 2]
INFO  jepsen.util - 1	:invoke	:write	2
INFO  jepsen.util - 1	:ok	:write	2
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	0
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[3 1]
INFO  jepsen.util - 0	:fail	:cas	[3 1]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[1 4]
INFO  jepsen.util - 2	:fail	:cas	[1 4]
INFO  jepsen.util - 0	:invoke	:cas	[1 4]
INFO  jepsen.util - 0	:fail	:cas	[1 4]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:cas	[4 0]
INFO  jepsen.util - 2	:fail	:cas	[4 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:cas	[3 1]
INFO  jepsen.util - 4	:fail	:cas	[3 1]
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 3	:invoke	:cas	[3 2]
INFO  jepsen.util - 3	:fail	:cas	[3 2]
INFO  jepsen.util - :nemesis	:info	:start	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 0	:invoke	:cas	[1 2]
INFO  jepsen.util - 0	:fail	:cas	[1 2]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 1	:invoke	:cas	[2 1]
INFO  jepsen.util - 1	:fail	:cas	[2 1]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - :nemesis	:info	:start	"Cut off {:n1 #{:n4 :n2}, :n5 #{:n4 :n2}, :n3 #{:n4 :n2}, :n2 #{:n3 :n5 :n1}, :n4 #{:n3 :n5 :n1}}"
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[1 3]
INFO  jepsen.util - 0	:fail	:cas	[1 3]
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:cas	[2 4]
INFO  jepsen.util - 0	:fail	:cas	[2 4]
INFO  jepsen.util - 4	:invoke	:cas	[4 2]
INFO  jepsen.util - 4	:fail	:cas	[4 2]
INFO  jepsen.util - 2	:invoke	:cas	[3 1]
INFO  jepsen.util - 2	:fail	:cas	[3 1]
INFO  jepsen.util - 0	:invoke	:cas	[2 0]
INFO  jepsen.util - 0	:fail	:cas	[2 0]
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:cas	[2 3]
INFO  jepsen.util - 0	:fail	:cas	[2 3]
INFO  jepsen.util - 4	:invoke	:cas	[0 0]
INFO  jepsen.util - 4	:fail	:cas	[0 0]
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[1 4]
INFO  jepsen.util - 4	:fail	:cas	[1 4]
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 2	:invoke	:cas	[3 0]
INFO  jepsen.util - 2	:fail	:cas	[3 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:cas	[1 4]
INFO  jepsen.util - 2	:fail	:cas	[1 4]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[3 3]
INFO  jepsen.util - 4	:fail	:cas	[3 3]
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[1 2]
INFO  jepsen.util - 4	:fail	:cas	[1 2]
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:cas	[2 2]
INFO  jepsen.util - 2	:fail	:cas	[2 2]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	3
INFO  jepsen.util - 4	:invoke	:cas	[0 2]
INFO  jepsen.util - 4	:fail	:cas	[0 2]
INFO  jepsen.util - 2	:invoke	:cas	[0 3]
INFO  jepsen.util - 2	:fail	:cas	[0 3]
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:cas	[4 0]
INFO  jepsen.util - 4	:fail	:cas	[4 0]
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:cas	[0 0]
INFO  jepsen.util - 0	:fail	:cas	[0 0]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:cas	[2 1]
INFO  jepsen.util - 0	:fail	:cas	[2 1]
INFO  jepsen.util - 4	:invoke	:cas	[3 3]
INFO  jepsen.util - 4	:fail	:cas	[3 3]
INFO  jepsen.util - 2	:invoke	:cas	[4 2]
INFO  jepsen.util - 2	:fail	:cas	[4 2]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	0
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:cas	[4 0]
INFO  jepsen.util - 2	:fail	:cas	[4 0]
INFO  jepsen.util - :nemesis	:info	:stop	nil
INFO  jepsen.util - 0	:invoke	:cas	[2 2]
INFO  jepsen.util - 0	:fail	:cas	[2 2]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 1	:invoke	:cas	[1 3]
INFO  jepsen.util - 1	:fail	:cas	[1 3]
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:fail	:write	1
INFO  jepsen.util - :nemesis	:info	:stop	"fully connected"
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:fail	:write	4
INFO  jepsen.util - 1	:invoke	:cas	[0 4]
INFO  jepsen.util - 1	:fail	:cas	[0 4]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[3 2]
INFO  jepsen.util - 2	:fail	:cas	[3 2]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[4 0]
INFO  jepsen.util - 3	:fail	:cas	[4 0]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[4 4]
INFO  jepsen.util - 0	:fail	:cas	[4 4]
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:fail	:write	1
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[0 0]
INFO  jepsen.util - 0	:fail	:cas	[0 0]
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:fail	:write	3
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:write	3
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 3	:fail	:write	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[2 4]
INFO  jepsen.util - 0	:fail	:cas	[2 4]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:write	3
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:write	2
INFO  jepsen.util - 0	:fail	:write	2
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:fail	:write	4
INFO  jepsen.util - 1	:invoke	:cas	[1 3]
INFO  jepsen.util - 1	:fail	:cas	[1 3]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[0 3]
INFO  jepsen.util - 0	:fail	:cas	[0 3]
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:fail	:write	1
INFO  jepsen.util - 1	:invoke	:cas	[4 3]
INFO  jepsen.util - 1	:fail	:cas	[4 3]
INFO  jepsen.util - 3	:invoke	:cas	[0 3]
INFO  jepsen.util - 3	:fail	:cas	[0 3]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:write	2
INFO  jepsen.util - 0	:fail	:write	2
INFO  jepsen.util - 4	:invoke	:cas	[3 3]
INFO  jepsen.util - 4	:fail	:cas	[3 3]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 3	:fail	:write	0
INFO  jepsen.util - 2	:invoke	:cas	[0 0]
INFO  jepsen.util - 2	:fail	:cas	[0 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[2 1]
INFO  jepsen.util - 1	:fail	:cas	[2 1]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:cas	[0 0]
INFO  jepsen.util - 4	:fail	:cas	[0 0]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 3	:fail	:write	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:cas	[3 3]
INFO  jepsen.util - 4	:fail	:cas	[3 3]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:fail	:write	2
INFO  jepsen.util - 2	:invoke	:cas	[4 4]
INFO  jepsen.util - 2	:fail	:cas	[4 4]
INFO  jepsen.util - 0	:invoke	:cas	[2 0]
INFO  jepsen.util - 0	:fail	:cas	[2 0]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:cas	[0 0]
INFO  jepsen.util - 0	:fail	:cas	[0 0]
INFO  jepsen.util - 4	:invoke	:cas	[2 3]
INFO  jepsen.util - 4	:fail	:cas	[2 3]
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	1
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	1
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	1
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 1	:invoke	:cas	[1 3]
INFO  jepsen.util - 1	:fail	:cas	[1 3]
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 0	:ok	:write	1
INFO  jepsen.util - 4	:invoke	:cas	[4 4]
INFO  jepsen.util - 4	:fail	:cas	[4 4]
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 1	:ok	:write	3
INFO  jepsen.util - 3	:invoke	:cas	[2 3]
INFO  jepsen.util - 3	:fail	:cas	[2 3]
INFO  jepsen.util - 2	:invoke	:cas	[1 2]
INFO  jepsen.util - 2	:fail	:cas	[1 2]
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 1	:invoke	:cas	[1 0]
INFO  jepsen.util - 1	:fail	:cas	[1 0]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	0
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 1	:invoke	:cas	[0 3]
INFO  jepsen.util - 1	:fail	:cas	[0 3]
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:cas	[1 2]
INFO  jepsen.util - 0	:fail	:cas	[1 2]
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	0
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - :nemesis	:info	:start	nil
INFO  jepsen.util - 4	:invoke	:cas	[0 2]
INFO  jepsen.util - 4	:fail	:cas	[0 2]
INFO  jepsen.util - 1	:invoke	:cas	[3 1]
INFO  jepsen.util - 1	:fail	:cas	[3 1]
INFO  jepsen.util - 3	:invoke	:cas	[3 0]
INFO  jepsen.util - 3	:fail	:cas	[3 0]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:cas	[2 2]
INFO  jepsen.util - 0	:fail	:cas	[2 2]
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - :nemesis	:info	:start	"Cut off {:n1 #{:n3 :n2}, :n5 #{:n3 :n2}, :n4 #{:n3 :n2}, :n2 #{:n4 :n5 :n1}, :n3 #{:n4 :n5 :n1}}"
INFO  jepsen.util - 0	:invoke	:cas	[2 4]
INFO  jepsen.util - 0	:fail	:cas	[2 4]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:cas	[3 2]
INFO  jepsen.util - 0	:fail	:cas	[3 2]
INFO  jepsen.util - 4	:invoke	:cas	[3 2]
INFO  jepsen.util - 4	:fail	:cas	[3 2]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	0
INFO  jepsen.util - 0	:invoke	:cas	[0 1]
INFO  jepsen.util - 0	:fail	:cas	[0 1]
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 0	:ok	:write	1
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 3	:invoke	:cas	[1 3]
INFO  jepsen.util - 3	:fail	:cas	[1 3]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 3	:invoke	:cas	[0 4]
INFO  jepsen.util - 3	:fail	:cas	[0 4]
INFO  jepsen.util - 0	:invoke	:cas	[0 3]
INFO  jepsen.util - 0	:fail	:cas	[0 3]
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:cas	[2 0]
INFO  jepsen.util - 0	:fail	:cas	[2 0]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:cas	[0 0]
INFO  jepsen.util - 3	:fail	:cas	[0 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:cas	[0 3]
INFO  jepsen.util - 4	:fail	:cas	[0 3]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:write	2
INFO  jepsen.util - 0	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 3	:invoke	:cas	[2 0]
INFO  jepsen.util - 3	:fail	:cas	[2 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:cas	[1 1]
INFO  jepsen.util - 4	:fail	:cas	[1 1]
INFO  jepsen.util - 3	:invoke	:cas	[2 1]
INFO  jepsen.util - 3	:fail	:cas	[2 1]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:cas	[4 2]
INFO  jepsen.util - 4	:fail	:cas	[4 2]
INFO  jepsen.util - 3	:invoke	:cas	[0 0]
INFO  jepsen.util - 3	:fail	:cas	[0 0]
INFO  jepsen.util - 0	:invoke	:cas	[0 4]
INFO  jepsen.util - 0	:fail	:cas	[0 4]
INFO  jepsen.util - 4	:invoke	:cas	[2 3]
INFO  jepsen.util - 4	:fail	:cas	[2 3]
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:cas	[3 0]
INFO  jepsen.util - 0	:fail	:cas	[3 0]
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:ok	:write	1
INFO  jepsen.util - 3	:invoke	:cas	[3 2]
INFO  jepsen.util - 3	:fail	:cas	[3 2]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 3	:invoke	:cas	[0 1]
INFO  jepsen.util - 3	:fail	:cas	[0 1]
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:cas	[0 2]
INFO  jepsen.util - 4	:fail	:cas	[0 2]
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 3	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:ok	:write	1
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:fail	:write	3
INFO  jepsen.util - 0	:invoke	:cas	[2 2]
INFO  jepsen.util - 0	:fail	:cas	[2 2]
INFO  jepsen.util - :nemesis	:info	:stop	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[0 2]
INFO  jepsen.util - 0	:fail	:cas	[0 2]
INFO  jepsen.util - :nemesis	:info	:stop	"fully connected"
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:fail	:write	2
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:fail	:write	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:write	4
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[0 3]
INFO  jepsen.util - 0	:fail	:cas	[0 3]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[0 1]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:cas	[0 1]
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:cas	[2 4]
INFO  jepsen.util - 2	:fail	:cas	[2 4]
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:fail	:write	0
INFO  jepsen.util - 4	:invoke	:cas	[2 4]
INFO  jepsen.util - 4	:fail	:cas	[2 4]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[3 3]
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[3 3]
INFO  jepsen.util - 2	:invoke	:cas	[2 3]
INFO  jepsen.util - 2	:fail	:cas	[2 3]
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 0	:fail	:write	1
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[2 0]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:cas	[2 0]
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 2	:fail	:write	2
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:fail	:write	0
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:fail	:write	4
INFO  jepsen.util - 1	:invoke	:cas	[1 3]
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:cas	[1 3]
INFO  jepsen.util - 3	:fail	:write	4
INFO  jepsen.util - 2	:invoke	:cas	[2 2]
INFO  jepsen.util - 2	:fail	:cas	[2 2]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:fail	:write	4
INFO  jepsen.util - 1	:invoke	:cas	[4 0]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:cas	[4 0]
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:cas	[4 2]
INFO  jepsen.util - 2	:fail	:cas	[4 2]
INFO  jepsen.util - 0	:invoke	:cas	[2 2]
INFO  jepsen.util - 0	:fail	:cas	[2 2]
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:fail	:write	3
INFO  jepsen.util - 1	:invoke	:cas	[2 4]
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:cas	[2 4]
INFO  jepsen.util - 3	:fail	:write	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:write	1
INFO  jepsen.util - 0	:fail	:write	1
INFO  jepsen.util - 4	:invoke	:cas	[2 2]
INFO  jepsen.util - 4	:fail	:cas	[2 2]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:fail	:write	4
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 3	:invoke	:cas	[3 1]
INFO  jepsen.util - 3	:fail	:cas	[3 1]
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 0	:invoke	:cas	[1 2]
INFO  jepsen.util - 0	:fail	:cas	[1 2]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 3	:invoke	:cas	[2 0]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[2 0]
INFO  jepsen.util - 1	:ok	:read	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 3	:invoke	:cas	[3 2]
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 3	:fail	:cas	[3 2]
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 2	:invoke	:cas	[3 4]
INFO  jepsen.util - 2	:fail	:cas	[3 4]
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[1 4]
INFO  jepsen.util - 0	:fail	:cas	[1 4]
INFO  jepsen.util - 4	:invoke	:cas	[1 4]
INFO  jepsen.util - 4	:fail	:cas	[1 4]
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:write	3
INFO  jepsen.util - 0	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 3	:ok	:write	4
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	4
INFO  jepsen.util - 0	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[3 4]
INFO  jepsen.util - 4	:fail	:cas	[3 4]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	4
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	0
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 1	:ok	:read	0
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:cas	[3 4]
INFO  jepsen.util - 2	:fail	:cas	[3 4]
INFO  jepsen.util - 0	:invoke	:cas	[1 1]
INFO  jepsen.util - 0	:fail	:cas	[1 1]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[0 2]
INFO  jepsen.util - 1	:ok	:read	2
INFO  jepsen.util - 3	:fail	:cas	[0 2]
INFO  jepsen.util - 2	:invoke	:cas	[0 4]
INFO  jepsen.util - 2	:fail	:cas	[0 4]
INFO  jepsen.util - 0	:invoke	:cas	[3 2]
INFO  jepsen.util - 0	:fail	:cas	[3 2]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 1	:invoke	:cas	[0 4]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:cas	[0 4]
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:ok	:write	0
INFO  jepsen.util - :nemesis	:info	:start	nil
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	0
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	0
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 3	:ok	:read	0
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:write	2
INFO  jepsen.util - 0	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - :nemesis	:info	:start	"Cut off {:n5 #{:n2 :n1}, :n4 #{:n2 :n1}, :n3 #{:n2 :n1}, :n1 #{:n3 :n4 :n5}, :n2 #{:n3 :n4 :n5}}"
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 2	:ok	:read	1
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 1	:invoke	:cas	[1 4]
INFO  jepsen.util - 3	:invoke	:cas	[0 4]
INFO  jepsen.util - 1	:fail	:cas	[1 4]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[0 4]
INFO  jepsen.util - 2	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:write	4
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:cas	[3 1]
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 2	:fail	:cas	[3 1]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	3
INFO  jepsen.util - 1	:invoke	:cas	[2 1]
INFO  jepsen.util - 1	:fail	:cas	[2 1]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - 1	:invoke	:cas	[4 0]
INFO  jepsen.util - 1	:fail	:cas	[4 0]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 4]
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 3	:fail	:cas	[2 4]
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 2	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:cas	[3 1]
INFO  jepsen.util - 4	:fail	:cas	[3 1]
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:fail	:write	0
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 1	:invoke	:cas	[4 0]
INFO  jepsen.util - 1	:fail	:cas	[4 0]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:cas	[2 1]
INFO  jepsen.util - 4	:fail	:cas	[2 1]
INFO  jepsen.util - 1	:invoke	:cas	[4 1]
INFO  jepsen.util - 1	:fail	:cas	[4 1]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[4 2]
INFO  jepsen.util - 3	:fail	:cas	[4 2]
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:ok	:write	4
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:ok	:write	4
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 2	:invoke	:cas	[4 0]
INFO  jepsen.util - 3	:invoke	:cas	[2 4]
INFO  jepsen.util - 3	:fail	:cas	[2 4]
INFO  jepsen.util - 2	:fail	:cas	[4 0]
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[3 0]
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 3	:fail	:cas	[3 0]
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[1 2]
INFO  jepsen.util - 4	:fail	:cas	[1 2]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[4 2]
INFO  jepsen.util - 2	:invoke	:cas	[3 0]
INFO  jepsen.util - 3	:fail	:cas	[4 2]
INFO  jepsen.util - 2	:fail	:cas	[3 0]
INFO  jepsen.util - 4	:invoke	:write	4
INFO  jepsen.util - 4	:ok	:write	4
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 1	:fail	:write	1
INFO  jepsen.util - 3	:invoke	:write	1
INFO  jepsen.util - 2	:invoke	:cas	[2 3]
INFO  jepsen.util - 2	:fail	:cas	[2 3]
INFO  jepsen.util - 3	:ok	:write	1
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 1	:invoke	:write	3
INFO  jepsen.util - 1	:fail	:write	3
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 4	:invoke	:cas	[3 1]
INFO  jepsen.util - 4	:fail	:cas	[3 1]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:cas	[2 4]
INFO  jepsen.util - 1	:fail	:cas	[2 4]
INFO  jepsen.util - 3	:invoke	:cas	[0 3]
INFO  jepsen.util - 2	:invoke	:write	1
INFO  jepsen.util - 3	:fail	:cas	[0 3]
INFO  jepsen.util - 2	:ok	:write	1
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:cas	[0 4]
INFO  jepsen.util - 1	:fail	:cas	[0 4]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 3	:ok	:read	3
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 0]
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 3	:fail	:cas	[2 0]
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:ok	:write	1
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:fail	:write	4
INFO  jepsen.util - 3	:invoke	:cas	[4 1]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[4 1]
INFO  jepsen.util - 2	:ok	:read	1
INFO  jepsen.util - :nemesis	:info	:stop	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 1	:invoke	:cas	[1 4]
INFO  jepsen.util - 1	:fail	:cas	[1 4]
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:cas	[3 2]
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 2	:fail	:cas	[3 2]
INFO  jepsen.util - :nemesis	:info	:stop	"fully connected"
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:cas	[1 2]
INFO  jepsen.util - 4	:fail	:cas	[1 2]
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:fail	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 1]
INFO  jepsen.util - 2	:invoke	:cas	[2 4]
INFO  jepsen.util - 3	:fail	:cas	[2 1]
INFO  jepsen.util - 2	:fail	:cas	[2 4]
INFO  jepsen.util - 0	:invoke	:cas	[1 2]
INFO  jepsen.util - 0	:fail	:cas	[1 2]
INFO  jepsen.util - 4	:invoke	:cas	[3 1]
INFO  jepsen.util - 4	:fail	:cas	[3 1]
INFO  jepsen.util - 1	:invoke	:cas	[0 3]
INFO  jepsen.util - 1	:fail	:cas	[0 3]
INFO  jepsen.util - 3	:invoke	:cas	[1 0]
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 2	:fail	:write	0
INFO  jepsen.util - 3	:fail	:cas	[1 0]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:fail	:write	0
INFO  jepsen.util - 1	:invoke	:cas	[2 3]
INFO  jepsen.util - 1	:fail	:cas	[2 3]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 0]
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[2 0]
INFO  jepsen.util - 0	:invoke	:cas	[0 4]
INFO  jepsen.util - 0	:fail	:cas	[0 4]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[4 2]
INFO  jepsen.util - 1	:fail	:cas	[4 2]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:write	0
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:fail	:write	0
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[1 2]
INFO  jepsen.util - 1	:fail	:cas	[1 2]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:cas	[2 4]
INFO  jepsen.util - 0	:fail	:cas	[2 4]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[3 3]
INFO  jepsen.util - 1	:fail	:cas	[3 3]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:invoke	:cas	[2 1]
INFO  jepsen.util - 2	:fail	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[2 1]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:cas	[0 4]
INFO  jepsen.util - 4	:fail	:cas	[0 4]
INFO  jepsen.util - 1	:invoke	:write	1
INFO  jepsen.util - 1	:fail	:write	1
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:fail	:write	2
INFO  jepsen.util - 3	:fail	:read	nil
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:fail	:read	nil
INFO  jepsen.util - 1	:invoke	:cas	[4 1]
INFO  jepsen.util - 1	:fail	:cas	[4 1]
INFO  jepsen.util - 2	:invoke	:cas	[1 4]
INFO  jepsen.util - 3	:invoke	:cas	[4 2]
INFO  jepsen.util - 3	:fail	:cas	[4 2]
INFO  jepsen.util - 2	:fail	:cas	[1 4]
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	1
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	1
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 2	:invoke	:cas	[3 2]
INFO  jepsen.util - 2	:fail	:cas	[3 2]
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:cas	[0 3]
INFO  jepsen.util - 0	:fail	:cas	[0 3]
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	2
INFO  jepsen.util - 2	:invoke	:write	0
INFO  jepsen.util - 3	:invoke	:cas	[4 4]
INFO  jepsen.util - 3	:fail	:cas	[4 4]
INFO  jepsen.util - 2	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:cas	[1 0]
INFO  jepsen.util - 0	:fail	:cas	[1 0]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	0
INFO  jepsen.util - 1	:invoke	:write	4
INFO  jepsen.util - 1	:ok	:write	4
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:cas	[1 3]
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 2	:fail	:cas	[1 3]
INFO  jepsen.util - 0	:invoke	:cas	[1 2]
INFO  jepsen.util - 0	:fail	:cas	[1 2]
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 1	:invoke	:cas	[3 2]
INFO  jepsen.util - 1	:fail	:cas	[3 2]
INFO  jepsen.util - 3	:invoke	:write	4
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:write	4
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[4 2]
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:fail	:cas	[4 2]
INFO  jepsen.util - 0	:invoke	:cas	[0 4]
INFO  jepsen.util - 0	:fail	:cas	[0 4]
INFO  jepsen.util - 4	:invoke	:write	1
INFO  jepsen.util - 4	:ok	:write	1
INFO  jepsen.util - 1	:invoke	:write	2
INFO  jepsen.util - 1	:ok	:write	2
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 2	:invoke	:cas	[3 3]
INFO  jepsen.util - 2	:fail	:cas	[3 3]
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	0
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	0
INFO  jepsen.util - 1	:invoke	:write	2
INFO  jepsen.util - 1	:ok	:write	2
INFO  jepsen.util - 2	:invoke	:write	3
INFO  jepsen.util - 3	:invoke	:cas	[3 0]
INFO  jepsen.util - 3	:fail	:cas	[3 0]
INFO  jepsen.util - 2	:ok	:write	3
INFO  jepsen.util - 0	:invoke	:cas	[0 1]
INFO  jepsen.util - 0	:fail	:cas	[0 1]
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:write	0
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 3	:ok	:write	0
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 0	:invoke	:write	0
INFO  jepsen.util - 0	:ok	:write	0
INFO  jepsen.util - 4	:invoke	:write	3
INFO  jepsen.util - 4	:ok	:write	3
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.util - 3	:invoke	:cas	[2 2]
INFO  jepsen.util - 3	:fail	:cas	[2 2]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 2	:ok	:read	3
INFO  jepsen.util - 0	:invoke	:cas	[4 2]
INFO  jepsen.util - 0	:fail	:cas	[4 2]
INFO  jepsen.util - 4	:invoke	:cas	[3 0]
INFO  jepsen.util - 4	:fail	:cas	[3 0]
INFO  jepsen.util - 1	:invoke	:cas	[4 2]
INFO  jepsen.util - 1	:fail	:cas	[4 2]
INFO  jepsen.util - 3	:invoke	:write	2
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:write	2
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - 0	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 3	:invoke	:cas	[4 1]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[4 1]
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 0	:invoke	:cas	[4 3]
INFO  jepsen.util - 0	:fail	:cas	[4 3]
INFO  jepsen.util - 4	:invoke	:write	0
INFO  jepsen.util - 4	:ok	:write	0
INFO  jepsen.util - :nemesis	:info	:start	nil
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 1	:ok	:write	0
INFO  jepsen.util - 3	:invoke	:cas	[1 2]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[1 2]
INFO  jepsen.util - 2	:ok	:read	0
INFO  jepsen.util - 0	:invoke	:cas	[0 1]
INFO  jepsen.util - 0	:fail	:cas	[0 1]
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 1	:invoke	:write	0
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:write	2
INFO  jepsen.util - 3	:ok	:read	2
INFO  jepsen.util - 2	:ok	:write	2
INFO  jepsen.util - 0	:invoke	:read	nil
INFO  jepsen.util - :nemesis	:info	:start	"Cut off {:n5 #{:n2 :n1}, :n4 #{:n2 :n1}, :n3 #{:n2 :n1}, :n1 #{:n3 :n4 :n5}, :n2 #{:n3 :n4 :n5}}"
INFO  jepsen.util - 4	:invoke	:write	2
INFO  jepsen.util - 4	:ok	:write	2
INFO  jepsen.util - 3	:invoke	:cas	[3 2]
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:fail	:cas	[3 2]
INFO  jepsen.util - 2	:ok	:read	2
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	2
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:read	nil
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 2	:ok	:read	4
INFO  jepsen.util - 4	:invoke	:read	nil
INFO  jepsen.util - 4	:ok	:read	4
INFO  jepsen.util - 3	:invoke	:read	nil
INFO  jepsen.util - 2	:invoke	:write	4
INFO  jepsen.util - 3	:ok	:read	4
INFO  jepsen.util - 2	:ok	:write	4
INFO  jepsen.util - 4	:invoke	:cas	[0 4]
INFO  jepsen.util - 4	:fail	:cas	[0 4]
INFO  jepsen.util - 3	:invoke	:write	3
INFO  jepsen.util - 2	:invoke	:cas	[0 1]
INFO  jepsen.util - 3	:ok	:write	3
INFO  jepsen.util - 2	:fail	:cas	[0 1]
INFO  jepsen.util - 4	:invoke	:cas	[3 3]
INFO  jepsen.util - 4	:fail	:cas	[3 3]
INFO  jepsen.util - :nemesis	:info	:stop	nil
INFO  jepsen.util - :nemesis	:info	:stop	"fully connected"
INFO  jepsen.util - 0	:fail	:read	nil
INFO  jepsen.util - 1	:fail	:write	0
INFO  jepsen.util - :nemesis	:info	:stop	nil
INFO  jepsen.util - :nemesis	:info	:stop	"fully connected"
INFO  jepsen.core - nemesis done
INFO  jepsen.core - Worker 3 done
INFO  jepsen.util - 1	:invoke	:read	nil
INFO  jepsen.core - Worker 2 done
INFO  jepsen.core - Worker 4 done
INFO  jepsen.core - Worker 0 done
INFO  jepsen.util - 1	:ok	:read	3
INFO  jepsen.core - Worker 1 done
INFO  jepsen.core - Run complete, writing
INFO  jepsen.core - Analyzing
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 9 :degenerate 6
:world-size 30 :degenerate 16
:world-size 61 :degenerate 28
:world-size 14 :degenerate 14
:world-size 8 :degenerate 8
:world-size 5 :degenerate 5
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 6 :degenerate 5
:world-size 9 :degenerate 7
:world-size 24 :degenerate 17
:world-size 9 :degenerate 9
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 10 :degenerate 10
:world-size 5 :degenerate 5
:world-size 10 :degenerate 10
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 4 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 10 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 3 :degenerate 3
:world-size 8 :degenerate 8
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 4 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 8 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 8 :degenerate 7
:world-size 4 :degenerate 4
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 3 :degenerate 3
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 10 :degenerate 10
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 11 :degenerate 7
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 11 :degenerate 8
:world-size 3 :degenerate 3
:world-size 11 :degenerate 8
:world-size 5 :degenerate 5
:world-size 14 :degenerate 11
:world-size 5 :degenerate 5
:world-size 14 :degenerate 11
:world-size 5 :degenerate 5
:world-size 12 :degenerate 10
:world-size 5 :degenerate 5
:world-size 17 :degenerate 10
:world-size 5 :degenerate 5
:world-size 17 :degenerate 10
:world-size 5 :degenerate 5
:world-size 12 :degenerate 10
:world-size 5 :degenerate 5
:world-size 14 :degenerate 11
:world-size 5 :degenerate 5
:world-size 6 :degenerate 5
:world-size 5 :degenerate 5
:world-size 6 :degenerate 5
:world-size 5 :degenerate 5
:world-size 6 :degenerate 5
:world-size 5 :degenerate 5
:world-size 17 :degenerate 10
:world-size 5 :degenerate 5
:world-size 9 :degenerate 8
:world-size 5 :degenerate 5
:world-size 6 :degenerate 5
:world-size 5 :degenerate 5
:world-size 12 :degenerate 10
:world-size 5 :degenerate 5
:world-size 12 :degenerate 10
:world-size 5 :degenerate 5
:world-size 9 :degenerate 8
:world-size 5 :degenerate 5
:world-size 6 :degenerate 5
:world-size 5 :degenerate 5
:world-size 6 :degenerate 5
:world-size 5 :degenerate 5
:world-size 11 :degenerate 9
:world-size 5 :degenerate 5
:world-size 9 :degenerate 8
:world-size 5 :degenerate 5
:world-size 14 :degenerate 11
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 14 :degenerate 10
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 9 :degenerate 9
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 8 :degenerate 8
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 8 :degenerate 8
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 18 :degenerate 13
:world-size 6 :degenerate 6
:world-size 9 :degenerate 9
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 15 :degenerate 13
:world-size 6 :degenerate 6
:world-size 9 :degenerate 9
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 5
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 12 :degenerate 8
:world-size 4 :degenerate 4
:world-size 6 :degenerate 4
:world-size 4 :degenerate 4
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 12 :degenerate 8
:world-size 4 :degenerate 4
:world-size 7 :degenerate 4
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 9 :degenerate 6
:world-size 4 :degenerate 4
:world-size 9 :degenerate 6
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 21 :degenerate 10
:world-size 4 :degenerate 4
:world-size 6 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 16 :degenerate 10
:world-size 4 :degenerate 4
:world-size 16 :degenerate 10
:world-size 6 :degenerate 6
:world-size 18 :degenerate 12
:world-size 6 :degenerate 6
:world-size 21 :degenerate 12
:world-size 6 :degenerate 6
:world-size 17 :degenerate 12
:world-size 6 :degenerate 6
:world-size 10 :degenerate 6
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 11 :degenerate 8
:world-size 6 :degenerate 6
:world-size 7 :degenerate 6
:world-size 6 :degenerate 6
:world-size 21 :degenerate 12
:world-size 6 :degenerate 6
:world-size 17 :degenerate 12
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 21 :degenerate 12
:world-size 6 :degenerate 6
:world-size 25 :degenerate 12
:world-size 6 :degenerate 6
:world-size 25 :degenerate 12
:world-size 6 :degenerate 6
:world-size 12 :degenerate 8
:world-size 6 :degenerate 6
:world-size 23 :degenerate 12
:world-size 6 :degenerate 6
:world-size 21 :degenerate 12
:world-size 6 :degenerate 6
:world-size 21 :degenerate 10
:world-size 4 :degenerate 4
:world-size 7 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 16 :degenerate 10
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 18 :degenerate 10
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 16 :degenerate 10
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 21 :degenerate 10
:world-size 4 :degenerate 4
:world-size 21 :degenerate 10
:world-size 4 :degenerate 4
:world-size 16 :degenerate 10
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 10
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 16 :degenerate 10
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 7 :degenerate 6
:world-size 6 :degenerate 6
:world-size 7 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 21 :degenerate 12
:world-size 6 :degenerate 6
:world-size 21 :degenerate 12
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 4 :degenerate 4
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 26 :degenerate 8
:world-size 4 :degenerate 4
:world-size 28 :degenerate 8
:world-size 4 :degenerate 4
:world-size 8 :degenerate 4
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 7 :degenerate 4
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 8 :degenerate 4
:world-size 4 :degenerate 4
:world-size 8 :degenerate 4
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 8 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 28 :degenerate 8
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 8 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 7 :degenerate 4
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 26 :degenerate 8
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 6 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 6 :degenerate 4
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 23 :degenerate 8
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 10 :degenerate 4
:world-size 4 :degenerate 4
:world-size 8 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 9 :degenerate 4
:world-size 4 :degenerate 4
:world-size 26 :degenerate 8
:world-size 4 :degenerate 4
:world-size 10 :degenerate 4
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 26 :degenerate 8
:world-size 4 :degenerate 4
:world-size 9 :degenerate 4
:world-size 4 :degenerate 4
:world-size 8 :degenerate 4
:world-size 4 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 18 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 6 :degenerate 6
:world-size 14 :degenerate 11
:world-size 6 :degenerate 6
:world-size 14 :degenerate 12
:world-size 6 :degenerate 6
:world-size 21 :degenerate 14
:world-size 8 :degenerate 8
:world-size 16 :degenerate 11
:world-size 8 :degenerate 8
:world-size 14 :degenerate 11
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 30 :degenerate 16
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 21 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 14
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 21 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 14 :degenerate 11
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 20 :degenerate 14
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 16 :degenerate 11
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 15 :degenerate 11
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 10 :degenerate 8
:world-size 8 :degenerate 8
:world-size 28 :degenerate 14
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 20 :degenerate 12
:world-size 6 :degenerate 6
:world-size 23 :degenerate 14
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 28 :degenerate 14
:world-size 6 :degenerate 6
:world-size 21 :degenerate 12
:world-size 6 :degenerate 6
:world-size 21 :degenerate 14
:world-size 8 :degenerate 8
:world-size 30 :degenerate 16
:world-size 8 :degenerate 8
:world-size 20 :degenerate 16
:world-size 8 :degenerate 8
:world-size 28 :degenerate 16
:world-size 8 :degenerate 8
:world-size 30 :degenerate 16
:world-size 8 :degenerate 8
:world-size 3 :degenerate 3
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 5
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 3 :degenerate 3
:world-size 7 :degenerate 6
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 4 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 10 :degenerate 10
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 4 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 12 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 2
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 21 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 12 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 2
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 16 :degenerate 10
:world-size 6 :degenerate 6
:world-size 4 :degenerate 4
:world-size 8 :degenerate 6
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 2
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 6 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 16 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 12 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 16 :degenerate 10
:world-size 6 :degenerate 6
:world-size 4 :degenerate 4
:world-size 8 :degenerate 6
:world-size 4 :degenerate 4
:world-size 12 :degenerate 8
:world-size 4 :degenerate 4
:world-size 8 :degenerate 6
:world-size 17 :degenerate 10
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 3 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 12 :degenerate 8
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 7 :degenerate 4
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 4
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 4 :degenerate 3
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 3 :degenerate 3
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 4 :degenerate 4
:world-size 13 :degenerate 10
:world-size 5 :degenerate 5
:world-size 3 :degenerate 3
:world-size 9 :degenerate 6
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 6 :degenerate 6
:world-size 10 :degenerate 6
:world-size 19 :degenerate 12
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 23 :degenerate 12
:world-size 6 :degenerate 6
:world-size 30 :degenerate 12
:world-size 91 :degenerate 28
:world-size 16 :degenerate 16
:world-size 10 :degenerate 10
:world-size 32 :degenerate 16
:world-size 6 :degenerate 6
:world-size 23 :degenerate 12
:world-size 68 :degenerate 24
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 24 :degenerate 12
:world-size 6 :degenerate 6
:world-size 23 :degenerate 12
:world-size 85 :degenerate 24
:world-size 12 :degenerate 12
:world-size 6 :degenerate 6
:world-size 20 :degenerate 8
:world-size 6 :degenerate 6
:world-size 26 :degenerate 12
:world-size 47 :degenerate 18
:world-size 10 :degenerate 10
:world-size 6 :degenerate 6
:world-size 23 :degenerate 12
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 6 :degenerate 6
:world-size 3 :degenerate 3
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 1 :degenerate 1
:world-size 2 :degenerate 2
:world-size 1 :degenerate 1
INFO  jepsen.core - Analysis complete
INFO  jepsen.system.consul - :n3 consul nuked
INFO  jepsen.system.consul - :n2 consul nuked
INFO  jepsen.system.consul - :n4 consul nuked
INFO  jepsen.system.consul - :n1 consul nuked
INFO  jepsen.system.consul - :n5 consul nuked
1964 element history linearizable. :D

Ran 1 tests containing 1 assertions.
0 failures, 0 errors.
```

<!--googleon: all-->
