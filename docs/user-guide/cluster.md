# Clustering / High Availability (beta)

This guide explains how to use Traefik in high availability mode.

In order to deploy and configure multiple Traefik instances, without copying the same configuration file on each instance, we will use a distributed Key-Value store.

## Prerequisites

You will need a working KV store cluster.
_(Currently, we recommend [Consul](https://consul.io) .)_

## File configuration to KV store migration

We created a special Traefik command to help configuring your Key Value store from a Traefik TOML configuration file.

Please refer to [this section](/user-guide/kv-config/#store-configuration-in-key-value-store) to get more details.

## Deploy a Traefik cluster

Once your Traefik configuration is uploaded on your KV store, you can start each Traefik instance.

A Traefik cluster is based on a manager/worker model.

When starting, Traefik will elect a manager.
If this instance fails, another manager will be automatically elected.

## Traefik cluster and Let's Encrypt

**In cluster mode, ACME certificates have to be stored in [a KV Store entry](/configuration/acme/#as-a-key-value-store-entry).**

Thanks to the Traefik cluster mode algorithm (based on [the Raft Consensus Algorithm](https://raft.github.io/)), only one instance will contact Let's encrypt to solve the challenges.

The others instances will get ACME certificate from the KV Store entry.
