# Consul Benchmark Results

As part of a benchmark, we started a 4 node DigitalOcean cluster to benchmark.
There are 3 servers, meaning writes must commit to at least 2 servers.
The cluster uses the 16GB DigitalOcean droplet which has the following specs:

 * 8 CPU Cores, 2Ghz
 * 16GB RAM
 * 160GB SSD disk
 * 1Gbps NIC

# Output

Below is the output for a test run on a benchmark cluster. We ran the benchmark
several times to warm up the nodes, and this is just a single representative sample.

Note, that a single worker was running the benchmark. This means the "stale" test
is not representative of total throughput, as the client was only routing to a
single server.

We also did an initial run where we got lots of noise in the results, so we
increased the number of requests to try to get a better sample.

```
===== PUT test =====
GOMAXPROCS=4 boom -m PUT -d "74a31e96-1d0f-4fa7-aa14-7212a326986e" -n 262144 -c 64 http://127.0.0.1:8500/v1/kv/bench
262144 / 262144 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

Summary:
  Total:	69.3512 secs.
  Slowest:	0.0966 secs.
  Fastest:	0.0026 secs.
  Average:	0.0169 secs.
  Requests/sec:	3779.9491
  Total Data Received:	1048576 bytes.
  Response Size per Request:	4 bytes.

Status code distribution:
  [200]	262144 responses

Response time histogram:
  0.003 [1]	|
  0.012 [66586]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.021 [146064]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.031 [34189]	|∎∎∎∎∎∎∎∎∎
  0.040 [9178]	|∎∎
  0.050 [3682]	|∎
  0.059 [1773]	|
  0.068 [464]	|
  0.078 [124]	|
  0.087 [63]	|
  0.097 [20]	|

Latency distribution:
  10% in 0.0095 secs.
  25% in 0.0119 secs.
  50% in 0.0151 secs.
  75% in 0.0195 secs.
  90% in 0.0260 secs.
  95% in 0.0323 secs.
  99% in 0.0489 secs.

===== GET default test =====
GOMAXPROCS=4 boom -n 262144 -c 64 http://127.0.0.1:8500/v1/kv/bench
262144 / 262144 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

Summary:
  Total:	34.8371 secs.
  Slowest:	0.9568 secs.
  Fastest:	0.0014 secs.
  Average:	0.0085 secs.
  Requests/sec:	7524.8570
  Total Data Received:	36175872 bytes.
  Response Size per Request:	138 bytes.

Status code distribution:
  [200]	262144 responses

Response time histogram:
  0.001 [1]	|
  0.097 [261977]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.192 [38]	|
  0.288 [64]	|
  0.384 [0]	|
  0.479 [0]	|
  0.575 [0]	|
  0.670 [0]	|
  0.766 [0]	|
  0.861 [38]	|
  0.957 [26]	|

Latency distribution:
  10% in 0.0044 secs.
  25% in 0.0055 secs.
  50% in 0.0072 secs.
  75% in 0.0098 secs.
  90% in 0.0130 secs.
  95% in 0.0157 secs.
  99% in 0.0228 secs.

===== GET stale test =====
GOMAXPROCS=4 boom -n 262144 -c 64 http://127.0.0.1:8500/v1/kv/bench?stale
262144 / 262144 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

Summary:
  Total:	26.8200 secs.
  Slowest:	0.0838 secs.
  Fastest:	0.0005 secs.
  Average:	0.0065 secs.
  Requests/sec:	9774.1922
  Total Data Received:	36175872 bytes.
  Response Size per Request:	138 bytes.

Status code distribution:
  [200]	262144 responses

Response time histogram:
  0.001 [1]	|
  0.009 [214210]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.017 [42999]	|∎∎∎∎∎∎∎∎
  0.026 [3709]	|
  0.034 [589]	|
  0.042 [313]	|
  0.050 [166]	|
  0.059 [102]	|
  0.067 [42]	|
  0.075 [11]	|
  0.084 [2]	|

Latency distribution:
  10% in 0.0031 secs.
  25% in 0.0041 secs.
  50% in 0.0056 secs.
  75% in 0.0079 secs.
  90% in 0.0109 secs.
  95% in 0.0134 secs.
  99% in 0.0203 secs.

===== GET consistent test =====
GOMAXPROCS=4 boom -n 262144 -c 64 http://127.0.0.1:8500/v1/kv/bench?consistent
262144 / 262144 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

Summary:
  Total:	35.6962 secs.
  Slowest:	0.0826 secs.
  Fastest:	0.0016 secs.
  Average:	0.0087 secs.
  Requests/sec:	7343.7475
  Total Data Received:	36175872 bytes.
  Response Size per Request:	138 bytes.

Status code distribution:
  [200]	262144 responses

Response time histogram:
  0.002 [1]	|
  0.010 [183123]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.018 [70460]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.026 [6955]	|∎
  0.034 [657]	|
  0.042 [391]	|
  0.050 [229]	|
  0.058 [120]	|
  0.066 [121]	|
  0.074 [68]	|
  0.083 [19]	|

Latency distribution:
  10% in 0.0047 secs.
  25% in 0.0059 secs.
  50% in 0.0077 secs.
  75% in 0.0104 secs.
  90% in 0.0137 secs.
  95% in 0.0162 secs.
  99% in 0.0227 secs.
```

# Profile

In order to probe performance a bit, we ran the get-stale benchmark on the
leader itself and collected pprof data. Here's the output of the benchmark:

```
===== GET stale test =====
GOMAXPROCS=4 boom -n 262144 -c 64 http://127.0.0.1:8500/v1/kv/bench?stale
262144 / 262144 Booooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo! 100.00 %

Summary:
  Total:        16.3139 secs.
  Slowest:      0.0815 secs.
  Fastest:      0.0001 secs.
  Average:      0.0040 secs.
  Requests/sec: 16068.7946
  Total Data Received:  36175872 bytes.
  Response Size per Request:    138 bytes.

Status code distribution:
  [200] 262144 responses

Response time histogram:
  0.000 [1]     |
  0.008 [240221]        |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.016 [18761] |∎∎∎
  0.025 [1937]  |
  0.033 [496]   |
  0.041 [293]   |
  0.049 [131]   |
  0.057 [162]   |
  0.065 [127]   |
  0.073 [10]    |
  0.081 [5]     |

Latency distribution:
  10% in 0.0013 secs.
  25% in 0.0019 secs.
  50% in 0.0030 secs.
  75% in 0.0046 secs.
  90% in 0.0074 secs.
  95% in 0.0109 secs.
  99% in 0.0174 secs.
```

And here's the [resulting flame graph](results-0.7.1.svg).
