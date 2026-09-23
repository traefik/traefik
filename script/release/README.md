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
- Current branch must be the release branch and up to date with upstream

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
git fetch upstream
git checkout v2.11
git pull upstream v2.11
git checkout -b prepare-release-v2.11.51
```

**2. Generate the changelog**

```bash
export GITHUB_TOKEN=$(gh auth token)
git cliff --config script/release/cliff.toml --tag v2.11.51 v2.11.50..v2.11
```

**3. Update CHANGELOG.md**

Prepend the output to `CHANGELOG.md`.

Within each section (Bug fixes, Enhancement, Documentation, Misc), sort entries alphabetically:
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
git fetch upstream
git checkout master
git pull upstream master
git checkout -b prepare-release-v3.8.0-rc.1
```

**2. Bump version in documentation**

```bash
git ls-files docs/ | grep -v '^docs/dist/' | xargs sed -i '' 's/v3\.7/v3\.8/g'
# Also update cmd/traefik/traefik.go
sed -i '' 's/v3\.7/v3\.8/g' cmd/traefik/traefik.go
git add $(git ls-files docs/ | grep -v '^docs/dist/') cmd/traefik/traefik.go
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
git cliff --config script/release/cliff.toml --tag v3.8.0-rc.1 ${PREV_TAG}..master
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
- git-cliff range: `v3.8.0-rc.1..v3.8`
- Only CHANGELOG.md is modified (no docs bump, no codename change)
