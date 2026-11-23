---
name: merge-back
description: Merge a Traefik maintenance branch back into the next one (oldest vX.Y → ... → master) — branch, merge, resolve conflicts, verify, commit, push, open and label the PR
user-invocable: true
---

# Merge-back skill

Usage: `/merge-back <source> <target>`. `<source>` is a maintenance branch (`vX.Y`). `<target>` is the next maintenance branch or `master`.

## Rules

- The chain goes from the oldest maintained branch to `master`, one minor at a time. The PR template of the repository lists the maintained branches. A step starts only when the PR of the previous step is merged.
- Merge `upstream/<source>` into `upstream/<target>`. Never use the local or `origin` branches as a merge side.
- The work branch is `merge-<source>`. Reuse it for every merge of that source and force-push it.
- Do not commit while a failure caused by the merge remains.
- Ask the user when a conflict is hard to resolve, when no rule below covers it, or when you have a doubt. Show both sides and the resolution you propose. Do not guess.
- Run outside the sandbox: network git commands, `gh`, branch checkouts, and the Go commands that use the build cache (`go build`, `go test`, `go mod tidy`, `go generate`).
- Write command output to the scratchpad and grep it. Do not re-run a command to read its output again.

## 1. Pre-flight

1. The working tree is clean: `git status --porcelain` prints nothing.
2. Fetch and fast-forward the local branches. Stop and report when `--ff-only` fails, do not reset the branch.

   ```
   git fetch upstream <source> <target>
   git checkout <source> && git merge --ff-only upstream/<source>
   git checkout <target> && git merge --ff-only upstream/<target>
   ```

3. No merge-back PR into `<source>` is open: `gh pr list -R traefik/traefik --search "Merge branch into <source> in:title"` prints nothing.
4. There is something to merge: `git log --oneline upstream/<target>..upstream/<source>` is not empty.

## 2. Merge

```
git checkout -B merge-<source> upstream/<target>
git merge --no-ff --no-commit upstream/<source>
```

No conflict: go to step 4.

## 3. Resolve conflicts

Before you edit a file, read what each side changed since the merge base:

```
MB=$(git merge-base upstream/<target> upstream/<source>)
git diff $MB upstream/<source> -- <file>
git diff $MB upstream/<target> -- <file>
```

### Textual conflicts

| File | Rule |
|------|------|
| `CHANGELOG.md` | Keep both sides. Order the sections by release date, newest first, not by version. Same day: highest version first. |
| `docs/content/migrate/v3.md` | Keep both sides, target section first, separated by a `---` line. Then apply the migration notes rules. |
| Other docs | Keep the target text. Apply the source link, typo and wording fixes to it. |
| `go.mod` | Take the higher version of each module, then `git checkout --ours go.sum && go mod tidy`. Check the result with `git diff upstream/<target> -- go.mod`: GoLand can revert a bump in the background. |
| Go code | Keep the target addition, apply the source move or rename. |
| Tests | Keep the source wording. Update the target-only test cases for the source behavior change, with the source test helpers. |

### Migration notes

The source adds its notes under a `## <source patch>` heading in `docs/content/migrate/v3.md`. Keep that section as is, even when the merge is automatic.

- Target with GA releases: copy each note under the target patch that ships the same fix. The target patch is the `CHANGELOG.md` section that lists the PR of the note, so the notes of one source patch can go to different target patches. In the copy, replace the version and the branch of the links (`/<source>/` to `/<target>/`). Create the target section at the top of the file when it does not exist, in version order, separated by `---` lines. The copy is part of the merge commit.
- Target without GA release (RC branch or `master`): do not copy. The release preparation adds the notes under the `## vX.Y.0` section.

### Semantic conflicts

The merge is clean but the tests fail. Known shapes:

- The source added a validation, the target added a syntax: extend the validation.
- The target added an option that gates a code path, the source added a call on that path: route the call through the gate.
- The source changed a generated output, the target added tests on the old output: update the target expectations.

## 4. Verify

1. No conflict marker: `git grep -nE '^(<<<<<<<|>>>>>>>) ' --cached` prints nothing.
2. `go build ./...` passes.
3. The tests of the touched packages pass. Use this while you resolve conflicts:

   ```
   go test $(git diff --cached --name-only -- '*.go' | xargs -n1 dirname | sort -u | sed 's#^#./#')
   ```

4. `golangci-lint run` passes on the packages edited by hand. Skip when the merge was automatic.
5. `go generate` leaves the tree unchanged.
6. All unit tests pass:

   ```
   go test ./pkg/... ./cmd/... > <scratchpad>/test-unit.log 2>&1; echo $?
   grep -E '^(FAIL|--- FAIL|panic:)' <scratchpad>/test-unit.log
   ```

A failing test in a package that no side changed is pre-existing until proven otherwise. Confirm it in a clean worktree of `upstream/<target>` (`git worktree add --detach`), report it, do not fix it.

## 5. Commit

```
git add -A
git commit -m "Merge branch <source> into <target>"
```

Subject only, no body, no trailer. `git log -1 --format='%p'` shows `upstream/<target>` then `upstream/<source>`.

## 6. Pull request

```
git push --force-with-lease -u origin merge-<source>
gh pr create -R traefik/traefik --base <target> --head <login>:merge-<source> \
  --title "Merge branch <source> into <target>" --body "..."
gh pr edit <number> -R traefik/traefik \
  --add-label status/2-needs-review --add-label bot/light-review --add-label bot/merge-method-ff \
  --milestone <milestone>
gh pr view <number> -R traefik/traefik --json labels,milestone
```

- `<milestone>`: the target branch without the `v` (`X.Y` for `vX.Y`), or `next` when the target is `master`.
- `status/2-needs-review`: enters the review queue.
- `bot/light-review`: one LGTM is enough instead of three.
- `bot/merge-method-ff`: the bot fast-forwards, so the merge commit lands as is on `<target>`.

Body (repository template without the HTML comment):

```
### What does this PR do?

This pull request merges the current <source> branch into <target>.

### Motivation

To be up to date.

### More

- ~[ ] Added/updated tests~
- ~[ ] Added/updated documentation~

### Additional Notes

<One bullet per conflict resolved by hand. Nothing when the merge was automatic.>

Assisted-by: <model full name>
```

A conflict bullet is one line: the file or package, then what was done. No diff, no reasoning, no command output.

```
- `pkg/muxer/http/matcher.go`: the Host rule regex now accepts the syntax added in <target>.
- `go.mod`: kept the higher quic-go version from <source> and tidied.
```

## 7. Report

Commit and parents, PR link, hand-resolved conflicts, skipped pre-existing failures.
