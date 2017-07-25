# Memcached http provider

Publishes challenges into memcached where they can be retrieved by nginx. Allows
specifying multiple memcached servers and the responses will be published to all
of them, making it easier to verify when your domain is hosted on a cluster of
servers.

Example nginx config:

```
    location /.well-known/acme-challenge/ {
        set $memcached_key "$uri";
        memcached_pass 127.0.0.1:11211;
    }
```
