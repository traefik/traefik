# Roadmap

The goal of the roadmap is to provide some insight on what to expect from Vulcand in the short term
and long term.

## Longer-term

### Routing

* Support pods/consistent hash-based routing
* Fan-In, Fan-Out support

### Reliability and performance

* TLS session caching
* Connection control for HTTP transports
* Reusing memory buffers with sync.Pool
* Profiling and benchmarking
* HTTP/2 support

### Reporting and UI

* Structured logging, ES connectors
* Dashboard with real-time metrics & CRUD
* Dependency analysis and visualization
* Bottleneck detection

### API support

* IP blacklists/whitelists with pluggable backends
* Request HMAC signing/checking

### Clustering

* Implementing Leader/Follower pattern, IP takeover
* Centralized metrics collection
* Rate limiting with pluggable backends
* Pluggable caching - Cassandra
* Integration with Kubernetes


