# PRD: `wt` v1 — Branch-centric Git worktree assistant (Go)

## 0) Summary

`wt` is a fast, branch-centric CLI that simplifies `git worktree` by letting users operate primarily on **branch names**. It deterministically manages worktree paths, safely removes/prunes worktrees, supports repo-local configuration, validates setup via a health check, generates shell completions, and can execute commands inside a worktree.
 Designed to work well with zsh “ghost” suggestions by exposing completion-friendly behavior.

---

## 1) Goals / Non-goals

### 1.1 Goals

- **Branch-first UX**: users think in branches, not folders.
- **Deterministic mapping**: branch → worktree directory is stable and predictable.
- **Low cognitive load**: no manual path management.
- **Snappy CLI**: suitable for frequent invocation (Go binary).
- **Safe lifecycle**: warn/refuse on dirty worktrees unless forced.
- **Repo-local configuration**: JSON config in repo root.
- **Shell-friendly**: scriptable outputs; completion can suggest subcommands + existing worktree branches.

### 1.2 Non-goals (v1)

- GitHub/GitLab API integration (PR-aware merge status).
- Multi-remote support beyond `origin`.
- `fzf` dependency.
- `wt <branch> <script>` shorthand; v1 uses explicit `wt exec`.
- Ongoing file sync; file copy is bootstrap-only.

---

## 2) Principles & Definitions

### 2.1 Repo root

`$REPO_PATH` is the git top-level directory from:

- `git rev-parse --show-toplevel`

### 2.2 Default branch

Default branch is derived from Git (no provider APIs).

**Detection rules (v1):**

1. Config override: `defaultBranch` in `.wt.config.json` (if present)
2. Remote default branch via `refs/remotes/origin/HEAD` symbolic ref target (STRICT)
3. If both above missing: **ERROR** (manual input required during `init`)

`wt health` must error when `origin/HEAD` is missing and no config override is set.

### 2.3 Worktree base path & mapping

Default base directory:

- `$REPO_PATH.wt/`

Worktree directory for a branch:

- `$REPO_PATH.wt/<sanitized-branch>`

**Sanitization rules (strict):**

- Replace `/` with `-`
- Fail if branch name contains illegal characters (anything other than alphanumeric, `-`, `_`, `.`)
- Collision policy (v1): FAIL

- If two distinct branch names map to the same sanitized directory name, `wt` must:
  - fail with a clear collision error
  - not create or reuse the colliding directory
- `wt health` must report this as **ERROR**.

### 2.4 “Dirty” worktree definition

A worktree is **dirty** if `git status --porcelain` in that worktree returns any output, including untracked files.

### 2.5 Output contract

- Errors go to stderr.
- `wt <branch>` prints only the resolved path to stdout on success.
- `wt` list output is stable and parseable.

---

## 3) Configuration

### 3.1 Filename & location

- `$REPO_PATH/.wt.config.json`

### 3.2 Supported keys (v1)

```json
{
  "defaultBranch": "main",
  "worktreePathTemplate": "$REPO_PATH.wt",
  "worktreeCopyPatterns": [".env*", ".vscode/**"],
  "postCreateCmd": ["bun install"],
  "deleteBranchWithWorktree": true
}
```

### 3.3 Semantics (v1)

- `defaultBranch` (optional): overrides detection.
- `worktreePathTemplate`: expands `$REPO_PATH` and yields the worktree base directory. Default `$REPO_PATH.wt`.
- `worktreeCopyPatterns`: applied **only when creating a new worktree**; copy **only if missing** at destination; no overwrites.
- `postCreateCmd`: commands executed after worktree creation, in the new worktree directory.
- `deleteBranchWithWorktree`: if true, `remove` and `prune` delete the local branch after removing its worktree, with safeguards:
  - never delete the default branch
  - never delete the branch currently checked out in the main worktree

### 3.4 Post-create atomicity (v1)

If any `postCreateCmd` step fails, the operation must be treated as failed and **rollback must be attempted**.

**Rollback definition (best-effort, tool-owned resources):**

- Remove the newly created worktree (so no worktree remains registered or on disk).
- If `wt` created the branch during this invocation (it did not exist beforehand), delete that branch too.
- Emit a clear message indicating:
  - which step failed
  - that rollback was attempted
  - whether rollback succeeded

**Explicit limitation:**
Rollback does not undo external side effects performed by postCreate commands outside the worktree (e.g., global caches, network actions). It only guarantees cleanup of the worktree/branch created by `wt`.

---

## 4) Command Specification (v1)

### 4.1 `wt` (list)

**Purpose:** list existing worktrees.

**Behavior:**

- List all worktrees from `git worktree list --porcelain`.
- Output to stdout:  
  `branch<TAB>path` per line  
  Detached worktrees use branch `(detached)`.

**Exit codes:**

- `0` success
- non-zero if not in a git repo or git fails

---

### 4.2 `wt <branch> [--from <base-branch>]` (ensure)

**Purpose:** ensure a worktree exists for `<branch>`, then print its path.

**Default-branch special case:**

- If `<branch>` equals default branch: print `$REPO_PATH` and exit `0`.
- Must not create `$REPO_PATH.wt/<default-branch>`.

**Normal behavior:**

1. If worktree already exists for `<branch>`:
   - print its existing path and exit `0`
   - do not run copy hooks or postCreate
2. Else:
   - Resolve/ensure the branch exists locally:
     - If local `<branch>` exists: use it
     - Else if remote-tracking `origin/<branch>` exists: create local tracking branch
     - Else create a new local branch from:
       - `--from <base-branch>` if provided
       - else default branch
   - Compute target path from template + sanitization:
     - base: `worktreePathTemplate` (default `$REPO_PATH.wt`)
     - leaf: `<sanitized-branch>`
   - If sanitized leaf collides with an existing mapping for another branch: **fail**.
   - Create worktree at computed path.
   - Apply `worktreeCopyPatterns` (copy only if missing).
   - Run `postCreateCmd`.
   - If postCreate fails: rollback as per §3.4.
   - Print worktree path to stdout.

**Constraints:**

- Idempotent.
- Must not change the caller’s working directory.

**Exit codes:**

- `0` success
- non-zero on any failure, including collisions and postCreate failure (after rollback attempt)

---

### 4.3 `wt exec <branch> -- <command...>`

**Purpose:** run an arbitrary command inside a branch’s worktree.

**Behavior:**

- Requires `--` delimiter.
- Resolves target directory:
  - default branch → `$REPO_PATH`
  - other branch → ensure worktree exists (same semantics as `wt <branch>`)
- Execute `<command...>` with working directory set to resolved path.
- Inherit stdin/stdout/stderr.
- Exit code equals executed command’s exit code.

**Exit codes:**

- child exit code
- non-zero for usage errors (missing `--`, missing command) or ensure failures

---

### 4.4 `wt remove [branch] [--force]`

**Purpose:** remove a worktree and optionally delete its branch.

**Behavior:**

- If `branch` provided: remove that worktree.
- If `branch` omitted: interactive selection from existing worktrees (simple numbered list prompt).
- Refuse to remove default branch / main worktree in v1:
  - `wt remove <default-branch>` must refuse.
- Dirty check:
  - If target worktree is dirty and `--force` not provided:
    - warn and require confirmation
    - if declined: do not remove; exit non-zero
- Remove worktree via `git worktree remove` (forced if needed).
- If `deleteBranchWithWorktree=true`:
  - delete local branch unless:
    - it is default branch, or
    - it is checked out in main worktree

**Exit codes:**

- `0` success
- non-zero on refusal, cancellation, or git errors

---

### 4.5 `wt prune [--dry-run] [--force] [--fetch]`

**Purpose:** remove worktrees whose branches are merged into default branch.

**Behavior:**

- Determine default branch per §2.2.
- Optional `--fetch`: run `git fetch --prune` first.
- Determine merged branches (git ancestry-based) relative to default branch.
- For each worktree:
  - skip `$REPO_PATH` main worktree
  - skip `(detached)`
  - skip default branch (shouldn’t exist as a separate worktree anyway)
  - if branch is merged:
    - if dirty and not `--force`: skip and report
    - if `--dry-run`: print candidate line(s)
    - else: remove worktree and optionally delete branch (with safeguards)
- Output:
  - without `--dry-run`: print number pruned
  - with `--dry-run`: list candidates (stable line format) + count

**Exit codes:**

- `0` success (even if nothing to prune)
- non-zero on git errors

---

### 4.6 `wt init [--yes]`

**Purpose:** create `.wt.config.json`.

**Behavior:**

- If config exists:
  - print path and exit `0` (no overwrite)
- If missing:
  - with `--yes`: write defaults without prompts
  - without `--yes`: interactive prompts for v1 keys; write file
- After writing: print config path.

**Exit codes:**

- `0` success
- non-zero on cancellation or write failure (no partial file)

---

### 4.7 `wt health`

**Purpose:** validate config and highlight issues that would otherwise be silently skipped.

**Minimum checks (v1):**

- Repo root resolvable.
- Config (if present):
  - valid JSON (ERROR if not)
  - unknown keys → WARN
- Default branch:
  - ERROR if cannot determine and no override
  - WARN if `origin/HEAD` missing and fallback/override used
- Worktree base directory:
  - writable/creatable (ERROR if not)
- Copy patterns:
  - WARN if patterns configured but cannot be applied
  - WARN if patterns match nothing in repo
- Post-create commands:
  - WARN on empty command strings
- Collision detection:
  - ERROR if two branches sanitize to same leaf

**Output:**

- Human-readable list of OK/WARN/ERROR.

**Exit codes:**

- `0` if no ERROR
- non-zero if any ERROR

---

### 4.8 `wt completion <shell>`

**Purpose:** generate shell completion script.

**Behavior:**

- Outputs completion script to stdout for the specified shell (v1: `zsh`).
- Does not modify any system files.

**Exit codes:**

- `0` success
- non-zero for unsupported shell

---

## 5) Zsh completion / autosuggestion requirements (v1)

### Scope

Completion candidates should include only:

- subcommands: `init`, `exec`, `remove`, `prune`, `health`
- existing worktree branch names (from `git worktree list --porcelain`)

Specific:

- `wt remove` completes only existing worktree branches.
- `wt exec` completes existing worktree branches for `<branch>`.
- Must be compatible with zsh-autocomplete so ghost suggestions can appear while typing (Tab not required).

---
