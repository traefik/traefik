# Feature Deprecation Notices

This page is maintained and updated periodically to reflect our roadmap and any decisions around feature deprecation.

| Feature                                                                                             | Deprecated | End of Support | Removal |
|-----------------------------------------------------------------------------------------------------|------------|----------------|---------|
| [Kubernetes CRDs API Group `traefik.containo.us`](#kubernetes-crds-api-group-traefikcontainous)     | 2.10       | N/A            | 3.0     |
| [Kubernetes CRDs API Version `traefik.io/v1alpha1`](#kubernetes-crds-api-version-traefikiov1alpha1) | N/A        | N/A            | 3.0     |

## Impact

### Kubernetes CRDs API Group `traefik.containo.us`

In v2.10, the Kubernetes CRDs API Group `traefik.containo.us` is deprecated, and its support will end starting with Traefik v3. Please use the API Group `traefik.io` instead.

### Kubernetes CRDs API Version `traefik.io/v1alpha1`

The newly introduced Kubernetes CRD API Version `traefik.io/v1alpha1` will subsequently be removed in Traefik v3. The following version will be `traefik.io/v1`.
