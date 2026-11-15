# Release procedure

This directory contains tooling for preparing Traefik releases.

## Automated (preferred)

Use the Claude Code skill:

```
/release <new-tag>
```

Examples:
- `/release v2.11.51`     — patch release
- `/release v3.8.0-rc.1`  — minor/major RC1
- `/release v3.8.0-rc.2`  — minor/major RC2+

## Manual procedure

Use this when Claude Code is unavailable.

### Prerequisites

- `git-cliff` installed (`cargo install git-cliff` or via package manager)
- `gh` CLI authenticated
- `upstream` remote pointing to `traefik/traefik`

### Release types

| Tag | Branch | Prev tag |
|-----|--------|----------|
| `v2.11.51` (patch) | `v2.11` | last `v2.11.*` tag |
| `v3.8.0-rc.1` (minor RC1) | `master` | last `v3.7.0-rc.*` or `ea.*` tag |
| `v3.8.0-rc.2` (minor RC2+) | `v3.8` | `v3.8.0-rc.1` |

---

### Patch release (`v2.11.51`)

**1. Check out a release branch**

```bash
git fetch upstream --tags
RANGE_END=$(git rev-parse upstream/v2.11)
git checkout -b prepare-release-v2.11.51 --no-track ${RANGE_END}
```

**2. Generate the changelog**

git-cliff reads the local refs: use the upstream commit as the range end, not a local branch that may be stale.
Do not use `upstream/v2.11` either: git-cliff sends it to the GitHub API, which rejects it.

```bash
export GITHUB_TOKEN=$(gh auth token)
git cliff --config script/release/cliff.toml --tag v2.11.51 v2.11.50..${RANGE_END}
```

**3. Update CHANGELOG.md**

Prepend the output to `CHANGELOG.md`.

Within each section (Enhancement, Bug fixes, Documentation, Misc), sort entries alphabetically:
- Entries with `**[area]**`: sort by tag text
- Entries without tag: sort by full line text, place after tagged entries

Remove blank lines between entries within a section.

**4. Commit, push and open PR**

```bash
git add CHANGELOG.md
git commit -m "Prepare release v2.11.51"
git push origin prepare-release-v2.11.51
```

PR title: `Prepare release v2.11.51`
PR base: `v2.11` on `traefik/traefik`
Labels: `area/documentation`, `size/S`

PR body:
```
### What does this PR do?

Prepare release v2.11.51.

aka `<codename from .github/workflows/release.yaml>`

### Motivation

To create a new release.

### More

- [ ] Added/updated tests
- [x] Added/updated documentation
```

---

### Minor/Major RC1 (`v3.8.0-rc.1`)

**1. Check out master and create release branch**

```bash
git fetch upstream --tags
RANGE_END=$(git rev-parse upstream/master)
git checkout -b prepare-release-v3.8.0-rc.1 --no-track ${RANGE_END}
```

**2. Bump version in documentation**

Replace only the doc URLs (`traefik/v3.7/` → `traefik/v3.8/`) and the Docker images (`traefik:v3.7` → `traefik:v3.8`).
Replacing every `v3.7` would also rewrite the `## v3.7.x` sections of the migration guide and unrelated versions in `go.mod`.
The URLs are also in Go sources and generated CRD files, not only in `docs/`.
In `docs/content/migrate/`, keep the `raw.githubusercontent.com/traefik/traefik/v3.7/` links:
each one belongs to the section of the release that introduced the CRD change.

```bash
git grep -l -E 'traefik/v3\.7/|traefik:v3\.7' -- . ':!CHANGELOG.md' ':!docs/dist/' > /tmp/bump-files.txt
python3 - /tmp/bump-files.txt <<'EOF'
import sys
for path in open(sys.argv[1]).read().split():
    s = open(path, encoding="utf-8").read()
    if path.startswith("docs/content/migrate/"):
        s = s.replace("doc.traefik.io/traefik/v3.7/", "doc.traefik.io/traefik/v3.8/")
    else:
        s = s.replace("traefik/v3.7/", "traefik/v3.8/").replace("traefik:v3.7", "traefik:v3.8")
    open(path, "w", encoding="utf-8").write(s)
EOF
```

Look for doc URLs to older minors brought back by merges from maintenance branches, and bump them too:

```bash
git grep -n -E 'doc\.traefik\.io/traefik/v[0-9]+\.[0-9]+/' -- '*.go' docs/ ':!docs/content/migrate/' ':!docs/dist/' | grep -v '/v3.8/'
```

Check that the generated files are in sync (on macOS, `make generate-crd` fails at its last `sed` step and leaves a leading `---` in `kubernetes-crd-definition-v1.yml`: revert it):

```bash
make generate && git diff --exit-code
make generate-crd && git diff --exit-code
```

```bash
git add $(git diff --name-only)
git commit -m "Bump documentation references from v3.7 to v3.8"
```

**3. Update CODENAME in `.github/workflows/release.yaml`**

Edit the `CODENAME:` field to the new codename for v3.8.

```bash
git add .github/workflows/release.yaml
git commit -m "Update release codename to <new_codename>"
```

**4. Generate the changelog**

```bash
export GITHUB_TOKEN=$(gh auth token)
# Detect previous RC1 tag
PREV_TAG=$(git tag --sort=-version:refname | grep -E "^v3\.7\.0-(rc|ea)\." | head -1)
# Ignore the v3.7 patch tags merged into master, otherwise git-cliff starts a release block at each of them.
git cliff --config script/release/cliff.toml --ignore-tags 'v2\..*|v3\.7\.[0-9]+$' --tag v3.8.0-rc.1 ${PREV_TAG}..${RANGE_END}
```

**5. Update CHANGELOG.md, commit, push and open PR**

Same as patch procedure above, with:
- Commit message: `Prepare release v3.8.0-rc.1`
- PR base: `master`
- PR title: `Prepare release v3.8.0-rc.1`

---

### Minor/Major RC2+ (`v3.8.0-rc.2`)

Same as patch procedure but:
- Branch: `v3.8`
- Prev tag: `v3.8.0-rc.1`
- git-cliff range: `v3.8.0-rc.1..$(git rev-parse upstream/v3.8)`
- Only CHANGELOG.md is modified (no docs bump, no codename change)
