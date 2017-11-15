# API Definition

```toml
# API definition
[api]
  # Name of the related entry point
  #
  # Optional
  # Default: "traefik"
  #
  entryPoint = "traefik"
  
  # Enabled Dashboard
  #
  # Optional
  # Default: true
  #
  dashboard = true
  
  # Enabled debug mode
  #
  # Optional
  # Default: false
  #
  debug = true
```

## Web UI

![Web UI Providers](/img/web.frontend.png)

![Web UI Health](/img/traefik-health.png)

## API

| Path                                                            | Method           | Description                               |
|-----------------------------------------------------------------|------------------|-------------------------------------------|
| `/`                                                             |     `GET`        | Provides a simple HTML frontend of Træfik |
| `/health`                                                       |     `GET`        | json health metrics                       |
| `/api`                                                          |     `GET`        | Configuration for all providers           |
| `/api/providers`                                                |     `GET`        | Providers                                 |
| `/api/providers/{provider}`                                     |     `GET`, `PUT` | Get or update provider                    |
| `/api/providers/{provider}/backends`                            |     `GET`        | List backends                             |
| `/api/providers/{provider}/backends/{backend}`                  |     `GET`        | Get backend                               |
| `/api/providers/{provider}/backends/{backend}/servers`          |     `GET`        | List servers in backend                   |
| `/api/providers/{provider}/backends/{backend}/servers/{server}` |     `GET`        | Get a server in a backend                 |
| `/api/providers/{provider}/frontends`                           |     `GET`        | List frontends                            |
| `/api/providers/{provider}/frontends/{frontend}`                |     `GET`        | Get a frontend                            |
| `/api/providers/{provider}/frontends/{frontend}/routes`         |     `GET`        | List routes in a frontend                 |
| `/api/providers/{provider}/frontends/{frontend}/routes/{route}` |     `GET`        | Get a route in a frontend                 |

!!! warning
    For compatibility reason, when you activate the rest provider, you can use `web` or `rest` as `provider` value.
    But be careful, in the configuration for all providers the key is still `web`.

### Provider configurations

```shell
curl -s "http://localhost:8080/api" | jq .
```
```json
{
  "file": {
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
}
```

### Health

```shell
curl -s "http://localhost:8080/health" | jq .
```
```json
{
  // Træfik PID
  "pid": 2458,
  // Træfik server uptime (formated time)
  "uptime": "39m6.885931127s",
  //  Træfik server uptime in seconds
  "uptime_sec": 2346.885931127,
  // current server date
  "time": "2015-10-07 18:32:24.362238909 +0200 CEST",
  // current server date in seconds
  "unixtime": 1444235544,
  // count HTTP response status code in realtime
  "status_code_count": {
    "502": 1
  },
  // count HTTP response status code since Træfik started
  "total_status_code_count": {
    "200": 7,
    "404": 21,
    "502": 13
  },
  // count HTTP response
  "count": 1,
  // count HTTP response
  "total_count": 41,
  // sum of all response time (formated time)
  "total_response_time": "35.456865605s",
  // sum of all response time in seconds
  "total_response_time_sec": 35.456865605,
  // average response time (formated time)
  "average_response_time": "864.8016ms",
  // average response time in seconds
  "average_response_time_sec": 0.8648016000000001,

  // request statistics [requires --web.statistics to be set]
  // ten most recent requests with 4xx and 5xx status codes
  "recent_errors": [
    {
      // status code
      "status_code": 500,
      // description of status code
      "status": "Internal Server Error",
      // request HTTP method
      "method": "GET",
      // request hostname
      "host": "localhost",
      // request path
      "path": "/path",
      // RFC 3339 formatted date/time
      "time": "2016-10-21T16:59:15.418495872-07:00"
    }
  ]
}
```

## Metrics

You can enable Traefik to export internal metrics to different monitoring systems.
```toml
[api]
  # ...

  # Enable more detailed statistics.
  [api.statistics]

    # Number of recent errors logged.
    #
    # Default: 10
    #
    recentErrors = 10

  # ...
```

| Path       | Method        | Description             |
|------------|---------------|-------------------------|
| `/metrics` |     `GET`     | Export internal metrics |
