---
title: "Traefik Coraza Documentation"
description: "The HTTP coraza middleware in Traefik Proxy implements web application firewall capability to Traefik. Read the technical documentation."
---

# Coraza

The HTTP [Coraza](https://coraza.io/) middleware in Traefik Proxy implements web application firewall capability to Traefik.
To get help writing rules https://coraza.io/docs/tutorials/introduction/ and https://coreruleset.org/docs/ are excellent places to start. 

## Configuration Examples

```yaml tab="Kubernetes"
# Denying /admin path
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: waf
spec:
  coraza:
    directives: |
      SecRuleEngine On
      SecRule REQUEST_URI "@streq /admin" "id:101,phase:1,t:lowercase,log,deny"
```

## Configuration Options

### `directives`

The `directives` string contains configurable waf rules for the middleware.

As an example we might want to block /admin path in production.

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: waf
spec:
  coraza:
    directives: |
      SecRuleEngine On
      SecRule REQUEST_URI "@streq /admin" "id:101,phase:1,t:lowercase,log,deny"
```

### `crsEnabled`

Set the `crsEnabled` option to `true` to enable [CRS rulesets](https://github.com/corazawaf/coraza-coreruleset/tree/main/rules/%40owasp_crs). (Default value is `false`.).

After the ruleset is enabled, it can be used in middleware.
Example: allow `GET` method only and deny others.


```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: wafcrs
spec:
  coraza:
    crsEnabled: true
    directives: |
      SecDefaultAction "phase:1,log,auditlog,deny,status:403"
      SecDefaultAction "phase:2,log,auditlog,deny,status:403"
      SecAction \
          "id:900110,\
          phase:1,\
          pass,\
          t:none,\
          nolog,\
          setvar:tx.inbound_anomaly_score_threshold=5,\
          setvar:tx.outbound_anomaly_score_threshold=4"
      SecAction \
          "id:900200,\
          phase:1,\
          pass,\
          t:none,\
          nolog,\
          setvar:'tx.allowed_methods=GET'"
      Include @owasp_crs/REQUEST-911-METHOD-ENFORCEMENT.conf
      Include @owasp_crs/REQUEST-949-BLOCKING-EVALUATION.conf
```
