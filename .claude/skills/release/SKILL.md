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

1. **Clean working tree**: `git status --porcelain` must be empty. If not: `Error: working tree has uncommitted changes.`
2. **Branch does not exist**: `git branch --list ${RELEASE_BRANCH}` must be empty. If not: `Error: branch ${RELEASE_BRANCH} already exists.`
3. **Upstream tip**: run `git fetch upstream --tags`, then set `RANGE_END=$(git rev-parse upstream/${BRANCH})`.
   The release branch is created from this commit and git-cliff uses it as the range end.
   Do not use the local `${BRANCH}`: it may be stale, and git-cliff reads local refs.
   Do not use `upstream/${BRANCH}` as the range end either: git-cliff sends it to the GitHub API, which rejects it (404).
4. **CODENAME**: read `.github/workflows/release.yaml` at `${RANGE_END}`, extract value after `CODENAME:`. If missing: ask "What is the current release codename? (check .github/workflows/release.yaml)"
5. **GITHUB_TOKEN**: use `$GITHUB_TOKEN` if set, else run `gh auth token`. Export for use in git-cliff.
6. **Fork remote**: run `git remote -v` and `gh api user --jq .login` to find the remote whose push URL contains the GitHub username. If ambiguous: ask "Which remote is your fork?"

Show summary and ask **"Proceed? (y to continue)"**:
```
Release type: <Patch | Minor RC1 | Minor RC2+ | Major RC1>
New tag:      ${NEW_TAG}
Prev tag:     ${PREV_TAG}
Branch:       upstream/${BRANCH} (${RANGE_END}) → ${RELEASE_BRANCH}
Codename:     ${CODENAME}
```

## Step 3 — Create release branch

```bash
git checkout -b ${RELEASE_BRANCH} --no-track ${RANGE_END}
```

## Step 4 — [RC1 only] Bump version in documentation

Detect old minor version: `OLD_MINOR=v${MAJOR}.$(( MINOR - 1 ))`, `NEW_MINOR=v${MAJOR}.${MINOR}`.

Replace only the doc URLs and image tags, in all git-tracked files except `CHANGELOG.md` and `docs/dist/`:
- `traefik/${OLD_MINOR}/` → `traefik/${NEW_MINOR}/` (`doc.traefik.io/traefik/…` and `raw.githubusercontent.com/traefik/traefik/…` URLs)
- `traefik:${OLD_MINOR}` → `traefik:${NEW_MINOR}` (Docker images)

Do not replace every `${OLD_MINOR}` occurrence:
it would rewrite the `## ${OLD_MINOR}.x` sections of `docs/content/migrate/v3.md` and unrelated dependency versions in `go.mod`.

In `docs/content/migrate/`, keep the `raw.githubusercontent.com/traefik/traefik/${OLD_MINOR}/` links:
each one belongs to the section of the release that introduced the CRD change, and must keep pointing to that minor's manifest
(e.g. the `## v3.7.8` section says "manifest for v3.7").

The URLs are not only in `docs/`: they are also in Go sources (`cmd/`, `pkg/cli/`, `pkg/config/dynamic/`, the CRD types and their generated applyconfiguration),
and in `integration/fixtures/k8s/01-traefik-crd.yml`.
Bumping only `docs/` leaves the generated CRD definitions out of sync with the Go types.

Use a literal string replacement on the listed files (no regex `sed`, where `.` matches any character):

```bash
git grep -l -E "traefik/${OLD_MINOR//./\\.}/|traefik:${OLD_MINOR//./\\.}" -- . ':!CHANGELOG.md' ':!docs/dist/' > /tmp/bump-files.txt
python3 - "${OLD_MINOR}" "${NEW_MINOR}" /tmp/bump-files.txt <<'EOF'
import sys
old, new, files = sys.argv[1:]
for path in open(files).read().split():
    s = open(path, encoding="utf-8").read()
    if path.startswith("docs/content/migrate/"):
        s = s.replace(f"doc.traefik.io/traefik/{old}/", f"doc.traefik.io/traefik/{new}/")
    else:
        s = s.replace(f"traefik/{old}/", f"traefik/{new}/").replace(f"traefik:{old}", f"traefik:{new}")
    open(path, "w", encoding="utf-8").write(s)
EOF
```

Check that every raw CRD link in the migration guide points to the minor of its section:

```bash
awk '/^## /{s=$2} /raw.githubusercontent.com\/traefik\/traefik\/v[0-9]/{match($0,/traefik\/traefik\/v[0-9]+\.[0-9]+/); print NR, s, substr($0,RSTART+16,RLENGTH-16)}' docs/content/migrate/v3.md
```

Then look for doc URLs to older minors in the Go sources and in the docs outside `docs/content/migrate/`.
Merges from older maintenance branches can bring them back (e.g. a `v3.6` URL on master while preparing `v3.8`):

```bash
git grep -n -E 'doc\.traefik\.io/traefik/v[0-9]+\.[0-9]+/' -- '*.go' docs/ ':!docs/content/migrate/' ':!docs/dist/' | grep -v "/${NEW_MINOR}/"
```

Show the matches and ask the user whether to bump them to `${NEW_MINOR}` too.
The URLs to older minors in `docs/content/migrate/` belong to historical sections: leave them.

Build the changed Go packages (`go build ./cmd/... ./pkg/...`).

The bump edits generated files (CRD definitions, applyconfiguration) by hand.
Run the generators the `validate-generate` CI job runs, and check that they produce no diff:

```bash
make generate && git diff --exit-code
make generate-crd && git diff --exit-code
```

On macOS, the last step of `script/code-gen.sh` (`sed -i '1{/^---$/d;}'`) only works with GNU sed:
with BSD sed it fails, and leaves a leading `---` in `kubernetes-crd-definition-v1.yml`.
That leading `---` is the only expected diff in that case: revert it.

Show the list of changed files. **Pause 1a** — ask: "Review the version bump changes in your editor. (y to commit)"

Commit only the changed files:
```bash
git add $(git diff --name-only)
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
GITHUB_TOKEN=${GITHUB_TOKEN} git cliff --config script/release/cliff.toml --tag ${NEW_TAG} ${PREV_TAG}..${RANGE_END}
```

**[RC1 only]** The `tag_pattern` of `cliff.toml` matches the patch tags of the previous minor (e.g. `v3.7.16`), which are merged into master.
git-cliff would start a new release block at each of them.
Ignore them so that the output is a single `${NEW_TAG}` block:

```bash
GITHUB_TOKEN=${GITHUB_TOKEN} git cliff --config script/release/cliff.toml \
  --ignore-tags "v2\..*|v${MAJOR}\.$(( MINOR - 1 ))\.[0-9]+$" \
  --tag ${NEW_TAG} ${PREV_TAG}..${RANGE_END}
```

Check that the output contains exactly one `## [` header, and that its `[All Commits]` link compares `${PREV_TAG}...${NEW_TAG}`.
If git-cliff panics or prints nothing, check its stderr (e.g. a 404 from the GitHub API).

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
