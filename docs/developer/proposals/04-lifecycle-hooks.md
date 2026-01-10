# Proposal 4: Lifecycle Hooks System

**Status:** Proposed
**Priority:** COULD HAVE
**Effort:** 6-8 hours
**Confidence:** 85%

## Quick Reference

| Aspect | Rating | Notes |
|--------|--------|-------|
| Impact | 7/10 | Powerful for power users and teams |
| Pragmatism | 6/10 | Moderate complexity (~12 hours) |
| Risk | 4/10 | Executing arbitrary user code |
| User Adoption | 6/10 | Advanced feature, not for all users |

---

## The Problem

Different teams and projects have specialized workflows beyond `wt`'s built-in capabilities:

### Common Use Cases

- Send Slack notifications when worktrees are created
- Update project management systems (Jira, Linear)
- Validate branch naming conventions before creation
- Check for open pull requests before removal
- Clean up Docker containers associated with worktrees
- Update database entries tracking development environments
- Run security scans before worktree creation
- Verify dependencies before executing commands

### Current Limitation

The only extension point is `postCreateCmd`, which:
- Only runs after creation (no pre-creation validation)
- Cannot prevent creation on validation failure
- Has no hooks for remove/prune operations
- Cannot access context about the operation

This forces teams to:
- Wrap `wt` with custom scripts
- Fork the project to add custom logic
- Accept reduced automation capabilities

---

## The Solution

Add a comprehensive lifecycle hooks system via `.wt.config.json`:

```json
{
  "defaultBranch": "main",
  "worktreePathTemplate": "$REPO_PATH.wt",
  "hooks": {
    "pre-create": [
      "./scripts/validate-branch-name.sh",
      "./scripts/check-disk-space.sh"
    ],
    "post-create": [
      "bun install",
      "./scripts/notify-team.sh 'Created worktree for $BRANCH'"
    ],
    "pre-remove": [
      "./scripts/check-open-prs.sh $BRANCH"
    ],
    "post-remove": [
      "./scripts/cleanup-docker.sh $BRANCH",
      "./scripts/update-dashboard.sh"
    ],
    "pre-prune": [
      "./scripts/audit-prune-candidates.sh"
    ],
    "post-prune": [
      "./scripts/notify-cleanup.sh"
    ]
  },
  "hookTimeout": 30
}
```

---

## Hook Types and Behavior

### Pre-hooks (Validation)

- Execute **before** the operation
- Non-zero exit code **aborts** the operation
- Can access operation context via environment variables
- Output goes to stderr
- Use cases: validation, permissions checks, external system queries

### Post-hooks (Notification/Cleanup)

- Execute **after** the operation completes successfully
- Errors are logged but don't fail the operation
- Can access result context via environment variables
- Output goes to stdout/stderr normally
- Use cases: notifications, cleanup, external system updates

---

## Environment Variables Available to Hooks

### All hooks receive:

```bash
WT_HOOK              # Hook name (pre-create, post-create, etc.)
WT_BRANCH            # Branch name
WT_REPO_ROOT         # Repository root path
WT_WORKTREE_PATH     # Worktree path (empty for pre-create)
WT_BASE_BRANCH       # Base branch (for create operations with --from)
WT_OPERATION         # Operation type (create, remove, prune)
```

### Additional per-hook:

```bash
# pre-create / post-create
WT_IS_NEW_BRANCH     # "true" if creating new branch, "false" if checking out existing

# pre-prune / post-prune
WT_PRUNE_CANDIDATES  # Comma-separated list of branches to be pruned
WT_PRUNE_COUNT       # Number of worktrees being pruned
```

---

## Implementation

### Phase 1: Config Schema

Update `internal/config/config.go`:

```go
type Hooks struct {
    PreCreate  []string `json:"pre-create"`
    PostCreate []string `json:"post-create"`
    PreRemove  []string `json:"pre-remove"`
    PostRemove []string `json:"post-remove"`
    PrePrune   []string `json:"pre-prune"`
    PostPrune  []string `json:"post-prune"`
}

type Config struct {
    DefaultBranch            string   `json:"defaultBranch"`
    WorktreePathTemplate     string   `json:"worktreePathTemplate"`
    WorktreeCopyPatterns     []string `json:"worktreeCopyPatterns"`
    PostCreateCmd            []string `json:"postCreateCmd"` // Kept for backwards compat
    DeleteBranchWithWorktree bool     `json:"deleteBranchWithWorktree"`
    Hooks                    Hooks    `json:"hooks"`
    HookTimeout              int      `json:"hookTimeout"` // seconds, default 30
}
```

### Phase 2: Hook Execution Engine

Add to `internal/core/hooks.go`:

```go
package core

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"
)

type HookContext struct {
    Hook          string
    Branch        string
    RepoRoot      string
    WorktreePath  string
    BaseBranch    string
    Operation     string
    IsNewBranch   bool
    PruneCandidates []string
}

// ExecuteHooks runs hooks with the given context
// isPre determines behavior: pre-hooks abort on error, post-hooks just log
func ExecuteHooks(hooks []string, ctx HookContext, cfg *config.Config, isPre bool) error {
    if len(hooks) == 0 {
        return nil
    }

    timeout := 30 * time.Second
    if cfg.HookTimeout > 0 {
        timeout = time.Duration(cfg.HookTimeout) * time.Second
    }

    // Build environment
    env := os.Environ()
    env = append(env,
        "WT_HOOK="+ctx.Hook,
        "WT_BRANCH="+ctx.Branch,
        "WT_REPO_ROOT="+ctx.RepoRoot,
        "WT_WORKTREE_PATH="+ctx.WorktreePath,
        "WT_BASE_BRANCH="+ctx.BaseBranch,
        "WT_OPERATION="+ctx.Operation,
    )

    if ctx.IsNewBranch {
        env = append(env, "WT_IS_NEW_BRANCH=true")
    } else {
        env = append(env, "WT_IS_NEW_BRANCH=false")
    }

    if len(ctx.PruneCandidates) > 0 {
        env = append(env,
            "WT_PRUNE_CANDIDATES="+strings.Join(ctx.PruneCandidates, ","),
            fmt.Sprintf("WT_PRUNE_COUNT=%d", len(ctx.PruneCandidates)),
        )
    }

    // Execute each hook
    for _, hookCmd := range hooks {
        if hookCmd == "" {
            continue
        }

        // Expand environment variables in command
        hookCmd = os.ExpandEnv(hookCmd)

        // Parse command
        parts := strings.Fields(hookCmd)
        if len(parts) == 0 {
            continue
        }

        // Create command with timeout
        ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()

        cmd := exec.CommandContext(ctxTimeout, parts[0], parts[1:]...)
        cmd.Env = env

        // Set working directory
        if ctx.WorktreePath != "" && isPre == false {
            cmd.Dir = ctx.WorktreePath
        } else {
            cmd.Dir = ctx.RepoRoot
        }

        // Capture output
        output, err := cmd.CombinedOutput()

        if err != nil {
            if ctxTimeout.Err() == context.DeadlineExceeded {
                err = fmt.Errorf("hook timed out after %v", timeout)
            }

            if isPre {
                // Pre-hooks abort on error
                return fmt.Errorf("%s hook failed: %s\nOutput: %s", ctx.Hook, err, output)
            } else {
                // Post-hooks just log errors
                fmt.Fprintf(os.Stderr, "Warning: %s hook failed: %s\nOutput: %s\n", ctx.Hook, err, output)
            }
        } else if len(output) > 0 {
            // Show hook output if any
            fmt.Fprintf(os.Stderr, "[%s] %s\n", ctx.Hook, output)
        }
    }

    return nil
}

// Helper to check if hooks should be skipped
var skipHooks bool

func ShouldSkipHooks() bool {
    return skipHooks || os.Getenv("WT_SKIP_HOOKS") != ""
}

func SetSkipHooks(skip bool) {
    skipHooks = skip
}
```

### Phase 3: Integrate Hooks into Operations

Update `internal/core/core.go` in `EnsureWorktree`:

```go
// In EnsureWorktree function
func EnsureWorktree(branch, base string) (string, error) {
    // ... existing setup code ...

    cfg, err := config.LoadConfig(root)
    if err != nil {
        return "", fmt.Errorf("failed to load config: %w", err)
    }

    // ... existing logic to determine if new branch ...

    // Execute pre-create hooks
    if !ShouldSkipHooks() {
        hookCtx := HookContext{
            Hook:        "pre-create",
            Branch:      branch,
            RepoRoot:    root,
            BaseBranch:  base,
            Operation:   "create",
            IsNewBranch: isNewBranch,
        }
        if err := ExecuteHooks(cfg.Hooks.PreCreate, hookCtx, cfg, true); err != nil {
            return "", fmt.Errorf("pre-create hook failed: %w", err)
        }
    }

    // ... existing worktree creation logic ...

    if err := git.CreateWorktree(targetPath, branch, base); err != nil {
        return "", err
    }

    // Execute post-create hooks (includes old postCreateCmd for backwards compat)
    if !ShouldSkipHooks() {
        hookCtx := HookContext{
            Hook:         "post-create",
            Branch:       branch,
            RepoRoot:     root,
            WorktreePath: targetPath,
            BaseBranch:   base,
            Operation:    "create",
            IsNewBranch:  isNewBranch,
        }

        // Run new hooks
        _ = ExecuteHooks(cfg.Hooks.PostCreate, hookCtx, cfg, false)

        // Run old postCreateCmd for backwards compatibility
        if len(cfg.PostCreateCmd) > 0 {
            fmt.Fprintf(os.Stderr, "Warning: postCreateCmd is deprecated, use hooks.post-create instead\n")
            _ = ExecuteHooks(cfg.PostCreateCmd, hookCtx, cfg, false)
        }
    }

    return targetPath, nil
}
```

Similar integration needed in `RemoveWorktree` and `PruneWorktrees`.

### Phase 4: CLI Flags

Update `cmd/wt/main.go`:

```go
var skipHooks bool

func init() {
    // Add global flag to all commands
    rootCmd.PersistentFlags().BoolVar(&skipHooks, "skip-hooks", false, "skip all hooks")

    // Pass to core package
    if skipHooks {
        core.SetSkipHooks(true)
    }
}
```

---

## Implementation Plan

### Phase 1: Config and Engine (3 hours)

1. Extend Config struct with Hooks
2. Implement ExecuteHooks function
3. Add timeout handling
4. Test hook execution in isolation

### Phase 2: Integration (3 hours)

1. Integrate pre/post-create hooks into EnsureWorktree
2. Integrate pre/post-remove hooks into RemoveWorktree
3. Integrate pre/post-prune hooks into PruneWorktrees
4. Maintain backwards compatibility with postCreateCmd

### Phase 3: CLI and Safety (2 hours)

1. Add `--skip-hooks` flag to all commands
2. Add environment variable support (WT_SKIP_HOOKS)
3. Test hook abortion scenarios
4. Test timeout scenarios

### Phase 4: Documentation (2 hours)

1. Create `docs/user/guides/hooks.md` with examples
2. Update configuration reference
3. Add security warnings
4. Create example hook scripts

### Phase 5: Examples and Testing (2 hours)

1. Create example hooks in `examples/hooks/`
2. Add integration tests for hook execution
3. Test pre-hook abortion
4. Test post-hook error handling

---

## Files to Modify/Create

- `internal/config/config.go` - Add Hooks struct
- `internal/core/hooks.go` - **New file** - Hook execution engine
- `internal/core/core.go` - Integrate hooks into operations
- `cmd/wt/main.go` - Add --skip-hooks flag
- `docs/user/guides/hooks.md` - **New file** - Hook documentation
- `docs/user/api-references/configuration.md` - Update with hooks
- `examples/hooks/` - **New directory** - Example hook scripts

---

## Example Hook Scripts

### Pre-create: Validate Branch Name

```bash
#!/bin/bash
# examples/hooks/validate-branch-name.sh

BRANCH="$WT_BRANCH"

# Enforce naming convention: type/description
if [[ ! "$BRANCH" =~ ^(feature|bugfix|hotfix|release)/.+ ]]; then
    echo "Error: Branch name must start with feature/, bugfix/, hotfix/, or release/"
    echo "Got: $BRANCH"
    exit 1
fi

echo "✓ Branch name is valid"
exit 0
```

### Post-create: Notify Slack

```bash
#!/bin/bash
# examples/hooks/notify-slack.sh

SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

curl -X POST "$SLACK_WEBHOOK" \
    -H 'Content-Type: application/json' \
    -d "{
        \"text\": \"🌿 New worktree created\",
        \"blocks\": [{
            \"type\": \"section\",
            \"text\": {
                \"type\": \"mrkdwn\",
                \"text\": \"*Branch:* \`$WT_BRANCH\`\\n*Path:* \`$WT_WORKTREE_PATH\`\"
            }
        }]
    }"
```

### Pre-remove: Check Open PRs

```bash
#!/bin/bash
# examples/hooks/check-open-prs.sh

BRANCH="$WT_BRANCH"

# Check if PR exists using gh CLI
if gh pr view "$BRANCH" --json state --jq '.state' | grep -q "OPEN"; then
    echo "Error: Branch $BRANCH has an open PR"
    echo "Please close or merge the PR before removing the worktree"
    exit 1
fi

exit 0
```

### Post-prune: Update Dashboard

```bash
#!/bin/bash
# examples/hooks/update-dashboard.sh

API_URL="https://dev-dashboard.company.com/api/worktrees"
PRUNED_BRANCHES="$WT_PRUNE_CANDIDATES"

curl -X POST "$API_URL/prune" \
    -H "Content-Type: application/json" \
    -d "{\"branches\": \"$PRUNED_BRANCHES\", \"count\": $WT_PRUNE_COUNT}"
```

---

## Usage Examples

### Basic Configuration

```json
{
  "hooks": {
    "pre-create": ["./scripts/validate.sh"],
    "post-create": ["npm install"],
    "post-remove": ["./scripts/cleanup.sh $WT_BRANCH"]
  }
}
```

### Bypassing Hooks

```bash
# Skip all hooks in emergency
$ wt remove feature/broken --skip-hooks

# Or via environment variable
$ WT_SKIP_HOOKS=1 wt remove feature/broken
```

### Debugging Hooks

```bash
# Hooks print their output to stderr
$ wt feature/test
[pre-create] ✓ Branch name is valid
[pre-create] ✓ Disk space sufficient
/Users/dev/repo.wt/feature-test
[post-create] Installing dependencies...
[post-create] ✓ Slack notification sent
```

---

## Why This is Priority #4

### Impact: 7/10

- Unlocks powerful customization
- Enables enterprise workflows
- Future-proofs the tool
- But: Only needed by power users and teams

### Pragmatism: 6/10

- Moderate complexity (~12 hours implementation)
- Requires careful error handling
- Need to handle timeouts and security
- Good pattern to follow (Git hooks, npm scripts)

### Risk: 4/10

- Medium risk: executing arbitrary user code
- Timeout handling adds complexity
- Backwards compatibility with postCreateCmd
- Security: hooks run with user's permissions

### User Adoption: 6/10

- Power users will love it
- Most users won't use it initially
- Requires scripting knowledge
- Discoverability: needs good documentation

### Confidence: 85%

- Command execution is well-understood
- Main uncertainty: edge cases in timeout handling
- Risk: users writing complex hooks that break
- Mitigation: Good documentation and examples

---

## User Perception

### Power users and teams will perceive this as:

- "Now it's a real platform, not just a tool"
- "We can integrate it into our workflows"
- "This is production-ready and enterprise-friendly"
- "The developers thought about extensibility"

### Casual users will perceive it as:

- "Nice to have the option if I need it"
- "Shows the tool is mature and well-designed"

This feature positions `wt` as a **platform** rather than just a utility, opening doors for community contributions of hook libraries.

---

## Alternatives Considered

### Alternative 1: Plugin system with Go plugins

- ❌ Rejected: Go plugin system is fragile and platform-specific
- ❌ Much more complex implementation
- ✅ Current solution: Shell scripts are universal

### Alternative 2: Webhooks to external services

- ❌ Rejected: Requires network, adds latency, complex auth
- ❌ Doesn't handle pre-hooks (validation)
- ✅ Current solution: Local scripts, synchronous execution

---

## Security Considerations

### Important Notes for Documentation

1. **Hooks run with user's permissions** - Can do anything the user can do
2. **Path handling** - Always use absolute paths or validate relative paths
3. **Input validation** - Don't trust branch names blindly (injection risks)
4. **Timeout protection** - Default 30s timeout prevents infinite loops
5. **Emergency bypass** - `--skip-hooks` flag for when hooks break

### Recommended Best Practices

```bash
# Good: Validate inputs
if [[ "$WT_BRANCH" =~ [^a-zA-Z0-9/_-] ]]; then
    echo "Invalid branch name"
    exit 1
fi

# Good: Use timeouts for external calls
timeout 10 curl ...

# Good: Check exit codes
if ! command; then
    echo "Command failed"
    exit 1
fi
```

---

## Success Metrics

### Quantitative

- Number of community hook libraries created
- Enterprise adoption mentions
- Questions about hook functionality

### Qualitative

- Power user feedback
- Integration examples shared
- Community contributions

---

## Related Proposals

- [Proposal 2: wt run command](02-wt-run-command.md) - Could trigger hooks
- [Proposal 3: Rich status display](03-rich-status-display.md) - Could show hook status

---

## Future Enhancements

### Potential additions (not in scope for v1):

- Hook templates library
- Hook validation/linting
- Async hooks (background execution)
- Hook composition (pre-defined hook chains)
- Hook marketplace/registry

---

**Status:** Ready for implementation (after core features)
**Next Steps:** Create config schema and hook execution engine
**Recommendation:** Implement after Proposals 1, 2, and 5 are complete
