#!/bin/sh

# backend 1
curl -i -H "Accept: application/json" -X PUT -d "NetworkErrorRatio() > 0.5"   http://localhost:8500/v1/kv/traefik/backends/backend1/circuitbreaker/expression
curl -i -H "Accept: application/json" -X PUT -d "http://172.17.0.2:80"        http://localhost:8500/v1/kv/traefik/backends/backend1/servers/server1/url
curl -i -H "Accept: application/json" -X PUT -d "10"                          http://localhost:8500/v1/kv/traefik/backends/backend1/servers/server1/weight
curl -i -H "Accept: application/json" -X PUT -d "http://172.17.0.3:80"        http://localhost:8500/v1/kv/traefik/backends/backend1/servers/server2/url
curl -i -H "Accept: application/json" -X PUT -d "1"                           http://localhost:8500/v1/kv/traefik/backends/backend1/servers/server2/weight

# backend 2
curl -i -H "Accept: application/json" -X PUT -d "drr"                         http://localhost:8500/v1/kv/traefik/backends/backend2/loadbalancer/method
curl -i -H "Accept: application/json" -X PUT -d "http://172.17.0.4:80"        http://localhost:8500/v1/kv/traefik/backends/backend2/servers/server1/url
curl -i -H "Accept: application/json" -X PUT -d "1"                           http://localhost:8500/v1/kv/traefik/backends/backend2/servers/server1/weight
curl -i -H "Accept: application/json" -X PUT -d "http://172.17.0.5:80"        http://localhost:8500/v1/kv/traefik/backends/backend2/servers/server2/url
curl -i -H "Accept: application/json" -X PUT -d "2"                           http://localhost:8500/v1/kv/traefik/backends/backend2/servers/server2/weight

# frontend 1
curl -i -H "Accept: application/json" -X PUT -d "backend2"                    http://localhost:8500/v1/kv/traefik/frontends/frontend1/backend
curl -i -H "Accept: application/json" -X PUT -d "http"                        http://localhost:8500/v1/kv/traefik/frontends/frontend1/entrypoints
curl -i -H "Accept: application/json" -X PUT -d "Host:test.localhost"         http://localhost:8500/v1/kv/traefik/frontends/frontend1/routes/test_1/rule

# frontend 2
curl -i -H "Accept: application/json" -X PUT -d "backend1"                    http://localhost:8500/v1/kv/traefik/frontends/frontend2/backend
curl -i -H "Accept: application/json" -X PUT -d "http"                  http://localhost:8500/v1/kv/traefik/frontends/frontend2/entrypoints
curl -i -H "Accept: application/json" -X PUT -d "Path:/test"                  http://localhost:8500/v1/kv/traefik/frontends/frontend2/routes/test_2/rule


# certificate 1
curl -i -H "Accept: application/json" -X PUT -d "https"                  http://localhost:8500/v1/kv/traefik/tls/pair1/entrypoints
curl -i -H "Accept: application/json" -X PUT -d "/tmp/test1.crt"                  http://localhost:8500/v1/kv/traefik/tls/pair1/certificate/certfile
curl -i -H "Accept: application/json" -X PUT -d "/tmp/test1.key"                  http://localhost:8500/v1/kv/traefik/tls/pair1/certificate/keyfile

# certificate 2
curl -i -H "Accept: application/json" -X PUT -d "http,https"                  http://localhost:8500/v1/kv/traefik/tls/pair2/entrypoints
curl -i -H "Accept: application/json" -X PUT -d "/tmp/test2.crt"                  http://localhost:8500/v1/kv/traefik/tls/pair2/certificate/certfile
curl -i -H "Accept: application/json" -X PUT -d "/tmp/test2.key"                  http://localhost:8500/v1/kv/traefik/tls/pair2/certificate/keyfile
