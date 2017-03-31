# Clustering / High Availability

This guide explains how tu use Træfik in high availability mode.
In order to deploy and configure multiple Træfik instances, without copying the same configuration file on each instance, we will use a distributed Key-Value store.

## Prerequisites

You will need a working KV store cluster.

## File configuration to KV store migration

We created a special Træfik command to help configuring your Key Value store from a Træfik TOML configuration file.
Please refer to [this section](/user-guide/kv-config/#store-configuration-in-key-value-store) to get more details.

## Deploy a Træfik cluster

Once your Træfik configuration is uploaded on your KV store, you can start each Træfik instance.
A Træfik cluster is based on a master/slave model. When starting, Træfik will elect a master. If this instance fails, another master will be automatically elected.
 