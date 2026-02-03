# Multi-Agent and Parallel Development Workflows

One of `wt`'s core strengths is providing a safe, predictable environment for parallel, non-interactive processes. This makes it ideal for managing complex, multi-task development sessions, especially when utilizing CLI agents or multiple terminal sessions (like tmux or Ghostty splits).

## The Core Value: Isolation and Context

When managing multiple tasks or features, each worktree provides a completely isolated environment with its own working directory, staged changes, and checked-out branch.

### Scenario: Running Two Tasks in Parallel

Imagine you are using a tool like **opencode** or **claudecode** in your main terminal to manage development, while simultaneously needing to run isolated setup, tests, or servers for different branches (`feature/a-fix` and `feature/b-refactor`).

#### 1. Setup in Isolation

From your main terminal (where your agent is running), create two worktrees.

```bash
# wt ensures the branch exists, copies config files, and runs the postCreateCmd (e.g., npm install)
wt feature/a-fix
wt feature/b-refactor
```

#### 2. Parallel Development with Split Terminals

Use split terminal panels (tmux, Ghostty, iTerm) to work in different worktrees simultaneously.

| Terminal 1 | Terminal 2 |
| :--- | :--- |
| **Goal:** Run tests for the fix. | **Goal:** Run a local development server. |
| **Command:** `wt cd feature/a-fix && npm test` | **Command:** `wt cd feature/b-refactor && npm run dev` |
| **Context:** Changes to `/repo.wt/feature-a-fix` and runs tests there. | **Context:** Changes to `/repo.wt/feature-b-refactor` and starts server. |

Each terminal operates in a completely isolated worktree with its own git state.

#### 3. Alternative: Run Commands Without Changing Directory

If you need to run a command in a worktree without changing your current directory, use subshell syntax:

```bash
# Run tests in feature/a-fix worktree from any location
(cd "$(wt feature/a-fix)" && npm test)

# Run server in background
(cd "$(wt feature/b-refactor)" && npm run dev) &
```

#### 4. Cleanup and Synchronization

When development is complete, cleanup is simple:

```bash
# This removes all worktrees whose branches have been merged into the default branch
wt prune
```

## Configuration for Automation

For robust multi-task environments, ensure your configuration is optimized for automation via `.wt.config.json` (created by `wt init`):

| Configuration Key | Value for Agent/CI Workflow | Benefit |
| :--- | :--- | :--- |
| `worktreeCopyPatterns` | `[".env.agent", ".vscode/**"]` | Ensures configuration files necessary for tools/CI are present immediately. |
| `postCreateCmd` | `["npm install", "go mod download"]` | Guarantees dependencies are installed atomically. If this fails, `wt` performs an automatic rollback, leaving your system clean. |
| `deleteBranchWithWorktree` | `true` | Simplifies cleanup by ensuring `wt remove` and `wt prune` also delete the local branch after a task is complete. |

See the [Configuration Reference](../api-references/configuration.md) for full details.
