---
name: release
description: Prepare a Traefik release (patch, minor RC1, minor RC2+, major RC) — generates changelog, bumps versions, commits, pushes, opens PR
user-invocable: true
---

# Release skill

Invocation: `/release <new-tag>`

Examples:
- `/release v2.11.51`      — patch release
- `/release v3.8.0-rc.1`   — minor RC1 (from master)
- `/release v3.8.0-rc.2`   — minor RC2+ (from v3.8 branch)
- `/release v4.0.0-rc.1`   — major RC1 (from master)

## Step 1 — Parse tag and detect release type

From `NEW_TAG`, extract:
- `MAJOR`, `MINOR`, `PATCH`, `RC_NUM` (0 if not an RC)
- `IS_RC`: true if tag contains `-rc.`
- `IS_RC1`: true if `RC_NUM == 1`

**Release types and their properties:**

| Type | Example | Branch | Prev tag detection |
|------|---------|--------|-------------------|
| Patch | `v2.11.51` | `v{major}.{minor}` | last `v{major}.{minor}.*` tag |
| RC1 | `v3.8.0-rc.1` | `master` | last `v{major}.{minor-1}.0-rc.*` or `ea.*` tag |
| RC2+ | `v3.8.0-rc.2` | `v{major}.{minor}` | `v{major}.{minor}.0-rc.{RC_NUM-1}` |

Set `BRANCH` and `RELEASE_BRANCH=prepare-release-${NEW_TAG}`.

Detect `PREV_TAG`:
- **Patch**: `git tag --sort=-version:refname | grep "^v${MAJOR}\.${MINOR}\." | grep -v "^${NEW_TAG}$" | grep -v "\-rc\." | head -1`
- **RC1**: `git tag --sort=-version:refname | grep -E "^v${MAJOR}\.$(( MINOR - 1 ))\.0-(rc|ea)\." | head -1`
- **RC2+**: construct directly as `v${MAJOR}.${MINOR}.0-rc.$(( RC_NUM - 1 ))`

If `PREV_TAG` is empty: ask the user "Could not auto-detect previous tag. What is the previous tag?"

## Step 2 — Pre-flight checks (fail fast, clear messages)

Run these checks in order:

1. **Current branch**: `git branch --show-current` must equal `BRANCH`. If not: `Error: must be on branch ${BRANCH}, currently on <current>.`
2. **Branch does not exist**: `git branch --list ${RELEASE_BRANCH}` must be empty. If not: `Error: branch ${RELEASE_BRANCH} already exists.`
3. **Up to date**: run `git fetch upstream` then `git rev-list HEAD..upstream/${BRANCH} --count`. If count > 0: `Error: branch is N commits behind upstream/${BRANCH}. Run: git pull upstream ${BRANCH}`
4. **CODENAME**: read `.github/workflows/release.yaml`, extract value after `CODENAME:`. If missing: ask "What is the current release codename? (check .github/workflows/release.yaml)"
5. **GITHUB_TOKEN**: use `$GITHUB_TOKEN` if set, else run `gh auth token`. Export for use in git-cliff.
6. **Fork remote**: run `git remote -v` and `gh api user --jq .login` to find the remote whose push URL contains the GitHub username. If ambiguous: ask "Which remote is your fork?"

Show summary and ask **"Proceed? (y to continue)"**:
```
Release type: <Patch | Minor RC1 | Minor RC2+ | Major RC1>
New tag:      ${NEW_TAG}
Prev tag:     ${PREV_TAG}
Branch:       ${BRANCH} → ${RELEASE_BRANCH}
Codename:     ${CODENAME}
```

## Step 3 — Create release branch

```bash
git checkout -b ${RELEASE_BRANCH}
```

## Step 4 — [RC1 only] Bump version in documentation

Detect old minor version: `OLD_MINOR=v${MAJOR}.$(( MINOR - 1 ))`, `NEW_MINOR=v${MAJOR}.${MINOR}`.

Replace all occurrences of `${OLD_MINOR}` with `${NEW_MINOR}` in all git-tracked files under `docs/`, excluding `docs/dist/`:

```bash
git ls-files docs/ | grep -v '^docs/dist/' | xargs sed -i '' "s/${OLD_MINOR}/${NEW_MINOR}/g"
```

Also update `cmd/traefik/traefik.go` (contains a doc URL with the version).

Show the list of changed files. **Pause 1a** — ask: "Review the version bump changes in your editor. (y to commit)"

Commit:
```bash
git add $(git ls-files docs/ | grep -v '^docs/dist/') cmd/traefik/traefik.go
git commit -m "Bump documentation references from ${OLD_MINOR} to ${NEW_MINOR}"
```

## Step 5 — [RC1 only] Update CODENAME

Ask the user: "What is the new codename for ${NEW_MINOR}?"

Update the `CODENAME:` line in `.github/workflows/release.yaml`:
- Find the line matching `CODENAME:` and replace its value with the new codename.

**Pause 1b** — ask: "CODENAME updated to <new_codename> in release.yaml. (y to commit)"

```bash
git add .github/workflows/release.yaml
git commit -m "Update release codename to <new_codename>"
```

## Step 6 — Generate changelog

```bash
GITHUB_TOKEN=${GITHUB_TOKEN} git cliff --config script/release/cliff.toml --tag ${NEW_TAG} ${PREV_TAG}..${BRANCH}
```

## Step 7 — Process changelog output

Apply to the raw git-cliff output:

1. **Sort entries within each section** alphabetically:
   - Entries with `**[area]**` tag: sort by the tag text inside `**[...]**`
   - Entries without a tag: sort by full line text, place after tagged entries
2. **Remove blank lines between entries** within a section (blank lines between sections are preserved)

Prepend the processed result to the top of `CHANGELOG.md`.

Show the added lines (first ~30 lines of the new block).

**Pause 2** — ask: "CHANGELOG.md updated. Review in your editor, adjust if needed. (y to commit)"

```bash
git add CHANGELOG.md
git commit -m "Prepare release ${NEW_TAG}"
```

## Step 8 — Push

**Pause 3** — ask: "Ready to push branch ${RELEASE_BRANCH} to ${FORK_REMOTE}. (y to push)"

```bash
git push ${FORK_REMOTE} ${RELEASE_BRANCH}
```

## Step 9 — Open pull request

**Pause 4** — ask: "Ready to open PR against traefik/traefik:${BRANCH}. (y to open PR)"

```bash
gh pr create \
  --repo traefik/traefik \
  --base ${BRANCH} \
  --head ${GITHUB_USER}:${RELEASE_BRANCH} \
  --title "Prepare release ${NEW_TAG}" \
  --label "area/documentation" \
  --label "size/S" \
  --body "$(cat <<'PRBODY'
### What does this PR do?

Prepare release ${NEW_TAG}.

aka `${CODENAME}`

### Motivation

To create a new release.

### More

- [ ] Added/updated tests
- [x] Added/updated documentation
PRBODY
)"
```

Print the PR URL.

---

## Reference: script/release/README.md

For manual procedure when this skill is unavailable, see `script/release/README.md`.
