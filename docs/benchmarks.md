# Benchmarks

Here are some early Benchmarks between Nginx, HA-Proxy and Træfɪk acting as simple load balancers between two servers.

- Nginx:

```sh
$ docker run -d -e VIRTUAL_HOST=test.nginx.localhost emilevauge/whoami
$ docker run -d -e VIRTUAL_HOST=test.nginx.localhost emilevauge/whoami
$ docker run --log-driver=none -d -p 80:80 -v /var/run/docker.sock:/tmp/docker.sock:ro jwilder/nginx-proxy
$ wrk -t12 -c400 -d60s -H "Host: test.nginx.localhost" --latency http://127.0.0.1:80
Running 1m test @ http://127.0.0.1:80
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   162.61ms  203.34ms   1.72s    91.07%
    Req/Sec   277.57    107.67   790.00     67.53%
  Latency Distribution
     50%  128.19ms
     75%  218.22ms
     90%  342.12ms
     99%    1.08s 
  197991 requests in 1.00m, 82.32MB read
  Socket errors: connect 0, read 0, write 0, timeout 18
Requests/sec:   3296.04
Transfer/sec:      1.37MB
```

- HA-Proxy:

```sh
$ docker run -d --name web1 -e VIRTUAL_HOST=test.haproxy.localhost emilevauge/whoami
$ docker run -d --name web2 -e VIRTUAL_HOST=test.haproxy.localhost emilevauge/whoami
$ docker run -d -p 80:80 --link web1:web1 --link web2:web2 dockercloud/haproxy
$ wrk -t12 -c400 -d60s -H "Host: test.haproxy.localhost" --latency http://127.0.0.1:80
Running 1m test @ http://127.0.0.1:80
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   158.08ms  187.88ms   1.75s    89.61%
    Req/Sec   281.33    120.47     0.98k    65.88%
  Latency Distribution
     50%  121.77ms
     75%  227.10ms
     90%  351.98ms
     99%    1.01s 
  200462 requests in 1.00m, 59.65MB read
Requests/sec:   3337.66
Transfer/sec:      0.99MB
```

- Træfɪk:

```sh
$ docker run -d -l traefik.backend=test1 -l traefik.frontend.rule=Host -l traefik.frontend.value=test.traefik.localhost emilevauge/whoami
$ docker run -d -l traefik.backend=test1 -l traefik.frontend.rule=Host -l traefik.frontend.value=test.traefik.localhost emilevauge/whoami
$ docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/traefik.toml -v /var/run/docker.sock:/var/run/docker.sock traefik
$ wrk -t12 -c400 -d60s -H "Host: test.traefik.localhost" --latency http://127.0.0.1:80
Running 1m test @ http://127.0.0.1:80
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   132.93ms  121.89ms   1.20s    66.62%
    Req/Sec   280.95    104.88   740.00     68.26%
  Latency Distribution
     50%  128.71ms
     75%  214.15ms
     90%  281.45ms
     99%  498.44ms
  200734 requests in 1.00m, 80.02MB read
Requests/sec:   3340.13
Transfer/sec:      1.33MB
```