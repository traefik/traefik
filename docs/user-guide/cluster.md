# Clustering / High Availability

This guide explains how tu use Træfɪk in high availability mode.
In order to deploy and configure multiple Træfɪk instances, without copying the same configuration file on each instance, we will use a distributed Key-Value store.

## Prerequisites

You will need a working KV store cluster.

## File configuration to KV store migration

We created a special Træfɪk command to help configuring your Key Value store from a Træfɪk TOML configuration file.
Please refer to [this section](/user-guide/kv-config/#store-configuration-in-key-value-store) to get more details.

## Deploy a Træfɪk cluster

Once your Træfɪk configuration is uploaded on your KV store, you can start each Træfɪk instance.
A Træfɪk cluster is based on a master/slave model. When starting, Træfɪk will elect a master. If this instance fails, another master will be automatically elected.
 