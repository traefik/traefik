# Clustering / High Availability (beta)

This guide explains how to use Træfik in high availability mode.

In order to deploy and configure multiple Træfik instances, without copying the same configuration file on each instance, we will use a distributed Key-Value store.

## Prerequisites

You will need a working KV store cluster.
_(Currently, we recommend [Consul](https://consul.io) .)_

## File configuration to KV store migration

We created a special Træfik command to help configuring your Key Value store from a Træfik TOML configuration file.

Please refer to [this section](../kv-config/#store-configuration-in-key-value-store) to get more details.

## Deploy a Træfik cluster

Once your Træfik configuration is uploaded on your KV store, you can start each Træfik instance.

A Træfik cluster is based on a manager/worker model.

When starting, Træfik will elect a manager.
If this instance fails, another manager will be automatically elected.

## Træfik cluster and Let's Encrypt

**In cluster mode, ACME certificates have to be stored in [a KV Store entry](../../configuration/acme/#as-a-key-value-store-entry).**

Thanks to the Træfik cluster mode algorithm (based on [the Raft Consensus Algorithm](https://raft.github.io/)), only one instance will contact Let's encrypt to solve the challenges.

The others instances will get ACME certificate from the KV Store entry.
