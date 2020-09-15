# Traefik & CRD & Let's Encrypt

Traefik with an IngressRoute Custom Resource Definition for Kubernetes, and TLS Through Let's Encrypt.
{: .subtitle }

This document is intended to be a fully working example demonstrating how to set up Traefik in [Kubernetes](https://kubernetes.io),
with the dynamic configuration coming from the [IngressRoute Custom Resource](../../providers/kubernetes-crd.md),
and TLS setup with [Let's Encrypt](https://letsencrypt.org).
However, for the sake of simplicity, we're using [k3s](https://github.com/rancher/k3s)  docker image for the Kubernetes cluster setup.

Please note that for this setup, given that we're going to use ACME's TLS-ALPN-01 challenge, the host you'll be running it on must be able to receive connections from the outside on port 443.
And of course its internet facing IP address must match the domain name you intend to use.

In the following, the Kubernetes resources defined in YAML configuration files can be applied to the setup in two different ways:

- the first, and usual way, is simply with the `kubectl apply` command.
- the second, which can be used for this tutorial, is to directly place the files in the directory used by the k3s docker image for such inputs (`/var/lib/rancher/k3s/server/manifests`).

!!! important "Kubectl Version"

    With the `rancher/k3s` version used in this guide (`0.8.0`), the kubectl version needs to be >= `1.11`.

## k3s Docker-compose Configuration

Our starting point is the docker-compose configuration file, to start the k3s cluster.
You can start it with:

```bash
docker-compose -f k3s.yml up
```

```yaml
--8<-- "content/user-guides/crd-acme/k3s.yml"
```

## Cluster Resources

Let's now have a look (in the order they should be applied, if using `kubectl apply`) at all the required resources for the full setup.

### IngressRoute Definition

First, the definition of the `IngressRoute` and the `Middleware` kinds.
Also note the RBAC authorization resources; they'll be referenced through the `serviceAccountName` of the deployment, later on.

```yaml
--8<-- "content/reference/dynamic-configuration/kubernetes-crd-definition.yml"

---
--8<-- "content/reference/dynamic-configuration/kubernetes-crd-rbac.yml"
```

### Services

Then, the services. One for Traefik itself, and one for the app it routes for, i.e. in this case our demo HTTP server: [whoami](https://github.com/traefik/whoami).

```yaml
--8<-- "content/user-guides/crd-acme/02-services.yml"
```

### Deployments

Next, the deployments, i.e. the actual pods behind the services.
Again, one pod for Traefik, and one for the whoami app.

```yaml
--8<-- "content/user-guides/crd-acme/03-deployments.yml"
```

### Port Forwarding

Now, as an exception to what we said above, please note that you should not let the ingressRoute resources below be applied automatically to your cluster.
The reason is, as soon as the ACME provider of Traefik detects we have TLS routers, it will try to generate the certificates for the corresponding domains.
And this will not work, because as it is, our Traefik pod is not reachable from the outside, which will make the ACME TLS challenge fail.
Therefore, for the whole thing to work, we must delay applying the ingressRoute resources until we have port-forwarding set up properly, which is the next step.

```bash
kubectl port-forward --address 0.0.0.0 service/traefik 8000:8000 8080:8080 443:4443 -n default
```

Also, and this is out of the scope if this guide, please note that because of the privileged ports limitation on Linux, the above command might fail to listen on port 443.
In which case you can use tricks such as elevating caps of `kubectl` with `setcaps`, or using `authbind`, or setting up a NAT between your host and the WAN.
Look it up.

### Traefik Routers

We can now finally apply the actual ingressRoutes, with:

```bash
kubectl apply -f 04-ingressroutes.yml
```

```yaml
--8<-- "content/user-guides/crd-acme/04-ingressroutes.yml"
```

Give it a few seconds for the ACME TLS challenge to complete, and you should then be able to access your whoami pod (routed through Traefik), from the outside.
Both with or (just for fun, do not do that in production) without TLS:

```bash
curl [-k] https://your.example.com/tls
```

```bash
curl http://your.example.com:8000/notls
```

Note that you'll have to use `-k` as long as you're using the staging server of Let's Encrypt, since it is not an authorized certificate authority on systems where it hasn't been manually added.
