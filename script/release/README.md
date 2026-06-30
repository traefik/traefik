# Release procedure

This directory contains tooling for preparing Traefik patch releases.

## Automated (preferred)

Use the Claude Code skill:

```
/release v2.11.51
```

The skill handles all steps below automatically.

## Manual procedure

Use this when Claude Code is unavailable.

### Prerequisites

- `git-cliff` installed (`cargo install git-cliff` or via package manager)
- `gh` CLI authenticated
- Current branch must be the release branch (e.g. `v2.11`) and up to date with upstream

### Steps

**1. Check out a release branch**

```bash
git fetch upstream
git checkout v2.11
git pull upstream v2.11
git checkout -b prepare-release-v2.11.51
```

**2. Generate the changelog**

Set a GitHub token to avoid rate limiting:

```bash
export GITHUB_TOKEN=$(gh auth token)
```

Run git-cliff:

```bash
git cliff --config script/release/cliff.toml --tag v2.11.51 v2.11.50..v2.11
```

**3. Update CHANGELOG.md**

Copy the output and prepend it to `CHANGELOG.md`.

Within each section (Bug fixes, Enhancement, Documentation, Misc), sort entries alphabetically:
- Entries with a `**[area]**` tag: sort by the tag text
- Entries without a tag: sort by full line text, place after tagged entries

Remove blank lines between entries within a section.

**4. Commit and push**

```bash
git add CHANGELOG.md
git commit -m "Prepare release v2.11.51"
git push origin prepare-release-v2.11.51
```

**5. Open a pull request**

Target branch: `v2.11` on `traefik/traefik`.

Read the current codename from `.github/workflows/release.yaml` (`CODENAME:` field).

PR title: `Prepare release v2.11.51`

PR body:
```
### What does this PR do?

Prepare release v2.11.51.

aka `<codename>`

### Motivation

To create a new release.

### More

- [ ] Added/updated tests
- [x] Added/updated documentation
```

Labels: `area/documentation`, `size/S`
