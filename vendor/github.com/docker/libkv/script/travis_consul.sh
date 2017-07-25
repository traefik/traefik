#!/bin/bash

if [  $# -gt 0 ] ; then
    CONSUL_VERSION="$1"
else
    CONSUL_VERSION="0.5.2"
fi

# install consul
wget "https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip"
unzip "consul_${CONSUL_VERSION}_linux_amd64.zip"

# make config for minimum ttl
touch config.json
echo "{\"session_ttl_min\": \"1s\"}" >> config.json

# check
./consul --version
