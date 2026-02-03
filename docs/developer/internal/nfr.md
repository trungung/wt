## Non‑Functional Requirements (NFR) — `wt` v1 (Go)

### 1) Performance & Responsiveness

- **P1 Startup latency:** binary startup should feel instant; avoid heavy init on every run.
- **P2 Git calls:** minimize number of `git` invocations per command; prefer single porcelain call + parse when possible.
- **P3 Completion performance:** any completion-related mode must run fast (<100ms typical). Avoid scanning filesystem trees.
- **P4 No unnecessary network:** never hit network implicitly except when user passes `--fetch` (explicit).
- **P5 Large repos:** commands must remain responsive in repos with:
  - many branches (1k+)
  - many worktrees (50+)
  - large working directories
    (Do not walk the repo tree except when required for copyPatterns matches.)

### 2) Reliability, Correctness, and Safety

- **R1 Idempotency:** repeated runs yield same result without surprises (especially `wt <branch>`).
- **R2 No silent destructive actions:** any deletion must be explicit (`remove`, `prune`) and guarded (dirty checks, default-branch safeguards).
- **R3 Best-effort rollback:** on postCreate failure, rollback must be attempted and the rollback status must be reported.
- **R4 Deterministic mapping:** branch→path mapping stable; collisions detected and treated as errors.
- **R5 Clear exit codes:** non-zero on failure; `exec` returns child exit code. Distinguish “user cancelled” vs “hard error” if feasible (documented).
- **R6 Concurrency safety:** if two `wt <branch>` run concurrently, tool must not corrupt state; must fail cleanly or serialize (file lock recommended).

### 3) UX & CLI Design Best Practices

- **U1 Consistent command grammar:** verbs for actions (`list`, `remove`, `prune`, `exec`, `health`); keep `wt <branch>` as the single intentional exception.
- **U2 Predictable stdout/stderr:** machine-usable stdout; human messages to stderr when appropriate.
- **U3 Discoverability:** `wt --help` and `wt <cmd> --help` must be clear and concise.
- **U4 Non-interactive first:** every interactive operation must have a non-interactive equivalent (`--yes`, `--force`, etc.).
- **U5 Stable output modes:** list/dry-run outputs should be stable across versions (or provide `--porcelain`/`--json` to guarantee stability).
- **U6 Good errors:** errors should include:
  - what failed
  - why (if known)
  - what to do next (actionable hint)

### 4) Maintainability & “Develop for scale”

- **M1 Clean architecture:** separate layers:
  - `cmd/` (CLI parsing, help, flags)
  - `internal/git/` (git command execution + parsing)
  - `internal/core/` (business rules: mapping, collisions, default branch, pruning logic)
  - `internal/config/` (read/validate config)
  - `internal/ui/` (prompts, formatting, printing)
- **M2 Testability:** core logic must be unit-testable without a real git repo by:
  - abstracting git runner interface
  - using fixtures for porcelain outputs
- **M3 Small, composable functions:** avoid god functions; explicit data structures for worktrees, branches, config, health results.
- **M4 Backward compatibility:** config schema changes must be additive in v1.x; unknown keys warn, not error.
- **M5 Minimal dependencies:** prefer stdlib; choose a small CLI framework (Cobra is fine) and avoid heavy stacks.
- **M6 Observability for debugging:** optional `WT_DEBUG=1` to print executed git commands + timings to stderr.

### 5) Portability & Packaging

- **X1 OS support:** macOS + Linux (v1). Windows optional but not required.
- **X2 Installability:** single static-ish binary; no runtime dependencies besides `git` and a POSIX shell for postCreate (documented).
- **X3 Shell compatibility:** zsh, bash, and fish completions supported via `wt completion` and `wt shell-setup`.

### 6) Security & Trust

- **S1 Safe command execution:** postCreate commands must not do unexpected shell interpolation beyond what user configured.
- **S2 Avoid leaking secrets:** do not print `.env` contents; avoid verbose logs unless debug enabled.
- **S3 No network calls unless requested:** reinforces trust.
- **S4 Config trust model:** config is repo-local; treat it like code—document that running postCreate executes commands from the repo config.

### 7) Data Integrity & Filesystem Behavior

- **F1 Atomic config writes:** `wt init` writes config via temp file + rename.
- **F2 No partial state:** if a create operation fails, tool should not leave behind:
  - registered worktree entries
  - stray directories (best effort cleanup)
- **F3 Respect permissions:** clear error messages when unable to create dirs or files.
- **F4 Copy semantics:** copyPatterns are “best effort”; any skipped copies must be surfaced via `wt health` (and optionally during creation if verbose).

### 8) Compatibility with Git behavior

- **G1 Use porcelain outputs:** parse `git worktree list --porcelain` and other stable formats.
- **G2 Don’t assume branch uniqueness:** handle detached worktrees; handle branch names with unusual chars (sanitization rules apply).
- **G3 Don’t rely on current branch:** prune uses default branch reference, not whatever is checked out.

---

## Suggested “Quality Gates” for v1 release

- **QG1**: Unit tests for mapping + collision detection + default branch detection logic.
- **QG2**: Integration tests in temp repos for: create, re-run, remove dirty, prune dry-run, postCreate rollback.
- **QG3**: Benchmark: `wt` list and `wt <existing-branch>` under an average repo runs within target latency.
- **QG4**: Lint/format: `gofmt`, `golangci-lint` (reasonable set).

---
