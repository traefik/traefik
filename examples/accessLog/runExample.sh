#!/bin/bash
# Script to run a three-server example.  This script runs the three servers and restarts Traefik
# Once it is running, use the command:
#
# curl http://127.0.0.1:80/test{1,2,2}
#
# to send requests to send test requests to the servers.  You should see a response like:
#
# Handler1: received query test1!
# Handler2: received query test2!
# Handler3: received query test2!
#
# and can then inspect log/access.log to see frontend, backend, and timing

# Kill traefik and any running example processes
sudo pkill -f traefik
pkill -f exampleHandler
[ ! -d log ] && mkdir log

# Start new example processes
cd examples/accessLog
go build exampleHandler.go
[ $? -ne 0 ] && exit $?
./exampleHandler -n Handler1 -p 8081 &
[ $? -ne 0 ] && exit $?
./exampleHandler -n Handler2 -p 8082 &
[ $? -ne 0 ] && exit $?
./exampleHandler -n Handler3 -p 8083 &
[ $? -ne 0 ] && exit $?

# Wait a couple of seconds for handlers to initialize and start Traefik
cd ../..
sleep 2s
echo Starting Traefik...
sudo ./traefik -c examples/accessLog/traefik.example.toml &
[ $? -ne 0 ] && exit $?

echo Sample handlers and traefik started successfully!
echo 'Use command curl http://127.0.0.1:80/test{1,2,2} to drive test'
echo Then inspect log/access.log to verify it contains frontend, backend, and timing
