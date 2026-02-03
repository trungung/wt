# wt completion

Generate shell completion scripts.

## Usage

```bash
wt completion <shell>
```

## Description

Outputs completion script to stdout for the specified shell. Save or source this script to enable tab completion.

**Note:** If you use `eval "$(wt shell-setup)"`, completions are already included. You don't need to set up completions separately.

## Arguments

### `<shell>`

Shell to generate completions for.

**Supported values:**

- `zsh`
- `bash`
- `fish`

## Installation

### Recommended: Use shell-setup (includes completions)

The easiest way to set up completions is via `wt shell-setup`, which includes both the `wt cd` wrapper and completions:

```bash
# Add to your shell config file
eval "$(wt shell-setup)"
```

This works for zsh, bash, and fish.

### Standalone Completion Setup

If you only want completions (without the `wt cd` wrapper):

#### Zsh

```zsh
# Add to ~/.zshrc
source <(wt completion zsh)
```

#### Bash

```bash
# Add to ~/.bashrc
source <(wt completion bash)
```

#### Fish

```fish
# Add to ~/.config/fish/config.fish
wt completion fish | source
```

## Completion Behavior

### Command Completion

After entering `wt` and pressing Tab:

```bash
$ wt <TAB>
completion  health     init       prune      remove     shell-setup  --help  --version
```

### Branch Completion

For `wt <branch>`, Tab shows all local branches:

```bash
$ wt <TAB>
feature/new-auth  feature/payment  main  develop
```

### Subcommand Completion

For subcommands like `remove`, Tab suggests existing worktree branches:

```bash
$ wt remove <TAB>
feature/new-auth  feature/payment
```

### Flag Completion

Flags are also completed:

```bash
$ wt --<TAB>
--from  --help  --version

$ wt remove --<TAB>
--force  --help
```

## Examples

### Generate completions for your shell

```bash
# Zsh
wt completion zsh

# Bash
wt completion bash

# Fish
wt completion fish
```

### Save to file (alternative installation)

```bash
# Zsh
mkdir -p ~/.zsh/completions
wt completion zsh > ~/.zsh/completions/_wt

# Bash
wt completion bash > ~/.wt-completion.bash
echo 'source ~/.wt-completion.bash' >> ~/.bashrc

# Fish
wt completion fish > ~/.config/fish/completions/wt.fish
```

## Exit Codes

- `0`: Success
- `1`: Unsupported shell

## See Also

- [Shell Setup](shell-setup.md) - Combined wrapper + completions setup
- [Quickstart Guide](../guides/quickstart.md) - Installation instructions
