---
layout: "docs"
page_title: "External Services"
sidebar_current: "docs-guides-external"
description: |-
  Very few infrastructures are entirely self-contained. Most rely on a multitude of external service providers. Consul supports this by allowing for the definition of external services, services that are not provided by a local node.
---

# Registering an External Service

Very few infrastructures are entirely self-contained. Most rely on a multitude
of external service providers. Consul supports this by allowing for the definition
of external services, services that are not provided by a local node.

Most services are registered in Consul through the use of a
[service definition](/docs/agent/services.html). However, this approach registers
the local node as the service provider. In the case of external services, we must
instead register the service with the catalog rather than as part of a standard
node service definition.

Once registered, the DNS interface will be able to return the appropriate A
records or CNAME records for the service. The service will also appear in standard
queries against the API.

Let us suppose we want to register a "search" service that is provided by
"www.google.com". We might accomplish that like so:

```text
$ curl -X PUT -d '{"Datacenter": "dc1", "Node": "google",
   "Address": "www.google.com",
   "Service": {"Service": "search", "Port": 80}}'
   http://127.0.0.1:8500/v1/catalog/register
```

If we do a DNS lookup now, we can see the new search service:

```text
; <<>> DiG 9.8.3-P1 <<>> @127.0.0.1 -p 8600 search.service.consul.
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 13313
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 4, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;search.service.consul.		IN	A

;; ANSWER SECTION:
search.service.consul.	0	IN	CNAME	www.google.com.
www.google.com.		264	IN	A	74.125.239.114
www.google.com.		264	IN	A	74.125.239.115
www.google.com.		264	IN	A	74.125.239.116

;; Query time: 41 msec
;; SERVER: 127.0.0.1#8600(127.0.0.1)
;; WHEN: Tue Feb 25 17:45:12 2014
;; MSG SIZE  rcvd: 178
```

If at any time we want to deregister the service, we simply do:

```text
$ curl -X PUT -d '{"Datacenter": "dc1", "Node": "google"}' http://127.0.0.1:8500/v1/catalog/deregister
```

This will deregister the `google` node along with all services it provides.

For more information, please see the [HTTP Catalog API](/api/catalog.html).
