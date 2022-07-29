---
title: "Traefik TLS on Kubernetes Documentation"
description: "Learn how to use Traefik Proxy on Kubernetes with TLS certificates from Let's Encrypt."
---

# Traefik & TLS on Kubernetes

This document is intended to be a complete guide demonstrating how to set up Traefik
on [Kubernetes](https://kubernetes.io) 1.18+, with dynamic configuration coming either from regular
Kubernetes [Ingress](../../providers/kubernetes-ingress.md) or from Traefik
[IngressRoute Custom Resource](../../providers/kubernetes-crd.md), and TLS setup with
[Let's Encrypt](https://letsencrypt.org).

Gateway API is also [supported](../../providers/kubernetes-gateway.md) on Traefik, as an
experimental feature. It is not covered by this guide.

In order to ease understanding, this user-guide is using `default` namespace.

## Pre-requisites

In order to configure TLS on Kubernetes, you'll need a Kubernetes cluster, a web application and Traefik.

### Kubernetes Cluster

First, you'll need a Kubernetes Cluster up & running. If you don't have one or
want to test locally, you can use [kind](https://kind.sigs.k8s.io/).

To forward http requests to `kind`, you'll need to configure network ports:

```bash
wget https://raw.githubusercontent.com/traefik/traefik/master/docs/content/user-guides/crd-acme/kind.config
kind create cluster --config=kind.config
```

```yaml
--8<-- "content/user-guides/crd-acme/kind.config"
```

One can see it takes around 30 seconds before it is ready.

```bash
$ kubectl get nodes
NAME                 STATUS     ROLES                  AGE   VERSION
kind-control-plane   NotReady   control-plane,master   11s   v1.21.1
# wait 30 seconds
$ kubectl get nodes
NAME                 STATUS   ROLES                  AGE   VERSION
kind-control-plane   Ready    control-plane,master   38s   v1.21.1
```

### Web application

You need a web application to secure. In this user-guide, we'll use [whoami](https://github.com/traefik/whoami), a demo HTTP server.

```bash
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.8/docs/content/user-guides/crd-acme/whoami.yml
```

```yaml
--8<-- "content/user-guides/crd-acme/whoami.yml"
```

### Traefik

Traefik can be installed with its own set of CRDs in order to ease configuration and
provides more features, like routing on different headers.

**CRDs & RBAC**

First, you will need to install Traefik CRDs and the RBAC authorization resources which will
be referenced through the `serviceAccountName` of the deployment.

In those files, the `ServiceAccount` is named `traefik-ingress-controller`.

```bash
# Install Traefik Resource Definitions:
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.8/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml
```
```yaml
customresourcedefinition.apiextensions.k8s.io/ingressroutes.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/ingressroutetcps.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/ingressrouteudps.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/middlewares.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/middlewaretcps.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/serverstransports.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/tlsoptions.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/tlsstores.traefik.containo.us created
customresourcedefinition.apiextensions.k8s.io/traefikservices.traefik.containo.us created
```
```bash
# Install RBAC and ServiceAccount for Traefik:
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.8/docs/content/reference/dynamic-configuration/kubernetes-crd-rbac.yml
```
```yaml
clusterrole.rbac.authorization.k8s.io/traefik-ingress-controller created
clusterrolebinding.rbac.authorization.k8s.io/traefik-ingress-controller created
serviceaccount/traefik-ingress-controller created
```

**Static Configuration**

We needs to enable Kubernetes [dynamic configuration providers](../../providers/overview.md)
and define the [entrypoints](../../routing/entrypoints.md) Traefik will listen to.

This static configuration also enables you to change log level verbosity and to see what is
current configuration with Traefik dashboard.

```bash
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.8/docs/content/user-guides/crd-acme/01-static-config.yml
```

```yaml
--8<-- "content/user-guides/crd-acme/01-static-config.yml"
```

**Deployment**

Next, the deployments, i.e. the actual pods behind the services.

```bash
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.8/docs/content/user-guides/crd-acme/02-deployment.yml
```

```yaml
--8<-- "content/user-guides/crd-acme/02-deployment.yml"
```

**Service**

For `traefik`, the service needs to be available outside of Kubernetes cluster.

* With `kind`, one can use `type: NodePort` service.
* With cloud providers, one can use `type: LoadBalancer` service. It usually creates a Load Balancer on the cloud provider.

!!! info

      On a cloud provider, you'll have to adapt `traefik` service. Please refer to their documentation.

The following is working on a local setup with `kind`.

```bash
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.8/docs/content/user-guides/crd-acme/03-service.yml
```

```yaml
--8<-- "content/user-guides/crd-acme/03-service.yml"
```

**Routing**

On Kubernetes, with Traefik you can either use regular `Ingress` or more powerful
Traefik `IngressRoute` for routing.

In this user-guide, we'll use `IngressRoute` with `basicAuth` middleware for
Traefik `dashboard` and both `Ingress` and `IngressRoute` for `whoami`.

!!! info

    Before applying this, you need to have a dns record pointing to your public IP.


```yaml
--8<-- "content/user-guides/crd-acme/04-routing.yml"
```


## Configure a TLS provider

With a Kubernetes cluster, a web application and Traefik up & running, you can now improve this setup with TLS.

One can use either Let's encrypt certificates provided by Traefik or use a more versatile
dedicated component like `cert-manager`.

### Let's encrypt

You'll need to configure a [certificate resolver](../../https/acme.md) in the static configuration.

This user guide shows how to configure dns challenge with `cloudflare`. There are many
[providers](../../https/acme.md#providers) supported, and you can also use TLS or HTTP challenge.

```diff
--- 01-static-config.ym
+++ tls/01-static-config.yml
@@ -10,6 +10,9 @@
         address: ':80'
       websecure:
         address: ':443'
+        http:
+          tls:
+            certResolver: myresolver

     log:
       level: INFO
@@ -20,6 +23,11 @@

     providers:
       kubernetesCRD: {}
+
+    certificatesResolvers:
+      myresolver:
+        acme:
+          email: you@example.com
+          storage: acme.json
+          dnsChallenge:
+            provider: cloudflare
```
```diff
--- 02-deployment.yml
+++ tls/02-deployment.yml
@@ -19,6 +19,9 @@
           image: traefik:v2.8
           args:
             - --configfile=/config/static.yaml
+          env:
+            - name: CLOUDFLARE_DNS_API_TOKEN
+              value: YOUR_API_TOKEN
           ports:
             - name: web
               containerPort: 80
```

After traefik restart, you should see in the logs this new provider:
```bash
kubectl logs -f -l app=traefik | grep acme
```
```yaml
level=info msg="Starting provider *acme.ChallengeTLSALPN"
level=info msg="Starting provider *acme.Provider"
level=info msg="Testing certificate renew..." providerName=myresolver.acme ACME CA="https://acme-v02.api.letsencrypt.org/directory"
level=info msg=Register... providerName=myresolver.acme
```

!!! warning

    This deployment is simplified.
    On production, you should follow Kubernetes security recommandations and secure your API
    token, at least in a `secret`. You may also deploy Traefik on a statefulset, in order to
    keep your Certificates when pod restart.

**Troubleshooting**

One can test configuration of Let's Encrypt DNS provider with lego CLI.
With `cloudflare`, one can use this command:

```bash
$ docker run -it -e CLOUDFLARE_DNS_API_TOKEN="YOUR_API_TOKEN" \
    goacme/lego --email you@example.com \
    --dns cloudflare --domains whoami.example.com run
```
```yaml
No key found for account you@example.com. Generating a P256 key.
Saved key to /.lego/accounts/acme-v02.api.letsencrypt.org/you@example.com/keys/you@example.com.key
Please review the TOS at https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf
Do you accept the TOS? Y/n
Y
[INFO] acme: Registering account for you@example.com
!!!! HEADS UP !!!!

Your account credentials have been saved in your Let's Encrypt
configuration directory at "/.lego/accounts".

You should make a secure backup of this folder now. This
configuration directory will also contain certificates and
private keys obtained from Let's Encrypt so making regular
backups of this folder is ideal.
[INFO] [whoami.traefik.win] acme: Obtaining bundled SAN certificate
[INFO] [whoami.traefik.win] AuthURL: https://acme-v02.api.letsencrypt.org/acme/authz-v3/135595617346
[INFO] [whoami.traefik.win] acme: Could not find solver for: tls-alpn-01
[INFO] [whoami.traefik.win] acme: Could not find solver for: http-01
[INFO] [whoami.traefik.win] acme: use dns-01 solver
[INFO] [whoami.traefik.win] acme: Preparing to solve DNS-01
[INFO] cloudflare: new record for whoami.traefik.win, ID 0fcf5063d00a18764ed9894bb2f5226f
[INFO] [whoami.traefik.win] acme: Trying to solve DNS-01
[INFO] [whoami.traefik.win] acme: Checking DNS record propagation using [10.3.0.1:53]
[INFO] Wait for propagation [timeout: 2m0s, interval: 2s]
[INFO] [whoami.traefik.win] acme: Waiting for DNS record propagation.
[INFO] [whoami.traefik.win] acme: Waiting for DNS record propagation.
[INFO] [whoami.traefik.win] The server validated our request
[INFO] [whoami.traefik.win] acme: Cleaning DNS-01 challenge
[INFO] [whoami.traefik.win] acme: Validations succeeded; requesting certificates
[INFO] [whoami.traefik.win] Server responded with a certificate.
```

#### Use it

To use this provider, you just have to add the `websecure` entrypoints on `IngressRoute` or `Ingress`.

**IngressRoute**

```diff
--- 04-routing.yml
+++ tls/04-routing.yml
@@ -40,6 +40,7 @@
 spec:
   entryPoints:
     - web
+    - websecure
   routes:
   - match: Host(`whoami.example.com`) && PathPrefix(`/whoami`)
     kind: Rule
```

**Ingress**

To use this provider, you just have to add the `websecure` entrypoints on `Ingress`.

```diff
--- 04-routing.yml
+++ tls/04-routing.yml
@@ -53,6 +54,7 @@
   name: whoami
   annotations:
     traefik.ingress.kubernetes.io/router.entrypoints: web
+    traefik.ingress.kubernetes.io/router.entrypoints: websecure
 spec:
  rules:
   - host: whoami.example.com
```

Give it a few seconds for the ACME provider to complete, and you should then be
able to access your whoami pod (routed through Traefik), from the outside.

```bash
curl [-k] https://whoami.example.com/whoami
curl [-k] https://whoami.example.com/whoami-ingress
```

Note that you'll have to use `-k` if you're using the staging server of Let's Encrypt,
since it is not an authorized certificate authority on systems where it hasn't been manually added.

### Cert Manager

cert-manager is a powerful and extensible X.509 certificate controller. It can be used
with Traefik quite easily.

**Installation**

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
```

It will install all the software components of cert-manager in `cert-manager` namespace:
```yaml
namespace/cert-manager created
customresourcedefinition.apiextensions.k8s.io/certificaterequests.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/certificates.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/challenges.acme.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/clusterissuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/issuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/orders.acme.cert-manager.io created
serviceaccount/cert-manager-cainjector created
serviceaccount/cert-manager created
serviceaccount/cert-manager-webhook created
configmap/cert-manager-webhook created
clusterrole.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
clusterrole.rbac.authorization.k8s.io/cert-manager-view created
clusterrole.rbac.authorization.k8s.io/cert-manager-edit created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-approve:cert-manager-io created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-certificatesigningrequests created
clusterrole.rbac.authorization.k8s.io/cert-manager-webhook:subjectaccessreviews created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-approve:cert-manager-io created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-certificatesigningrequests created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-webhook:subjectaccessreviews created
role.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
role.rbac.authorization.k8s.io/cert-manager:leaderelection created
role.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
rolebinding.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
rolebinding.rbac.authorization.k8s.io/cert-manager:leaderelection created
rolebinding.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
service/cert-manager created
service/cert-manager-webhook created
deployment.apps/cert-manager-cainjector created
deployment.apps/cert-manager created
deployment.apps/cert-manager-webhook created
mutatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created
validatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created
```

After, you can configure `cert-manager`. We'll use `ClusterIssuer` in order to provide
certificates on multiple namespaces. With a dns challenge on CloudFlare, one can
configure cert-manager with this kind of `ClusterIssuer`:

```yaml
--8<-- "content/user-guides/crd-acme/cm/issuer.yml"
```

#### Use it

**Generate your certificate**

For `Ingress`, you'll need first to create your certificate :

```yaml
--8<-- "content/user-guides/crd-acme/cm/certificate.yml"
```

This certificate will be stored in `whoami-tls` secret.
Once your `Certificate` is **Ready**, you can add it on your `Ingress` or your `IngressRoute`

**IngressRoute**

```diff
--- 04-routing.yml
+++ cm/04-routing.yml
@@ -46,6 +46,8 @@
     services:
     - name: whoami
       port: 80
+  tls:
+    secretName: whoami-tls
 ---
 apiVersion: networking.k8s.io/v1
 kind: Ingress
```

**Ingress**

```diff
@@ -65,3 +67,5 @@
             name: whoami
             port:
               number: 80
+  tls:
+  - secretName: whoami-tls
```

## Configure TLS

### Force TLS v1.2+

Nowadays, TLS v1.0 and v1.1 are deprecated.
In order to force TLS v1.2 or later on all URLs with TLS, you can define
the `default` [TLSOption](../../routing/providers/kubernetes-crd.md#kind-tlsoption):

```bash
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.8/docs/content/user-guides/crd-acme/05-tlsoption.yml
```

```yaml
--8<-- "content/user-guides/crd-acme/05-tlsoption.yml"
```

You're starting here to see Traefik magic : you were able to change configuration
dynamically without losing any connection.

### Force HTTPS usage

You can force HTTPS on a specific URL using a Redirect middleware or even for everything,
if you do it on entrypoints. A convenient way to be sure that every request will use HTTPS
is to configure the redirection at the entrypoint level, in static configuration.

```diff
--- 01-static-config.yml
+++ 05-force-https.yml
@@ -8,6 +8,11 @@
     entryPoints:
       web:
         address: ':80'
+        http:
+          redirections:
+            entryPoint:
+              to: websecure
+              scheme: https
       websecure:
         address: ':443'

```

After a restart of Traefik, you can confirm the redirection with curl:

```bash
curl -v http://whoami.example.com/whoami
```
```yaml
*   Trying 127.0.0.1:80...
* Connected to whoami.example.com (127.0.0.1) port 80 (#0)
> GET /whoami HTTP/1.1
> Host: whoami.example.com
> User-Agent: curl/7.74.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 301 Moved Permanently
< Location: https://whoami.traefik.win/whoami
< Content-Length: 17
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host whoami.example.com left intact
Moved Permanently
```
