# Proposal 3: Rich Status Display with Git Awareness

**Status:** Proposed
**Priority:** SHOULD HAVE
**Effort:** 4-6 hours
**Confidence:** 90%

## Quick Reference

| Aspect | Rating | Notes |
|--------|--------|-------|
| Impact | 8/10 | Significantly improves visibility |
| Pragmatism | 7/10 | Moderate implementation complexity |
| Risk | 3/10 | Performance with many worktrees |
| User Adoption | 8/10 | Immediately noticeable improvement |

---

## The Problem

The current `wt` list output is minimalist and script-friendly:

```
main    /Users/dev/repo
feature/auth    /Users/dev/repo.wt/feature-auth
feature/payments    /Users/dev/repo.wt/feature-payments
```

This provides no context about worktree state. Users cannot answer these questions without manually checking each worktree:

- Which worktrees have uncommitted changes?
- Which are ahead/behind their remote branches?
- Which branches have been merged and can be pruned?
- Which worktrees haven't been used recently?
- Which is my current worktree (based on cwd)?

### Pain Points

- Managing 5+ worktrees requires manual inspection of each
- No visibility into "what needs attention"
- Users run `git status` manually in each worktree
- No way to quickly identify stale or merged worktrees
- Prune candidates not visible until running `wt prune --dry-run`

---

## The Solution

Add rich, colorized status output while preserving script-friendly plain output:

```bash
$ wt status
# or just: wt

Worktrees (5 total, 2 need attention):

● main                   /Users/dev/repo                       [clean] [↑1]
● feature/auth           /Users/dev/repo.wt/feature-auth       [dirty] [↑3 ↓1] ⚠️
  feature/api-v2         /Users/dev/repo.wt/feature-api-v2     [clean] [↓2]
● feature/payments       /Users/dev/repo.wt/feature-payments   [dirty] [merged] 💡
  hotfix/security-patch  /Users/dev/repo.wt/hotfix-security-patch [clean] [stale: 14d]

Legend:
  ● = current worktree (based on cwd)
  [clean] = no uncommitted changes
  [dirty] = uncommitted changes present
  [↑N] = N commits ahead of remote
  [↓N] = N commits behind remote
  [merged] = branch merged into main (candidate for pruning)
  [stale: Nd] = no commits in N days
  ⚠️  = needs attention (dirty or conflicts)
  💡 = can be pruned (merged)
```

### Plain Output Mode (for scripts)

```bash
$ wt --plain
# or: wt status --plain
main    /Users/dev/repo
feature/auth    /Users/dev/repo.wt/feature-auth
feature/api-v2  /Users/dev/repo.wt/feature-api-v2
feature/payments    /Users/dev/repo.wt/feature-payments
hotfix/security-patch   /Users/dev/repo.wt/hotfix-security-patch
```

---

## Implementation

### Phase 1: Status Data Structure

Add to `internal/core/core.go`:

```go
// WorktreeStatus represents the detailed status of a worktree
type WorktreeStatus struct {
    Worktree    git.Worktree
    IsCurrent   bool   // based on current working directory
    IsDirty     bool   // has uncommitted changes
    Ahead       int    // commits ahead of remote
    Behind      int    // commits behind remote
    IsMerged    bool   // merged into default branch
    LastCommit  time.Time
    StaleDays   int
}

// GetWorktreeStatuses returns detailed status for all worktrees
func GetWorktreeStatuses() ([]WorktreeStatus, error) {
    root, err := git.GetRepoRoot()
    if err != nil {
        return nil, err
    }

    cfg, err := config.LoadConfig(root)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    defaultBranch := cfg.DefaultBranch
    if defaultBranch == "" {
        defaultBranch, err = git.GetDefaultBranch()
        if err != nil {
            return nil, err
        }
    }

    worktrees, err := git.ListWorktrees()
    if err != nil {
        return nil, err
    }

    // Get merged branches once
    merged, _ := git.GetMergedBranches(defaultBranch)
    mergedSet := make(map[string]bool)
    for _, b := range merged {
        mergedSet[b] = true
    }

    // Get current directory to detect current worktree
    cwd, _ := os.Getwd()

    var statuses []WorktreeStatus
    for _, wt := range worktrees {
        status := WorktreeStatus{
            Worktree:  wt,
            IsMerged:  mergedSet[wt.Branch],
        }

        // Check if current
        if cwd != "" {
            cleanWtPath := filepath.Clean(wt.Path)
            if strings.HasPrefix(cwd, cleanWtPath) {
                status.IsCurrent = true
            }
        }

        // Check dirty status
        if dirty, err := git.IsDirty(wt.Path); err == nil {
            status.IsDirty = dirty
        }

        // Check ahead/behind (skip for detached)
        if wt.Branch != "(detached)" && wt.Branch != defaultBranch {
            ahead, behind := git.GetAheadBehind(wt.Path, wt.Branch)
            status.Ahead = ahead
            status.Behind = behind
        }

        // Check last commit time
        if lastCommit, err := git.GetLastCommitTime(wt.Path); err == nil {
            status.LastCommit = lastCommit
            status.StaleDays = int(time.Since(lastCommit).Hours() / 24)
        }

        statuses = append(statuses, status)
    }

    return statuses, nil
}
```

### Phase 2: Git Helper Functions

Add to `internal/git/git.go`:

```go
// GetAheadBehind returns commits ahead and behind remote
func GetAheadBehind(path, branch string) (int, int) {
    // Get remote tracking branch
    out, err := run(path, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
    if err != nil {
        return 0, 0 // No remote tracking
    }

    remote := strings.TrimSpace(string(out))

    // Get ahead count
    aheadOut, err := run(path, "rev-list", "--count", remote+"..HEAD")
    ahead := 0
    if err == nil {
        fmt.Sscanf(string(aheadOut), "%d", &ahead)
    }

    // Get behind count
    behindOut, err := run(path, "rev-list", "--count", "HEAD.."+remote)
    behind := 0
    if err == nil {
        fmt.Sscanf(string(behindOut), "%d", &behind)
    }

    return ahead, behind
}

// GetLastCommitTime returns the timestamp of the last commit
func GetLastCommitTime(path string) (time.Time, error) {
    out, err := run(path, "log", "-1", "--format=%ct")
    if err != nil {
        return time.Time{}, err
    }

    timestamp := strings.TrimSpace(string(out))
    unixTime, err := strconv.ParseInt(timestamp, 10, 64)
    if err != nil {
        return time.Time{}, err
    }

    return time.Unix(unixTime, 0), nil
}
```

### Phase 3: Rich Output with Lipgloss

Add to `internal/ui/ui.go`:

```go
import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/trungung/wt/internal/core"
)

var (
    // Color styles (respects NO_COLOR)
    currentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
    branchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
    pathStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
    cleanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
    dirtyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
    mergedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
    staleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// RenderWorktreeStatus renders a rich status view
func RenderWorktreeStatus(statuses []core.WorktreeStatus) string {
    var sb strings.Builder

    // Header
    total := len(statuses)
    needsAttention := 0
    for _, s := range statuses {
        if s.IsDirty || s.Behind > 0 {
            needsAttention++
        }
    }

    sb.WriteString(fmt.Sprintf("Worktrees (%d total", total))
    if needsAttention > 0 {
        sb.WriteString(fmt.Sprintf(", %d need attention", needsAttention))
    }
    sb.WriteString("):\n\n")

    // Find max branch length for alignment
    maxBranchLen := 0
    for _, s := range statuses {
        if len(s.Worktree.Branch) > maxBranchLen {
            maxBranchLen = len(s.Worktree.Branch)
        }
    }

    // Render each worktree
    for _, s := range statuses {
        // Current indicator
        indicator := "  "
        if s.IsCurrent {
            indicator = currentStyle.Render("● ")
        }

        // Branch name (padded)
        branch := s.Worktree.Branch
        paddedBranch := branch + strings.Repeat(" ", maxBranchLen-len(branch))

        // Path
        path := pathStyle.Render(s.Worktree.Path)

        // Status indicators
        var indicators []string

        // Dirty/clean
        if s.IsDirty {
            indicators = append(indicators, dirtyStyle.Render("[dirty]"))
        } else {
            indicators = append(indicators, cleanStyle.Render("[clean]"))
        }

        // Ahead/behind
        if s.Ahead > 0 {
            indicators = append(indicators, fmt.Sprintf("[↑%d]", s.Ahead))
        }
        if s.Behind > 0 {
            indicators = append(indicators, fmt.Sprintf("[↓%d]", s.Behind))
        }

        // Merged
        if s.IsMerged {
            indicators = append(indicators, mergedStyle.Render("[merged]"))
        }

        // Stale
        if s.StaleDays > 7 {
            indicators = append(indicators, staleStyle.Render(fmt.Sprintf("[stale: %dd]", s.StaleDays)))
        }

        // Attention/action emojis
        if s.IsDirty || s.Behind > 0 {
            indicators = append(indicators, "⚠️")
        }
        if s.IsMerged {
            indicators = append(indicators, "💡")
        }

        // Assemble line
        line := fmt.Sprintf("%s%-*s  %s  %s",
            indicator,
            maxBranchLen,
            branchStyle.Render(paddedBranch),
            path,
            strings.Join(indicators, " "),
        )

        sb.WriteString(line + "\n")
    }

    // Legend
    sb.WriteString("\nLegend:\n")
    sb.WriteString("  ● = current worktree (based on cwd)\n")
    sb.WriteString("  [clean] = no uncommitted changes\n")
    sb.WriteString("  [dirty] = uncommitted changes present\n")
    sb.WriteString("  [↑N] = N commits ahead of remote\n")
    sb.WriteString("  [↓N] = N commits behind remote\n")
    sb.WriteString("  [merged] = branch merged into default (candidate for pruning)\n")
    sb.WriteString("  [stale: Nd] = no commits in N days\n")
    sb.WriteString("  ⚠️  = needs attention (dirty or behind remote)\n")
    sb.WriteString("  💡 = can be pruned (merged into default branch)\n")

    return sb.String()
}
```

### Phase 4: Update Root Command

Modify `cmd/wt/main.go`:

```go
var plainOutput bool

var rootCmd = &cobra.Command{
    Use:     "wt [branch]",
    Short:   "wt is a branch-centric git worktree helper",
    Long:    `A fast, branch-addressable git worktree manager.`,
    Version: version,
    Args:    cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            // wt: list worktrees with rich status
            if plainOutput || os.Getenv("NO_COLOR") != "" {
                // Plain output for scripts
                worktrees, err := git.ListWorktrees()
                if err != nil {
                    return err
                }
                for _, wt := range worktrees {
                    fmt.Printf("%s\t%s\n", wt.Branch, wt.Path)
                }
            } else {
                // Rich status output
                statuses, err := core.GetWorktreeStatuses()
                if err != nil {
                    return err
                }
                fmt.Print(ui.RenderWorktreeStatus(statuses))
            }
            return nil
        }

        // wt <branch>: ensure worktree (unchanged)
        branch := args[0]
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
        fmt.Println(path)
        return nil
    },
}

// Optional: Add explicit status command
var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "show detailed worktree status",
    RunE: func(cmd *cobra.Command, args []string) error {
        statuses, err := core.GetWorktreeStatuses()
        if err != nil {
            return err
        }

        if plainOutput {
            for _, s := range statuses {
                fmt.Printf("%s\t%s\n", s.Worktree.Branch, s.Worktree.Path)
            }
        } else {
            fmt.Print(ui.RenderWorktreeStatus(statuses))
        }

        return nil
    },
}

func init() {
    rootCmd.Flags().StringVarP(&fromBase, "from", "f", "", "base branch to create from")
    rootCmd.Flags().BoolVar(&plainOutput, "plain", false, "plain output (tab-separated)")

    // Add status command
    statusCmd.Flags().BoolVar(&plainOutput, "plain", false, "plain output (tab-separated)")
    rootCmd.AddCommand(statusCmd)

    // ... rest of init
}
```

---

## Implementation Plan

### Phase 1: Git Status Functions (2 hours)

1. Add `GetAheadBehind` to `internal/git/git.go`
2. Add `GetLastCommitTime` to `internal/git/git.go`
3. Test git functions in isolation

### Phase 2: Core Status Logic (2 hours)

1. Add `WorktreeStatus` struct to `internal/core/core.go`
2. Implement `GetWorktreeStatuses` function
3. Test with various worktree configurations

### Phase 3: Rich UI Rendering (2 hours)

1. Add `RenderWorktreeStatus` to `internal/ui/ui.go`
2. Use lipgloss for colors (already in dependencies)
3. Respect NO_COLOR environment variable
4. Test color output across terminals

### Phase 4: CLI Integration (1 hour)

1. Update root command to use rich output by default
2. Add `--plain` flag for script compatibility
3. Add optional `wt status` alias command
4. Test backwards compatibility

### Phase 5: Documentation (1 hour)

1. Create `docs/user/api-references/status.md`
2. Update quickstart guide with status examples
3. Add screenshots to documentation
4. Update README with new output format

---

## Files to Modify

- `internal/core/core.go` - Add WorktreeStatus, GetWorktreeStatuses
- `internal/git/git.go` - Add GetAheadBehind, GetLastCommitTime
- `internal/ui/ui.go` - Add RenderWorktreeStatus
- `cmd/wt/main.go` - Update root command, add status command
- `docs/user/api-references/status.md` - New file
- `docs/user/api-references/index.md` - Update
- `docs/user/guides/quickstart.md` - Add status examples

---

## Example Usage Scenarios

### Scenario 1: Daily Status Check

```bash
$ wt
Worktrees (4 total, 1 needs attention):

● main                /Users/dev/myproject                  [clean] [↑1]
  feature/auth        /Users/dev/myproject.wt/feature-auth  [dirty] [↑2] ⚠️
  feature/api         /Users/dev/myproject.wt/feature-api   [clean] [merged] 💡
  hotfix/bug-123      /Users/dev/myproject.wt/hotfix-bug-123 [clean] [↑1 ↓0]
```

**User sees immediately:**
- Currently in main worktree
- feature/auth needs attention (dirty)
- feature/api can be pruned (merged)
- All worktrees' sync status with remote

### Scenario 2: Scripting (Plain Output)

```bash
$ wt --plain | while IFS=$'\t' read -r branch path; do
    echo "Checking $branch..."
    cd "$path" && npm test
done
```

### Scenario 3: Before Pruning

```bash
$ wt status
# See which worktrees have [merged] indicator
# Verify they're safe to remove
$ wt prune
```

---

## Why This is Priority #3

### Impact: 8/10

- Significantly improves visibility into worktree state
- Reduces manual `git status` checking
- Makes prune decisions more informed
- Helps identify what needs attention

### Pragmatism: 7/10

- Moderate implementation: ~6 hours total
- Leverages existing dependencies (lipgloss)
- Git operations are well-understood
- Main complexity: aggregating multiple git queries

### Risk: 3/10

- Low risk of breakage (rich output is additive)
- Git parsing is battle-tested
- Performance concern: multiple git operations per worktree
- Mitigation: Add `--plain` for fast, simple output

### User Adoption: 8/10

- Immediately noticeable improvement
- Visual appeal makes tool feel modern
- Plain output preserves scriptability
- Optional `status` command for discoverability

### Confidence: 90%

- Git operations are well-understood
- Lipgloss is mature and widely used
- Main uncertainty: performance with many worktrees (>20)
- Mitigation: Could add caching or parallel queries

---

## User Perception

Users will perceive this as:

- "Wow, this looks professional and modern"
- "Finally I can see what's going on at a glance"
- "This feels like a real worktree manager, not just a wrapper"
- "The attention to detail is impressive"

This elevates `wt` from "useful CLI" to "polished, production-ready tool."

---

## Alternatives Considered

### Alternative 1: Separate `wt status` command only

- ❌ Rejected: Discoverability issue, users won't find it
- ✅ Modified: Rich status by default, `--plain` for scripts

### Alternative 2: Use fzf for interactive selection

- ❌ Rejected: PRD explicitly excludes fzf dependency (section 1.2)
- ❌ Adds external dependency
- ✅ Current solution: Static rich output, scriptable

---

## Success Metrics

### Quantitative

- Reduced "how do I check status?" questions
- User screenshots shared on social media
- Adoption in team workflows

### Qualitative

- User feedback: "This looks great"
- Feeling of professional quality
- Increased daily usage

---

## Related Proposals

- [Proposal 1: Shell cd integration](01-shell-cd-integration.md) - Workflow enhancement
- [Proposal 2: wt run command](02-wt-run-command.md) - Automation improvement

---

**Status:** Ready for implementation
**Next Steps:** Implement git helper functions and test
