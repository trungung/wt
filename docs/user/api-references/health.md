# wt health

Validate configuration and environment to diagnose potential issues.

## Usage

```bash
wt health
```

## Description

Performs comprehensive checks on repository configuration, git setup, and worktree state. Reports issues as OK, WARN, or ERROR.

## Checks Performed

### 1. Repository Root

**Check:** Can determine git top-level directory via `git rev-parse --show-toplevel`.

**Level:** ERROR

**Error:** "not a git repository: <error-details>"

### 2. Configuration File

**Check:** `.wt.config.json` exists and is valid JSON.

**Level:** ERROR

**Error:** "failed to read config: <error-details>" or "invalid JSON: <error-details>"

**Level:** WARN

**Warning:** "Unknown config keys: <key1>, <key2>"

### 3. Default Branch

**Check:** Default branch can be determined.

**Level:** ERROR

**Error:** "could not determine default branch via origin/HEAD. Please set 'defaultBranch' in config."

**Detection rules:**

1. Check `defaultBranch` in `.wt.config.json`
2. Else: read `refs/remotes/origin/HEAD` symbolic ref
3. Else: ERROR

### 4. Worktree Base Directory

**Check:** Worktree base directory (from `worktreePathTemplate`) is writable or can be created.

**Level:** ERROR

**Error:** "Cannot create/write to worktree base directory: <path>"

### 5. Copy Patterns

**Check:** `worktreeCopyPatterns` matches at least one existing file in repository.

**Level:** WARN

**Warning:** "Copy patterns match no files in repository"

**Behavior:** Still valid configuration (empty patterns allowed), just warns user.

### 6. Post-Create Commands

**Check:** `postCreateCmd` does not contain empty command strings.

**Level:** WARN

**Warning:** "Empty command string in postCreateCmd: <index>"

### 7. Branch Name Collisions

**Check:** No two branches sanitize to the same directory name.

**Level:** WARN

**Warning:** "branches <branch1> and <branch2> will both map to directory <directory>"

**Collision examples:**

- `feature/user-api` and `feature/user_api` → both `feature-user-api`
- `feature/new-auth` and `feature/new_auth` → both `feature-new-auth`

## Output Format

Human-readable list of checks, one per line:

```
[OK] Repository root: /path/to/repo
[OK] Default branch: main
[OK] Worktree base directory: /path/to/repo.wt (writable)
[WARN] Copy patterns match no files in repository
[OK] Post-create commands: valid
[OK] No branch collisions detected
```

## Exit Codes

- `0`: No errors (WARNs are acceptable)
- `1`: At least one ERROR detected

## Examples

### All checks passing

```bash
$ wt health
[OK] Repository root: /Users/dev/myproject
[OK] Config: valid JSON
[OK] Default branch: main
[OK] Worktree base directory: /Users/dev/myproject.wt (writable)
[OK] Copy patterns: match 2 files
[OK] Post-create commands: valid
[OK] No branch collisions detected
```

### Configuration errors

```bash
$ wt health
[OK] Repository root: /Users/dev/myproject
[ERROR] failed to read config: <error-details>
[ERROR] could not determine default branch via origin/HEAD. Please set 'defaultBranch' in config.
[OK] Worktree base directory: /Users/dev/myproject.wt (writable)
```

### Warnings only

```bash
$ wt health
[OK] Repository root: /Users/dev/myproject
[WARN] Unknown config keys: foo, bar
[OK] Default branch: main
[OK] Worktree base directory: /Users/dev/myproject.wt (writable)
[OK] Post-create commands: valid
[OK] No branch collisions detected
```

### Branch collision

```bash
$ wt health
[OK] Repository root: /Users/dev/myproject
[OK] Config: valid JSON
[OK] Default branch: main
[OK] Worktree base directory: /Users/dev/myproject.wt (writable)
[WARN] branches "feature/user-api" and "feature/user_api" will both map to directory "feature-user-api"
[OK] No other collisions detected
```

## Troubleshooting

### "Cannot determine default branch"

**Cause:** `origin/HEAD` is missing and no config override.

**Solutions:**

1. Set up default branch on remote:

   ```bash
   git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/main
   ```

2. Add to `.wt.config.json`:

   ```json
   {
     "defaultBranch": "main"
   }
   ```

### "Invalid JSON in .wt.config.json"

**Cause:** Syntax error in config file.

**Solution:** Fix JSON syntax (missing commas, trailing commas, quotes, etc.). Use a JSON linter or online validator.

### "Branch collision detected"

**Cause:** Two different branch names sanitize to the same directory name.

**Solutions:**

1. Rename one of the branches
2. Accept this as a v1 limitation (collisions cause errors)

### "Cannot create/write to worktree base directory"

**Cause:** Permission issue or disk full.

**Solution:**

- Check directory permissions: `ls -ld <path>`
- Choose different `worktreePathTemplate` location

## See Also

- [Configuration Reference](configuration.md) - All configuration options
- [wt init](init.md) - Initialize configuration file
