# Quickstart Guide

Get up and running with `wt` in under 5 minutes.

## What is wt?

`wt` is a fast CLI for managing git worktrees:

- Work with branches: Use `wt feature/name` instead of paths
- Execute in isolation: Run commands in any branch from your current terminal
- Safe operations: Won't delete worktrees with uncommitted changes
- Automatic setup: Copies config files and runs commands when creating new worktrees
- Prunes merged branches automatically

## Prerequisites

- Git (any version that supports worktrees)
- macOS or Linux
- Go 1.25+ (if building from source)

## Installation

### Homebrew (Recommended)

Homebrew installs `wt`, sets PATH, and installs completions automatically:

```bash
brew tap trungung/wt
brew install wt
```

### Go Install

Install via Go (manages version yourself):

```bash
go install github.com/trungung/wt/cmd/wt@latest
```

After Go or binary installs, ensure `wt` is on your PATH and set up completions (see below).

### Add `wt` to PATH (Go/binary installs)

```bash
# If installed with Go
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc

# If installed from binary to /usr/local/bin
echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.zshrc
source ~/.zshrc
```

### Binary Download

For manual installation, download the latest release from [GitHub Releases](https://github.com/trungung/wt/releases):

```bash
# macOS ARM64 (Apple Silicon)
curl -sSL https://github.com/trungung/wt/releases/latest/download/wt_Darwin_arm64.tar.gz | tar xz
sudo install wt /usr/local/bin/

# macOS Intel
curl -sSL https://github.com/trungung/wt/releases/latest/download/wt_Darwin_x86_64.tar.gz | tar xz
sudo install wt /usr/local/bin/

# Linux ARM64
curl -sSL https://github.com/trungung/wt/releases/latest/download/wt_Linux_arm64.tar.gz | tar xz
sudo install wt /usr/local/bin/

# Linux x86_64
curl -sSL https://github.com/trungung/wt/releases/latest/download/wt_Linux_x86_64.tar.gz | tar xz
sudo install wt /usr/local/bin/
```

## Quick Start

### 1. Initialize Configuration

In your git repository:

```bash
cd /path/to/your/repo
wt init
```

This creates `.wt.config.json` with sensible defaults. The interactive prompt asks for:

- Default branch (auto-detected from `origin/HEAD`)
- Worktree base path (default: `$REPO_PATH.wt`)
- File patterns to copy to new worktrees
- Post-create commands
- Whether to delete branches with worktrees

**Use defaults with `-y`:**

```bash
wt init --yes
```

### 2. Create Your First Worktree

```bash
# Create worktree for feature branch
wt feature/new-auth

# Output: /path/to/repo.wt/feature-new-auth
```

The first creation:

- Creates the worktree directory
- Copies configured files (`.env*`, `.vscode/**`, etc.)
- Runs post-create commands (e.g., `bun install`)
- Prints the absolute path

### 3. List Existing Worktrees

```bash
wt
```

Output format: `branch<TAB>path` (tab-separated)

```
main /path/to/repo
feature/new-auth /path/to/repo.wt/feature-new-auth
feature/payment /path/to/repo.wt/feature-payment
```

### 4. Execute Commands in Worktree

```bash
wt exec feature/new-auth -- npm test
```

This runs `npm test` with working directory set to `feature/new-auth`'s worktree.

### 5. Remove Worktrees

```bash
# Remove specific worktree
wt remove feature/new-auth

# Interactive selection (select from list)
wt remove
```

`wt` refuses to remove dirty worktrees unless forced:

```bash
wt remove feature/new-auth --force
```

### 6. Prune Merged Worktrees

Remove worktrees whose branches are merged into default branch:

```bash
# Preview what would be removed
wt prune --dry-run

# Actually remove
wt prune
```

Fetch before pruning:

```bash
wt prune --fetch
```

## Shell Completions

- **Homebrew installs**: PATH and zsh completions are configured automatically.
- **Go or binary installs**: add `wt` to your PATH and set up zsh completions manually.

Enable tab completion and ghost suggestions in Zsh:

```zsh
# Add to ~/.zshrc
source <(wt completion zsh)
```

If completions don't load, manually add to fpath:

```zsh
mkdir -p ~/.zsh/completions
wt completion zsh > ~/.zsh/completions/_wt
fpath=(~/.zsh/completions $fpath)
autoload -Uz _wt
compdef _wt wt
```

## Easy Navigation with `wt cd`

Setup the shell wrapper for seamless navigation:

```bash
# Add shell wrapper to your config
wt shell-setup zsh >> ~/.zshrc
source ~/.zshrc

# Use wt cd to navigate (creates worktree if needed)
wt cd feature/new-auth
```

The `wt cd` command:

- Creates the worktree if it doesn't exist
- Changes your shell's working directory to the worktree
- Silently navigates (like normal `cd`)

Supported shells: zsh, bash, fish

```bash
# Before: Manual navigation
$ wt feature/new-auth
/path/to/repo.wt/feature-new-auth
$ cd /path/to/repo.wt/feature-new-auth

# After: Seamless navigation
$ wt cd feature/new-auth
```

## Next Steps

- [Configuration Reference](../api-references/configuration.md) - All configuration options
- [API Reference](../api-references/index.md) - Detailed command documentation
- [**Multi-Agent Workflow**](multi-agent-workflow.md) - How to use `wt` with CLI agents and parallel terminal sessions.

## Troubleshooting

### `wt: command not found`

Add `wt` to your PATH or create a symlink:

```bash
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.zshrc
source ~/.zshrc
```

### `wt init` fails with "could not auto-detect default branch"

Set the default branch manually during init:

```bash
wt init
# Enter: main (or master, or your default branch name)
```

### Worktree collision error

Two branches map to the same directory name. Example:

- `feature/user-api` → `feature-user-api`
- `feature/user_api` → `feature-user-api` (collision)

**Error message:**
```
Error: collision: branch "feature/user_api" maps to same directory "feature-user-api" as existing worktree for branch "feature/user-api"
```

**Solution:**

Rename one of the branches to use a different directory mapping:

```bash
# Option 1: Rename the new branch before creating worktree
git branch -m feature/user_api feature/user-api-fix
wt feature/user-api-fix

# Option 2: Remove the existing worktree first
wt remove feature/user-api
wt feature/user_api
```

**Note:** The `worktreePathTemplate` config only changes the base directory (where `.wt/` is located), not the individual worktree naming. It cannot resolve branch name collisions.

See [Configuration Reference](../api-references/configuration.md) for sanitization rules.
