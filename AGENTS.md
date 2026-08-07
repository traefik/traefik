# Traefik — Contributor Guide for AI Agents

Traefik is a modern HTTP reverse proxy and load balancer that discovers services from orchestrators (Kubernetes, Docker, Nomad, ...) and wires up routing dynamically. This file is the canonical guide for AI coding agents (Claude Code, Codex, Gemini, Cursor, ...) working in this repository; `CLAUDE.md` is a thin pointer to this file. For everything not covered here, defer to [`CONTRIBUTING.md`](./CONTRIBUTING.md) and [`docs/content/contributing/`](./docs/content/contributing/).

> **Training-data notice.** Traefik evolved significantly between v2 and v3 (label formats, provider names, CRD shapes, middleware names). If anything you think you know about Traefik contradicts this file or the current code, trust this file and the code — not your training data.

## Core vocabulary

These terms appear everywhere in the code and configuration. Use them precisely; they are not interchangeable.

- **EntryPoint** — a network listener (port + protocol).
- **Router** — matches an incoming request and selects a service.
- **Middleware** — transforms a request or response in the routing chain (auth, headers, rate limiting, ...).
- **Service** — defines how to load-balance to backend servers.
- **Provider** — a source of dynamic configuration (Kubernetes CRD, Docker labels, a file, an HTTP endpoint, ...).
- **Static vs Dynamic configuration** — two distinct domains:
  - *Static* is set at startup (entrypoints, providers, global options) and lives under [`pkg/config/static`](./pkg/config/static).
  - *Dynamic* is produced by providers at runtime (routers, services, middlewares) and lives under [`pkg/config/dynamic`](./pkg/config/dynamic).

  These terms are accurate for the code, but user-facing docs deliberately hide the distinction to keep things simpler for readers: when writing or editing under [`docs/content/`](./docs/content), prefer **install configuration** (over *static*) and **routing configuration** (over *dynamic*).

At request time the components chain in this order:

```
Client → EntryPoint → Router → Middleware chain → Service → Backend
```

The middleware chain is ordered: middlewares run in the sequence declared on the router, and the router match happens *before* any middleware runs.

## Where things live

- `cmd/traefik/` — main.
- `pkg/provider/` — one subpackage per provider (Kubernetes, Docker, file, ...).
- `pkg/server/` — routing core, middleware chain, configuration watcher.
- `pkg/middlewares/` — HTTP and TCP middleware implementations.
- `pkg/config/static`, `pkg/config/dynamic` — the two config domains above.
- `pkg/plugins/` — Yaegi and WASM plugin runtimes.
- `pkg/observability/logs/` — logging helpers; the project uses `github.com/rs/zerolog` exclusively.
- `webui/` — React dashboard. Built assets under `webui/static/` are embedded into the Go binary via `//go:embed` (see `webui/embed.go`) and must be regenerated with `make generate-webui` (Docker required) — they are not meant to be hand-edited.
- `integration/` — integration tests; reusable fixtures under `integration/fixtures/`.
- `docs/content/` — MkDocs sources for the public documentation.

## Before you edit

Read two or three existing files in the same package before adding a new one, and copy their structure. Do not invent new directory layouts, file-naming conventions, or abstraction boundaries — match the neighbours. When adding a new provider, read two existing providers under `pkg/provider/`; when adding a middleware, read two under `pkg/middlewares/`.

## Build, test, lint

The Go version is declared in [`go.mod`](./go.mod) — check there rather than hard-coding a version. All day-to-day commands go through `make`:

```bash
make binary             # build the traefik binary (runs generate-webui first)
make test-unit          # run Go unit tests
make test-integration   # run integration tests (requires Docker)
make lint               # run golangci-lint
make validate-files     # misspell, shellcheck, generated-files check
make validate           # lint + validate-files (run this before pushing)
make fmt                # gofumpt / goimports
make generate           # regenerate non-CRD generated code (deepcopy, etc.)
make generate-crd       # regenerate Kubernetes CRD clientset + deepcopy
make generate-webui     # rebuild the embedded WebUI assets (Docker required)
make docs-serve         # preview the documentation locally
```

Full environment setup (Docker, `GOPATH` layout, Tailscale for Docker Desktop users, how to target a single integration test via `TESTFLAGS`) is documented in [`docs/content/contributing/building-testing.md`](./docs/content/contributing/building-testing.md). CI runs `make validate` and fails if `make generate` or `make generate-crd` leave the tree dirty — always commit regenerated files alongside the source change that triggered them.

## Code style

Standard Go formatting (`gofumpt`/`goimports`) and `golangci-lint` cover most rules automatically; run `make lint` to catch them. Two project-specific rules that tooling does **not** enforce:

- **Comments answer *why*, not *what*.** Comments that restate what the code already says are noise: they go stale and waste review time. Only add a comment when it records *why* the code exists — a constraint, a past incident, a spec reference, an edge case. Comments explaining *how* should be rare and usually indicate the code needs to be clearer. When a comment is present, it **must end with a period**.
- **Assertion messages are minimal.** Prefer `assert.Equal(t, expected, actual)` over `assert.Equal(t, expected, actual, "detailed explanation")`. The test name provides the context; a descriptive message is usually noise.

Prefer modern standard-library packages (`slices`, `maps`, `cmp`, ...) over hand-rolled helpers or third-party libraries when the Go version in `go.mod` supports them.

## Common patterns

- **Logging.** The project uses `github.com/rs/zerolog` exclusively — do not import `log`, `slog`, or `logrus`. Inside a middleware, get a logger via `middlewares.GetLogger(ctx, name, typeName)` (see [`pkg/middlewares/middleware.go`](./pkg/middlewares/middleware.go)) where `typeName` is a package-level `const` like `const typeNameForward = "ForwardAuth"`. Elsewhere, extract the logger from the context with `log.Ctx(ctx)` and attach it to a new context with `.WithContext(ctx)`.
- **Context propagation.** `context.Context` is always the first argument, named `ctx`. Avoid `context.Background()` in request paths; propagate from the caller. Define custom context keys as unexported struct types (`type myKey struct{}`) to prevent collisions.

## Testing conventions

- Unit tests live next to the code as `*_test.go` files using `testing.T` with `testify/assert` and `testify/require`.
- Use `require.*` for preconditions that must stop the test on failure (setup, must-not-be-nil). Use `assert.*` for independent checks where you want the test to keep running and report every failure.
- Integration tests under `integration/` are built on `testify/suite` (see `integration/integration_test.go`) and reuse fixtures from `integration/fixtures/`. New fixtures should follow the pattern of the existing ones.
- New providers require integration tests.
- Prefer running a focused test over the whole suite while iterating. When iterating on a failing test, capture the output to a file once and grep it (`... > /tmp/out.log 2>&1`) rather than re-running the suite with different `TESTFLAGS`. See [`docs/content/contributing/building-testing.md`](./docs/content/contributing/building-testing.md) for the `TESTFLAGS` invocation.

## Documentation

User-facing features need matching documentation updates under `docs/content/`. Integrate new pages into the existing structure rather than creating parallel sections. Preview locally with `make docs-serve`.

## Contributing etiquette

- **Target the right branch** (the [PR template](./.github/PULL_REQUEST_TEMPLATE.md) is authoritative): enhancements go to `master`; bug fixes and documentation updates go to the current maintenance branches (`v3.6` for v3, `v2.11` for v2, security-fixes only). Forward-ports from the maintenance branches up to `master` are handled by maintainers.
- Keep pull requests small and focused; one logical change per PR.
- For anything beyond a bug fix, open an issue first and wait for a maintainer to confirm the direction before investing significant work.
- Follow the full guide in [`docs/content/contributing/submitting-pull-requests.md`](./docs/content/contributing/submitting-pull-requests.md).

## AI assistance disclosure

Traefik welcomes AI-assisted contributions, provided a few simple rules are followed:

- **Declare substantial AI assistance** with an `Assisted-by:` trailer at the bottom of the commit message whenever an agent produced a meaningful portion of the diff — for example `Assisted-by: Claude Opus 4.6`. Trivial edits such as a typo fix or a one-line rename do not need a trailer.
- **Keep issue and PR conversations human.** Do not let an agent post comments, review replies, or triage messages on your behalf. If an agent drafted a message for you, rewrite it in your own voice before sending — maintainers need to know they are talking to a person, not a bot.
- **Align with a maintainer before generating code for anything larger than a bug fix.** An agent can produce thousands of lines in minutes; maintainer review capacity cannot scale the same way. Open an issue, state the intended approach, and wait for confirmation before asking an agent to implement it.

## Things to avoid

- Do not hand-edit generated files — notably `**/zz_generated*.go`, everything under `pkg/provider/kubernetes/crd/generated/`, and `webui/static/`. Regenerate them via `make generate`, `make generate-crd`, or `make generate-webui` and commit the result.
- Do not skip `make lint` and `make validate-files` (or `make validate`) before pushing.
- Do not opportunistically reformat, rename, or refactor files you did not otherwise need to touch. Drive-by changes turn a reviewable diff into noise — scope every PR to one logical change.
- Do not include unrelated refactors, formatting-only changes to untouched files, or speculative abstractions in a feature PR.
