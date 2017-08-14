#! /usr/bin/env bash

# Initialize variables
readonly traefik_url="traefik.localhost.com"
readonly basedir=$(dirname $0)
readonly doc_file=$basedir"/compose-acme.yml"

# Stop and remove Docker environment
down_environment() {
    echo "STOP Docker environment"
    ! docker-compose -f $doc_file down -v &>/dev/null && \
        echo "[ERROR] Impossible to stop the Docker environment" && exit 11
}

# Create and start Docker-compose environment or subpart of its services (if services are listed)
# $@ : List of services to start (optional)
up_environment() {
    echo "START Docker environment"
    ! docker-compose -f $doc_file up -d $@ &>/dev/null && \
        echo "[ERROR] Impossible to start Docker environment" && exit 21
}

# Init the environment : get IP address and create needed files
init_environment() {
    for netw in $(ip addr show | grep -v "LOOPBACK" | grep -v docker | grep -oE "^[0-9]{1}: .*:" | cut -d ':' -f2); do
        ip_addr=$(ip addr show $netw | grep -E "inet " | grep -Eo "[0-9]*\.[0-9]*\.[0-9]*\.[0-9]*" | head -n 1)
        [[ ! -z $ip_addr ]] && break
    done

    [[ -z $ip_addr ]] && \
        echo "[ERROR] Impossible to find an IP address for the Docker host" && exit 31

    # The $traefik_url entry must exist into /etc/hosts file
    # It has to refer to the $ip_addr IP address
        [[ $(cat /etc/hosts | grep $traefik_url | grep -vE "^#" | grep -oE "([0-9]+(\.)?){4}") != $ip_addr ]] && \
            echo "[ERROR] Domain ${traefik_url} has to refer to ${ip_addr} into /etc/hosts file." && exit 32
    # Export IP_HOST to use it in the DOcker COmpose file
    export IP_HOST=$ip_addr

    echo "CREATE empty acme.json file"
    rm -f $basedir/acme.json && \
    touch $basedir/acme.json && \
    chmod 600 $basedir/acme.json # Needed for ACME
}

# Start all the environement
start() {
    init_environment
    echo "Start boulder environment"
    up_environment bmysql brabbitmq bhsm boulder
    waiting_counter=12
    # Not start Traefik if boulder is not started
    echo "WAIT for boulder..."
    while [[ -z $(curl -s http://$traefik_url:4000/directory) ]]; do
        sleep 5
        let waiting_counter-=1
        if [[ $waiting_counter -eq 0 ]]; then
            echo "[ERROR] Impossible to start boulder container in the allowed time, the Docker environment will be stopped"
            down_environment
            exit 41
        fi
    done
    echo "START Traefik container"
    up_environment traefik
}

# Script usage
show_usage() {
    echo
    echo "USAGE : manage_acme_docker_environment.sh [--start|--stop|--restart]"
    echo
}

# Main method
# $@ All parameters given
main() {

    [[ $# -ne 1 ]] && show_usage && exit 1

    case $1 in
        "--start")
            # Start boulder environment
            start
            echo "ENVIRONMENT SUCCESSFULLY STARTED"
            ;;
        "--stop")
            ! down_environment
            echo "ENVIRONMENT SUCCESSFULLY STOPPED"
            ;;
        "--restart")
            down_environment
            start
            echo "ENVIRONMENT SUCCESSFULLY STARTED"
            ;;
        *)
            show_usage && exit 2
            ;;
    esac
}

main $@