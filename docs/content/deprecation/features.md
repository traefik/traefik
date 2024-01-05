# Feature Deprecation Notices

This page is maintained and updated periodically to reflect our roadmap and any decisions around feature deprecation.

| Feature                                                                                                              | Deprecated | End of Support | Removal |
|----------------------------------------------------------------------------------------------------------------------|------------|----------------|---------|
| [Kubernetes CRD Provider API Version `traefik.io/v1alpha1`](#kubernetes-crd-provider-api-version-traefikiov1alpha1)  | 3.0        | N/A            | 4.0     |
| [Kubernetes Ingress API Version `networking.k8s.io/v1beta1`](#kubernetes-ingress-api-version-networkingk8siov1beta1) | N/A        | N/A            | 3.0     |
| [CRD API Version `apiextensions.k8s.io/v1beta1`](#kubernetes-ingress-api-version-networkingk8siov1beta1)             | N/A        | N/A            | 3.0     |

## Impact

### Kubernetes CRD Provider API Version `traefik.io/v1alpha1`

The Kubernetes CRD provider API Version `traefik.io/v1alpha1` is deprecated in Traefik v3.
Please use the API Group `traefik.io/v1` instead.

### Kubernetes Ingress API Version `networking.k8s.io/v1beta1`

The Kubernetes Ingress API Version `networking.k8s.io/v1beta1` support is removed in v3. 
Please use the API Group `networking.k8s.io/v1` instead.

### Traefik CRD Definitions API Version `apiextensions.k8s.io/v1beta1`

The Traefik CRD definitions API Version `apiextensions.k8s.io/v1beta1` support is removed in v3.
Please use the API Group `apiextensions.k8s.io/v1` instead.
