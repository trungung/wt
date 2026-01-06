# Configuration Reference

`wt` stores configuration in `.wt.config.json` at the root of your git repository.

## File Location

```
/path/to/repo/.wt.config.json
```

Run `wt init` to create this file interactively, or create it manually.

## Configuration Options

### `defaultBranch` (string, optional)

Default branch name. Overrides auto-detection from `origin/HEAD`.

**Default:** Auto-detected from `refs/remotes/origin/HEAD`

**Example:**

```json
{
  "defaultBranch": "main"
}
```

**Detection rules:**

1. If `defaultBranch` is set in config: use it
2. Else: read from `origin/HEAD` symbolic ref
3. If both missing: `wt init` requires manual input, `wt health` errors

### `worktreePathTemplate` (string, optional)

Template for worktree base directory. Supports `$REPO_PATH` variable expansion.

**Default:** `$REPO_PATH.wt`

**Examples:**

```json
{
  "worktreePathTemplate": "$REPO_PATH.wt"
}
```

```json
{
  "worktreePathTemplate": "$REPO_PATH-worktrees"
}
```

```json
{
  "worktreePathTemplate": "/tmp/wt-worktrees"
}
```

**Variable expansion:**

- `$REPO_PATH` expands to absolute path of git root
- No other variables supported in v1

**Result:**

- If repo is `/Users/dev/myproject`
- Template `$REPO_PATH.wt` → `/Users/dev/myproject.wt`

### `worktreeCopyPatterns` (array of strings, optional)

Glob patterns for files to copy to new worktrees. Files are copied **only if missing** at destination (no overwrites).

**Default:** `[]` (empty array)

**Examples:**

```json
{
  "worktreeCopyPatterns": [".env*", ".vscode/**"]
}
```

```json
{
  "worktreeCopyPatterns": [
    ".env*",
    ".env.local",
    ".vscode/**",
    "scripts/**",
    "*.config.js"
  ]
}
```

**Behavior:**

- Applied only when creating a new worktree
- Uses Go's `filepath.Match` for glob matching
- Copies from repo root to worktree root
- Skips existing files in destination
- Errors if source file doesn't exist (except globs that match nothing)

**Common patterns:**

- `.env*` - All .env files
- `.vscode/**` - VSCode settings (recursive)
- `scripts/**` - Scripts directory
- `*.config.js` - All .config.js files

### `postCreateCmd` (array of strings, optional)

Commands to execute after creating a new worktree. Runs in the new worktree directory.

**Default:** `[]` (empty array)

**Examples:**

```json
{
  "postCreateCmd": ["bun install"]
}
```

```json
{
  "postCreateCmd": ["npm install", "npm run dev:setup"]
}
```

```json
{
  "postCreateCmd": ["make dependencies"]
}
```

**Behavior:**

- Runs only when creating a new worktree
- Executed in worktree directory (not repo root)
- Commands run sequentially (in order)
- If any command fails: rollback is attempted (worktree removed, branch deleted if created)
- Stdout/stderr shown in terminal
- Exit code of last command determines overall success

**Error handling:**

- If `postCreateCmd` fails, `wt` performs automatic rollback:
  - Removes newly created worktree
  - Deletes local branch (if `wt` created it during this invocation)
  - Reports rollback status to stderr

### `deleteBranchWithWorktree` (boolean, optional)

Whether to delete the local branch when removing its worktree.

**Default:** `false`

**Examples:**

```json
{
  "deleteBranchWithWorktree": true
}
```

**Safeguards (always applied, regardless of setting):**

- Never deletes default branch
- Never deletes branch currently checked out in main worktree

**Commands affected:**

- `wt remove <branch>` (with or without `--force`)
- `wt prune` (when removing merged worktrees)

## Complete Example

```json
{
  "defaultBranch": "main",
  "worktreePathTemplate": "$REPO_PATH.wt",
  "worktreeCopyPatterns": [
    ".env*",
    ".vscode/**"
  ],
  "postCreateCmd": [
    "bun install"
  ],
  "deleteBranchWithWorktree": false
}
```

## Branch Sanitization Rules (Strict)

Branch names are sanitized when mapping to directory names:

**Rules:**

1. Replace `/` with `-`
2. Fail if branch name contains illegal characters (anything except: `a-z`, `A-Z`, `0-9`, `-`, `_`, `.`)
3. Fail if two different branches sanitize to same directory name

**Examples:**

| Branch Name | Sanitized Directory | Valid? |
|-------------|-------------------|---------|
| `feature/new-auth` | `feature-new-auth` | ✅ Yes |
| `feature/user-api` | `feature-user-api` | ✅ Yes |
| `feature@2024` | - | ❌ No (illegal `@`) |
| `bug fix` | - | ❌ No (illegal space) |

**Collision examples:**

| Branch A | Branch B | Both Sanitize To | Collision? |
|-----------|-----------|------------------|------------|
| `feature/user-api` | `feature/user_api` | `feature-user-api` | ✅ Collision (ERROR) |
| `feature/new-auth` | `feature/new_auth` | `feature-new-auth` | ✅ Collision (ERROR) |

## Environment Variables

### `WT_DEBUG`

Enable debug logging for git commands.

**Usage:**

```bash
WT_DEBUG=1 wt feature/new-auth
```

**Output:**

```
[DEBUG] git worktree add ... (took 12ms)
[DEBUG] git branch ... (took 5ms)
```

Logs to stderr, does not affect command output.

## Configuration Validation

Run `wt health` to validate configuration:

```bash
wt health
```

Checks:

- Config is valid JSON (ERROR if not)
- Config contains only known keys (WARN if unknown keys)
- Default branch can be determined (ERROR if not)
- Worktree base path is writable/creatable (ERROR if not)
- Copy patterns match existing files (WARN if nothing matches)
- No branch name collisions (ERROR if collision detected)

See [Getting Started](../getting-started.md#troubleshooting) for common issues.
