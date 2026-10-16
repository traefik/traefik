# ingress-nginx e2e tests for the Kubernetes Ingress NGINX provider

This directory vendors the [ingress-nginx e2e tests](https://github.com/kubernetes/ingress-nginx/tree/main/test/e2e)
so that they can be run against Traefik's `kubernetesIngressNGINX` provider.
Each upstream test creates Ingress resources and asserts on HTTP requests and responses,
which makes it a behavioral reference for the provider.

## Layout

It is a dedicated Go module named `k8s.io/ingress-nginx`, so that the upstream files build without import rewrites,
and so that its dependencies stay out of Traefik's `go.mod`.

- `test/e2e/annotations`: upstream annotation tests, verbatim.
- `test/e2e/framework`: upstream framework, patched by `framework.patch` to deploy Traefik instead of ingress-nginx.
  The Traefik specific code lives in `test/e2e/framework/traefik.go`.
- `internal/k8s`: upstream package imported by the framework, verbatim.
- `UPSTREAM`: the vendored ingress-nginx commit.
- `known-failures.txt`: ginkgo skip regexes for specs that are expected to fail.

Never edit the vendored files directly, except the framework: edit it, then regenerate the patch with `./sync.sh diff`.
To update the vendored code, run `./sync.sh <commit>`.

## How the harness works

- `integration/ingress_nginx_e2e_test.go` starts a k3s cluster, loads the `traefik/traefik:latest` image,
  and runs the tests with ginkgo on the k3s node, where the controller Service ClusterIP is reachable, like upstream.
- For every spec, the framework deploys Traefik in the test namespace, under the names and labels upstream expects
  (`nginx-ingress-controller`), with the `kubernetesIngressNGINX` provider watching that namespace.
- Waiting for an nginx reload is replaced by waiting for the Traefik configuration (`/api/rawdata`) to change and settle.
- `AllowSnippetConfiguration` restarts Traefik with `allowSnippetAnnotations` enabled.
- Helpers that inspect nginx directly (`ExecCommand`, `NginxLogs`, the nginx ConfigMap helpers, ...) skip the spec.
- The nginx.conf matchers (`WaitForNginxServer`, `WaitForNginxConfiguration`, ...) cannot be evaluated against Traefik:
  they only wait for the configuration to settle.
  A spec that uses them and sends no HTTP request through the framework client is reported as skipped rather than passed.
  Specs that combine both keep their HTTP assertions only.
- The vendored `httpexpect` fork is fixed so that a failed body assertion is reported instead of panicking.

## Known failures

`known-failures.txt` lists, one exact spec per line, the specs expected to fail, grouped by reason:

- Harness: the spec asserts on nginx.conf captured by a matcher.
- Documented: the annotation or behavior is listed as unsupported or different in the provider documentation.
- Undocumented: a behavior difference to either fix in the provider, or document.

When a provider change fixes a spec, remove its line.

## Running

```bash
make test-ingress-nginx-e2e

# Focus, skip and parallelism.
FOCUS="cors" SKIP="canary" E2E_NODES=2 make test-ingress-nginx-e2e
```

The backend images used by the tests (`registry.k8s.io/ingress-nginx/e2e-test-echo`, ...) are pulled by k3s.
