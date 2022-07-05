---
title: "Traefik Getting Started Quickly With Kubernetes"
description: "Looking to get started with Traefik Proxy quickly? Read the technical documentation to learn a simple use case that leverages Kubernetes."
---

# Quick Start

A Simple Use Case Using Kubernetes
{: .subtitle }

Using Traefik in a Kubernetes environment could be difficult to start with.
This Getting-Started guide shows how to run an HTTP application in Kubernetes along
with Traefik as Ingress Controller.

## Permissions and Accesses

Traefik consumes the Kubernetes API to discover the running services and to unlock its power.

Before being able to consume the Kubernetes API, Traefik needs some permissions.
This permission mechanism is based on roles defines by the cluster administrator.
The role is then bound on an account used by an application, Traefik is this case.  

The first step is to create the role.
The [`ClusterRole`](https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRole) resource enumerates the resources and actions
available for the role. In a file called `00-role.yml`, put the following `ClusterRole`:

```yaml tab="00-role.yml"
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: traefik-role

rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
      - secrets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
      - networking.k8s.io
    resources:
      - ingresses
      - ingressclasses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - ingresses/status
    verbs:
      - update
```

The next step is to create an account for Traefik.
In a file called `00-account.yml`, put the following [`ClusterAccount`](https://kubernetes.io/docs/reference/kubernetes-api/authentication-resources/service-account-v1/#ServiceAccount) resource:

```yaml tab="00-account.yml"
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-account
```

And then, the last operation consists in binding the role on the account to apply the
permissions and rules on the latter. In a file called `01-role-binding.yml`, put the
following [`ClusterRoleBinding`](https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-binding-v1/#ClusterRoleBinding) resource:

```yaml tab="01-role-binding.yml"
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: traefik-role-binding

roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik-role
subjects:
  - kind: ServiceAccount
    name: traefik-account
    namespace: default # Using "default" because we didn't specify a namespace when creating the ClusterAccount.
```

!!! info
    
    `roleRef` is the kubernetes reference to the role created in `00-role.yml`.
    
    `subjects` is the list of accounts reference. In the guilde, it
    only contains the account created in `00-account.yml`

## Deployment and Exposition

The ingress controller (reverse proxy) is a software that runs in the same way as any other application on a cluster.
In order to start Traefik on the Kubernetes cluster,
a [`Deployment`](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/) resource must exist to describe how to configure
and scale containers horizontally to support larger workloads.

Start by creating a file called `02-traefik.yml` and paste the following `Deployment` resource:

```yaml tab="02-traefik.yml"
kind: Deployment
apiVersion: apps/v1
metadata:
  name: traefik-deployment
  labels:
    app: traefik

spec:
  replicas: 1
  selector:
    matchLabels:
      app: traefik
  template:
    metadata:
      labels:
        app: traefik
    spec:
      serviceAccountName: traefik-account
      containers:
        - name: traefik
          image: traefik:v2.7
          args:
            - --api.insecure
            - --providers.kubernetesingress
          ports:
            - name: web
              containerPort: 80
            - name: dashboard
              containerPort: 8080
```

The deployment contains an important attribute for customizing Traefik: `args`.
These arguments are the static configuration for Traefik.
From here, it is possible to enable the dashboard, configure entry points,
select dynamic configuration providers, and [more](../reference/static-configuration/cli.md)...

In this deployment, the static configuration enables the Traefik dashboard and uses Kubernetes native
ingress resources as router definitions to route incoming requests.

!!! info
    - If there is no entry points in the static configuration,
      Traefik creates a default one called `web` using the port `80` routing HTTP requests.
    - When enabling the `api.insecure` mode, Traefik exposes the dashboard on the port `8080`.

A deployment manages scaling and then can create a tons of containers, called Pods.
Each pod is configured following the `spec` field in the deployment.
Given that it can exist a lot of Traefik instances, a piece is required to forward the traffic
to any of the instance: a [`Service`](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/#Service).
This process is called load balancing.
Create a file called `02-traefik-services.yml` and insert the two `Service` resources:

```yaml tab="02-traefik-services.yml"
apiVersion: v1
kind: Service
metadata:
  name: traefik-dashboard-service

spec:
  type: LoadBalancer
  ports:
    - port: 8080
      targetPort: dashboard
  selector:
    app: traefik
---
apiVersion: v1
kind: Service
metadata:
  name: traefik-web-service

spec:
  type: LoadBalancer
  ports:
    - targetPort: web
      port: 80
  selector:
    app: traefik
```

!!! warning
    It is possible to expose a service in different manners. Depending on your working environment and use case,
    the `spec.type` might change. It is strongly recommended understanding the available [service types](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types).

Traefik has an account and permissions.
Traefik has a deployment.
Traefik can be exposed outside the cluster.
Traefik is ready to start.

Apply the files on your cluster to start Traefik.

```shell
kubectl apply -f 00-role.yml \
              -f 00-account.yml \
              -f 01-role-binding.yml \
              -f 02-traefik.yml \
              -f 02-traefik-services.yml
```

## Proxying applications

The missing part, now, is the business application behind the reverse-proxy.
For this guide, [traefik/whoami](https://github.com/traefik/whoami) is used,
but the concepts are equivalent for any other application.

Whoami is a simple HTTP server running on the port 80 which answers host related information
to the incoming requests. As usual, start by creating a file called `03-whoami.yml` and
paste the following `Deployment` resource:

```yaml tab="03-whoami.yml"
kind: Deployment
apiVersion: apps/v1
metadata:
  name: whoami
  labels:
    app: whoami

spec:
  replicas: 1
  selector:
    matchLabels:
      app: whoami
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
        - name: whoami
          image: traefik/whoami
          ports:
            - name: web
              containerPort: 80
```

And continue by creating the following `Service` resource in a file called `03-whoami-services.yml`:

```yaml tab="03-whoami-services.yml"
apiVersion: v1
kind: Service
metadata:
  name: whoami

spec:
  ports:
    - name: web
      port: 80
      targetPort: web
      
  selector:
    app: whoami
```

Traefik is an ingress controller.
An ingress controller is a software that understands Ingress resources and adapts to them.
Thanks to the Kubernetes API, Traefik is notified when an Ingress resource is created, updated or deleted.
It makes the process very dynamic. The ingresses are, in a way,
the [dynamic configuration](https://doc.traefik.io/traefik/providers/kubernetes-ingress/) for Traefik.

!!! tip
    Find more information on [IngressController](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) and
    [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) on the Kubernetes documentation.

Create a file called `04-whoami-ingresses.yml` and insert the `Ingress` resource:

```yaml tab="04-whoami-ingresses.yml"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: whoami-ingress
spec:
  rules:
  - http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: whoami
            port:
              name: web
```

This `Ingress` configures Traefik to redirect any incoming request starting by `/` to the `whoami:80` service.

At this point, all the configurations are ready, apply the new files:

```shell
kubectl apply -f 03-whoami.yml \
              -f 03-whoami-services.yml \
              -f 04-whoami-ingresses.yml
```

If everything is applied correctly, it should be possible to access to the `whoami` application and the traefik dashboard.
Try to load the dashboard in a web browser: `http://<external-ip>:8080`.

Then try to access to call the `whoami` application:

```shell
curl -v http://<external-ip>/
```

!!! question "Going further"
    - [Filter the ingresses](../providers/kubernetes-ingress.md#ingressclass) to use with [IngressClass](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)
    - Use [IngressRoute CRD](../providers/kubernetes-crd.md)
    - Protect [ingresses with TLS](../routing/providers/kubernetes-ingress.md#enabling-tls-via-annotations)
