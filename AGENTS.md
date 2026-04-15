<!-- crag:auto-start -->
# AGENTS.md

> Generated from governance.md by crag. Regenerate: `crag compile --target agents-md`

## Project: traefik


## Quality Gates

All changes must pass these checks before commit:

### Lint
1. `go vet ./...`
2. `golangci-lint run`
3. `make lint`
4. `make fmt`

### Test
1. `go test ./...`
2. `make test`
3. `make validate`

### Ci (inferred from workflow)
1. `make binary`
2. `make generate binary`
3. `make multi-arch-image-experimental-${GITHUB_REF##*/}`
4. `yarn tsc`
5. `yarn lint`
6. `yarn build`
7. `make test-gateway-api-conformance`
8. `make binary-linux-amd64`

## Coding Standards

- Stack: go, docker
- Follow project commit conventions

## Architecture

- Type: monolith

## Key Directories

- `.github/` — CI/CD
- `cmd/` — executables
- `docs/` — documentation
- `integration/` — tests
- `pkg/` — source

## Testing

- Framework: go test
- Layout: flat

## Code Style

- Linter: golangci-lint

## Anti-Patterns

Do not:
- Do not ignore returned errors — handle or explicitly discard with `_ =`
- Do not use `panic()` in library code — return errors instead
- Do not use `init()` functions unless absolutely necessary
- Do not use `latest` tag in FROM — pin to a specific version
- Do not run containers as root — use a non-root USER

## Security

- No hardcoded secrets — grep for sk_live, AKIA, password= before commit

## Workflow

1. Read `governance.md` at the start of every session — it is the single source of truth.
2. Run all mandatory quality gates before committing.
3. If a gate fails, fix the issue and re-run only the failed gate.
4. Use the project commit conventions for all changes.

<!-- crag:auto-end -->
