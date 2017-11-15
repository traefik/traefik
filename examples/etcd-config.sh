#!/bin/sh

# This shell script allows adding data in a ETCD V3 which is directly installed on a host
# or in container which binds its port 2379 on a host in the way to allows etcd_client container to access it

# $1 : ETCD IP Address (not 127.0.0.1)

if [[ $# -ne 1 || $1 == "127.0.0.1" || -z $(echo $1 | grep -oE "([0-9]+(\.)?){4}") ]]; then
    echo "USAGE : etcd-config.sh ETCD_IP_ADDRESS"
    echo "        ETCD_IP_ADDRESS : Host ETCD IP address (not 127.0.0.1)"
    exit 1
else
    readonly etcd_ip=$1
fi


# backend 1
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend1/circuitbreaker/expression" "NetworkErrorRatio() > 0.5"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend1/servers/server1/url" "http://172.17.0.2:80"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend1/servers/server1/weight" "10"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend1/servers/server2/url" "http://172.17.0.3:80"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend1/servers/server2/weight" "1"

# backend 2
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend2/loadbalancer/method" "drr"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend2/servers/server1/url" "http://172.17.0.4:80"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend2/servers/server1/weight" "1"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend2/servers/server2/url" "http://172.17.0.5:80"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/backends/backend2/servers/server2/weight" "2"

# frontend 1
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/frontends/frontend1/backend" "backend2"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik//frontends/frontend1/entrypoints" "http"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/frontends/frontend1/routes/test_1/rule" "Host:test.localhost"

# frontend 2
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/frontends/frontend2/backend" "backend1"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/frontends/frontend2/entrypoints" "http"
docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/frontends/frontend2/routes/test_2/rule" "Path:/test"
