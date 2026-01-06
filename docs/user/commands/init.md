# wt init

Initialize `.wt.config.json` configuration file.

## Usage

```bash
wt init [--yes]
```

## Description

Creates `.wt.config.json` at repository root with interactive prompts. If config already exists, prints path and exits (no overwrite).

## Options

### `--yes`, `-y`

Write default configuration without prompts.

```bash
wt init --yes
```

**Default values with `--yes`:**

- `defaultBranch`: Auto-detected from `origin/HEAD`
- `worktreePathTemplate`: `$REPO_PATH.wt`
- `worktreeCopyPatterns`: `[]`
- `postCreateCmd`: `[]`
- `deleteBranchWithWorktree`: `false`

## Behavior

### Interactive Mode (default)

Prompts for each configuration key:

1. **Default branch** (with auto-detection)
   - Shows auto-detected value if available
   - If detection fails: requires manual input

2. **Worktree path template**
   - Default: `$REPO_PATH.wt`

3. **Worktree copy patterns**
   - Accepts comma-separated list or empty
   - Press Enter for empty array

4. **Post-create commands**
   - Accepts comma-separated list or empty
   - Press Enter for empty array

5. **Delete branch with worktree**
   - Yes/no prompt
   - Default: `false`

### Non-Interactive Mode

With `--yes`, uses all defaults without prompts. Fails if default branch cannot be auto-detected.

### Config Exists

If `.wt.config.json` already exists:

```bash
wt init
```

**Output:**

```
/path/to/repo/.wt.config.json
```

Exit code: 0 (success, no changes made)

## Examples

### Interactive with defaults

```bash
$ wt init
Initializing .wt.config.json
Default branch [main]: 
Worktree path template [$REPO_PATH.wt]: 
Worktree copy patterns []: .env*, .vscode/**
Post-create commands []: bun install
Delete branch with worktree [false]: 
Configuration generated
/path/to/repo/.wt.config.json
```

### Use defaults silently

```bash
$ wt init --yes
/path/to/repo/.wt.config.json
```

### Manual default branch

If `origin/HEAD` is missing:

```bash
$ wt init
Warning: could not determine default branch from origin/HEAD
Default branch: main
[continues with prompts...]
```

## Exit Codes

- `0`: Success
- `1`: Cancellation or write failure (no partial file written)
- `2`: Cannot auto-detect default branch (with `--yes`)

## See Also

- [Configuration Reference](../reference/configuration.md) - All configuration options
- [wt health](health.md) - Validate configuration
