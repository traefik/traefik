# Feature Deprecation Notices

This page is maintained and updated periodically to reflect our roadmap and any decisions around feature deprecation.

| Feature                                                                                                              | Deprecated | End of Support | Removal |
|----------------------------------------------------------------------------------------------------------------------|------------|----------------|---------|
| [Kubernetes Ingress API Version `networking.k8s.io/v1beta1`](#kubernetes-ingress-api-version-networkingk8siov1beta1) | N/A        | N/A            | 3.0     |
| [CRD API Version `apiextensions.k8s.io/v1beta1`](#kubernetes-ingress-api-version-networkingk8siov1beta1)             | N/A        | N/A            | 3.0     |
| [Entry point `underscoreHeadersStrategy` option](#entry-point-underscoreheadersstrategy-option)                      | 3.7.12     | N/A            | N/A     |

## Impact

### Kubernetes Ingress API Version `networking.k8s.io/v1beta1`

The Kubernetes Ingress API Version `networking.k8s.io/v1beta1` support is removed in v3. 
Please use the API Group `networking.k8s.io/v1` instead.

### Traefik CRD Definitions API Version `apiextensions.k8s.io/v1beta1`

The Traefik CRD definitions API Version `apiextensions.k8s.io/v1beta1` support is removed in v3.
Please use the API Group `apiextensions.k8s.io/v1` instead.

### Entry Point `underscoreHeadersStrategy` Option

Starting on 3.7.12 the `underscoreHeadersStrategy` entry point option is deprecated,
please use the [`aliasHeadersStrategy`](../reference/install-configuration/entrypoints.md#opt-http-aliasHeadersStrategy) option instead.
Underscores are only one of the characters making a header name alias another one, and `aliasHeadersStrategy` handles them all.
An entry point cannot configure both options with different values.
