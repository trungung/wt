## Day 1 — Skeleton + core behavior (happy path)

**Goal:** make the CLI usable end-to-end for the main flow and match the core contract.

1. **Project structure + interfaces**

- Establish clean layers (cmd / git runner / core / config / ui) to keep scale-friendly separation .
- Implement `wt --help` / subcommand wiring (`list`, `init`, `health`, `remove`, `prune`, `exec`) per scope contract .

2. **Implement `wt` list + `wt <branch>` ensure (no extras)**

- Output `branch<TAB>path` .
- Ensure logic for:
  - mapping `feature/x → feature-x`
  - collision detection = fail
  - default branch special-case prints `$REPO_PATH`
- Keep stdout clean: `wt <branch>` prints only the path .

3. **Local manual smoke test**

- In a throwaway repo, validate: list, ensure, re-run idempotency .

Deliverable end of Day 1: you can run `wt feature/x` and get a worktree at `$REPO_PATH.wt/feature-x` reliably and idempotently , and `wt main` prints `$REPO_PATH` .

---

## Day 2 — Safety-critical flows + rollback

**Goal:** make destructive operations trustworthy.

1. **`wt remove` with dirty detection + confirmation**

- Dirty includes untracked .
- Must warn/confirm when dirty; `--force` bypasses .
- Refuse removing default branch/main worktree .

2. **Post-create “atomic rollback”**

- If `postCreateCmd` fails: exit non-zero and attempt rollback:
  - remove newly created worktree
  - delete newly created branch if created during this run
- Add clear error reporting including rollback status (also aligns with “high trust” setup) .

3. **Start integration tests (5–8) for risky behavior**
   Implement tests for:

- mapping + idempotency
- default branch behavior
- collision failure
- remove dirty confirmation/force
- postCreate rollback
  This matches the recommended high ROI test focus and quality gates .

Deliverable end of Day 2: remove is safe, rollback works, and you have integration tests preventing regressions .

---

## Day 3 — Prune/health + CI + release readiness

**Goal:** make it shippable and maintainable.

1. **`wt prune` (merged into default branch)**

- Prune only merged branches .
- Refuse dirty unless `--force` .
- Support `--dry-run` and `--fetch` .

2. **`wt health`**

- Validate config, warn on missing origin/HEAD fallback usage, and detect collisions .
- Error on invalid JSON and on undeterminable default branch (unless overridden) .

3. **CI pipeline (PR checks) + release workflow**

- Add PR CI: gofmt, go test, golangci-lint, build targets (as per your “balanced setup”) .
- Add GoReleaser + tag-based release workflow and do a local snapshot release run .

4. **Update README install + quickstart**

- Include the GitHub Releases install snippet and basic usage (list/ensure/remove/prune/health/exec)

Deliverable end of Day 3: CI gates + release pipeline exist , prune/health are correct , and you’re ready to tag a `v0.1.0` when you feel good.
