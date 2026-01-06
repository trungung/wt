# Multi-Agent and Parallel Development Workflows

One of `wt`'s core strengths is providing a safe, predictable environment for parallel, non-interactive processes. This makes it ideal for managing complex, multi-task development sessions, especially when utilizing CLI agents or multiple terminal sessions (like tmux or Ghostty splits).

## The Core Value: Isolation and Context

When managing multiple tasks or features, the `wt exec` command allows you to execute commands in a fresh, isolated worktree *without leaving your main repository directory*.

### Scenario: Running Two Tasks in Parallel

Imagine you are using a tool like **opencode** or **claudecode** in your main terminal to manage development, while simultaneously needing to run isolated setup, tests, or servers for different branches (`feature/a-fix` and `feature/b-refactor`).

#### 1. Setup in Isolation

From your main terminal (where your agent is running), create two worktrees.

```bash
# wt ensures the branch exists, copies config files, and runs the postCreateCmd (e.g., npm install)
wt feature/a-fix
wt feature/b-refactor
```

#### 2. Parallel Development via `wt exec`

You can now use split terminal panels (tmux, Ghostty, iTerm) or different agent instances, all rooted in your main worktree, to operate on the isolated environments.

| Task (Agent A) | Task (Agent B) |
| :--- | :--- |
| **Goal:** Run tests for the fix. | **Goal:** Run a local development server. |
| **Command:** `wt exec feature/a-fix -- npm test` | **Command:** `wt exec feature/b-refactor -- npm run dev` |
| **Context:** Runs tests in `/repo.wt/feature-a-fix` without changing your current directory. | **Context:** Starts a server in `/repo.wt/feature-b-refactor` (often run in the background with `&`). |

This provides the agent (or developer) with the full file structure context of the main repo while guaranteeing that execution occurs in a clean, dependency-ready environment dedicated to a single branch.

#### 3. Cleanup and Synchronization

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
