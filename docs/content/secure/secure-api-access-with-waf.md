---
title: 'Secure API Access with WAF'
description: 'Traefik Hub API Gateway - Learn how to configure the Coraza Web Application Firewall middleware to protect your applications from common web attacks.'
---

# Secure API Access with WAF

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

The [Coraza Web Application Firewall](https://coraza.io/) middleware in Traefik Hub API Gateway provides comprehensive protection against common web application attacks.

There are two ways to implement Coraza WAF protection with Traefik:

- **Native Middleware** (Traefik Hub exclusive): A high-performance native implementation available exclusively in Traefik Hub API Gateway that provides at least 23 times better performance compared to the community plugin version.
- **Community WASM Plugin** (Open Source Traefik Proxy): Available as a [community plugin](https://plugins.traefik.io/plugins/65f2aea146079255c9ffd1ec/coraza-waf) in the Traefik Plugin Catalog for open-source Traefik Proxy users through the [coraza-http-wasm-traefik](https://github.com/jcchavezs/coraza-http-wasm-traefik) project.

!!! note "Performance Advantage"
    The native middleware implementation in Traefik Hub API Gateway delivers significantly superior performance compared to the WASM-based community plugin available for open-source Traefik Proxy. This performance boost ensures that security enforcement does not compromise application responsiveness.

!!! note "Rule Compatibility"
    The native middleware supports the Coraza rule syntax and is compatible with [OWASP Core Rule Set (CRS)](https://coreruleset.org/docs/), allowing you to leverage proven security rules maintained by the security community.

## Basic WAF Protection

To protect your applications with custom security rules, apply the following configuration:

```yaml tab="Middleware WAF"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: waf-protection
  namespace: apps
spec:
  plugin:
    coraza:
      directives:
        - SecRuleEngine On
        - SecRule REQUEST_URI "@streq /admin" "id:101,phase:1,t:lowercase,log,deny"
        - SecRule ARGS "@detectSQLi" "id:102,phase:2,block,msg:'SQL Injection Attack Detected',logdata:'Matched Data: %{MATCHED_VAR} found within %{MATCHED_VAR_NAME}'"
```

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: protected-app
  namespace: apps
spec:
  entryPoints:
    - websecure
  routes:
  - match: Path(`/my-app`)
    kind: Rule
    services:
    - name: whoami
      port: 80
    middlewares:
    - name: waf-protection
```

```yaml tab="Service & Deployment"
kind: Deployment
apiVersion: apps/v1
metadata:
  name: whoami
  namespace: apps
spec:
  replicas: 3
  selector:
    matchLabels:
      app: whoami
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
      - name: whoami
        image: traefik/whoami

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: apps
spec:
  ports:
  - port: 80
    name: whoami
  selector:
    app: whoami
```

## Advanced Protection with OWASP Core Rule Set

To implement comprehensive protection using the OWASP Core Rule Set, which provides battle-tested rules against common attack patterns, apply the following configuration:

```yaml tab="Middleware WAF with CRS"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: waf-crs-protection
  namespace: apps
spec:
  plugin:
    coraza:
      crsEnabled: true
      directives:
        - SecDefaultAction "phase:1,log,auditlog,deny,status:403"
        - SecDefaultAction "phase:2,log,auditlog,deny,status:403"
        - SecAction "id:900110, phase:1, pass, t:none, nolog, setvar:tx.inbound_anomaly_score_threshold=5, setvar:tx.outbound_anomaly_score_threshold=4"
        - SecAction "id:900200, phase:1, pass, t:none, nolog, setvar:'tx.allowed_methods=GET POST'"
        - Include @owasp_crs/REQUEST-911-METHOD-ENFORCEMENT.conf
        - Include @owasp_crs/REQUEST-949-BLOCKING-EVALUATION.conf
```

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: crs-protected-app
  namespace: apps
spec:
  entryPoints:
    - websecure
  routes:
  - match: Path(`/my-app`)
    kind: Rule
    services:
    - name: whoami
      port: 80
    middlewares:
    - name: waf-crs-protection
```

```yaml tab="Service & Deployment"
kind: Deployment
apiVersion: apps/v1
metadata:
  name: whoami
  namespace: apps
spec:
  replicas: 3
  selector:
    matchLabels:
      app: whoami
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
      - name: whoami
        image: traefik/whoami

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: apps
spec:
  ports:
  - port: 80
    name: whoami
  selector:
    app: whoami
```

!!! warning
    Starting with Traefik Hub v3.11.0, Coraza requires read/write permissions to `/tmp`. This requirement stems from upstream changes in the Coraza engine.

!!! note "Advanced Configuration"
    Advanced options and detailed rule configuration are described in the [reference page](../reference/routing-configuration/http/middlewares/waf.md).

    The WAF middleware supports extensive customization through Coraza directives. You can create custom rules, tune detection thresholds, configure logging levels, and integrate with external threat intelligence feeds. For comprehensive rule writing guidance, consult the [Coraza documentation](https://coraza.io/docs/tutorials/introduction/) and [OWASP CRS documentation](https://coreruleset.org/docs/).

{!traefik-for-business-applications.md!}
