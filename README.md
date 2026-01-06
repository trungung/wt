# wt

**Fast, branch-centric git worktree manager**

`wt` makes git worktrees branch-addressable. Think in branches, not directories.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8E.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-darwin%20%7C%20Linux-lightgrey.svg)](https://github.com/trungung/wt/releases)

## Overview

`wt` is a command-line tool that simplifies `git worktree` by making worktrees branch-addressable. Instead of managing paths manually, you operate on branch names and `wt` handles the rest:

- **Deterministic mapping**: Branch names map to predictable directories
- **Safe lifecycle**: Automatic rollback on failures, dirty worktree protection
- **Shell-friendly**: Zsh completions with ghost suggestions
- **Automated**: Prune merged branches, post-create commands

## Quick Start

### Install

Download the latest release from [GitHub Releases](https://github.com/trungung/wt/releases):

```bash
# macOS ARM64 (Apple Silicon)
curl -sSL https://github.com/trungung/wt/releases/download/v0.0.1/wt_Darwin_arm64.tar.gz | tar xz
sudo install wt /usr/local/bin/
```

Or build from source:

```bash
go install github.com/trungung/wt/cmd/wt@latest
```

### Initialize

In your git repository:

```bash
wt init
```

This creates `.wt.config.json` with sensible defaults (interactive prompts or `--yes` for defaults).

### Create Worktree

```bash
wt feature/new-auth
# Output: /path/to/repo.wt/feature-new-auth
```

First creation:
- Creates worktree directory
- Copies configured files (`.env*`, `.vscode/**`, etc.)
- Runs post-create commands (e.g., `bun install`)
- Prints absolute path

### Execute Command in Worktree

```bash
wt exec feature/new-auth -- npm test
```

### Remove Worktree

```bash
wt remove feature/new-auth
```

Refuses to remove dirty worktrees unless forced.

### Prune Merged Worktrees

```bash
wt prune --dry-run  # Preview
wt prune            # Remove
```

## Documentation

- [Getting Started](docs/user/getting-started.md) - Install and basic usage
- [Command Reference](docs/user/commands/) - Detailed command documentation
- [Configuration Reference](docs/user/reference/configuration.md) - All configuration options
- [CHANGELOG](CHANGELOG.md) - Version history

## Features

### Branch-First UX

Think in branches, not directories:

```bash
# Instead of:
git worktree add ../worktrees/feature-auth feature/auth

# Use:
wt feature/auth
```

### Deterministic Path Mapping

Branch names map to consistent, flat directory structures:

- `feature/new-auth` → `$REPO.wt/feature-new-auth`
- `bugfix/header-fix` → `$REPO.wt/bugfix-header-fix`
- `feature/user-api` → `$REPO.wt/feature-user-api`

**Rules:**
- Replace `/` with `-`
- Fail on illegal characters (alphanumeric, `-`, `_`, `.` only)
- Fail on collisions (two branches → same directory)

### Safety Features

- **Automatic rollback**: Remove worktree and branch if post-create commands fail
- **Dirty worktree protection**: Refuse removal of worktrees with uncommitted changes
- **Safeguards**: Never remove default branch or currently checked out branches
- **Concurrent safety**: File locking prevents race conditions

### Shell Integration

Enable Zsh completions:

```zsh
# Add to ~/.zshrc
source <(wt completion zsh)
```

Supports tab completion and ghost suggestions (with zsh-autosuggestions).

### Configuration

Repo-local `.wt.config.json`:

```json
{
  "defaultBranch": "main",
  "worktreePathTemplate": "$REPO_PATH.wt",
  "worktreeCopyPatterns": [".env*", ".vscode/**"],
  "postCreateCmd": ["bun install"],
  "deleteBranchWithWorktree": false
}
```

## Commands

| Command | Description |
|---------|-------------|
| `wt` | List all worktrees |
| `wt <branch>` | Ensure/create worktree for branch |
| `wt exec <branch> -- <cmd>` | Execute command in worktree |
| `wt remove [branch]` | Remove worktree |
| `wt prune` | Remove merged worktrees |
| `wt init` | Initialize configuration |
| `wt health` | Validate config and environment |
| `wt completion zsh` | Generate Zsh completion script |

## Examples

### Interactive Worktree Management

```bash
# Initialize with defaults
wt init --yes

# Create worktree (runs bun install after creation)
wt feature/payment
# Output: /path/to/repo.wt/feature-payment

# List worktrees
wt
# Output: branch<TAB>path (tab-separated)

# Run tests in worktree
wt exec feature/payment -- npm test

# Remove worktree (interactive selection)
wt remove
# Select: feature/payment
# Worktree removed
```

### Advanced Configuration

```json
{
  "worktreePathTemplate": "$REPO_PATH-worktrees",
  "worktreeCopyPatterns": [
    ".env*",
    ".env.local",
    ".vscode/**",
    "scripts/**"
  ],
  "postCreateCmd": [
    "npm install",
    "npm run dev:setup"
  ]
}
```

### Debug Mode

Enable git command tracing:

```bash
WT_DEBUG=1 wt feature/new-auth
```

## Troubleshooting

### `wt: command not found`

Add to PATH or install globally:

```bash
sudo install wt /usr/local/bin
```

### `wt init` fails: "could not auto-detect default branch"

Set manually during init or in config:

```bash
wt init
# Enter: main
```

### Worktree collision error

Two branches map to same directory (e.g., `feature/user-api` and `feature/user_api`). Solutions:
- Rename one branch
- Use custom `worktreePathTemplate`
- Accept as v1 limitation (collisions fail)

Run `wt health` to diagnose issues.

## Requirements

- Git (any version with worktree support)
- macOS or Linux
- Go 1.25+ (building from source)

## License

[MIT License](LICENSE) - see LICENSE file for details.

## Contributing

Contributions welcome! See [Documentation](docs/developer/) for guidelines.

## Version

Current: **0.0.1**

[CHANGELOG](CHANGELOG.md) for details.

## Links

- [GitHub Repository](https://github.com/trungung/wt)
- [Issues](https://github.com/trungung/wt/issues)
- [Releases](https://github.com/trungung/wt/releases)
