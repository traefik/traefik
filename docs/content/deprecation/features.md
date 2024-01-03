# Feature Deprecation Notices

This page is maintained and updated periodically to reflect our roadmap and any decisions around feature deprecation.

| Feature                                                                                                              | Deprecated | End of Support | Removal |
|----------------------------------------------------------------------------------------------------------------------|------------|----------------|---------|
| [Kubernetes CRDs API Version `traefik.io/v1alpha1`](#kubernetes-crds-api-version-traefikiov1alpha1)                  | 3.0        | N/A            | 4.0     |
| [Kubernetes Ingress API Version `networking.k8s.io/v1beta1`](#kubernetes-ingress-api-version-networkingk8siov1beta1) | N/A        | N/A            | 3.0     |
| [CRD API Version `apiextensions.k8s.io/v1beta1`](#kubernetes-ingress-api-version-networkingk8siov1beta1)             | N/A        | N/A            | 3.0     |

## Impact

### Kubernetes CRDs API Version `traefik.io/v1alpha1`

The Kubernetes CRD provider API Version `traefik.io/v1alpha1` will subsequently be deprecated in Traefik v3. The next version will be `traefik.io/v1`.

### Kubernetes Ingress API Version `networking.k8s.io/v1beta1`

The Kubernetes Ingress API Version `networking.k8s.io/v1beta1` is removed in v3. Please use the API Group `networking.k8s.io/v1` instead.

### Traefik CRD API Version `apiextensions.k8s.io/v1beta1`

The Traefik CRD API Version `apiextensions.k8s.io/v1beta1` is removed in v3. Please use the API Group `apiextensions.k8s.io/v1` instead.
