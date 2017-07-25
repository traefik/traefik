# Changelog

## 0.8.0-beta.2 (2015-01-16)

* Roll a fix for "Out of memory bug" https://github.com/vulcand/vulcand/issues/156

## 0.8.0-beta.1 (2015-01-14)

### Bugfixes

* Rewrite plugin should be able to rewrite HTTP to HTTPS https://github.com/vulcand/vulcand/issues/120

### Features

* OCSP support for cert revocation checking
* Expose TLS settings for listeners and backends
* Add trace plugin for structured logging of HTTP requests

## 0.8.0-alpha.3 (2014-12-31)

### Bugfixes

* Fix output when upserting middleware
* Fix missing response bodies with Transfer-Ecncoding: chunked

### Features

* Scoped listeners. See http://docs.vulcand.io/proxy.html#listeners
