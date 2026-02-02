# API Reference

This section provides a detailed, flag-by-flag reference for every `wt` command and configuration option.

## Commands

| Command         | Description                                                                                   | Reference                   |
| :-------------- | :-------------------------------------------------------------------------------------------- | :-------------------------- |
| `wt`            | List all existing worktrees.                                                                  | [List](list.md)             |
| `wt <branch>`   | Ensure a worktree exists for a branch (creates if needed). Supports `--from <base>` flag. | [Ensure](ensure.md)         |
| `wt init`       | Initializes the `.wt.config.json` file in the repository root.                                | [Init](init.md)             |
| `wt exec`       | Executes a command inside the specified worktree's directory.                                 | [Exec](exec.md)             |
| `wt remove`     | Removes a worktree and optionally its associated branch.                                      | [Remove](remove.md)         |
| `wt prune`      | Removes worktrees whose branches have been merged into the default branch.                    | [Prune](prune.md)           |
| `wt health`     | Validates the configuration and environment, diagnosing potential issues.                     | [Health](health.md)         |
| `wt completion` | Generates shell completion scripts for Zsh.                                                   | [Completion](completion.md) |
| `wt shell-setup`| Generates shell wrapper function for easier navigation (supports zsh, bash, fish).        | [Shell Setup](shell-setup.md)   |

---

## Configuration

- [**Configuration Reference**](configuration.md) - Detailed explanation of every field in `.wt.config.json`, including defaults, examples, and branch sanitization rules.
