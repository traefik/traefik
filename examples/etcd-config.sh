#! /usr/bin/env bash

#
# Insert data in ETCD V3
function insert_etcd2_data() {
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

    # certificate 1
    curl -i -H "Accept: application/json" -X PUT -d value="https"                       http://localhost:2379/v2/keys/traefik/tls/pair1/entrypoints
    curl -i -H "Accept: application/json" -X PUT -d value="/tmp/test1.crt"              http://localhost:2379/v2/keys/traefik/tls/pair1/certificate/certfile
    curl -i -H "Accept: application/json" -X PUT -d value="/tmp/test1.key"              http://localhost:2379/v2/keys/traefik/tls/pair1/certificate/keyfile

    # certificate 2
    curl -i -H "Accept: application/json" -X PUT -d value="http,https"                  http://localhost:2379/v2/keys/traefik/tls/pair2/entrypoints
    curl -i -H "Accept: application/json" -X PUT -d value="/tmp/test2.crt"              http://localhost:2379/v2/keys/traefik/tls/pair2/certificate/certfile
    curl -i -H "Accept: application/json" -X PUT -d value="/tmp/test2.key"              http://localhost:2379/v2/keys/traefik/tls/pair2/certificate/keyfile
}

#
# Insert data in ETCD V3
# $1 = ECTD IP address
# Note : This function allows adding data in a ETCD V3 which is directly installed on a host
#        or in container which binds its port 2379 on a host in the way to allows etcd_client container to access it.
function insert_etcd3_data() {

    readonly etcd_ip=$1
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

    # certificate 1
    docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/tls/pair1/entrypoints" "https"
    docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/tls/pair1/certificate/certfile" "/tmp/test1.crt"
    docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/tls/pair1/certificate/keyfile" "/tmp/test1.key"

    # certificate 2
    docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/tls/pair2/entrypoints" "https"
    docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/tls/pair2/certificate/certfile" "/tmp/test2.crt"
    docker container run --rm -ti -e ETCDCTL_DIAL_="TIMEOUT 10s" -e ETCDCTL_API="3" tenstartups/etcdctl --endpoints=[$etcd_ip:2379] put "/traefik/tls/pair2/certificate/keyfile" "/tmp/test2.key"
}

function show_usage() {
        echo "USAGE : etcd-config.sh ETCD_API_VERSION [ETCD_IP_ADDRESS]"
        echo "        ETCD_API_VERSION : Values V2 or V3 (V3 requires ETCD_IP_ADDRESS)"
        echo "        ETCD_IP_ADDRESS : Host ETCD IP address (not 127.0.0.1)"
}

function main() {
    case $# in
        1)
            if [[ $1 == "V2" ]]; then
                insert_etcd2_data
            else
                show_usage
                exit 1
            fi
            ;;
        2)
            if [[ $1 == "V3" && $2 != "127.0.0.1" && ! -z $(echo $2 | grep -oE "([0-9]+(\.)?){4}") ]]; then
                insert_etcd3_data $2
            else
                show_usage
                exit 1
            fi
            ;;
        *)
            show_usage
            exit 1
            ;;
    esac
}

main $@
