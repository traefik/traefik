#!/usr/bin/env bash
set -e

# Read the address to join from the file we provisioned
JOIN_ADDRS=$(cat /tmp/consul-server-addr | tr -d '\n')

# consul version to install
CONSUL_VERSION=0.6.4

sudo sh -c 'echo "127.0.0.1 consul-node-'$(cat /tmp/consul-server-index)'" >> /etc/hosts'

echo "Installing dependencies..."
sudo apt-get update -y
sudo apt-get install -y unzip

echo "Fetching Consul..."
cd /tmp
wget "https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip" -O consul.zip

echo "Installing Consul..."
unzip consul.zip >/dev/null
sudo chmod +x consul
sudo mv consul /usr/local/bin/consul
sudo mkdir -p /etc/consul.d
sudo mkdir -p /mnt/consul
sudo mkdir -p /etc/service

# Setup the join address
cat >/tmp/consul-join << EOF
export CONSUL_JOIN="${JOIN_ADDRS}"
EOF
sudo mv /tmp/consul-join /etc/service/consul-join
chmod 0644 /etc/service/consul-join

echo "Installing Upstart service..."
sudo mv /tmp/upstart.conf /etc/init/consul.conf
sudo mv /tmp/upstart-join.conf /etc/init/consul-join.conf
