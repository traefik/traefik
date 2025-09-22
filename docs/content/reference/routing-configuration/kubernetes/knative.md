---
title: "Traefik Knative Documentation"
description: "The Knative provider can be used for routing and load balancing in Traefik Proxy. View examples in the technical documentation."
---

# Traefik & Knative

When using the Knative provider, Traefik leverages Knative's Custom Resource Definitions (CRDs) to obtain its routing configuration. 
For detailed information on Knative concepts and resources, refer to the official [documentation](https://knative.dev/docs/).

The Knative provider supports version [v1.17.0](https://github.com/knative/serving/releases/tag/knative-v1.17.0) of the 
specification.

It fully supports all `Service` and `Ingress` resources.

## Deploying a Knative Service

A `Service` is a core resource in the Knative specification that defines the entry point for traffic into a Knative application. 
It is linked to a `Ingress`, which specifies the controller responsible for managing and handling the traffic, ensuring that it is directed to the appropriate Knative backend services.

The following `Service` manifest configures the running Traefik controller to handle the incoming traffic.

```yaml tab="Service"
---
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
  namespace: default
spec:
  template:
    metadata:
      annotations:
        annotations:
          networking.knative.dev/ingress-class: "traefik.ingress.networking.knative.dev"
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
          env:
            - name: TARGET
              value: "Go Sample v1"
```

Once everything is deployed, sending a `GET` request to the HTTP endpoint should return the following response:

```shell
$ curl http://helloworld-go.default.example.com

Hello Go Sample v1!
```

!!! Note

    All functionalities from the [Knative documentation](https://knative.dev/docs/) are supported.

### Tag based routing
The following `Service` manifest configures the running Traefik controller to handle the incoming traffic based on 
percentage.

```yaml tab="Service"
---
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
  namespace: default
spec:
  template:
    metadata:
      annotations:
        annotations:
          networking.knative.dev/ingress-class: "traefik.ingress.networking.knative.dev"
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
          env:
            - name: TARGET
              value: "Go Sample v1"
    traffic:
    - revisionName: helloworld-go-00001
      percent: 100

```

In this example:
- The `traffic` section specifies one revision (`helloworld-go-00001`), receiving 100% of the traffic.

You can access the tagged revisions using URLs like

- `http://helloworld-go.default.example.com`.

### Tag based routing with percentage
To add tag-based routing with percentage in Knative, you can define the `traffic` section in your `Service` manifest to include different revisions with specific tags and percentages. Here is an example:

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
  namespace: default
spec:
  template:
    metadata:
      annotations:
        networking.knative.dev/ingress-class: "traefik.ingress.networking.knative.dev"
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
          env:
            - name: TARGET
              value: "Go Sample v1"
  traffic:
    - tag: v1
      revisionName: helloworld-go-00001
      percent: 50
    - tag: v2
      revisionName: helloworld-go-00002
      percent: 50
```

In this example:
- The `traffic` section specifies two revisions (`helloworld-go-00001` and `helloworld-go-00002`) with tags `v1` and `v2`, each receiving 50% of the traffic.
- The `tag` field allows you to route traffic to specific revisions using the tag.

You can access the tagged revisions using these URLs:

- `http://v1.helloworld-go.default.example.com`
- `http://v2.helloworld-go.default.example.com`

Use the default URL to access percentage-based routing:

- `http://helloworld-go.default.example.com`


### HTTP/HTTPS

The `Ingress` is a core resource in the Knative specification, designed to define how HTTP traffic should be routed 
within a Kubernetes cluster across Knative resources.
It allows the specification of routing rules that direct HTTP requests to the appropriate Knative backend services.

!!! Note "Ingress type"

    The `Service` resource should be pre-created before deploying the `Ingress` resource to ensure proper routing and traffic management.

For example, the following manifests configure a helloworld-go backend and its corresponding `Ingress`,
reachable through the deployed `Service` at the `https://helloworld-go.default.example.com` address.

```yaml tab="Ingress"
---
apiVersion: networking.internal.knative.dev/v1alpha1
kind: Ingress
metadata:
  name: helloworld-go
  namespace: default
spec:
  tls:
    - secretName: my-tls-secret
      secretNamespace: default
  rules:
    - hosts:
        - helloworld-go.default.example.com
      http:
        paths:
          - path: /
            splits:
              - backend:
                  serviceName: helloworld-go
                  serviceNamespace: default
                  servicePort: 80
                percent: 100
```

Once everything is deployed, sending a `GET` request to the HTTP endpoint should return the following response:

```shell
$ curl https://helloworld-go.default.example.com

Hello Go Sample v1!
```

{!traefik-for-business-applications.md!}
