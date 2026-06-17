---
name: review
description: Repository-specific guidance for the AI code review agent reviewing this repo
user-invocable: true
disable-model-invocation: true
---

# Code review guidance

This is Traefik (Go), a cloud-native reverse proxy and load balancer. The reviewer already fixes the
role, the tools, the severity definitions, the finding mechanics and the output format; this guidance
is **additive** and must not restate them. Record findings only with the severities the reviewer
accepts: `CRITICAL`, `IMPORTANT`, `MINOR`, `QUESTION`.

Review in this priority order, and keep the bar high at every level.

## Security

- Validate authentication, authorization and trust boundaries at every boundary crossing.
- Trace untrusted input through to dangerous sinks (exec, SQL, file I/O, template rendering).
- No hardcoded passwords, tokens or API keys; no secrets or sensitive data in logs.

## Correctness

- Nil dereferences, race conditions, off-by-one errors, inverted conditions, unhandled edge cases.
- Errors are wrapped with `fmt.Errorf` in gerund form, e.g. `fmt.Errorf("unmarshalling data: %w", err)`, and never silently dropped with `_`.
- Resource leaks: unclosed connections, leaked goroutines, contexts not propagated or cancelled.
- `context.Context` must be the first argument of every function that accepts one, named `ctx`.
- Do not use `context.Background()` in request paths — propagate the context from the caller.
- Custom context keys must be unexported struct types (`type myKey struct{}`), never bare strings or integers.
- Changes must not blur the static/dynamic configuration boundary: static configuration is read at startup only; dynamic configuration is produced by providers at runtime. Code that reads dynamic config at startup, or stores static config in a runtime struct, is a correctness bug.

## Breaking changes

- Flag breaking changes to exported Go APIs and to CRD schemas under `traefik.io/v1alpha1`.
- Flag configuration changes that affect existing deployments.

## Performance

- Superlinear algorithms where linear suffices, and avoidable allocations or I/O, but only in hot paths where it measurably matters.

## Maintainability

- Use `github.com/rs/zerolog` exclusively — do not import `log`, `slog`, or `logrus`.
- Interfaces: prefer single-method, `-er` suffix, declared at the usage site.
- Idiomatic Go: early returns and guard clauses, the standard library (`slices`, `maps`, `cmp`) over premature abstraction or third-party helpers, grouped import ordering, exported items with a doc comment starting with the item name.
- Comments must explain *why*, not *what*; the code already says what. Every comment must end with a period.
- Tests: new behaviour needs tests using `testify/assert` and `testify/require`; use `require` for preconditions that must stop the test, `assert` for independent checks. Tests must be table-driven with `t.Parallel()`. Blackbox testing (`package x_test`) is preferred. Do not add verbose message arguments to `assert`/`require` calls — the test name provides the context.

## Do not flag

- Generated code: files matching `zz_generated*.go`, everything under `pkg/provider/kubernetes/crd/generated/`, and `webui/static/`.
- Test mocks: files matching `mock_*.go` or `*_mock.go`.
- `//nolint:` directives — they are intentional.
- Integration test fixtures under `integration/fixtures/` — Docker-dependent behaviour cannot be verified statically.
- Patterns already used consistently across the codebase.

Before claiming a change is unsafe, check how the changed functions are used elsewhere in the repository: the call sites decide.
