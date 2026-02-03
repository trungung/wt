## `wt` — v1 Scope Contract (one page)

### Purpose

A fast, branch-centric `git worktree` helper that makes worktrees feel “branch-addressable”: users type branch names, `wt` manages paths, creation, cleanup, and safety checks.

---

## In Scope (v1) — MUST

### Core UX / Behavior

- **Branch-first**: `wt <branch>` ensures worktree exists and prints its path.
- **Deterministic paths**: worktrees live under **`$REPO_PATH.wt/`**.
- **Flattened branch mapping**: `feature/x` → `feature-x` (no nested dirs).
- **Collision policy = FAIL**: if two branches map to same directory name, `wt` errors (no auto-suffix).
- **Default branch special-case**: `wt <default-branch>` prints **`$REPO_PATH`** and does not create a separate default-branch worktree.
- **Dirty detection includes untracked files**.

### Commands

- `wt` — list worktrees (`branch<TAB>path`).
- `wt <branch> [--from <base>]` — ensure/create worktree; create new branch from base if needed.
- `wt remove [branch] [--force]` — remove worktree; warn/confirm if dirty.
- `wt prune [--dry-run] [--force] [--fetch]` — prune worktrees merged into default branch; refuse dirty unless forced.
- `wt init [--yes]` — create config interactively or with defaults; requires manual input if `origin/HEAD` missing.
- `wt health` — validate config + environment; error if default branch cannot be determined.
- `wt completion <shell>` — generate shell completion script (zsh, bash, fish).
- `wt shell-setup [shell]` — generate shell wrapper function + completions for easy navigation.

### Config

- Config file: **`$REPO_PATH/.wt.config.json`**
- Supported keys (v1):
  - `defaultBranch` (optional override)
  - `worktreePathTemplate` (default `$REPO_PATH.wt`)
  - `worktreeCopyPatterns` (copy **only if missing**, create-time only)
  - `postCreateCmd`
  - `deleteBranchWithWorktree`

### Post-create atomicity (best-effort rollback)

- If any `postCreateCmd` step fails:
  - `wt <branch>` exits non-zero
  - `wt` attempts rollback:
    - remove newly created worktree
    - delete newly created branch if it didn’t exist before this run
  - tool prints failure + rollback status

### Shell integration

- Provide completion support for zsh, bash, and fish that suggests:
  - subcommands
  - branch names (all local branches for `wt <branch>`, existing worktree branches for `wt remove`)
- `wt shell-setup` provides `wt cd` navigation and completions for easy setup.

---

## In Scope (v1) — SHOULD (recommended, but not hard blockers)

- `wt prune --fetch` implemented (already in MUST list above; treat as SHOULD if you want to simplify).
- Clear, actionable error messages (what happened, how to fix).
- Stable “porcelain-ish” output for list/dry-run modes.

---

## Out of Scope (v1) — MUST NOT

- Provider APIs (GitHub/GitLab) for PR-merged detection.
- Multi-remote support (anything beyond `origin`).
- `fzf` dependency (interactive selection must work without it).
- `wt run` command (use `wt cd` + shell commands instead).
- Background daemon; watching/syncing files continuously.
- Automatic collision resolution (no hashing/suffixing).

---

## Acceptance “Done” Checklist (v1 exit criteria)

- `wt feature/x` creates/prints `$REPO_PATH.wt/feature-x` (idempotent).
- Collision detected → errors (and `wt health` reports ERROR).
- Default branch (`main`) resolves to `$REPO_PATH` (no separate worktree).
- `remove` and `prune` refuse dirty worktrees unless `--force`.
- `prune` targets branches merged into default branch (git ancestry-based).
- Post-create failure triggers rollback (worktree removed; branch removed if created).
- `.wt.config.json` can be created via `wt init` and validated via `wt health`.
- Shell completions (zsh, bash, fish) suggest subcommands + branches.
- `wt shell-setup` provides `wt cd` navigation and completions.

---
