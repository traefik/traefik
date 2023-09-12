---
title: "Traefik Getting Started With Kubernetes"
description: "Looking to get started with Traefik Proxy? Read the technical documentation to learn a simple use case that leverages Kubernetes."
---

# Quick Start

A Simple Use Case of Traefik Proxy and Kubernetes
{: .subtitle }

This guide is an introduction to using Traefik Proxy in a Kubernetes environment.
The objective is to learn how to run an application behind a Traefik reverse proxy in Kubernetes.
It presents and explains the basic blocks required to start with Traefik such as Ingress Controller, Ingresses, Deployments, static, and dynamic configuration.

## Permissions and Accesses

Traefik uses the Kubernetes API to discover running services.

In order to use the Kubernetes API, Traefik needs some permissions.
This [permission mechanism](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) is based on roles defined by the cluster administrator.
The role is then bound to an account used by an application, in this case, Traefik Proxy.

The first step is to create the role.
The [`ClusterRole`](https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRole) resource enumerates the resources and actions available for the role.
In a file called `00-role.yml`, put the following `ClusterRole`:

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
      - networking.k8s.io
    resources:
      - ingresses/status
    verbs:
      - update
```

!!! info "You can find the reference for this file [there](../../reference/dynamic-configuration/kubernetes-crd/#rbac)."

The next step is to create a dedicated service account for Traefik.
In a file called `00-account.yml`, put the following [`ServiceAccount`](https://kubernetes.io/docs/reference/kubernetes-api/authentication-resources/service-account-v1/#ServiceAccount) resource:

```yaml tab="00-account.yml"
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-account
```

And then, bind the role on the account to apply the permissions and rules on the latter. In a file called `01-role-binding.yml`, put the
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
    namespace: default # Using "default" because we did not specify a namespace when creating the ClusterAccount.
```

!!! info "`roleRef` is the Kubernetes reference to the role created in `00-role.yml`."

!!! info "`subjects` is the list of accounts reference."

    In this guide, it only contains the account created in `00-account.yml`

## Deployment and Exposition

!!! info "This section can be managed with the help of the [Traefik Helm chart](../install-traefik/#use-the-helm-chart)."

The [ingress controller](https://traefik.io/glossary/kubernetes-ingress-and-ingress-controller-101/#what-is-a-kubernetes-ingress-controller)
is a software that runs in the same way as any other application on a cluster.
To start Traefik on the Kubernetes cluster,
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
          image: traefik:v2.10
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
From here, it is possible to enable the dashboard,
configure entry points,
select dynamic configuration providers,
and [more](../reference/static-configuration/cli.md)...

In this deployment,
the static configuration enables the Traefik dashboard,
and uses Kubernetes native Ingress resources as router definitions to route incoming requests.

!!! info "When there is no entry point in the static configuration"

    Traefik creates a default one called `web` using the port `80` routing HTTP requests.

!!! info "When enabling the [`api.insecure`](../../operations/api/#insecure) mode, Traefik exposes the dashboard on the port `8080`."

A deployment manages scaling and then can create lots of containers, called [Pods](https://kubernetes.io/docs/concepts/workloads/pods/).
Each Pod is configured following the `spec` field in the deployment.
Given that, a Deployment can run multiple Traefik Proxy Pods,
a piece is required to forward the traffic to any of the instance:
namely a [`Service`](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/#Service).
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

!!! warning "It is possible to expose a service in different ways."

    Depending on your working environment and use case, the `spec.type` might change.
    It is strongly recommended to understand the available [service types](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types) before proceeding to the next step.

It is now time to apply those files on your cluster to start Traefik.

```shell
kubectl apply -f 00-role.yml \
              -f 00-account.yml \
              -f 01-role-binding.yml \
              -f 02-traefik.yml \
              -f 02-traefik-services.yml
```

## Proxying applications

The only part still missing is the business application behind the reverse proxy.
For this guide, we use the example application [traefik/whoami](https://github.com/traefik/whoami),
but the principles are applicable to any other application.

The `whoami` application is a simple HTTP server running on port 80 which answers host-related information to the incoming requests.
As usual, start by creating a file called `03-whoami.yml` and paste the following `Deployment` resource:

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

Thanks to the Kubernetes API,
Traefik is notified when an Ingress resource is created, updated, or deleted.
This makes the process dynamic.
The ingresses are, in a way, the [dynamic configuration](../../providers/kubernetes-ingress/) for Traefik.

!!! tip

    Find more information on [ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/),
    and [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) in the official Kubernetes documentation.

Create a file called `04-whoami-ingress.yml` and insert the `Ingress` resource:

```yaml tab="04-whoami-ingress.yml"
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

This `Ingress` configures Traefik to redirect any incoming requests starting with `/` to the `whoami:80` service.

At this point, all the configurations are ready.
It is time to apply those new files:

```shell
kubectl apply -f 03-whoami.yml \
              -f 03-whoami-services.yml \
              -f 04-whoami-ingress.yml
```

Now you should be able to access the `whoami` application and the Traefik dashboard.
Load the dashboard on a web browser: [`http://localhost:8080`](http://localhost:8080).

And now access the `whoami` application:

```shell
curl -v http://localhost/
```

!!! question "Going further"

    - [Filter the ingresses](../providers/kubernetes-ingress.md#ingressclass) to use with [IngressClass](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)
    - Use [IngressRoute CRD](../providers/kubernetes-crd.md)
    - Protect [ingresses with TLS](../routing/providers/kubernetes-ingress.md#enabling-tls-via-annotations)
    
{!traefik-for-business-applications.md!}
