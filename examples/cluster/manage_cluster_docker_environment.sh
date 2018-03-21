#! /usr/bin/env bash

# Initialize variables
readonly basedir=$(dirname $0)
readonly doc_file=$basedir"/docker-compose.yml"
export COMPOSE_PROJECT_NAME="cluster"

# Stop and remove Docker environment
down_environment() {
    echo "DOWN Docker environment"
    ! docker-compose -f $doc_file down -v &>/dev/null && \
        echo "[ERROR] Unable to stop the Docker environment" && exit 11
    return 0
}

# Create and start Docker-compose environment or subpart of its services (if services are listed)
# $@ : List of services to start (optional)
up_environment() {
    echo "START Docker environment "$@
    ! docker-compose -f $doc_file up -d $@ &>/dev/null && \
        echo "[ERROR] Unable to start Docker environment ${@}" && exit 21
    return 0
}

# Stop and remove Docker environment
delete_services() {
    echo "DELETE services "$@
    ! docker-compose -f $doc_file stop $@ &>/dev/null && \
        echo "[ERROR] Unable to stop services "$@ && exit 31
    ! docker-compose -f $doc_file rm -vf $@ &>/dev/null && \
        echo "[ERROR] Unable to delete services "$@ && exit 31
    return 0
}

start_consul() {
    up_environment consul
    waiting_counter=12
    # Not start Traefik store config if consul is not started
    echo "WAIT for consul..."
    sleep 5
    while [[ -z $(curl -s http://10.0.1.2:8500/v1/status/leader) ]]; do
        sleep 5
        let waiting_counter-=1
        if [[ $waiting_counter -eq 0 ]]; then
            echo "[ERROR] Unable to start consul container in the allowed time, the Docker environment will be stopped"
            down_environment
            exit 41
        fi
    done

}

start_etcd3() {
    up_environment etcd3
    waiting_counter=12
    # Not start Traefik store config if consul is not started
    echo "WAIT for ETCD3..."
    while [[ -z $(curl -s --connect-timeout 2 http://10.0.1.12:2379/version) ]]; do
        sleep 5
        let waiting_counter-=1
        if [[ $waiting_counter -eq 0 ]]; then
            echo "[ERROR] Unable to start etcd3 container in the allowed time, the Docker environment will be stopped"
            down_environment
            exit 51
        fi
    done
}

start_storeconfig_consul() {
    # Create traefik.toml with consul provider
    cp $basedir/traefik.toml.tmpl $basedir/traefik.toml
    echo '
        [consul]
        endpoint = "10.0.1.2:8500"
        watch = true
        prefix = "traefik"' >> $basedir/traefik.toml
    up_environment storeconfig
    rm -f $basedir/traefik.toml
    waiting_counter=5
    delete_services storeconfig

}

start_storeconfig_etcd3() {
    # Create traefik.toml with consul provider
    cp $basedir/traefik.toml.tmpl $basedir/traefik.toml
    echo '
        [etcd]
        endpoint = "10.0.1.12:2379"
        watch = true
        prefix = "/traefik"
        useAPIV3 = true' >> $basedir/traefik.toml
    up_environment storeconfig
    rm -f $basedir/traefik.toml
    waiting_counter=5
    # Don't start Traefik store config if ETCD3 is not started
    echo "Delete storage file key..."
    while [[ $(docker-compose -f $doc_file up --exit-code-from etcdctl-ping etcdctl-ping &>/dev/null) -ne 0 &&  $waiting_counter -gt 0 ]]; do
        sleep 5
        let waiting_counter-=1
    done
    delete_services storeconfig etcdctl-ping
}

start_traefik() {
    up_environment traefik01
    # Waiting for the first instance which is mapped to the host as leader before to start the second one
    waiting_counter=5
    echo "WAIT for traefik leader..."
    sleep 10
    while [[ -z $(curl -s --connect-timeout 3 http://10.0.1.8:8080/ping) ]]; do
        sleep 2
        let waiting_counter-=1
        if [[ $waiting_counter -eq 0 ]]; then
            echo "[ERROR] Unable to start Traefik leader container in the allowed time, the Docker environment will be stopped"
            down_environment
            exit 51
        fi
    done
    up_environment whoami01
    waiting_counter=5
    echo "WAIT for whoami..."
    sleep 10
    while [[ -z $(curl -s --connect-timeout 3 http://10.0.1.10) ]]; do
        sleep 2
        let waiting_counter-=1
        if [[ $waiting_counter -eq 0 ]]; then
            echo "[ERROR] Unable to start whoami container in the allowed time, the Docker environment will be stopped"
            down_environment
            exit 52
        fi
    done
    up_environment traefik02 whoami02
}

# Start boulder services
start_boulder() {
    echo "Start boulder environment"
    up_environment bmysql bhsm boulder
    waiting_counter=12
    # Not start Traefik if boulder is not started
    echo "WAIT for boulder..."
    while [[ -z $(curl -s http://10.0.1.3:4001/directory) ]]; do
        sleep 5
        let waiting_counter-=1
        if [[ $waiting_counter -eq 0 ]]; then
            echo "[ERROR] Unable to start boulder container in the allowed time, the Docker environment will be stopped"
            down_environment
            exit 61
        fi
    done
    echo "Boulder started."
}

# Script usage
show_usage() {
    echo
    echo "USAGE : manage_cluster_docker_environment.sh [--start [--consul|--etcd3]|--stop|--restart [--consul|--etcd3]]"
    echo
}

# Main method
# $@ All parameters given
main() {

    [[ $# -lt 1 && $# -gt 2 ]] && show_usage && exit 1

    case $1 in
        "--start")
                [[ $# -ne 2 ]] && show_usage && exit 2
                # The domains who01.localhost.com and who02.localhost.com have to refer 127.0.0.1
                # I, the /etc/hosts file
                for whoami_idx in "01" "02"; do
                    [[ -z $(cat /etc/hosts | grep "127.0.0.1" | grep -vE "^#" | grep "who${whoami_idx}.localhost.com") ]] && \
                        echo "[ERROR] Domain who${whoami_idx}.localhost.com has to refer to 127.0.0.1 into /etc/hosts file." && \
                        exit 3
                done
                case $2 in
                    "--etcd3")
                        echo "USE ETCD V3 AS KV STORE"
                        export TRAEFIK_CMD="--etcd --etcd.endpoint=10.0.1.12:2379 --etcd.useAPIV3=true"
                        start_boulder && \
                        start_etcd3 && \
                        start_storeconfig_etcd3 && \
                        start_traefik
                        ;;
                    "--consul")
                        echo "USE CONSUL AS KV STORE"
                        export TRAEFIK_CMD="--consul --consul.endpoint=10.0.1.2:8500"
                        start_boulder && \
                        start_consul && \
                        start_storeconfig_consul && \
                        start_traefik
                        ;;
                    *)
                        show_usage && exit 4
                        ;;
                esac
            echo "ENVIRONMENT SUCCESSFULLY STARTED"
            ;;
        "--stop")
            ! down_environment
            echo "ENVIRONMENT SUCCESSFULLY STOPPED"
            ;;
        "--restart")
            [[ $# -ne 2 ]] && show_usage && exit 5
            down_environment
            main --start $2
            ;;
        *)
            show_usage && exit 6
            ;;
    esac
}

main $@