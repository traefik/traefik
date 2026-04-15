---
trigger: always_on
description: Governance rules for traefik — compiled from governance.md by crag
---

# Windsurf Rules — traefik

Generated from governance.md by crag. Regenerate: `crag compile --target windsurf`

## Project

(No description)

**Stack:** go, docker

## Runtimes

go

## Cascade Behavior

When Windsurf's Cascade agent operates on this project:

- **Always read governance.md first.** It is the single source of truth for quality gates and policies.
- **Run all mandatory gates before proposing changes.** Stop on first failure.
- **Respect classifications.** OPTIONAL gates warn but don't block. ADVISORY gates are informational.
- **Respect path scopes.** Gates with a `path:` annotation must run from that directory.
- **No destructive commands.** Never run rm -rf, dd, DROP TABLE, force-push to main, curl|bash, docker system prune.
- - No hardcoded secrets — grep for sk_live, AKIA, password= before commit
- Follow the project commit conventions.

## Quality Gates (run in order)

1. `go vet ./...`
2. `golangci-lint run`
3. `make lint`
4. `make fmt`
5. `go test ./...`
6. `make test`
7. `make validate`
8. `make binary`
9. `make generate binary`
10. `make multi-arch-image-experimental-${GITHUB_REF##*/}`
11. `yarn tsc`
12. `yarn lint`
13. `yarn build`
14. `make test-gateway-api-conformance`
15. `make binary-linux-amd64`

## Rules of Engagement

1. **Minimal changes.** Don't rewrite files that weren't asked to change.
2. **No new dependencies** without explicit approval.
3. **Prefer editing** existing files over creating new ones.
4. **Always explain** non-obvious changes in commit messages.
5. **Ask before** destructive operations (delete, rename, migrate schema).

---

**Tool:** crag — https://www.npmjs.com/package/@whitehatd/crag
