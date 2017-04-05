#!/usr/bin/env sh

trap 'echo "Caught SIGTERM"' TERM

sleep 10000 &
# Waiting for SIGTERM
wait $!

# Remove plugins
for plugin in "$@"
do
    echo "Removing plugin ${plugin}"
    rm $plugin
done