---
title: 'Providing Dynamic Configuration to Traefik'
description: 'Learn about the different methods for providing dynamic configuration to Traefik. Read the technical documentation.'
---

# Providing Dynamic Configuration to Traefik

Traefik relies on dynamic configuration to discover the services it needs to route traffic to. There are several ways to provide this dynamic configuration, depending on your environment and preferences. This guide covers the different methods available:

- File Provider: Use TOML or YAML files.
- Docker and ECS Providers: Use container labels.
- Other Providers: Use tags or Annotations

## Using the File Provider

The File provider allows you to define dynamic configuration in static files using either TOML or YAML syntax. This method is ideal for environments where services are not automatically discovered or when you prefer to manage configurations manually.

### Enabling the File Provider

To enable the File provider, add the following to your Traefik static configuration:

```yaml tab="YAML"
providers:
  file:
    directory: "/path/to/dynamic/conf"
```

```toml tab="TOML"
[providers.file]
  filename = "/path/to/your/configuration/file.toml"
```

Example using the file provider to declare routers & services:

```yaml tab="File (YAML)"
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      service: my-service

  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://localhost:8080"
```

```toml tab="File (TOML)"
[http]
  [http.routers]
    [http.routers.my-router]
      rule = "Host(`example.com`)"
      service = "my-service"

  [http.services]
    [http.services.my-service.loadBalancer]
      [[http.services.my-service.loadBalancer.servers]]
        url = "http://localhost:8080"
```

## Using Labels With Docker and ECS

When using Docker or Amazon ECS, you can define dynamic configuration using container labels. This method allows Traefik to automatically discover services and apply configurations without the need for additional files.

??? example "Example with Docker"

    When deploying a Docker container, you can specify labels to define routing rules and services:

    ```yaml
    version: '3'

    services:
      my-service:
        image: my-image
        labels:
          - "traefik.http.routers.my-router.rule=Host(`example.com`)"
          - "traefik.http.services.my-service.loadbalancer.server.port=80"
    ```

??? example "Example with ECS"

    In ECS, you can use task definition labels to achieve the same effect:

    ```yaml
    {
      "containerDefinitions": [
        {
          "name": "my-service",
          "image": "my-image",
          "dockerLabels": {
            "traefik.http.routers.my-router.rule": "Host(`example.com`)",
            "traefik.http.services.my-service.loadbalancer.server.port": "80"
          }
        }
      ]
    }
    ```

## Using Tags with Other Providers

For providers that do not support labels, such as Consul or Nomad, you can use tags or annotations to provide dynamic configuration.

??? example "Example with Consul"

    When using Consul, you can register services with specific tags that Traefik will recognize:

    ```json
    {
    "Name": "my-service",
    "Tags": [
      "traefik.http.routers.my-router.rule=Host(`example.com`)",
      "traefik.http.services.my-service.loadbalancer.server.port=80"
    ],
    "Address": "localhost",
    "Port": 8080
    }
    ```

??? example "Example with Kubernetes Ingress Annotations"

    ```yaml
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: whoami
      namespace: apps
      annotations:
        traefik.ingress.kubernetes.io/router.entrypoints: websecure
        traefik.ingress.kubernetes.io/router.priority: "42"
        traefik.ingress.kubernetes.io/router.tls: "true"
        traefik.ingress.kubernetes.io/router.tls.options: apps-opt@kubernetescrd
    spec:
      rules:
        - host: my-domain.example.com
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name:  whoami
                    namespace: apps
                    port:
                      number: 80
      tls:
      - secretName: supersecret    
    ```
