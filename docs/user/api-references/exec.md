# wt exec

Execute an arbitrary command inside a branch's worktree.

## Usage

```bash
wt exec <branch> -- <command...>
```

## Description

Runs `<command...>` with working directory set to the worktree for `<branch>`. Inherits stdin/stdout/stderr from parent process. Exit code matches executed command's exit code.

## Arguments

### `<branch>`

Branch name whose worktree to use as working directory.

**Requirement:** Worktree must exist for this branch. Fails if not found.

### `<command...>`

Command and arguments to execute.

## Options

### `--` (required delimiter)

Marks end of `wt` options and start of command to execute.

**Required:** Yes

## Behavior

1. **Resolve target directory:**
   - If `<branch>` equals default branch: use repository root (`git rev-parse --show-toplevel`)
   - Else: look up worktree path for `<branch>` from `git worktree list --porcelain`

2. **Validate worktree exists:**
   - Fail if worktree not found for `<branch>`
   - Show error message with available branches

3. **Execute command:**
   - Set working directory to resolved path
   - Execute first argument as command, rest as arguments
   - Inherit stdin/stdout/stderr (interactive commands work)

4. **Propagate exit code:**
   - Exit with same code as executed command

## Examples

### Run tests in worktree

```bash
$ wt exec feature/new-auth -- npm test
# Running: npm test
# In directory: /path/to/repo.wt/feature-new-auth
```

### Build in worktree

```bash
$ wt exec feature/payment -- make build
# In directory: /path/to/repo.wt/feature-payment
```

### Multiple arguments

```bash
$ wt exec feature/billing -- ./scripts/deploy.sh --env staging
# Running: ./scripts/deploy.sh --env staging
```

### Interactive command

```bash
$ wt exec feature/new-auth -- node
# Interactive REPL in worktree directory
```

### Use with shell pipes

```bash
wt exec feature/new-auth -- cat package.json | jq .scripts
```

### Default branch special case

```bash
$ wt exec main -- npm run build
# Runs in repository root (main worktree)
```

### Error: worktree doesn't exist

```bash
$ wt exec nonexistent-branch -- ls
Error: no worktree exists for branch "nonexistent-branch"
```

## Error Handling

### Missing `--` delimiter

```bash
$ wt exec feature/new-auth ls
Error: missing command after --
```

### Worktree not found

```bash
$ wt exec <branch-name> -- <command>
Error: no worktree exists for branch "<branch-name>"
```

Exit code: 1

### Command execution fails

If executed command fails, `wt exec` exits with the same code:

```bash
$ wt exec feature/new-auth -- npm run nonexistent-script
Error: Command "nonexistent-script" not found
# Exit code: 1 (from npm)
```

## Exit Codes

- Child command's exit code (success/failure)
- `1`: Missing `--` delimiter
- `1`: Worktree not found for branch
- `1`: Usage error (no command provided)

## See Also

- [wt <branch>](ensure.md) - Ensure worktree for a branch
- [Configuration Reference](configuration.md) - Default branch configuration
