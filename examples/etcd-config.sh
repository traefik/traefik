#!/bin/sh

# backend 1
curl -i -H "Accept: application/json" -X PUT -d value="NetworkErrorRatio() > 0.5"   http://localhost:2379/v2/keys/traefik/backends/backend1/circuitbreaker/expression
curl -i -H "Accept: application/json" -X PUT -d value="http://172.17.0.2:80"        http://localhost:2379/v2/keys/traefik/backends/backend1/servers/server1/url
curl -i -H "Accept: application/json" -X PUT -d value="10"                          http://localhost:2379/v2/keys/traefik/backends/backend1/servers/server1/weight
curl -i -H "Accept: application/json" -X PUT -d value="http://172.17.0.3:80"        http://localhost:2379/v2/keys/traefik/backends/backend1/servers/server2/url
curl -i -H "Accept: application/json" -X PUT -d value="1"                           http://localhost:2379/v2/keys/traefik/backends/backend1/servers/server2/weight

# backend 2
curl -i -H "Accept: application/json" -X PUT -d value="drr"                         http://localhost:2379/v2/keys/traefik/backends/backend2/loadbalancer/method
curl -i -H "Accept: application/json" -X PUT -d value="http://172.17.0.4:80"        http://localhost:2379/v2/keys/traefik/backends/backend2/servers/server1/url
curl -i -H "Accept: application/json" -X PUT -d value="1"                           http://localhost:2379/v2/keys/traefik/backends/backend2/servers/server1/weight
curl -i -H "Accept: application/json" -X PUT -d value="http://172.17.0.5:80"        http://localhost:2379/v2/keys/traefik/backends/backend2/servers/server2/url
curl -i -H "Accept: application/json" -X PUT -d value="2"                           http://localhost:2379/v2/keys/traefik/backends/backend2/servers/server2/weight

# frontend 1
curl -i -H "Accept: application/json" -X PUT -d value="backend2"                    http://localhost:2379/v2/keys/traefik/frontends/frontend1/backend
curl -i -H "Accept: application/json" -X PUT -d value="http"                        http://localhost:2379/v2/keys/traefik/frontends/frontend1/entrypoints
curl -i -H "Accept: application/json" -X PUT -d value="Host:test.localhost"         http://localhost:2379/v2/keys/traefik/frontends/frontend1/routes/test_1/rule

# frontend 2
curl -i -H "Accept: application/json" -X PUT -d value="backend1"                    http://localhost:2379/v2/keys/traefik/frontends/frontend2/backend
curl -i -H "Accept: application/json" -X PUT -d value="http"                        http://localhost:2379/v2/keys/traefik/frontends/frontend2/entrypoints
curl -i -H "Accept: application/json" -X PUT -d value="Path:/test"                  http://localhost:2379/v2/keys/traefik/frontends/frontend2/routes/test_2/rule
