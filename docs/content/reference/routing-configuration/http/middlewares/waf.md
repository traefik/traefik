---
title: 'Coraza Web Application Firewall'
description: 'Traefik Hub API Gateway - The HTTP Coraza in Traefik Hub API Gateway provides web application firewall capabilities'
---

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

The [Coraza WAF](https://coraza.io/) middleware in Traefik Hub API Gateway provides web application firewall capabilities.

The native middleware in Hub API Gateway provides at least 23 times more performance compared to the
WASM-based [Coraza plugin](https://plugins.traefik.io/plugins/65f2aea146079255c9ffd1ec/coraza-waf) available with the open-source Traefik Proxy.

To learn how to write rules, please visit [Coraza documentation](https://coraza.io/docs/tutorials/introduction/ "Link to Coraza introduction tutorial") and
[OWASP CRS documentation](https://coreruleset.org/docs/ "Link to the OWAP CRS project documentation").

!!! warning

    Starting with Traefik Hub v3.11.0, Coraza needs to have read/write permissions to `/tmp`. This is related to [this upstream PR](https://github.com/corazawaf/coraza/pull/1030).

---

## Configuration Examples

```yaml tab="Deny the /admin path"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: waf
spec:
  plugin:
    coraza:
      directives:
        - SecRuleEngine On
        - SecRule REQUEST_URI "@streq /admin" "id:101,phase:1,t:lowercase,log,deny"
```

```yaml tab="Allow only GET methods"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: wafcrs
  namespace: apps
spec:
  plugin:
    coraza:
      crsEnabled: true
      directives:
        - SecDefaultAction "phase:1,log,auditlog,deny,status:403"
        - SecDefaultAction "phase:2,log,auditlog,deny,status:403"
        - SecAction "id:900110, phase:1, pass, t:none, nolog, setvar:tx.inbound_anomaly_score_threshold=5, setvar:tx.outbound_anomaly_score_threshold=4"
        - SecAction "id:900200, phase:1, pass, t:none, nolog, setvar:'tx.allowed_methods=GET'"
        - Include @owasp_crs/REQUEST-911-METHOD-ENFORCEMENT.conf
        - Include @owasp_crs/REQUEST-949-BLOCKING-EVALUATION.conf
```

## Configuration Options

| Field    | Description   | Default | Required |
|:---------|:-----------------------|:--------|:----------------------------|
| `directives` | List of WAF rules to enforce. |  | Yes |
| `crsEnabled` | Enable [CRS rulesets](https://github.com/corazawaf/coraza-coreruleset/tree/main/rules/%40owasp_crs).<br /> Once the ruleset is enabled, it can be used in the middleware. | false |  False |

{!traefik-for-business-applications.md!}
