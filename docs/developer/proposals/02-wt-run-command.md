⚠️ **NOT IMPLEMENTED**

This document describes a proposed feature that is **not yet implemented** in `wt`.
For currently available features, see the [API Reference](../../user/api-references/index.md).

---

# Proposal 2: `wt run` Command (Ensure + Exec Combined)

**Status:** Proposed
**Priority:** SHOULD HAVE
**Effort:** 2-3 hours
**Confidence:** 95%

## Quick Reference

| Aspect | Rating | Notes |
|--------|--------|-------|
| Impact | 9/10 | Reduces cognitive load, enables automation |
| Pragmatism | 10/10 | Only ~40 lines of code |
| Risk | 1/10 | Low risk: combines existing functions |
| User Adoption | 9/10 | Immediately obvious value |

---

## The Problem

The current workflow separates worktree creation from command execution:

```bash
# Step 1: Ensure worktree exists
$ wt feature/test

# Step 2: Execute command
$ wt exec feature/test -- npm test
```

This creates several pain points:

1. **Two-step thinking** - Users must consciously think "does the worktree exist?"
2. **Non-idempotent scripts** - Scripts must check existence before running
3. **`exec` requires existence** - The `wt exec` command fails if worktree doesn't exist (see `cmd/wt/main.go:84` which calls `core.FindWorktree`)
4. **Friction in multi-agent workflows** - AI agents want "just run tests on branch X" without state management
5. **Not CI/CD friendly** - Automation prefers single, idempotent commands

### Current Implementation Issue

```go
// cmd/wt/main.go:83-86
path, err := core.FindWorktree(branch)
if err != nil {
    return err
}
```

The `exec` command will fail if the worktree doesn't exist, requiring users to manually ensure it first.

---

## The Solution

Add a `wt run` command that combines ensure + exec in one operation:

```bash
# New: Idempotent, "just works"
$ wt run feature/test -- npm test

# Creates worktree if needed, then runs command
# If worktree exists, just runs command
```

---

## Implementation

### Add Command to main.go

```go
var runCmd = &cobra.Command{
    Use:   "run <branch> -- <command...>",
    Short: "ensure worktree exists and run command in it",
    Long: `Ensures a worktree exists for the given branch, creating it if needed,
then executes the command in that worktree's directory.

This is an idempotent combination of 'wt <branch>' and 'wt exec'.

Examples:
  # Run tests (creates worktree if needed)
  wt run feature/test -- npm test

  # Build from specific base branch
  wt run feature/new --from main -- make build

  # Start dev server
  wt run feature/ui -- npm run dev`,
    Args:  cobra.MinimumNArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        branch := args[0]

        // Find the '--' delimiter
        dashIndex := -1
        for i, arg := range os.Args {
            if arg == "--" {
                dashIndex = i
                break
            }
        }

        if dashIndex == -1 || dashIndex+1 >= len(os.Args) {
            return fmt.Errorf("missing command after --")
        }

        commandArgs := os.Args[dashIndex+1:]

        // KEY DIFFERENCE: Use EnsureWorktree instead of FindWorktree
        path, err := core.EnsureWorktree(branch, fromBase)
        if err != nil {
            var rbErr *core.RollbackError
            if errors.As(err, &rbErr) {
                fmt.Fprintf(os.Stderr, "Error: %v\n", rbErr.OriginalErr)
                fmt.Fprintf(os.Stderr, "Rollback status: %s\n", rbErr.RollbackStatus)
                os.Exit(1)
            }
            return err
        }

        // Execute command (identical to exec command)
        c := exec.Command(commandArgs[0], commandArgs[1:]...)
        c.Dir = path
        c.Stdin = os.Stdin
        c.Stdout = os.Stdout
        c.Stderr = os.Stderr

        if err := c.Run(); err != nil {
            if exitErr, ok := err.(*exec.ExitError); ok {
                os.Exit(exitErr.ExitCode())
            }
            return err
        }

        return nil
    },
}

func init() {
    // ... existing init code

    // Add --from flag support to run command
    runCmd.Flags().StringVarP(&fromBase, "from", "f", "", "base branch to create from")
    rootCmd.AddCommand(runCmd)
}
```

---

## Implementation Plan

### Phase 1: Core Implementation (2 hours)

1. Add `runCmd` to `cmd/wt/main.go` (30-40 lines)
2. Wire up `--from` flag (already exists, reuse logic)
3. Test manually with various scenarios

### Phase 2: Testing (1 hour)

Add integration test in `test/integration_test.go`:

```go
t.Run("run command creates and executes", func(t *testing.T) {
    // Test 1: 'wt run' creates worktree if missing and executes
    output := runWt(t, repoPath, "run", "feature/run-test", "--", "pwd")
    if !strings.Contains(output, "feature-run-test") {
        t.Errorf("Expected pwd output to contain worktree path")
    }

    // Test 2: 'wt run' executes in existing worktree
    output = runWt(t, repoPath, "run", "feature/run-test", "--", "echo", "hello")
    if !strings.Contains(output, "hello") {
        t.Errorf("Expected command output")
    }

    // Test 3: Exit code propagates correctly
    _, err := runWtError(t, repoPath, "run", "feature/run-test", "--", "false")
    if err == nil {
        t.Error("Expected non-zero exit code")
    }
})
```

### Phase 3: Documentation (1 hour)

1. Create `docs/user/api-references/run.md`
2. Update `docs/user/api-references/index.md` to include run
3. Add examples to `docs/user/guides/multi-agent-workflow.md`
4. Update README.md quick reference table

---

## Files to Modify

- `cmd/wt/main.go` - Add runCmd (~40 lines)
- `test/integration_test.go` - Add test cases
- `docs/user/api-references/run.md` - New file
- `docs/user/api-references/index.md` - Add entry
- `docs/user/guides/multi-agent-workflow.md` - Update examples
- `README.md` - Update command table

---

## Example Usage Scenarios

### Scenario 1: Single Command Testing

```bash
# Current (2 commands, need to think):
$ wt feature/api-v2                    # Ensure exists
$ wt exec feature/api-v2 -- npm test   # Then run

# New (1 command, idempotent):
$ wt run feature/api-v2 -- npm test    # Just works!
```

### Scenario 2: Parallel Multi-Agent Builds

```bash
# Run builds in parallel for multiple features
$ wt run feature/auth -- npm run build &
$ wt run feature/payments -- npm run build &
$ wt run feature/api -- npm run build &
$ wait

# Worktrees created if needed, all builds run in isolation
```

### Scenario 3: CI/CD Pipeline

```yaml
# .github/workflows/test.yml
jobs:
  test-feature:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Test feature branch
        run: |
          wt init --yes
          wt run ${{ github.head_ref }} -- npm test
```

### Scenario 4: One-liner Scripts

```bash
#!/bin/bash
# test-all-features.sh
for branch in $(git branch --format='%(refname:short)' | grep '^feature/'); do
    echo "Testing $branch..."
    wt run "$branch" -- npm test || echo "FAILED: $branch"
done
```

---

## Why This is Priority #2

### Impact: 9/10

- Reduces cognitive load significantly
- Perfect for automation and scripting
- Makes the tool more intuitive for new users
- Enables powerful multi-agent workflows

### Pragmatism: 10/10

- Only ~40 lines of code
- Combines two existing, battle-tested functions
- No changes to core business logic
- Implementation is straightforward

### Risk: 1/10

- Very low risk: just combining existing functions
- EnsureWorktree is already well-tested
- Command execution logic copied from existing exec command
- Main risk is ensuring flag compatibility

### User Adoption: 9/10

- Immediately obvious value proposition
- Reduces "do I need to create first?" mental overhead
- Scripts become simpler and more reliable
- Natural evolution of the tool's interface

### Confidence: 95%

- Straightforward implementation
- Leverages existing, tested code paths
- Similar patterns exist in other tools
- Minor risk: edge cases with flag inheritance

---

## User Perception

Users will perceive this as:

- "This is exactly what I wanted from `exec`"
- "Makes scripting so much easier"
- "The tool is reading my mind - just do what I mean"
- "Production-ready for CI/CD"

This positions `wt` as a tool that understands real workflows, not just individual operations.

---

## Alternatives Considered

### Alternative 1: Make `exec` auto-ensure

- ❌ Rejected: Breaking change, violates principle of least surprise
- ❌ Users expect `exec` to fail if worktree missing
- ✅ Current solution: New command, backwards compatible

### Alternative 2: Add `wt <branch> <cmd>` shorthand

- ❌ Rejected: Parsing ambiguity (branch vs command)
- ❌ PRD explicitly excludes this (section 1.2)
- ✅ Current solution: Explicit `run` subcommand with `--`

---

## Success Metrics

### Quantitative

- Usage analytics (if implemented)
- Reduction in multi-step workflow issues
- CI/CD pipeline adoption

### Qualitative

- User feedback: "This makes scripting easier"
- Reduced questions about "how to ensure first?"
- Community scripts using `wt run`

---

## Related Proposals

- [Proposal 1: Shell cd integration](01-shell-cd-integration.md) - Complements workflow improvements
- [Proposal 4: Lifecycle hooks](04-lifecycle-hooks.md) - Hooks could run during `run` command

---

**Status:** Ready for implementation
**Next Steps:** Add runCmd to main.go and write tests
