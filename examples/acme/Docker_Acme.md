# ACME Testing environment

## Objectives

In our integration ACME tests, we use a simulated Let's Encrypt container based stack named boulder.

The goal of this directory is to provide to developers a Traefik-boulder full stack environment.
This environment may be used in order to quickly test developments on ACME certificates management.

The provided Boulder stack is based on the environment used during integration tests.

## Directory content

* **docker-compose.yml** : Docker-Compose file which contains the description of Traefik and all the boulder stack containers to get,
* **acme.toml** : Traefik configuration file used by the Traefik container described above,
* **manage_acme_docker_environment.sh**  Shell script which does all needed checks and manages the docker-compose environment.

## Shell script

### Description

To work fine, boulder needs a domain name, with a related IP and storage file. The shell script allows to check the environment before launching the Docker environment with the rights parameters and to managing this environment.

### Use

The script **manage_acme_docker_environment.sh** requires one argument. This argument can have 3 values :

* **--start** : Launch a new Docker environment Boulder + Traefik.
* **--stop** : Stop and delete the current Docker environment.
* **--restart--** : Concatenate **--stop** and **--start** actions.
* **--dev** : Launch a new Boulder Docker environment.