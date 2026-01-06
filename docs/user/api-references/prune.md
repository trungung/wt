# wt prune

Remove worktrees whose branches are merged into default branch.

## Usage

```bash
wt prune [--dry-run] [--force] [--fetch]
```

## Description

Scans all worktrees, identifies branches merged into default branch, and removes their worktrees. Optionally deletes branches if configured.

## Options

### `--dry-run`

Preview what would be removed without actually removing anything.

```bash
wt prune --dry-run
```

**Output:** Lists candidate worktrees that would be pruned.

### `--force`, `-f`

Force removal even if worktree is dirty (has uncommitted changes).

**Warning:** You may lose uncommitted work.

```bash
wt prune --force
```

**Without `--force`:**

- Skip dirty worktrees and report them
- Do not remove dirty worktrees

### `--fetch`

Run `git fetch --prune` before scanning for merged branches.

```bash
wt prune --fetch
```

Useful to ensure remote branches are up-to-date.

## Behavior

1. **Determine default branch:**
   - From config override (`defaultBranch`)
   - Or from `origin/HEAD` symbolic ref

2. **Optional fetch:**
   - If `--fetch`: run `git fetch --prune`
   - Updates remote branch tracking

3. **Identify merged branches:**
   - For each worktree (excluding main and detached):
   - Check if branch is merged into default branch (git ancestry check)
   - Branches ahead of default branch or unrelated: not merged

4. **Filter worktrees:**
   - Skip main worktree
   - Skip detached worktrees
   - Skip default branch (shouldn't exist as separate worktree)
   - Skip dirty worktrees unless `--force` provided

5. **For each merged worktree:**
   - If `--dry-run`: print to stdout
   - Else: remove worktree and optionally delete branch (if configured)

## Merged Branch Detection

A branch is considered "merged" if its commits are ancestors of default branch's HEAD.

```bash
# Internally runs (conceptually):
git merge-base --is-ancestor <branch> <default-branch>
```

**Examples:**

- Branch `feature/new-auth` merged into `main`: ✅ Merged → prune
- Branch `feature/billing` not yet merged: ❌ Not merged → skip
- Branch `feature/payment` diverged (different history): ❌ Not merged → skip

## Examples

### Dry-run preview

```bash
$ wt prune --dry-run
Candidates for pruning:
  feature/new-auth
  feature/old-user-ui
  bugfix/header-issue

Total candidates: 3 (run without --dry-run to prune)
```

### Prune merged worktrees

```bash
$ wt prune
Pruned 3 worktrees.
```

### Prune with fetch

```bash
$ wt prune --fetch
# First: git fetch --prune
# Then: scan and prune
Pruned 2 worktrees.
```

### Force prune (remove dirty worktrees)

```bash
$ wt prune --force
# Removes even worktrees with uncommitted changes
Pruned 4 worktrees.
```

### Dry-run + force (preview with force)

```bash
$ wt prune --dry-run --force
Candidates for pruning (including dirty):
  feature/new-auth (dirty)
  feature/payment (dirty)

Total candidates: 2 (run without --dry-run to prune)
```

### Prune and delete branches (configured)

Config:

```json
{
  "deleteBranchWithWorktree": true
}
```

Command:

```bash
$ wt prune
Pruned 3 worktrees, deleted 3 branches.
```

## Safeguards

`wt prune` never removes:

### Default Branch

Default branch worktree (or main worktree) is never pruned.

### Detached Worktrees

Worktrees with detached HEAD are skipped (cannot determine branch for merge check).

### Unmerged Branches

Branches not merged into default branch are skipped.

### Dirty Worktrees (without `--force`)

```bash
$ wt prune
Skipping feature/new-auth (dirty)
Skipping feature/payment (dirty)
Pruned 1 worktree.
```

## Exit Codes

- `0`: Success (even if nothing to prune)
- `1`: Git error during fetch or merge check
- `2`: Invalid option or usage error

## See Also

- [wt remove](remove.md) - Remove specific worktree
- [Configuration Reference](configuration.md#deletebranchwithworktree) - Branch deletion setting
- [wt](list.md) - List all worktrees
