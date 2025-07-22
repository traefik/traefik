---
title: "Traefik V3 Migration Documentation"
description: "Migrate from Traefik Proxy v2 to v3 and update all the necessary configurations to take advantage of all the improvements. Read the technical documentation."
---

# Migration Guide: From v2 to v3

How to Migrate from Traefik v2 to Traefik v3.
{: .subtitle }

With Traefik v3, we are introducing a streamlined transition process from v2. Minimal breaking changes have been made to specific options in the [static configuration](./v2-to-v3-details.md#static-configuration-changes "Link to static configuration changes"), and we are ensuring backward compatibility with v2 syntax in the [dynamic configuration](./v2-to-v3-details.md#dynamic-configuration-changes "Link to dynamic configuration changes"). This will offer a gradual path for adopting the v3 syntax, allowing users to progressively migrate their Kubernetes ingress resources, Docker labels, etc., to the new format.

Here are the steps to progressively migrate from Traefik v2 to v3:

1. [Prepare configurations and test v3](#step-1-prepare-configurations-and-test-v3)
1. [Migrate production instances to Traefik v3](#step-2-migrate-production-instances-to-traefik-v3)
1. [Progressively migrate dynamic configuration](#step-3-progressively-migrate-dynamic-configuration)

## Step 1: Prepare Configurations and Test v3

Check the changes in [static configurations](./v2-to-v3-details.md#static-configuration-changes "Link to static configuration changes") and [operations](./v2-to-v3-details.md#operations-changes "Link to operations changes") brought by Traefik v3.
Modify your configurations accordingly.

Then, add the following snippet to the static configuration:

```yaml
# static configuration
core:
  defaultRuleSyntax: v2
```

This snippet in the static configuration makes the [v2 format](../migration/v2-to-v3-details.md#configure-the-default-syntax-in-static-configuration "Link to configure default syntax in static config") the default rule matchers syntax.

Start Traefik v3 with this new configuration to test it.

If you don’t get any error logs while testing, you are good to go!
Otherwise, follow the remaining migration options highlighted in the logs.

Once your Traefik test instances are starting and routing to your applications, proceed to the next step.

## Step 2: Migrate Production Instances to Traefik v3

We strongly advise you to follow a progressive migration strategy ([Kubernetes rolling update mechanism](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/ "Link to the Kubernetes rolling update documentation"), for example) to migrate your production instances to v3.

!!! Warning
    Ensure you have a [real-time monitoring solution](https://traefik.io/blog/capture-traefik-metrics-for-apps-on-kubernetes-with-prometheus/ "Link to the blog on capturing Traefik metrics with Prometheus") for your ingress traffic to detect issues instantly.

During the progressive migration, monitor your ingress traffic for any errors. Be prepared to rollback to a working state in case of any issues.

If you encounter any issues, leverage debug and access logs provided by Traefik to understand what went wrong and how to fix it.

Once every Traefik instance is updated, you will be on Traefik v3!

## Step 3: Progressively Migrate Dynamic Configuration

!!! info
    This step can be done later in the process, as Traefik v3 is compatible with the v2 format for [dynamic configuration](./v2-to-v3-details.md#dynamic-configuration-changes "Link to dynamic configuration changes").
    Enable Traefik logs to get some help if any deprecated option is in use.

Check the changes in [dynamic configuration](./v2-to-v3-details.md#dynamic-configuration-changes "Link to dynamic configuration changes").

Then, progressively [switch each router to the v3 syntax](./v2-to-v3-details.md#configure-the-syntax-per-router "Link to configuring the syntax per router").

Test and update each Ingress resource and ensure that ingress traffic is not impacted.

Once a v3 Ingress resource migration is validated, deploy the resource and delete the v2 Ingress resource.
Repeat it until all Ingress resources are migrated.

Now, remove the following snippet added to the static configuration in Step 1:

```yaml
# static configuration
core:
  defaultRuleSyntax: v2
```

You are now fully migrated to Traefik v3 🎉
