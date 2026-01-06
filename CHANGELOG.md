# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2025-01-06

### Added

- Initial release of wt, a branch-centric git worktree manager
- `wt init`: Interactive configuration wizard with `--yes` for defaults
- `wt <branch>`: Ensure/create worktree for specified branch, print path
- `wt exec <branch> -- <cmd>`: Execute commands inside worktree context
- `wt remove [branch]`: Remove worktrees with dirty checks and interactive selection
- `wt prune`: Remove worktrees whose branches are merged into default branch
- `wt health`: Comprehensive validation of config and environment
- `wt`: List all worktrees (tab-separated format)
- `wt completion zsh`: Generate Zsh completion script
- Shell completions: Tab completion and ghost suggestions for Zsh
- Branch sanitization: Strict validation, collision detection
- File copying: Configurable `worktreeCopyPatterns` (copy only if missing)
- Post-create commands: Execute commands after worktree creation with automatic rollback
- Configuration: Repo-local `.wt.config.json` with 5 configurable keys
- Debug mode: `WT_DEBUG` environment variable for git command tracing
- Concurrent safety: File locking mechanism (5-second timeout)
- Rollback errors: Automatic cleanup with detailed status reporting
- Default branch special case: Return repo root path for default branch
- Git fetch in prune: `--fetch` option for `git fetch --prune`
- Force removal: `--force` in `remove` and `prune` to skip dirty checks
- Dry-run mode: `--dry-run` in `prune` to preview candidates

### Changed

- Branch sanitization is now strict (fails on illegal characters instead of replacing)

### Security

- File locking prevents race conditions during concurrent operations
- Dirty worktree checks prevent accidental data loss
- Safeguards prevent removal of default branch and currently checked out branches

### Documentation

- Comprehensive user documentation in Markdown format
- Command reference for all 8 commands
- Configuration reference with examples and validation rules
- Getting started guide with installation and quick start
- AI-friendly `llms.txt` for LLM discovery
- Internal development documentation preserved in `docs/developer/internal/`
- MIT License added

### Testing

- 11 integration tests covering mapping, idempotency, exec, remove, prune, collisions, and rollback
- Concurrency tests for file locking mechanism
- Health check validation tests

[Unreleased]: https://github.com/trungung/wt/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/trungung/wt/releases/tag/v0.0.1
