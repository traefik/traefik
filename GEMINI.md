<!-- crag:auto-start -->
# GEMINI.md

> Generated from governance.md by crag. Regenerate: `crag compile --target gemini`

## Project Context

- **Name:** traefik
- **Stack:** go, docker
- **Runtimes:** go

## Rules

### Quality Gates

Run these checks in order before committing any changes:

1. [lint] `go vet ./...`
2. [lint] `golangci-lint run`
3. [lint] `make lint`
4. [lint] `make fmt`
5. [test] `go test ./...`
6. [test] `make test`
7. [test] `make validate`
8. [ci (inferred from workflow)] `make binary`
9. [ci (inferred from workflow)] `make generate binary`
10. [ci (inferred from workflow)] `make multi-arch-image-experimental-${GITHUB_REF##*/}`
11. [ci (inferred from workflow)] `yarn tsc`
12. [ci (inferred from workflow)] `yarn lint`
13. [ci (inferred from workflow)] `yarn build`
14. [ci (inferred from workflow)] `make test-gateway-api-conformance`
15. [ci (inferred from workflow)] `make binary-linux-amd64`

### Security

- No hardcoded secrets — grep for sk_live, AKIA, password= before commit

### Workflow

- Follow project commit conventions
- Run quality gates before committing
- Review security implications of all changes

<!-- crag:auto-end -->
