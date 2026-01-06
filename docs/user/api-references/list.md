# wt

List all worktrees in the current repository.

## Usage

```bash
wt
```

## Description

Lists all worktrees registered with git, showing branch names and their corresponding paths.

## Output Format

Tab-separated: `branch<TAB>path` (one per line)

```
main /path/to/repo
feature/new-auth /path/to/repo.wt/feature-new-auth
feature/payment /path/to/repo.wt/feature-payment
feature/billing /path/to/repo.wt/feature-billing
(detached) /path/to/repo.wt/detached-head
```

**Notes:**

- First worktree is always the main repository (not in `.wt/`)
- Detached worktrees show branch as `(detached)`
- Main worktree always shows as your default branch name
- Output is parseable (use `cut -f1` for branch names, `cut -f2` for paths)

## Examples

### List all worktrees

```bash
$ wt
main /Users/dev/myproject
feature/new-auth /Users/dev/myproject.wt/feature-new-auth
feature/payment /Users/dev/myproject.wt/feature-payment
```

### Extract branch names

```bash
$ wt | cut -f1
main
feature/new-auth
feature/payment
```

### Extract paths

```bash
$ wt | cut -f2
/Users/dev/myproject
/Users/dev/myproject.wt/feature-new-auth
/Users/dev/myproject.wt/feature-payment
```

### Count worktrees (excluding main)

```bash
$ wt | tail -n +2 | wc -l
2
```

## Exit Codes

- `0`: Success
- `1`: Not in a git repository or git command failed

## See Also

- [wt <branch>](ensure.md) - Ensure worktree for a branch
- [wt remove](remove.md) - Remove a worktree
- [wt prune](prune.md) - Remove merged worktrees
