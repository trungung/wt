# Quickstart Guide

Get up and running with `wt` in under 5 minutes.

## What is wt?

`wt` is a fast, branch-centric CLI that simplifies `git worktree` by making worktrees branch-addressable. Instead of managing paths manually, you think in branches and `wt` handles the rest:

- Deterministic branch → directory mapping
- Safe worktree creation/removal
- Shell completion support
- Prune merged branches automatically

## Prerequisites

- Git (any version that supports worktrees)
- macOS or Linux
- Go 1.25+ (if building from source)

## Installation

### Binary Download (Recommended)

Download the latest release from [GitHub Releases](https://github.com/trungung/wt/releases):

```bash
# macOS ARM64 (Apple Silicon)
curl -sSL https://github.com/trungung/wt/releases/download/v0.0.1/wt_Darwin_arm64.tar.gz | tar xz
sudo install wt /usr/local/bin/

# macOS Intel
curl -sSL https://github.com/trungung/wt/releases/download/v0.0.1/wt_Darwin_x86_64.tar.gz | tar xz
sudo install wt /usr/local/bin/

# Linux ARM64
curl -sSL https://github.com/trungung/wt/releases/download/v0.0.1/wt_Linux_arm64.tar.gz | tar xz
sudo install wt /usr/local/bin/

# Linux x86_64
curl -sSL https://github.com/trungung/wt/releases/download/v0.0.1/wt_Linux_x86_64.tar.gz | tar xz
sudo install wt /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/trungung/wt.git
cd wt
go build ./cmd/wt
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

## Next Steps

- [Configuration Reference](../api-references/configuration.md) - All configuration options
- [API Reference](../api-references/index.md) - Detailed command documentation
- [**Multi-Agent Workflow**](multi-agent-workflow.md) - How to use `wt` with CLI agents and parallel terminal sessions.

## Troubleshooting

### `wt: command not found`

Add `wt` to your PATH or create a symlink:

```bash
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.zshrc
source ~/.zsh/completions/_wt
```

### `wt init` fails with "could not auto-detect default branch"

Set the default branch manually during init:

```bash
wt init
# Enter: main (or master, or your default branch name)
```

### Worktree collision error

Two branches sanitize to the same directory name. Example:

- `feature/user-api` → `feature-user-api`
- `feature/user_api` → `feature-user-api` (collision)

Solutions:

- Rename branches
- Use custom `worktreePathTemplate`
- Accept this as a limitation (v1 fails on collisions)

See [Configuration Reference](../api-references/configuration.md) for sanitization rules.
