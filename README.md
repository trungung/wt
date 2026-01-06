# wt

**Fast, branch-centric git worktree manager**

`wt` makes git worktrees branch-addressable, focusing on safety and configuration. Think in branches, not directories.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8E.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-darwin%20%7C%20Linux-lightgrey.svg)](https://github.com/trungung/wt/releases)

---

## Why use `wt`?

`wt` automates the complex, error-prone parts of `git worktree` so you can focus on development, not file paths.

### Key Benefits

- **Branch-First UX:** Create, list, and remove worktrees using only branch names (`wt feature/branch-name`).
- **Safety & Rollback:** Automatic rollback if post-create commands (like `npm install`) fail. Refuses to remove worktrees with dirty state.
- **Automated Setup:** Automatically copies config files (e.g., `.env`, `.vscode/`) and runs setup commands upon creation.
- **Maintenance:** Easily prune merged worktrees with a single command (`wt prune`).

## Quick Start

### 1. Install

Install via Go:

```bash
go install github.com/trungung/wt/cmd/wt@latest
```

For binary downloads (macOS/Linux), see [GitHub Releases](https://github.com/trungung/wt/releases).

### 2. Initialize

Run this in your main repository to create the local configuration file `.wt.config.json`:

```bash
wt init
```

### 3. Use Worktrees

Create a worktree for a new branch, execute commands inside it, and remove it when done:

| Command                               | Action                                         |
| :------------------------------------ | :--------------------------------------------- |
| `wt feature/payment`                  | Create worktree at `./repo.wt/feature-payment` |
| `wt exec feature/payment -- npm test` | Run command in the worktree directory          |
| `wt prune`                            | Automatically remove merged worktrees          |

---

## Documentation

For comprehensive guides, command references, and detailed configuration, please refer to the dedicated documentation:

- **[Quickstart Guide](docs/user/guides/quickstart.md)** – Full installation, setup, and core workflow.
- **[API Reference](docs/user/api-references/index.md)** – Detailed usage for `wt exec`, `wt remove`, `wt prune`, and more.
- **[Configuration Reference](docs/user/api-references/configuration.md)** – All options for `.wt.config.json`, including post-create commands and path templates.

## Contributing

Contributions are welcome! Please see the [developer documentation](docs/developer/) for guidelines.

## License

`wt` is released under the [MIT License](LICENSE).
