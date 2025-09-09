---
title: 'Secure API Access with WAF'
description: 'Traefik Hub API Gateway - Learn how to configure the Coraza Web Application Firewall middleware to protect your applications from common web attacks.'
---

# Secure API Access with WAF

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

The [Coraza Web Application Firewall](https://coraza.io/) middleware in Traefik Hub API Gateway provides comprehensive protection against common web application attacks. The middleware supports the Coraza rule syntax and is compatible with [OWASP Core Rule Set (CRS)](https://coreruleset.org/docs/), allowing you to leverage proven security rules maintained by the security community.

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

This configuration implements three security directives that work together to protect an application:

- **SecRuleEngine On**: Activates the WAF engine to begin processing incoming requests. Without this directive, all other rules remain inactive regardless of their configuration.

- **Admin Path Protection**: The second rule blocks all access to `/admin` paths by examining the request URI. This completely prevents access to administrative interfaces that often contain sensitive functionality like user management, system configuration, or database administration tools. The rule triggers during phase 1 (request headers processing) and applies lowercase transformation to catch variations like `/Admin` or `/ADMIN`.

- **SQL Injection Detection**: The third rule scans request parameters (query strings and form data) for SQL injection patterns using Coraza's built-in detection engine. The `ARGS` variable covers query string parameters like `?id=1` and form data from POST requests like `username=admin&password=123`, but does not include cookies. SQL injection attacks attempt to manipulate database queries by injecting malicious SQL code through user inputs. When detected, the rule blocks the request and logs detailed information about the attempted attack, including which parameter contained the malicious payload.

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
        - SecRuleEngine On
        - SecDefaultAction "phase:1,log,auditlog,deny,status:403"
        - SecDefaultAction "phase:2,log,auditlog,deny,status:403"
        - SecAction "id:900110, phase:1, pass, t:none, nolog, setvar:tx.inbound_anomaly_score_threshold=5, setvar:tx.outbound_anomaly_score_threshold=4"
        - SecAction "id:900200, phase:1, pass, t:none, nolog, setvar:'tx.allowed_methods=GET POST'"
        - Include @owasp_crs/REQUEST-911-METHOD-ENFORCEMENT.conf
        - Include @owasp_crs/REQUEST-949-BLOCKING-EVALUATION.conf
```

This advanced configuration implements [OWASP Core Rule Set (CRS)](https://coreruleset.org/docs/) protection with anomaly scoring:

- **SecDefaultAction for Phase 1 & 2**: Sets default behavior for request processing phases. Phase 1 processes request headers while Phase 2 processes request body. When rules match, they log the event to both standard and audit logs, then deny the request with a 403 status code.

- **Anomaly Score Configuration**: The first `SecAction` sets anomaly score thresholds where `inbound_anomaly_score_threshold=5` means requests scoring 5 or higher are blocked, and `outbound_anomaly_score_threshold=4` applies the same logic to responses. This scoring system allows multiple suspicious patterns to accumulate points rather than blocking on first detection, reducing false positives while maintaining security.

- **Allowed Methods Configuration**: The second `SecAction` restricts HTTP methods to only `GET` and `POST` requests. This prevents potentially dangerous methods like `PUT`, `DELETE`, `PATCH`, or `OPTIONS` that could modify server resources or reveal system information.

- **METHOD-ENFORCEMENT Rule Set**: The `REQUEST-911-METHOD-ENFORCEMENT.conf` file enforces the allowed HTTP methods policy defined above. It checks incoming requests against the permitted methods and contributes to the anomaly score for disallowed methods.

- **BLOCKING-EVALUATION Rule Set**: The `REQUEST-949-BLOCKING-EVALUATION.conf` file evaluates the accumulated anomaly score against the configured thresholds. If the total score exceeds the threshold, it triggers the blocking action, preventing the request from reaching your application.

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
