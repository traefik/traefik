# Ping Definition

```toml
# Ping definition
#
# Default:
# [ping]
#   entrypoint = "traefik"
#
[ping]
  entrypoint="traefik"
```


| Path                                                            |     Method    | Description                                                                                        |
|-----------------------------------------------------------------|:-------------:|----------------------------------------------------------------------------------------------------|
| `/ping`                                                         | `GET`, `HEAD` | A simple endpoint to check for Træfik process liveness. Return a code `200` with the content: `OK` |


!!!warning
    Even if you have authentication configure on entrypoint, the `/ping` path of the api is excluded from authentication.

### Example

```shell
curl -sv "http://localhost:8080/ping"
```
```shell
*   Trying ::1...
* Connected to localhost (::1) port 8080 (#0)
> GET /ping HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Thu, 25 Aug 2016 01:35:36 GMT
< Content-Length: 2
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host localhost left intact
OK
```