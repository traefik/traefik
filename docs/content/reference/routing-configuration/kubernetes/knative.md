---
title: "Traefik Knative Documentation"
description: "The Knative provider can be used for routing and load balancing in Traefik Proxy. View examples in the technical documentation."
---

# Traefik & Knative

When using the Knative provider, Traefik leverages Knative's Custom Resource Definitions (CRDs) to obtain its routing configuration. 
For detailed information on Knative concepts and resources, refer to the official [documentation](https://knative.dev/docs/).

The Knative provider supports version [v1.19.0](https://github.com/knative/serving/releases/tag/knative-v1.19.0) of the specification.

## Deploying a Knative Service

A `Service` is a core resource in the Knative specification that defines the entry point for traffic into a Knative application. 
It is linked to a `Ingress`, which specifies the Knative networking controller responsible for managing and handling the traffic, 
ensuring that it is directed to the appropriate Knative backend services.

The following `Service` manifest configures the running Traefik controller to handle the incoming traffic.

```yaml
---
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
  namespace: default
spec:
  template:
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

    The `example.com` domain is the public domain configured when deploying the Traefik controller.
    Check out [the install configuration](../../install-configuration/providers/kubernetes/knative.md) for more details.

### Tag based routing

To add tag-based routing with percentage in Knative, you can define the `traffic` section in your `Service` manifest to include different revisions with specific tags and percentages. 
Here is an example:

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
  namespace: default
spec:
  template:
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
          env:
            - name: TARGET
              value: "Go Sample v2"
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

- `http://v1-helloworld-go.default.example.com`
- `http://v2-helloworld-go.default.example.com`

Use the default URL to access percentage-based routing:

- `http://helloworld-go.default.example.com`

### HTTP/HTTPS

Check out the Knative documentation for [HTTP/HTTPS configuration](https://knative.dev/docs/serving/encryption/external-domain-tls/#configure-external-domain-encryption).

{!traefik-for-business-applications.md!}
