# Rest Provider

Traefik can be configured:

- using a RESTful api.

## Configuration

```toml
# Enable REST Provider.
[rest]
  # Name of the related entry point
  #
  # Optional
  # Default: "traefik"
  #
  entryPoint = "traefik"
```

## API

| Path                         | Method | Description     |
|------------------------------|--------|-----------------|
| `/api/providers/web`         | `PUT`  | update provider |
| `/api/providers/rest`        | `PUT`  | update provider |

!!! warning
    For compatibility reason, when you activate the rest provider, you can use `web` or `rest` as `provider` value.


```shell
curl -XPUT -d @file "http://localhost:8080/api/providers/rest"
```

with `@file`:
```json
{
    "frontends": {
      "frontend2": {
        "routes": {
          "test_2": {
            "rule": "Path:/test"
          }
        },
        "backend": "backend1"
      },
      "frontend1": {
        "routes": {
          "test_1": {
            "rule": "Host:test.localhost"
          }
        },
        "backend": "backend2"
      }
    },
    "backends": {
      "backend2": {
        "loadBalancer": {
          "method": "drr"
        },
        "servers": {
          "server2": {
            "weight": 2,
            "URL": "http://172.17.0.5:80"
          },
          "server1": {
            "weight": 1,
            "url": "http://172.17.0.4:80"
          }
        }
      },
      "backend1": {
        "loadBalancer": {
          "method": "wrr"
        },
        "circuitBreaker": {
          "expression": "NetworkErrorRatio() > 0.5"
        },
        "servers": {
          "server2": {
            "weight": 1,
            "url": "http://172.17.0.3:80"
          },
          "server1": {
            "weight": 10,
            "url": "http://172.17.0.2:80"
          }
        }
      }
    }
}
```
