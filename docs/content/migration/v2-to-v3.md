---
title: "Traefik V3 Migration Documentation"
description: "Migrate from Traefik Proxy v2 to v3 and update all the necessary configurations to take advantage of all the improvements. Read the technical documentation."
---

# Migration Guide: From v2 to v3

How to Migrate from Traefik v2 to Traefik v3.
{: .subtitle }

This guide shows you how to progressively migrate from Traefik v2 to v3.

## Step 1: Identify Changes in Static Configuration and Test v3
To check the changes in [static configuration](https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#static-configuration "Link to static configuration changes") brought by Traefik v3, follow the link: 

Once you have prepared your static configuration, add the following snippet to it:

```
# static configuration
core:
  defaultRuleSyntax: v2
```

The snippet makes static configuration use the default [v2 syntax](https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/?ref=traefik.io#configure-the-default-syntax-in-static-configuration "Link to configure default syntax in static config").

Start Traefik v3 with this new configuration to test it.

!!! info
    If you don’t get any error logs while testing, you are good to go!
    Otherwise, follow the remaining migration options highlighted in the logs.

Once your Traefik test instances are starting and routing to your applications, proceed to the next step.

## Step 2: Migrate Production Instances to Traefik v3
Use the Kubernetes [rolling update mechanism](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/ "Link to the Kubernetes rolling update documentation") to progressively migrate your production instances to v3.

!!! Warning
    Ensure you have a [real-time monitoring solution](https://traefik.io/blog/capture-traefik-metrics-for-apps-on-kubernetes-with-prometheus/ "Link to the blog on capturing Traefik metrics with Prometheus") for your ingress traffic to detect issues instantly.

While the rolling update is in progress, monitor your ingress traffic for any errors. Be prepared to rollback to a working state in case of any issues.

!!! Info
    If you encounter any issues, leverage debug and access logs provided by Traefik to understand what went wrong and how to fix it.

Once every pod is updated, you will be on Traefik v3!

## Step 3: Progressively Migrate Ingress Resources

!!! info
    You can do this later, as Traefik v3 is compatible with the v2 format for [dynamic configuration](https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#dynamic-configuration "Link to dynamic configuration changes").

[Switch each router to the v3 syntax](https://doc.traefik.io/traefik/v3.0/migration/v2-to-v3/#configure-the-syntax-per-router "Link to configuring the syntax per router") progressively.
Test and update each Ingress resource and ensure that ingress traffic is not impacted.

Once a v3 Ingress resource migration is validated, deploy the resource and delete the v2 Ingress resource.

Remove the snippet added to the static configuration in Step 1.

You are now fully migrated to Traefik v3 🎉
