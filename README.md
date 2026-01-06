# wt

**An opinionated git worktree manager**

`wt` manages git worktrees using branch names instead of paths. Create worktrees,
run commands in isolated environments, and clean up safely.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8E.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-darwin%20%7C%20Linux-lightgrey.svg)](https://github.com/trungung/wt/releases)

---

## What is wt?

`wt` is a CLI for git worktree management. Use branch names to create worktrees,
execute commands in them, and remove them when done.

Built for multi-branch workflows and AI-assisted development.

## Why use `wt`?

- **Multi-branch workflows:** Run commands in multiple worktrees from your main terminal
- **Safe cleanup:** Won't delete worktrees with uncommitted changes
- **Automatic setup:** New branches copy config files and run setup commands
- **Multi-agent support:** AI agents work in isolated environments while you stay in your main repo

See [Multi-Agent Workflows](docs/user/guides/multi-agent-workflow.md) for details.

## Quick Start

### 1. Install

**Recommended** - Install via Go:

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

For comprehensive setup and examples, see the [Quickstart Guide](docs/user/guides/quickstart.md).

---

## Documentation

- **[Quickstart Guide](docs/user/guides/quickstart.md)** – Installation, setup, and examples
- **[Multi-Agent Workflows](docs/user/guides/multi-agent-workflow.md)** – Parallel development and AI agent workflows
- **[API Reference](docs/user/api-references/index.md)** – Detailed command documentation
- **[Configuration Reference](docs/user/api-references/configuration.md)** – All configuration options explained

## Contributing

Contributions are welcome! Please see the [developer documentation](docs/developer/) for guidelines.

## License

`wt` is released under the [MIT License](LICENSE).
