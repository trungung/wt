# Acceptance Criteria (v1)

## A) General

- A1: Running `wt` outside a git repo exits non-zero and prints a clear error to stderr.
- A2: `wt <branch>` prints only the resolved path to stdout on success.

## B) Listing

- B1: `wt` outputs one line per worktree: `branch<TAB>path`.
- B2: Detached worktrees show branch as `(detached)`.

## C) Ensure (`wt <branch>`)

- C1: `wt feature/x` creates worktree at `$REPO_PATH.wt/feature-x` if missing.
- C2: Running `wt feature/x` again prints the same path and does not create a duplicate.
- C3: If local `feature/x` does not exist but `origin/feature/x` exists, `wt feature/x` creates a local tracking branch and a worktree.
- C4: If `feature/x` does not exist anywhere:
  - with `--from dev`, branch is created from `dev`
  - without `--from`, branch is created from default branch
- C5: Copy patterns are applied only on creation and do not overwrite existing destination files.
- C6: If two distinct branches sanitize to the same directory name, `wt <branch>` fails with a clear collision error and does not create/reuse that directory.

## D) Default branch behavior

- D1: `wt <default-branch>` prints `$REPO_PATH` and does not create `$REPO_PATH.wt/<default-branch>`.

## E) Post-create atomic rollback

- E1: If any `postCreateCmd` step fails during creation of a new worktree, `wt <branch>` exits non-zero.
- E2: On postCreate failure, `wt` attempts rollback and removes the newly created worktree (no registered worktree remains for that branch due to this attempt).
- E3: If `wt` created the branch during this invocation and postCreate fails, `wt` deletes that branch during rollback.
- E4: After a postCreate failure + rollback, rerunning `wt <branch>` behaves like a first-time creation attempt (not “half-created”).

## F) Remove

- F1: `wt remove feature/x` removes the worktree if clean.
- F2: If the worktree is dirty, `wt remove feature/x` warns and requires confirmation; if user declines, it does not remove and exits non-zero.
- F3: `wt remove feature/x --force` removes even if dirty.
- F4: If `deleteBranchWithWorktree=true`, the local branch is deleted after removal except:
  - the default branch
  - the branch currently checked out in the main worktree
- F5: `wt remove <default-branch>` refuses (v1 cannot remove the main worktree).

## G) Prune

- G1: `wt prune` prunes only worktrees whose branches are merged into default branch (git ancestry-based).
- G2: `wt prune` does not prune dirty merged worktrees unless `--force`.
- G3: `wt prune --dry-run` removes nothing and lists candidates in stable line-based output.
- G4: `wt prune --fetch` fetches before computing merged branches.
- G5: Default branch is never pruned (and v1 should not create a separate default-branch worktree).

## H) Init

- H1: `wt init --yes` creates `$REPO_PATH/.wt.config.json` with defaults if missing (fails if `origin/HEAD` cannot be detected).
- H2: `wt init` (interactive) requires manual input for default branch if `origin/HEAD` is missing; on cancellation it leaves no partial/invalid file.
- H3: If config exists, `wt init` does not overwrite and exits 0 printing the path.

## I) Health

- I1: If `.wt.config.json` exists but is invalid JSON, `wt health` reports ERROR and exits non-zero.
- I2: Unknown config keys produce WARN (not ERROR).
- I3: If default branch cannot be determined (no `origin/HEAD`) and no `defaultBranch` override exists, `wt health` reports ERROR and exits non-zero.
- I4: `wt health` reports OK if `defaultBranch` is configured or `origin/HEAD` is present.
- I5: If copy patterns are configured but cannot be applied, `wt health` warns.
- I6: If two branches collide to the same sanitized path leaf, `wt health` reports ERROR.

## J) Completion

- J1: `wt completion zsh` outputs a valid Zsh completion script to stdout.
- J2: `wt completion bash` outputs a valid Bash completion script to stdout.
- J3: `wt completion fish` outputs a valid Fish completion script to stdout.

## K) Shell Setup

- K1: `wt shell-setup` outputs a shell wrapper function and completion source command.
- K2: The wrapper enables `wt cd <branch>` to change the shell's working directory.

---
