# Internal Developer Documentation

## ⚠️ Important Notice

These files are historical development artifacts from the v1 planning phase. **Not all content may reflect the current implementation.**

For accurate information about `wt`:

- Refer to [User Documentation](../../user/)
- Refer to [Release Process](../release-process.md)
- Refer to the actual codebase in `internal/` and `cmd/`

## Overview

This directory contains planning documents, acceptance criteria, and design specifications that guided the development of `wt` v1.

## Documents

| File | Purpose | Status |
|------|----------|--------|
| [prd.md](prd.md) | Product Requirements Document - goals, features, and v1 scope | Planning complete |
| [contract.md](contract.md) | v1 Scope Contract - in-scope/out-of-scope decisions | Implemented |
| [acceptance.md](acceptance.md) | Acceptance Criteria - detailed test cases for v1 | Validated |
| [nfr.md](nfr.md) | Non-Functional Requirements - performance, reliability, security | Met |
| [plan1-3.md](plan1-3.md) | 3-Day Implementation Plan | Completed |
| [more-set-up.md](more-set-up.md) | Balanced CI/CD and process setup recommendations | Referenced |

## Key Decisions

These documents established:

- Branch-centric UX with deterministic path mapping
- Default branch special case behavior
- Strict collision detection (no auto-resolution)
- Post-create atomic rollback requirements
- Dirty worktree protection in remove/prune commands
- Zsh completion only (Bash/Fish deferred)
- File locking for concurrent safety

## Current Implementation

For the actual implementation, see:

- `cmd/wt/main.go` - CLI entry point
- `internal/core/` - Business logic
- `internal/git/` - Git operations wrapper
- `internal/config/` - Configuration handling
- `internal/ui/` - Interactive prompts
- `test/` - Integration and concurrency tests
