# Consul Benchmark Results

As part of a benchmark, we started a 4 node DigitalOcean cluster to benchmark.
There are 3 servers, meaning writes must commit to at least 2 servers.
The cluster uses the 16GB DigitalOcean droplet which has the following specs:

 * 8 CPU Cores, 2Ghz
 * 16GB RAM
 * 160GB SSD disk
 * 1Gbps NIC

We used `bonnie++` to benchmark the disk, and the key metrics are:

 * 188MB/s sequential write
 * 86MB/s sequential read-write-flush
 * 2636 random seeks per second

# Output

Below is the output for a test run on a benchmark cluster. We ran the benchmark
several times to warm up the nodes, and this is just a single representative sample.

Note, that a single worker was running the benchmark. This means the "stale" test is
not representative of total throughput, as the client was only routing to a single server.

    ===== PUT test =====
    GOMAXPROCS=4 boom -m PUT -d "74a31e96-1d0f-4fa7-aa14-7212a326986e" -n 20480 -c 64 http://localhost:8500/v1/kv/bench
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	5.0053 secs.
      Slowest:	0.0414 secs.
      Fastest:	0.0062 secs.
      Average:	0.0156 secs.
      Requests/sec:	4091.6699
      Total Data Received:	102400 bytes.
      Response Size per Request:	5 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.006 [1]	|
      0.010 [568]	|∎∎
      0.013 [6184]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.017 [7594]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.020 [3425]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.024 [1506]	|∎∎∎∎∎∎∎
      0.027 [775]	|∎∎∎∎
      0.031 [209]	|∎
      0.034 [142]	|
      0.038 [44]	|
      0.041 [32]	|

    Latency distribution:
      10% in 0.0111 secs.
      25% in 0.0126 secs.
      50% in 0.0148 secs.
      75% in 0.0174 secs.
      90% in 0.0218 secs.
      95% in 0.0243 secs.
      99% in 0.0310 secs.

    ===== GET default test =====
    GOMAXPROCS=4 boom -n 20480 -c 64 http://localhost:8500/v1/kv/bench
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	1.9562 secs.
      Slowest:	0.0330 secs.
      Fastest:	0.0010 secs.
      Average:	0.0061 secs.
      Requests/sec:	10469.5400
      Total Data Received:	2867200 bytes.
      Response Size per Request:	140 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.001 [1]	|
      0.004 [4866]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.007 [10998]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.011 [3520]	|∎∎∎∎∎∎∎∎∎∎∎∎
      0.014 [610]	|∎∎
      0.017 [280]	|∎
      0.020 [117]	|
      0.023 [52]	|
      0.027 [23]	|
      0.030 [12]	|
      0.033 [1]	|

    Latency distribution:
      10% in 0.0033 secs.
      25% in 0.0043 secs.
      50% in 0.0056 secs.
      75% in 0.0072 secs.
      90% in 0.0091 secs.
      95% in 0.0107 secs.
      99% in 0.0170 secs.

    ===== GET stale test =====
    GOMAXPROCS=4 boom -n 20480 -c 64 http://localhost:8500/v1/kv/bench?stale
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	1.8706 secs.
      Slowest:	0.0271 secs.
      Fastest:	0.0011 secs.
      Average:	0.0058 secs.
      Requests/sec:	10948.2819
      Total Data Received:	2867200 bytes.
      Response Size per Request:	140 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.001 [1]	|
      0.004 [3383]	|∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.006 [10080]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.009 [5110]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.011 [1227]	|∎∎∎∎
      0.014 [427]	|∎
      0.017 [141]	|
      0.019 [58]	|
      0.022 [30]	|
      0.025 [14]	|
      0.027 [9]	|

    Latency distribution:
      10% in 0.0032 secs.
      25% in 0.0041 secs.
      50% in 0.0054 secs.
      75% in 0.0070 secs.
      90% in 0.0087 secs.
      95% in 0.0103 secs.
      99% in 0.0146 secs.

    ===== GET consistent test =====
    GOMAXPROCS=4 boom -n 20480 -c 64 http://localhost:8500/v1/kv/bench?consistent
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	1.9989 secs.
      Slowest:	0.0272 secs.
      Fastest:	0.0013 secs.
      Average:	0.0062 secs.
      Requests/sec:	10245.5621
      Total Data Received:	2867200 bytes.
      Response Size per Request:	140 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.001 [1]	|
      0.004 [3176]	|∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.006 [9755]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.009 [5195]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.012 [1505]	|∎∎∎∎∎∎
      0.014 [499]	|∎∎
      0.017 [186]	|
      0.019 [53]	|
      0.022 [34]	|
      0.025 [36]	|
      0.027 [40]	|

    Latency distribution:
      10% in 0.0035 secs.
      25% in 0.0044 secs.
      50% in 0.0057 secs.
      75% in 0.0073 secs.
      90% in 0.0094 secs.
      95% in 0.0111 secs.
      99% in 0.0162 secs.

