# wt <branch>

Ensure a worktree exists for the specified branch, then print its path.

## Usage

```bash
wt <branch> [--from <base-branch>]
```

## Description

Creates a worktree for `<branch>` if it doesn't exist. If worktree already exists, prints its existing path and exits.

## Arguments

### `<branch>`

Branch name to ensure worktree for.

**Special case:** Default branch

If `<branch>` equals default branch, prints repo root path and exits immediately (no worktree created).

```bash
# Assuming default branch is "main"
$ wt main
/path/to/repo
```

## Options

### `--from`, `-f <base-branch>`

Base branch to create `<branch>` from (only used when creating new branch).

```bash
wt feature/new-auth --from develop
```

**Behavior:**

- If local `<branch>` exists: ignored
- If remote `origin/<branch>` exists: ignored (creates tracking branch from origin)
- If neither exists: creates branch from this base (or default branch if omitted)

## Behavior

### Worktree Already Exists

If worktree for `<branch>` already exists:

1. Print existing absolute path to stdout
2. Exit with code 0
3. Do not run copy hooks or postCreate commands

```bash
$ wt feature/new-auth
/path/to/repo.wt/feature-new-auth
# No copy, no post-create (already exists)
```

### Worktree Doesn't Exist

1. **Resolve/ensure branch:**
   - If local `<branch>` exists: use it
   - Else if remote `origin/<branch>` exists: create local tracking branch
   - Else create new local branch from:
     - `--from <base-branch>` if provided
     - Default branch if omitted

2. **Compute target path:**
   - Base: `worktreePathTemplate` from config (default: `$REPO_PATH.wt`)
   - Leaf: Sanitized branch name (`/` → `-`, strict character validation)
   - Full path: `<base>/<sanitized-leaf>`

3. **Check for collisions:**
   - Fail if sanitized name collides with another branch's worktree
   - Fail if directory already exists for different branch

4. **Create worktree:**
   - Execute `git worktree add <path> <branch>`

5. **Apply copy patterns:**
   - Copy files matching `worktreeCopyPatterns` from repo root to worktree root
   - Copy only if missing at destination (no overwrites)
   - Applied only when creating new worktree

6. **Run post-create commands:**
   - Execute commands from `postCreateCmd` array
   - Run in worktree directory (not repo root)
   - Run sequentially (in order)
   - If any fails: perform rollback

7. **Print path:**
   - Print absolute path to stdout

## Rollback on Failure

If `postCreateCmd` fails or worktree creation fails, `wt` performs automatic rollback:

**Cleanup steps:**

1. Remove newly created worktree (via `git worktree remove`)
2. Delete local branch (if `wt` created it during this invocation)
3. Report rollback status to stderr

```bash
$ wt feature/new-auth
Error: post-create command failed: npm install
Rollback status: Removed worktree, deleted branch feature/new-auth
```

**Limitation:** Rollback does not undo side effects from post-create commands (e.g., global caches, network requests).

## Examples

### Ensure worktree (already exists)

```bash
$ wt feature/new-auth
/path/to/repo.wt/feature-new-auth
```

### Create worktree from new branch

```bash
$ wt feature/payment
# Creates branch from default branch, creates worktree, copies files, runs post-create
/path/to/repo.wt/feature-payment
```

### Create from specific base branch

```bash
$ wt feature/billing --from develop
/path/to/repo.wt/feature-billing
```

### Default branch special case

```bash
$ wt main
/path/to/repo
# No worktree created
```

### Create with post-create (config example)

Config (`.wt.config.json`):

```json
{
  "postCreateCmd": ["bun install"]
}
```

Command:

```bash
$ wt feature/new-auth
# After worktree creation:
# 1. Copy configured files
# 2. Run: bun install
/path/to/repo.wt/feature-new-auth
```

## Exit Codes

- `0`: Success
- `1`: Failure (including collisions, dirty worktree checks, post-create failures after rollback)
- `2`: Invalid branch name (illegal characters)
- `3`: Collision detected (two branches sanitize to same directory)

## See Also

- [wt exec](exec.md) - Execute command in worktree
- [Configuration Reference](configuration.md#worktreecopypatterns) - Copy patterns
- [Configuration Reference](configuration.md#postcreatecmd) - Post-create commands
