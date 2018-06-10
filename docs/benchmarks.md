# Benchmarks

## Configuration

I would like to thanks [vincentbernat](https://github.com/vincentbernat) from [exoscale.ch](https://www.exoscale.ch) who kindly provided the infrastructure needed for the benchmarks.

I used 4 VMs for the tests with the following configuration:

- 30 GB RAM
- 8 CPU Cores
- 200 GB local SSD
- CentOS Linux release 7.5.1804 (Kernel 3.10.0-693.21.1.el7.x86_64)

## Setup

1. One VM used to launch the benchmarking tool [wrk](https://github.com/wg/wrk)
2. One VM for Traefik (v1.6.3 amd64) / nginx (v1.14.0)
3. Two VMs for 2 backend servers in go [whoami](https://github.com/emilevauge/whoamI/)

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

### Latency

Latency between each VM has been around 0.5 - 0.9ms

### Nginx

Here is the config Nginx file use `/etc/nginx/nginx.conf`:

```
user nginx;
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
    listen 80;
    server_name test.local;
    access_log off;
    error_log /dev/null crit;
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

[entryPoints]
  [entryPoints.http]
  address = ":80"

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
    rule = "Host: test.local"
```

## Results

### whoami:
```shell
wrk -t20 -c1000 -d60s -H "Host: test.local" --latency  http://IP-whoami:80/bench
Running 1m test @ http://IP-whoami/bench
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    15.48ms    5.86ms 267.96ms   93.09%
    Req/Sec     3.27k   398.14     4.88k    70.88%
  Latency Distribution
     50%   14.87ms
     75%   17.08ms
     90%   19.68ms
     99%   26.19ms
  3906591 requests in 1.00m, 469.43MB read
Requests/sec:  65083.24
Transfer/sec:      7.82MB
```

### nginx:
```shell
wrk -t20 -c1000 -d60s -H "Host: test.local" --latency  http://IP-nginx:80/bench
Running 1m test @ http://IP-nginx:80/bench
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    22.40ms   17.59ms   1.05s    99.60%
    Req/Sec     2.30k   381.60     3.47k    68.27%
  Latency Distribution
     50%   21.03ms
     75%   24.74ms
     90%   28.02ms
     99%   35.19ms
  2743560 requests in 1.00m, 368.92MB read
Requests/sec:  45708.20
Transfer/sec:      6.15MB
```

### Traefik:

```shell
/wrk -t20 -c1000 -d60s -H "Host: test.local" --latency  http://IP-traefik:80/bench
Running 1m test @ http://IP-traefik:80/bench
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    51.45ms   19.13ms 296.68ms   72.46%
    Req/Sec     0.98k   125.96     1.41k    68.61%
  Latency Distribution
     50%   53.88ms
     75%   63.35ms
     90%   72.72ms
     99%   97.32ms
  1167155 requests in 1.00m, 113.53MB read
Requests/sec:  19440.05
Transfer/sec:      1.89MB
```

## Conclusion

Traefik is slower than nginx:
- Traefik can serve 19.440 req/sec 
- nginx can serve 45.708 req/sec
- whoami can serve 65.083 req/sec (without proxy)

Some areas of possible improvements:

- Use [GO_REUSEPORT](https://github.com/kavu/go_reuseport) listener
- Run a separate server instance per CPU core with `GOMAXPROCS=1` (it appears during benchmarks that there is a lot more context switches with Traefik than with nginx)

