# Benchmarks

## Configuration

I would like to thanks [vincentbernat](https://github.com/vincentbernat) from [exoscale.ch](https://www.exoscale.ch) who kindly provided the infrastructure needed for the benchmarks.

I used 4 VMs for the tests with the following configuration:

- 32 GB RAM
- 8 CPU Cores
- 10 GB SSD
- Ubuntu 14.04 LTS 64-bit

## Setup

1. One VM used to launch the benchmarking tool [wrk](https://github.com/wg/wrk)
2. One VM for Traefik (v1.0.0-beta.416) / nginx (v1.4.6)
3. Two VMs for 2 backend servers in go [whoami](https://github.com/containous/whoami/)

Each VM has been tuned using the following limits:

```bash
sysctl -w fs.file-max="9999999"
sysctl -w fs.nr_open="9999999"
sysctl -w net.core.netdev_max_backlog="4096"
sysctl -w net.core.rmem_max="16777216"
sysctl -w net.core.somaxconn="65535"
sysctl -w net.core.wmem_max="16777216"
sysctl -w net.ipv4.ip_local_port_range="1025       65535"
sysctl -w net.ipv4.tcp_fin_timeout="30"
sysctl -w net.ipv4.tcp_keepalive_time="30"
sysctl -w net.ipv4.tcp_max_syn_backlog="20480"
sysctl -w net.ipv4.tcp_max_tw_buckets="400000"
sysctl -w net.ipv4.tcp_no_metrics_save="1"
sysctl -w net.ipv4.tcp_syn_retries="2"
sysctl -w net.ipv4.tcp_synack_retries="2"
sysctl -w net.ipv4.tcp_tw_recycle="1"
sysctl -w net.ipv4.tcp_tw_reuse="1"
sysctl -w vm.min_free_kbytes="65536"
sysctl -w vm.overcommit_memory="1"
ulimit -n 9999999
```

### Nginx

Here is the config Nginx file use `/etc/nginx/nginx.conf`:

```
user www-data;
worker_processes auto;
worker_rlimit_nofile 200000;
pid /var/run/nginx.pid;

events {
    worker_connections 10000;
    use epoll;
    multi_accept on;
}

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 300;
    keepalive_requests 10000;
    types_hash_max_size 2048;

    open_file_cache max=200000 inactive=300s;
    open_file_cache_valid 300s;
    open_file_cache_min_uses 2;
    open_file_cache_errors on;

    server_tokens off;
    dav_methods off;

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    access_log /var/log/nginx/access.log combined;
    error_log /var/log/nginx/error.log warn;

    gzip off;
    gzip_vary off;

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*.conf;
}
```

Here is the Nginx vhost file used:

```
upstream whoami {
    server IP-whoami1:80;
    server IP-whoami2:80;
    keepalive 300;
}

server {
    listen 8001;
    server_name test.traefik;
    access_log off;
    error_log /dev/null crit;
    if ($host != "test.traefik") {
        return 404;
    }
    location / {
        proxy_pass http://whoami;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
	proxy_set_header  X-Forwarded-Host $host;
    }
}
```

### Traefik

Here is the `traefik.toml` file used:

```toml
maxIdleConnsPerHost = 100000
defaultEntryPoints = ["http"]

[entryPoints]
  [entryPoints.http]
  address = ":8000"

[file]
[backends]
  [backends.backend1]
    [backends.backend1.servers.server1]
    url = "http://IP-whoami1:80"
    weight = 1
    [backends.backend1.servers.server2]
    url = "http://IP-whoami2:80"
    weight = 1

[frontends]
  [frontends.frontend1]
  backend = "backend1"
    [frontends.frontend1.routes.test_1]
    rule = "Host: test.traefik"
```

## Results

### whoami:
```shell
wrk -t20 -c1000 -d60s -H "Host: test.traefik" --latency  http://IP-whoami:80/bench
Running 1m test @ http://IP-whoami:80/bench
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    70.28ms  134.72ms   1.91s    89.94%
    Req/Sec     2.92k   742.42     8.78k    68.80%
  Latency Distribution
     50%   10.63ms
     75%   75.64ms
     90%  205.65ms
     99%  668.28ms
  3476705 requests in 1.00m, 384.61MB read
  Socket errors: connect 0, read 0, write 0, timeout 103
Requests/sec:  57894.35
Transfer/sec:      6.40MB
```

### nginx:
```shell
wrk -t20 -c1000 -d60s -H "Host: test.traefik" --latency  http://IP-nginx:8001/bench
Running 1m test @ http://IP-nginx:8001/bench
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   101.25ms  180.09ms   1.99s    89.34%
    Req/Sec     1.69k   567.69     9.39k    72.62%
  Latency Distribution
     50%   15.46ms
     75%  129.11ms
     90%  302.44ms
     99%  846.59ms
  2018427 requests in 1.00m, 298.36MB read
  Socket errors: connect 0, read 0, write 0, timeout 90
Requests/sec:  33591.67
Transfer/sec:      4.97MB
```

### Traefik:

```shell
wrk -t20 -c1000 -d60s -H "Host: test.traefik" --latency  http://IP-traefik:8000/bench
Running 1m test @ http://IP-traefik:8000/bench
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    91.72ms  150.43ms   2.00s    90.50%
    Req/Sec     1.43k   266.37     2.97k    69.77%
  Latency Distribution
     50%   19.74ms
     75%  121.98ms
     90%  237.39ms
     99%  687.49ms
  1705073 requests in 1.00m, 188.63MB read
  Socket errors: connect 0, read 0, write 0, timeout 7
Requests/sec:  28392.44
Transfer/sec:      3.14MB
```

## Conclusion

Traefik is obviously slower than Nginx, but not so much: Traefik can serve 28392 requests/sec and Nginx 33591 requests/sec which gives a ratio of 85%.
Not bad for young project :) !

Some areas of possible improvements:

- Use [GO_REUSEPORT](https://github.com/kavu/go_reuseport) listener
- Run a separate server instance per CPU core with `GOMAXPROCS=1` (it appears during benchmarks that there is a lot more context switches with Traefik than with nginx)

