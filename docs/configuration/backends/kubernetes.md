# Kubernetes Ingress Backend

Træfik can be configured to use Kubernetes Ingress as a backend configuration:

```toml
################################################################
# Kubernetes Ingress configuration backend
################################################################
# Enable Kubernetes Ingress configuration backend
[kubernetes]

# Kubernetes server endpoint
#
# When deployed as a replication controller in Kubernetes, Traefik will use
# the environment variables KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT
# to construct the endpoint.
# Secure token will be found in /var/run/secrets/kubernetes.io/serviceaccount/token
# and SSL CA cert in /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
#
# The endpoint may be given to override the environment variable values.
#
# When the environment variables are not found, Traefik will try to connect to
# the Kubernetes API server with an external-cluster client. In this case, the
# endpoint is required. Specifically, it may be set to the URL used by
# `kubectl proxy` to connect to a Kubernetes cluster from localhost.
#
# Optional for in-cluster configuration, required otherwise
# Default: empty
#
# endpoint = "http://localhost:8080"

# Bearer token used for the Kubernetes client configuration.
#
# Optional
# Default: empty
#
# token = "my token"

# Path to the certificate authority file used for the Kubernetes client
# configuration.
#
# Optional
# Default: empty
#
# certAuthFilePath = "/my/ca.crt"

# Array of namespaces to watch.
#
# Optional
# Default: all namespaces (empty array).
#
# namespaces = ["default", "production"]

# Ingress label selector to identify Ingress objects that should be processed.
# See https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors for details.
#
# Optional
# Default: empty (process all Ingresses)
#
# labelselector = "A and not B"
```

Annotations can be used on containers to override default behaviour for the whole Ingress resource:

- `traefik.frontend.rule.type: PathPrefixStrip`: override the default frontend rule type (Default: `PathPrefix`).
- `traefik.frontend.priority: 3`: override the default frontend rule priority (Default: `len(Path)`).

Annotations can be used on the Kubernetes service to override default behaviour:

- `traefik.backend.loadbalancer.method=drr`: override the default `wrr` load balancer algorithm
- `traefik.backend.loadbalancer.sticky=true`: enable backend sticky sessions

You can find here an example [ingress](https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheese-ingress.yaml) and [replication controller](https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik.yaml).

Additionally, an annotation can be used on Kubernetes services to set the [circuit breaker expression](https://docs.traefik.io/basics/#backends) for a backend.

- `traefik.backend.circuitbreaker: <expression>`: set the circuit breaker expression for the backend (Default: nil).

As known from nginx when used as Kubernetes Ingress Controller, a List of IP-Ranges which are allowed to access can be configured by using an ingress annotation:

- `ingress.kubernetes.io/whitelist-source-range: "1.2.3.0/24, fe80::/16"`

An unset or empty list allows all Source-IPs to access. If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access.


### Authentication

Is possible to add additional authentication annotations in the Ingress rule.
The source of the authentication is a secret that contains usernames and passwords inside the the key auth.

- `ingress.kubernetes.io/auth-type`: `basic`
- `ingress.kubernetes.io/auth-secret`: contains the usernames and passwords with access to the paths defined in the Ingress Rule.

The secret must be created in the same namespace as the Ingress rule.

Limitations:

- Basic authentication only.
- Realm not configurable; only `traefik` default.
- Secret must contain only single file.
