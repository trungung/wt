# Shell Setup

Generates shell wrapper function for easier worktree navigation.

## Overview

The `wt shell-setup` command generates a shell wrapper function that adds the `wt cd` command for seamless worktree navigation. This solves the shell limitation where a child process cannot change the parent directory.

## Usage

```bash
# Generate for your shell and append to shell config
wt shell-setup zsh >> ~/.zshrc
wt shell-setup bash >> ~/.bashrc
wt shell-setup fish >> ~/.config/fish/config.fish
```

## How It Works

The shell wrapper function intercepts the `wt cd` command and changes your shell's working directory to the worktree path:

```zsh
# Generated wrapper for zsh
wt() {
    if [ "$1" = "cd" ]; then
        local dir=$(command wt "$2")
        [ -d "$dir" ] && cd "$dir"
    else
        command wt "$@"
    fi
}
```

## Examples

### Setup and Use

```bash
# Step 1: Add shell wrapper to your config
wt shell-setup zsh >> ~/.zshrc
source ~/.zshrc

# Step 2: Use wt cd to navigate
wt cd feature/new-auth    # Creates if needed, cds, silent
pwd
# /path/to/repo.wt/feature-new-auth
```

### Daily Workflow

```bash
# Before: Manual navigation
$ wt feature/new-auth
/path/to/repo.wt/feature-new-auth
$ cd /path/to/repo.wt/feature-new-auth  # Tedious copy-paste

# After: Seamless navigation with wt cd
$ wt cd feature/new-auth
$ (you're now in the worktree directory)
$ git status
```

### Using with Other Commands

```bash
# Normal wt commands still work
wt                           # List worktrees
wt feature/new-auth          # Get path (creates if needed)
wt exec feature/bugfix -- npm test

# Navigation only happens with 'wt cd'
wt cd feature/bugfix
```

## Shell Detection

If you don't specify a shell, `wt shell-setup` attempts to auto-detect from the `SHELL` environment variable:

```bash
# Auto-detects from $SHELL
wt shell-setup    # Outputs wrapper for your current shell
```

Default shell: `zsh`

## Supported Shells

- `zsh` (recommended for macOS)
- `bash`
- `fish`

## How the Wrapper Works

The shell wrapper function:

1. **Intercepts `wt cd`**: When you run `wt cd <branch>`, the wrapper function catches it
2. **Creates worktree**: Calls `command wt <branch>` to ensure the worktree exists
3. **Changes directory**: Uses the shell's built-in `cd` to change to the worktree path
4. **Delegates other commands**: All other wt commands pass through to the binary using `command wt "$@"`

The `command` builtin ensures we call the actual `wt` binary, not the wrapper function itself.

## Troubleshooting

### Wrapper not working after setup

Make sure you've reloaded your shell configuration:

```bash
# For zsh
source ~/.zshrc

# For bash
source ~/.bashrc

# For fish
source ~/.config/fish/config.fish
```

### `wt cd` doesn't change directory

Verify the wrapper function is loaded:

```bash
# Check if wrapper function exists
type wt
# Should show a function definition, not just the binary path

# If not loaded, reload your shell config
source ~/.zshrc  # or ~/.bashrc, etc.
```

### Want to remove the wrapper

Edit your shell config file and remove the `wt()` function, then reload:

```bash
# Edit ~/.zshrc or ~/.bashrc
# Remove the wt() function block

# Reload shell
source ~/.zshrc
```

## Security

The shell wrapper function is minimal and safe:

- Only intercepts `wt cd` commands
- All other wt commands pass through unchanged
- Uses the shell's built-in `cd` for directory changes
- No arbitrary command execution
- No network access
