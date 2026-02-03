# Shell Setup

Generates shell wrapper function and completions for easier worktree navigation.

## Overview

The `wt shell-setup` command generates a shell configuration that includes:

1. **Shell wrapper function** - Adds the `wt cd` command for seamless directory navigation
2. **Tab completions** - Enables completion for commands, branches, and flags

This is the recommended way to set up `wt` for daily use.

## Usage

```bash
# Add to your shell config file
eval "$(wt shell-setup)"          # Auto-detects shell from $SHELL
eval "$(wt shell-setup zsh)"      # Explicit shell
eval "$(wt shell-setup bash)"
wt shell-setup fish | source      # Fish syntax
```

## Supported Shells

- `zsh` (default, recommended for macOS)
- `bash`
- `fish`

## What It Sets Up

### 1. The `wt cd` Command

The shell wrapper intercepts `wt cd <branch>` and changes your shell's working directory:

```bash
# Creates worktree if needed, then cd's into it
wt cd feature/new-auth
```

This solves the shell limitation where a child process cannot change the parent's directory.

### 2. Tab Completions

Enables completion for:

- Subcommands (`wt rem<TAB>` -> `wt remove`)
- Branch names (`wt feature/<TAB>`)
- Flags (`wt --<TAB>`, `wt remove --<TAB>`)

## Installation

Add one line to your shell configuration file:

### Zsh (~/.zshrc)

```zsh
eval "$(wt shell-setup)"
```

### Bash (~/.bashrc)

```bash
eval "$(wt shell-setup)"
```

### Fish (~/.config/fish/config.fish)

```fish
wt shell-setup fish | source
```

After adding, reload your shell:

```bash
source ~/.zshrc   # or ~/.bashrc
```

## Examples

### Daily Workflow

```bash
# Navigate to a worktree (creates if needed)
wt cd feature/new-auth
pwd
# /path/to/repo.wt/feature-new-auth

# All other wt commands work normally
wt                          # List worktrees
wt feature/payment          # Get/create worktree path
wt remove feature/old       # Remove worktree
```

### Tab Completion in Action

```bash
$ wt <TAB>
completion  health  init  prune  remove  shell-setup

$ wt feature/<TAB>
feature/new-auth  feature/payment  feature/billing

$ wt remove <TAB>
feature/new-auth  feature/payment   # Only shows existing worktrees

$ wt --<TAB>
--from  --help  --version
```

## How It Works

The generated shell function:

1. **Intercepts `wt cd`**: When you run `wt cd <branch>`, the wrapper catches it
2. **Creates worktree**: Calls `command wt <branch>` to ensure the worktree exists
3. **Changes directory**: Uses the shell's built-in `cd` to change to the worktree path
4. **Delegates other commands**: All other wt commands pass through to the binary
5. **Loads completions**: Sources the completion script for your shell

The `command` builtin ensures we call the actual `wt` binary, not the wrapper function itself.

## Shell Detection

If you don't specify a shell, `wt shell-setup` auto-detects from `$SHELL`:

```bash
wt shell-setup    # Outputs config for your current shell
```

Default: `zsh`

## Troubleshooting

### Wrapper not working after setup

Reload your shell configuration:

```bash
source ~/.zshrc   # or ~/.bashrc
```

### `wt cd` doesn't change directory

Verify the wrapper function is loaded:

```bash
type wt
# Should show a function definition, not just the binary path
```

### Completions not working

Ensure you're using `eval "$(wt shell-setup)"` not just `wt shell-setup >> ~/.zshrc`. The eval is important for proper shell integration.

### Want to remove the wrapper

Edit your shell config file, remove the `eval "$(wt shell-setup)"` line, and reload:

```bash
source ~/.zshrc
```

## Security

The shell wrapper is minimal and safe:

- Only intercepts `wt cd` commands
- All other wt commands pass through unchanged
- Uses the shell's built-in `cd` for directory changes
- No arbitrary command execution
- No network access

## See Also

- [Completion](completion.md) - Standalone completion setup (if you don't want the wrapper)
- [Quickstart Guide](../guides/quickstart.md) - Full installation guide
