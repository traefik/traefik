This directory is a point-in-time, unmodified copy of the `config/crd`,
`config/rbac` and `config/manager` directories of
[traefik/gateway-operator](https://github.com/traefik/gateway-operator), a
private repository.

It exists so that `go run ./cmd/internal/gatewayapioperatorfixture` can render
`integration/fixtures/gateway-api-conformance/01-operator.yml` without network
access or access to that repository. Refresh it, deliberately and rarely, with:

```
go run ./cmd/internal/gatewayapioperatorfixture -update [ref|path to a checkout]
```

Source: traefik/gateway-operator@2ca71c164570d551c6d068cb0f49cc5970752887
Vendored: 2026-09-04
