#!/usr/bin/env bash
set -e

echo "Starting Consul..."
if [ -x "$(command -v systemctl)" ]; then
  echo "using systemctl"
  sudo systemctl enable consul.service
  sudo systemctl start consul
else 
  echo "using upstart"
  sudo start consul
fi
