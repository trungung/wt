# wt completion

Generate shell completion scripts.

## Usage

```bash
wt completion <shell>
```

## Description

Outputs completion script to stdout for the specified shell. Save or source this script to enable tab completion.

## Arguments

### `<shell>`

Shell to generate completions for.

**Supported values:**
- `zsh` (only shell supported in v1)

## Supported Shells

### Zsh

Generate Zsh completion script:

```bash
wt completion zsh
```

**Installation methods:**

#### Method 1: Source directly (recommended)

Add to `~/.zshrc`:

```zsh
source <(wt completion zsh)
```

**Enable ghost suggestions (zsh-autosuggestions):**

```zsh
# In ~/.zshrc
ZSH_AUTOSUGGEST_STRATEGY=(completion history)
source <(wt completion zsh)
```

#### Method 2: Manually add to fpath

If completions aren't loading with Method 1:

```zsh
mkdir -p ~/.zsh/completions
wt completion zsh > ~/.zsh/completions/_wt
fpath=(~/.zsh/completions $fpath)
autoload -Uz _wt
compdef _wt wt
```

Then add to `~/.zshrc`:

```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz _wt
compdef _wt wt
```

## Completion Behavior

### Command Completion

After entering `wt` and pressing Tab:

```bash
$ wt <TAB>
completion  exec      health     init       prune      remove     --version
```

### Branch Completion

For commands that accept branch names (e.g., `wt <branch>`, `wt exec <branch>`), Tab shows existing worktree branches:

```bash
$ wt <TAB>
feature/new-auth  feature/payment  main
```

### Subcommand Completion

For subcommands (e.g., `remove`), Tab suggests worktree branches:

```bash
$ wt remove <TAB>
feature/new-auth  feature/payment
```

## Examples

### Generate Zsh completions

```bash
$ wt completion zsh
#compdef wt
# Output: complete completion script
```

### Save to file

```bash
wt completion zsh > ~/.zsh/completions/_wt
```

### Test completions

After installation:

```bash
$ wt comp<TAB>
# Completes to: wt completion

$ wt exec f<TAB>
# Completes to: wt exec feature/new-auth

$ wt exec feature/<TAB>
# Shows: feature/new-auth, feature/payment
```

## Exit Codes

- `0`: Success
- `1`: Unsupported shell

## Future Shells

Additional shells (Bash, Fish) may be added in future releases. Vote or contribute to track progress.

## See Also

- [Getting Started](../getting-started.md#shell-completions) - Installation instructions
- [Command Reference](list.md) - List all commands
