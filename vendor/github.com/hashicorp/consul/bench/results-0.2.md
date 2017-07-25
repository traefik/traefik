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
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	19.4302 secs.
      Slowest:	0.1715 secs.
      Fastest:	0.0157 secs.
      Average:	0.0606 secs.
      Requests/sec:	1054.0313
      Total Data Received:	102400 bytes.
      Response Size per Request:	5 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.016 [1]	|
      0.031 [233]	|∎
      0.047 [4120]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.062 [8079]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.078 [5082]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.094 [2045]	|∎∎∎∎∎∎∎∎∎∎
      0.109 [656]	|∎∎∎
      0.125 [200]	|
      0.140 [12]	|
      0.156 [31]	|
      0.172 [21]	|

    Latency distribution:
      10% in 0.0416 secs.
      25% in 0.0484 secs.
      50% in 0.0579 secs.
      75% in 0.0697 secs.
      90% in 0.0835 secs.
      95% in 0.0919 secs.
      99% in 0.1113 secs.

    ===== GET default test =====
    GOMAXPROCS=4 boom -n 20480 -c 64 http://localhost:8500/v1/kv/bench
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	9.6804 secs.
      Slowest:	0.0830 secs.
      Fastest:	0.0023 secs.
      Average:	0.0302 secs.
      Requests/sec:	2115.6096
      Total Data Received:	2560000 bytes.
      Response Size per Request:	125 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.002 [1]	|
      0.010 [143]	|
      0.018 [1666]	|∎∎∎∎∎∎∎∎∎
      0.026 [6009]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.035 [6732]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.043 [3857]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.051 [1389]	|∎∎∎∎∎∎∎∎
      0.059 [459]	|∎∎
      0.067 [154]	|
      0.075 [53]	|
      0.083 [17]	|

    Latency distribution:
      10% in 0.0189 secs.
      25% in 0.0233 secs.
      50% in 0.0291 secs.
      75% in 0.0358 secs.
      90% in 0.0427 secs.
      95% in 0.0476 secs.
      99% in 0.0597 secs.

    ===== GET stale test =====
    GOMAXPROCS=4 boom -n 20480 -c 64 http://localhost:8500/v1/kv/bench?stale
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	10.3082 secs.
      Slowest:	0.0972 secs.
      Fastest:	0.0015 secs.
      Average:	0.0322 secs.
      Requests/sec:	1986.7714
      Total Data Received:	2560000 bytes.
      Response Size per Request:	125 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.002 [1]	|
      0.011 [320]	|∎
      0.021 [2558]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.030 [6247]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.040 [6895]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.049 [3174]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.059 [971]	|∎∎∎∎∎
      0.068 [249]	|∎
      0.078 [52]	|
      0.088 [11]	|
      0.097 [2]	|

    Latency distribution:
      10% in 0.0187 secs.
      25% in 0.0246 secs.
      50% in 0.0317 secs.
      75% in 0.0387 secs.
      90% in 0.0461 secs.
      95% in 0.0511 secs.
      99% in 0.0618 secs.

    ===== GET consistent test =====
    GOMAXPROCS=4 boom -n 20480 -c 64 http://localhost:8500/v1/kv/bench?consistent
    20480 / 20480 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

    Summary:
      Total:	10.4835 secs.
      Slowest:	0.0991 secs.
      Fastest:	0.0024 secs.
      Average:	0.0327 secs.
      Requests/sec:	1953.5549
      Total Data Received:	2560000 bytes.
      Response Size per Request:	125 bytes.

    Status code distribution:
      [200]	20480 responses

    Response time histogram:
      0.002 [1]	|
      0.012 [137]	|
      0.022 [2405]	|∎∎∎∎∎∎∎∎∎∎∎∎
      0.031 [7754]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.041 [6382]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.051 [2629]	|∎∎∎∎∎∎∎∎∎∎∎∎∎
      0.060 [826]	|∎∎∎∎
      0.070 [245]	|∎
      0.080 [81]	|
      0.089 [17]	|
      0.099 [3]	|

    Latency distribution:
      10% in 0.0208 secs.
      25% in 0.0254 secs.
      50% in 0.0314 secs.
      75% in 0.0384 secs.
      90% in 0.0463 secs.
      95% in 0.0518 secs.
      99% in 0.0645 secs.

