# Governance — traefik
# Inferred by crag analyze — review and adjust as needed

## Identity
- Project: traefik
- Stack: go, docker

## Gates (run in order, stop on failure)
### Lint
- go vet ./...
- golangci-lint run
- make lint
- make fmt

### Test
- go test ./...
- make test
- make validate

### CI (inferred from workflow)
- make binary
- make generate binary
- make multi-arch-image-experimental-${GITHUB_REF##*/}
- yarn tsc
- yarn lint
- yarn build
- make test-gateway-api-conformance
- make binary-linux-amd64

## Advisories (informational, not enforced)
- hadolint Dockerfile  # [ADVISORY]
- actionlint  # [ADVISORY]

## Branch Strategy
- Trunk-based development
- Free-form commits
- Commit trailer: Co-Authored-By: Claude <noreply@anthropic.com>

## Security
- No hardcoded secrets — grep for sk_live, AKIA, password= before commit

## Autonomy
- Auto-commit after gates pass

## Deployment
- Target: docker
- CI: github-actions

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

## Dependencies
- Package manager: go (go.sum)
- Go: >=1.25.0

## Anti-Patterns

Do not:
- Do not ignore returned errors — handle or explicitly discard with `_ =`
- Do not use `panic()` in library code — return errors instead
- Do not use `init()` functions unless absolutely necessary
- Do not use `latest` tag in FROM — pin to a specific version
- Do not run containers as root — use a non-root USER

