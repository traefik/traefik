#! /usr/bin/env bash

# Initialize variables
readonly traefik_url="traefik.localhost.com"
readonly basedir=$(dirname $0)
readonly doc_file=$basedir"/docker-compose.yml"

# Stop and remove Docker environment
down_environment() {
    echo "STOP Docker environment"
    ! docker-compose -f $doc_file down -v &>/dev/null && \
        echo "[ERROR] Unable to stop the Docker environment" && exit 11
}

# Create and start Docker-compose environment or subpart of its services (if services are listed)
# $@ : List of services to start (optional)
up_environment() {
    echo "START Docker environment"
    ! docker-compose -f $doc_file up -d $@ &>/dev/null && \
        echo "[ERROR] Unable to start Docker environment" && exit 21
}

# Init the environment : get IP address and create needed files
init_environment() {
    echo "CREATE empty acme.json file"
    rm -f $basedir/acme.json && \
    touch $basedir/acme.json && \
    chmod 600 $basedir/acme.json # Needed for ACME
}

# Start all the environement
start_boulder() {
    init_environment
    echo "Start boulder environment"
    up_environment bmysql bhsm boulder
    waiting_counter=12
    # Not start Traefik if boulder is not started
    echo "WAIT for boulder..."
    while [[ -z $(curl -s http://127.0.0.1:4000/directory) ]]; do
        sleep 5
        let waiting_counter-=1
        if [[ $waiting_counter -eq 0 ]]; then
            echo "[ERROR] Unable to start boulder container in the allowed time, the Docker environment will be stopped"
            down_environment
            exit 41
        fi
    done
}

# Script usage
show_usage() {
    echo
    echo "USAGE : manage_acme_docker_environment.sh [--dev|--start|--stop|--restart]"
    echo
}

# Main method
# $@ All parameters given
main() {

    [[ $# -ne 1 ]] && show_usage && exit 1

    case $1 in
        "--dev")
            start_boulder
            ;;
        "--start")
            # Start boulder environment
            start_boulder
            echo "START Traefik container"
            up_environment traefik
            echo "ENVIRONMENT SUCCESSFULLY STARTED"
            ;;
        "--stop")
            ! down_environment
            echo "ENVIRONMENT SUCCESSFULLY STOPPED"
            ;;
        "--restart")
            down_environment
            start_boulder
            echo "START Traefik container"
            up_environment traefik
            echo "ENVIRONMENT SUCCESSFULLY RESTARTED"
            ;;
        *)
            show_usage && exit 2
            ;;
    esac
}

main $@