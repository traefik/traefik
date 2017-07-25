Consul Benchmark
================

This repo contains the automation necessary for the Consul benchmarks.

There is a single main Packer file `bench.json`. To use it, the variables
for `do_client_id` and `do_api_key` must be provided. These correspond to
your DigitalOcean client ID and API key.

When Packer runs, it will generate 3 images:

* bench-bootstrap - Consul server in bootstrap mode
* bench-server - Consul server
* bench-worker - Worker node

For the benchmark you should start 1 bootstrap instance, and 2 normal
servers. As many workers as desired can be started. Once the nodes are
up, you must SSH into one of the Consul servers.

Connect all the nodes with:

    $ consul join <n1> ... <n5>

This will connect all the nodes within the same datacenter.

To run the benchmarks, use the Makefile:

    $ cd /
    $ make # Runs all the benchmarks
    $ make put # Runs only the PUT benchmarks

There is no good way to currently cause multiple workers to run at the same
time, so I just type in the make command and rapidly start the test on all
workers. It is not perfect, but the test runs long enough that the calls
overlap.

