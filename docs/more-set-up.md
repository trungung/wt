Given “personal now, adoptable later” + “balance”, optimize for **low ongoing maintenance** and **high trust** in destructive commands, without building a heavyweight process.

## The balanced setup I’d do (in order)

### 1) Add a minimal but strong CI gate (PR-based)

**Why:** it keeps `main` healthy with almost zero mental overhead.

CI on `pull_request`:

- `gofmt` check (fail if formatting changes)
- `go test ./...`
- `golangci-lint` (small, curated config)
- build on `linux/amd64` and `darwin/arm64` (enough to catch portability issues)

Keep it under ~2–4 minutes.

### 2) Add 5–8 integration tests (only the risky flows)

**Why:** your tool wraps git and touches deletion; unit tests won’t protect you from regressions.

Cover:

- ensure/create path mapping (`feature/x` → `feature-x`)
- default branch special-case (`wt main` prints repo root)
- collision detection (fails)
- remove refuses when dirty; `--force` works
- prune refuses dirty unless `--force`
- postCreate failure triggers rollback (worktree removed; branch deleted if created)

This is the single biggest “adoptable later” investment.

### 3) Add branch protection (light)

Require only:

- PR required to merge to `main`
- CI must pass
  No need for code owner reviews since you’re solo.

### 4) Add “debuggability” hooks early

Implement:

- `WT_DEBUG=1` (prints git commands + resolved paths + default branch decisions to stderr)
- consistent error messages (optionally simple codes like `WT_COLLISION`, `WT_DIRTY`, `WT_ROLLBACK_FAILED`)

This makes your future self and potential users much happier.

### 5) Ship via GitHub Releases (GoReleaser) early, brew later

You already have this. For “balance”:

- release when you hit a coherent milestone (e.g. v0.1.0 with core commands)
- don’t over-release; churn kills adoption

### 6) Docs that matter (tiny but effective)

Add to README:

- 30-second quickstart
- “Safety model” (dirty checks, prune semantics)
- “Config trust model” (postCreate executes commands—treat config as code)
- “Troubleshooting” section: run `wt health` and `WT_DEBUG=1`

### 7) AI PR review: optional, keep it on-demand

Since you’re using opencode, I’d avoid always-on AI review (noise). Use either:

- manual “ask AI” when you touch prune/remove/rollback, or
- a workflow that runs only when you add a PR label like `ai-review`.

## What I would not do yet

- Heavy project management (milestones, complex roadmaps)
- Multi-OS CI matrix explosion
- Homebrew tap right away
- Provider API integration

## A practical “definition of done” for each PR

In your PR template, include:

- [ ] Acceptance criteria impacted updated/checked
- [ ] New/changed behavior has at least one test (unit or integration)
- [ ] `wt health` still passes in a clean repo
- [ ] If destructive logic changed: verified dirty/force flows
