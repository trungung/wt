# wt remove

Remove a worktree and optionally delete its branch.

## Usage

```bash
wt remove [branch] [--force]
```

## Description

Removes a worktree from the repository. If `deleteBranchWithWorktree` is true in config, also deletes the local branch (with safeguards).

## Arguments

### `[branch]` (optional)

Branch name of worktree to remove.

**Behavior:**

- If provided: remove that specific worktree
- If omitted: interactive selection from existing worktrees (excluding main worktree)

## Options

### `--force`, `-f`

Force removal even if worktree is dirty (has uncommitted changes).

**Warning:** Use with caution. You may lose uncommitted work.

```bash
wt remove feature/new-auth --force
```

**Without `--force`:**

- Check if worktree is dirty (`git status --porcelain` returns output)
- If dirty: warn and require confirmation
- If declined: do not remove, exit with non-zero code

## Behavior

### Interactive Selection (no branch argument)

Lists worktrees (excluding main worktree) and prompts for selection:

```bash
$ wt remove
Select a worktree to remove (Tab to complete): _
```

Supports:

- Tab completion
- Fuzzy matching (partial matches)
- Confirmation before removal

### Dirty Worktree Check

If target worktree is dirty and `--force` not provided:

1. Check `git status --porcelain` in worktree
2. If any output: worktree is dirty
3. Show warning message
4. Prompt for confirmation
5. If user declines: exit without changes (exit code 1)

```bash
$ wt remove feature/new-auth
Worktree feature/new-auth is dirty. Remove anyway? [y/N] n
# Exits without removing
```

### Force Removal

With `--force`, skip dirty check and remove regardless:

```bash
$ wt remove feature/new-auth --force
# Removes worktree even if dirty
```

### Branch Deletion (if configured)

If `deleteBranchWithWorktree` is true in `.wt.config.json`:

1. Check safeguards:
   - Never delete default branch
   - Never delete branch currently checked out in main worktree
2. If safeguards pass: delete local branch
3. Report: "Deleted branch: <branch-name>"

**Examples:**

```json
{
  "deleteBranchWithWorktree": true
}
```

```bash
$ wt remove feature/new-auth
Removing worktree: /path/to/repo.wt/feature-new-auth
Deleted branch: feature/new-auth
```

## Safeguards

`wt remove` refuses to remove:

### Default Branch

```bash
$ wt remove main
Error: refusing to remove default branch/main worktree
```

### Main Worktree

The main repository worktree (not in `.wt/`) is never removable via `wt remove`.



## Examples

### Remove specific worktree

```bash
$ wt remove feature/new-auth
Removing worktree: /path/to/repo.wt/feature-new-auth
```

### Interactive removal

```bash
$ wt remove
Select a worktree to remove: feature/payment
Removing worktree: /path/to/repo.wt/feature-payment
```

### Force remove dirty worktree

```bash
$ wt remove feature/new-auth --force
Removing worktree: /path/to/repo.wt/feature-new-auth
# Removes even with uncommitted changes
```

### Remove and delete branch (configured)

```bash
$ wt remove feature/new-auth
Removing worktree: /path/to/repo.wt/feature-new-auth
Deleted branch: feature/new-auth
```

## Exit Codes

- `0`: Success
- `1`: Refusal (dirty worktree without force, default branch, cancellation)
- `1`: Git error during removal
- `2`: Invalid worktree selection (interactive mode)

## See Also

- [Configuration Reference](configuration.md#deletebranchwithworktree) - Branch deletion setting
- [wt prune](prune.md) - Remove merged worktrees in bulk
- [wt](list.md) - List all worktrees
