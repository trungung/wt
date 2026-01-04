# wt - git worktree wrapper

`wt` is a fast, branch-centric `git worktree` helper that makes worktrees feel "branch-addressable". It manages paths, creation, cleanup, and safety checks, ensuring a deterministic and flattened worktree structure.

## Core Commands

- `wt`: List worktrees.
- `wt <branch>`: Ensure/create worktree for `<branch>`.
- `wt exec <branch> -- <cmd>`: Execute command inside worktree.
- `wt remove [branch]`: Remove worktree safely.
- `wt prune`: Prune merged worktrees.
- `wt init`: Initialize configuration.
- `wt health`: Validate config and environment.

## Documentation

- [PRD](docs/prd.md)
- [Scope Contract](docs/contract.md)
- [Acceptance Criteria](docs/acceptance.md)
- [Non-Functional Requirements (NFR)](docs/nfr.md)
- [Release Plan](docs/release.md)

## Development

- [Setup & Planning](docs/more-set-up.md)
- [Initial Plan](docs/plan1-3.md)
