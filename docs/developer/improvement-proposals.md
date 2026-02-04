# Improvement Proposals for `wt`

**Document Version:** 1.0
**Date:** 2026-01-10
**Status:** Proposed

> **Note:** This document provides a comprehensive overview of all improvement proposals. For detailed, individual proposals, see the [proposals directory](proposals/) which contains one document per feature.

## Quick Navigation

- [Proposal 1: Shell cd Integration](proposals/01-shell-cd-integration.md)
- [Proposal 2: wt run Command](proposals/02-wt-run-command.md)
- [Proposal 3: Rich Status Display](proposals/03-rich-status-display.md)
- [Proposal 4: Lifecycle Hooks](proposals/04-lifecycle-hooks.md)
- [Proposal 5: Bash/Fish Completions](proposals/05-bash-fish-completions.md)
- [Proposals Index](proposals/README.md) - Implementation guide and overview

---

## Executive Summary

This document presents the top 5 improvement proposals for `wt`, a git worktree manager. These proposals emerged from a comprehensive analysis of the codebase, documentation, and user workflows. Each proposal has been evaluated for:

- **Impact** - How much it improves the user experience
- **Pragmatism** - Feasibility and implementation complexity
- **Accretive value** - Adds functionality without breaking existing features
- **User perception** - How users will perceive and adopt the change

All proposals are independently implementable and require no breaking changes to existing functionality.

### Quick Reference

| Rank | Feature | Impact | Effort | Confidence | Priority |
|------|---------|--------|--------|------------|----------|
| 1 | Shell cd integration | Very High | 1h | 99% | Must Have |
| 2 | `wt run` command | High | 2-3h | 95% | Should Have |
| 3 | Rich status display | High | 4-6h | 90% | Should Have |
| 4 | Lifecycle hooks | Medium-High | 6-8h | 85% | Could Have |
| 5 | Bash/Fish completions | Medium | 1-2h | 95% | Should Have |

---

## Proposal 1: Shell Integration Function for Directory Navigation

### Priority: MUST HAVE

### The Problem

Currently, `wt <branch>` prints the worktree path to stdout, enabling scripting and piping. However, the most common use case is creating a worktree and immediately navigating into it. The current workflow requires:

```bash
$ wt feature/auth
/Users/dev/repo.wt/feature-auth
$ cd /Users/dev/repo.wt/feature-auth  # Manual copy-paste or retyping
```

This is the **#1 friction point** in daily usage. Unix processes cannot change their parent shell's directory, so this requires manual intervention after every worktree creation.

**Pain Points:**

- Manual copy-paste of paths is error-prone
- Breaks flow state during development
- Requires memorizing or typing long paths
- Particularly frustrating on macOS with long temp directory paths
- New users expect "just go there" behavior

### The Solution

Provide official shell function wrappers that intercept `wt` output and automatically `cd` when the command succeeds and outputs a valid directory path.

#### For Zsh

```zsh
# Add to ~/.zshrc
wtcd() {
    local result
    result=$(command wt "$@")
    local exit_code=$?

    # If wt succeeded and output looks like a directory path
    if [[ $exit_code -eq 0 && -d "$result" ]]; then
        cd "$result" || return 1
        echo "→ $result"
    else
        echo "$result"
    fi

    return $exit_code
}

# Optional: Replace wt entirely
alias wt=wtcd
```

#### For Bash

```bash
# Add to ~/.bashrc
wtcd() {
    local result
    result=$(command wt "$@")
    local exit_code=$?

    # If wt succeeded and output looks like a directory path
    if [[ $exit_code -eq 0 && -d "$result" ]]; then
        cd "$result" || return 1
        echo "→ $result"
    else
        echo "$result"
    fi

    return $exit_code
}

# Optional: Replace wt entirely
alias wt=wtcd
```

#### For Fish

```fish
# Add to ~/.config/fish/functions/wt.fish
function wt
    set result (command wt $argv)
    set exit_code $status

    # If wt succeeded and output looks like a directory path
    if test $exit_code -eq 0 -a -d "$result"
        cd "$result"; or return 1
        echo "→ $result"
    else
        echo "$result"
    end

    return $exit_code
end
```

### Advanced: CLI Command for Shell Integration

Add a `wt shell-init` command that outputs the appropriate function:

```go
var shellInitCmd = &cobra.Command{
    Use:   "shell-init [shell]",
    Short: "output shell integration code for cd functionality",
    Long: `Outputs shell function code that enables automatic cd when creating worktrees.

Usage:
  # Zsh
  eval "$(wt shell-init zsh)"

  # Bash
  eval "$(wt shell-init bash)"

  # Fish
  wt shell-init fish | source

To make permanent, add to your shell config:
  echo 'eval "$(wt shell-init zsh)"' >> ~/.zshrc
`,
    Args:      cobra.ExactArgs(1),
    ValidArgs: []string{"bash", "zsh", "fish"},
    RunE: func(cmd *cobra.Command, args []string) error {
        shell := args[0]
        script, ok := shellInitScripts[shell]
        if !ok {
            return fmt.Errorf("unsupported shell: %s", shell)
        }
        fmt.Print(script)
        return nil
    },
}

var shellInitScripts = map[string]string{
    "zsh": `wtcd() {
    local result
    result=$(command wt "$@")
    local exit_code=$?
    if [[ $exit_code -eq 0 && -d "$result" ]]; then
        cd "$result" || return 1
        echo "→ $result"
    else
        echo "$result"
    fi
    return $exit_code
}
alias wt=wtcd
`,
    "bash": `wtcd() {
    local result
    result=$(command wt "$@")
    local exit_code=$?
    if [[ $exit_code -eq 0 && -d "$result" ]]; then
        cd "$result" || return 1
        echo "→ $result"
    else
        echo "$result"
    fi
    return $exit_code
}
alias wt=wtcd
`,
    "fish": `function wt
    set result (command wt $argv)
    set exit_code $status
    if test $exit_code -eq 0 -a -d "$result"
        cd "$result"; or return 1
        echo "→ $result"
    else
        echo "$result"
    end
    return $exit_code
end
`,
}

func init() {
    rootCmd.AddCommand(shellInitCmd)
}
```

### Implementation Plan

#### Phase 1: Documentation (Immediate - 1 hour)

1. Add "Shell Integration" section to `docs/user/guides/quickstart.md`
2. Include copy-paste snippets for zsh, bash, and fish
3. Add to README.md "Quick Start" section
4. Update multi-agent workflow guide with shell integration example

#### Phase 2: CLI Command (Optional - 2 hours)

1. Add `shell-init` command to `cmd/wt/main.go`
2. Embed shell scripts as string constants
3. Add tests for output format
4. Update completion command documentation

#### Phase 3: Distribution (Homebrew - 1 hour)

1. Update Homebrew formula to suggest shell integration in post-install message
2. Consider automatic shell config detection and prompting

### Files to Modify

- `docs/user/guides/quickstart.md` - Add shell integration section
- `README.md` - Add to Quick Start
- `docs/user/guides/multi-agent-workflow.md` - Update examples
- `cmd/wt/main.go` - (Optional) Add shell-init command

### Example Usage

**Current Workflow:**

```bash
$ wt feature/auth
/Users/dev/repo.wt/feature-auth
$ cd /Users/dev/repo.wt/feature-auth
$ ls
```

**With Shell Integration:**

```bash
$ wt feature/auth
→ /Users/dev/repo.wt/feature-auth
$ ls  # Already in the directory!
```

**Still Script-Friendly:**

```bash
# Direct command usage preserves scriptability
$ WORKTREE_PATH=$(command wt feature/auth)
$ docker run -v "$WORKTREE_PATH:/app" myimage
```

### Why This is Priority #1

**Impact: 10/10**

- Eliminates the #1 daily friction point
- Used every single time someone creates or accesses a worktree
- Transforms "useful tool" into "indispensable tool"

**Pragmatism: 10/10**

- Zero Go code changes needed for basic version
- Pure shell functions, no binary modifications
- Can be implemented in documentation alone
- Optional CLI command adds convenience but not required

**Risk: 0/10**

- No breaking changes whatsoever
- Completely opt-in via shell configuration
- Falls back gracefully if function not installed
- Shell functions are well-understood technology

**User Adoption: 10/10**

- Every user will immediately understand the value
- Copy-paste installation takes 10 seconds
- Works with existing muscle memory (still type `wt <branch>`)
- Can be aliased or used as separate command (`wtcd`)

**Confidence: 99%**

- Shell functions are battle-tested patterns
- Used by tools like `z`, `autojump`, `direnv`
- No compatibility concerns across Unix shells
- Zero implementation risk

### User Perception

Users will perceive this as:

- "Finally! This is how it should have worked from the start"
- "This tool gets me - it removes the annoying parts"
- "Professional quality - they thought about the actual workflow"

This single feature can turn satisfied users into enthusiastic advocates who recommend the tool to others.

---

## Proposal 2: `wt run` Command (Ensure + Exec Combined)

### Priority: SHOULD HAVE

### The Problem

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

**Current Implementation Issue:**

```go
// cmd/wt/main.go:83-86
path, err := core.FindWorktree(branch)
if err != nil {
    return err
}
```

The `exec` command will fail if the worktree doesn't exist, requiring users to manually ensure it first.

### The Solution

Add a `wt run` command that combines ensure + exec in one operation:

```bash
# New: Idempotent, "just works"
$ wt run feature/test -- npm test

# Creates worktree if needed, then runs command
# If worktree exists, just runs command
```

### Implementation

#### Add Command to main.go

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

### Implementation Plan

#### Phase 1: Core Implementation (2 hours)

1. Add `runCmd` to `cmd/wt/main.go` (30-40 lines)
2. Wire up `--from` flag (already exists, reuse logic)
3. Test manually with various scenarios

#### Phase 2: Testing (1 hour)

1. Add integration test in `test/integration_test.go`:

   ```go
   t.Run("run command creates and executes", func(t *testing.T) {
       // Test that 'wt run' creates worktree if missing
       // Test that 'wt run' executes in existing worktree
       // Test that command exit codes propagate correctly
   })
   ```

#### Phase 3: Documentation (1 hour)

1. Create `docs/user/api-references/run.md`
2. Update `docs/user/api-references/index.md` to include run
3. Add examples to `docs/user/guides/multi-agent-workflow.md`
4. Update README.md quick reference table

### Files to Modify

- `cmd/wt/main.go` - Add runCmd (~40 lines)
- `test/integration_test.go` - Add test cases
- `docs/user/api-references/run.md` - New file
- `docs/user/api-references/index.md` - Add entry
- `docs/user/guides/multi-agent-workflow.md` - Update examples
- `README.md` - Update command table

### Example Usage Scenarios

#### Scenario 1: Single Command Testing

```bash
# Current (2 commands, need to think):
$ wt feature/api-v2                    # Ensure exists
$ wt exec feature/api-v2 -- npm test   # Then run

# New (1 command, idempotent):
$ wt run feature/api-v2 -- npm test    # Just works!
```

#### Scenario 2: Parallel Multi-Agent Builds

```bash
# Run builds in parallel for multiple features
$ wt run feature/auth -- npm run build &
$ wt run feature/payments -- npm run build &
$ wt run feature/api -- npm run build &
$ wait

# Worktrees created if needed, all builds run in isolation
```

#### Scenario 3: CI/CD Pipeline

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

#### Scenario 4: One-liner Scripts

```bash
#!/bin/bash
# test-all-features.sh
for branch in $(git branch --format='%(refname:short)' | grep '^feature/'); do
    echo "Testing $branch..."
    wt run "$branch" -- npm test || echo "FAILED: $branch"
done
```

### Why This is Priority #2

**Impact: 9/10**

- Reduces cognitive load significantly
- Perfect for automation and scripting
- Makes the tool more intuitive for new users
- Enables powerful multi-agent workflows

**Pragmatism: 10/10**

- Only ~40 lines of code
- Combines two existing, battle-tested functions
- No changes to core business logic
- Implementation is straightforward

**Risk: 1/10**

- Very low risk: just combining existing functions
- EnsureWorktree is already well-tested
- Command execution logic copied from existing exec command
- Main risk is ensuring flag compatibility

**User Adoption: 9/10**

- Immediately obvious value proposition
- Reduces "do I need to create first?" mental overhead
- Scripts become simpler and more reliable
- Natural evolution of the tool's interface

**Confidence: 95%**

- Straightforward implementation
- Leverages existing, tested code paths
- Similar patterns exist in other tools
- Minor risk: edge cases with flag inheritance

### User Perception

Users will perceive this as:

- "This is exactly what I wanted from `exec`"
- "Makes scripting so much easier"
- "The tool is reading my mind - just do what I mean"
- "Production-ready for CI/CD"

This positions `wt` as a tool that understands real workflows, not just individual operations.

---

## Proposal 3: Rich Status Display with Git Awareness

### Priority: SHOULD HAVE

### The Problem

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

**Pain Points:**

- Managing 5+ worktrees requires manual inspection of each
- No visibility into "what needs attention"
- Users run `git status` manually in each worktree
- No way to quickly identify stale or merged worktrees
- Prune candidates not visible until running `wt prune --dry-run`

### The Solution

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

**Plain Output Mode (for scripts):**

```bash
$ wt --plain
# or: wt status --plain
main    /Users/dev/repo
feature/auth    /Users/dev/repo.wt/feature-auth
feature/api-v2  /Users/dev/repo.wt/feature-api-v2
feature/payments    /Users/dev/repo.wt/feature-payments
hotfix/security-patch   /Users/dev/repo.wt/hotfix-security-patch
```

### Implementation

#### Phase 1: Status Data Structure

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

#### Phase 2: Git Helper Functions

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

#### Phase 3: Rich Output with Lipgloss

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

#### Phase 4: Update Root Command

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

### Implementation Plan

#### Phase 1: Git Status Functions (2 hours)

1. Add `GetAheadBehind` to `internal/git/git.go`
2. Add `GetLastCommitTime` to `internal/git/git.go`
3. Test git functions in isolation

#### Phase 2: Core Status Logic (2 hours)

1. Add `WorktreeStatus` struct to `internal/core/core.go`
2. Implement `GetWorktreeStatuses` function
3. Test with various worktree configurations

#### Phase 3: Rich UI Rendering (2 hours)

1. Add `RenderWorktreeStatus` to `internal/ui/ui.go`
2. Use lipgloss for colors (already in dependencies)
3. Respect NO_COLOR environment variable
4. Test color output across terminals

#### Phase 4: CLI Integration (1 hour)

1. Update root command to use rich output by default
2. Add `--plain` flag for script compatibility
3. Add optional `wt status` alias command
4. Test backwards compatibility

#### Phase 5: Documentation (1 hour)

1. Create `docs/user/api-references/status.md`
2. Update quickstart guide with status examples
3. Add screenshots to documentation
4. Update README with new output format

### Files to Modify

- `internal/core/core.go` - Add WorktreeStatus, GetWorktreeStatuses
- `internal/git/git.go` - Add GetAheadBehind, GetLastCommitTime
- `internal/ui/ui.go` - Add RenderWorktreeStatus
- `cmd/wt/main.go` - Update root command, add status command
- `docs/user/api-references/status.md` - New file
- `docs/user/api-references/index.md` - Update
- `docs/user/guides/quickstart.md` - Add status examples

### Example Usage Scenarios

#### Scenario 1: Daily Status Check

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

#### Scenario 2: Scripting (Plain Output)

```bash
$ wt --plain | while IFS=$'\t' read -r branch path; do
    echo "Checking $branch..."
    cd "$path" && npm test
done
```

#### Scenario 3: Before Pruning

```bash
$ wt status
# See which worktrees have [merged] indicator
# Verify they're safe to remove
$ wt prune
```

### Why This is Priority #3

**Impact: 8/10**

- Significantly improves visibility into worktree state
- Reduces manual `git status` checking
- Makes prune decisions more informed
- Helps identify what needs attention

**Pragmatism: 7/10**

- Moderate implementation: ~6 hours total
- Leverages existing dependencies (lipgloss)
- Git operations are well-understood
- Main complexity: aggregating multiple git queries

**Risk: 3/10**

- Low risk of breakage (rich output is additive)
- Git parsing is battle-tested
- Performance concern: multiple git operations per worktree
- Mitigation: Add `--plain` for fast, simple output

**User Adoption: 8/10**

- Immediately noticeable improvement
- Visual appeal makes tool feel modern
- Plain output preserves scriptability
- Optional `status` command for discoverability

**Confidence: 90%**

- Git operations are well-understood
- Lipgloss is mature and widely used
- Main uncertainty: performance with many worktrees (>20)
- Mitigation: Could add caching or parallel queries

### User Perception

Users will perceive this as:

- "Wow, this looks professional and modern"
- "Finally I can see what's going on at a glance"
- "This feels like a real worktree manager, not just a wrapper"
- "The attention to detail is impressive"

This elevates `wt` from "useful CLI" to "polished, production-ready tool."

---

## Proposal 4: Lifecycle Hooks System

### Priority: COULD HAVE

### The Problem

Different teams and projects have specialized workflows beyond `wt`'s built-in capabilities:

**Common Use Cases:**

- Send Slack notifications when worktrees are created
- Update project management systems (Jira, Linear)
- Validate branch naming conventions before creation
- Check for open pull requests before removal
- Clean up Docker containers associated with worktrees
- Update database entries tracking development environments
- Run security scans before worktree creation
- Verify dependencies before executing commands

**Current Limitation:**
The only extension point is `postCreateCmd`, which:

- Only runs after creation (no pre-creation validation)
- Cannot prevent creation on validation failure
- Has no hooks for remove/prune operations
- Cannot access context about the operation

This forces teams to:

- Wrap `wt` with custom scripts
- Fork the project to add custom logic
- Accept reduced automation capabilities

### The Solution

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

### Hook Types and Behavior

#### Pre-hooks (Validation)

- Execute **before** the operation
- Non-zero exit code **aborts** the operation
- Can access operation context via environment variables
- Output goes to stderr
- Use cases: validation, permissions checks, external system queries

#### Post-hooks (Notification/Cleanup)

- Execute **after** the operation completes successfully
- Errors are logged but don't fail the operation
- Can access result context via environment variables
- Output goes to stdout/stderr normally
- Use cases: notifications, cleanup, external system updates

### Environment Variables Available to Hooks

All hooks receive:

```bash
WT_HOOK              # Hook name (pre-create, post-create, etc.)
WT_BRANCH            # Branch name
WT_REPO_ROOT         # Repository root path
WT_WORKTREE_PATH     # Worktree path (empty for pre-create)
WT_BASE_BRANCH       # Base branch (for create operations with --from)
WT_OPERATION         # Operation type (create, remove, prune)
```

Additional per-hook:

```bash
# pre-create / post-create
WT_IS_NEW_BRANCH     # "true" if creating new branch, "false" if checking out existing

# pre-prune / post-prune
WT_PRUNE_CANDIDATES  # Comma-separated list of branches to be pruned
WT_PRUNE_COUNT       # Number of worktrees being pruned
```

### Implementation

#### Phase 1: Config Schema

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

#### Phase 2: Hook Execution Engine

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
```

#### Phase 3: Integrate Hooks into Operations

Update `internal/core/core.go`:

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

// Similar integration in RemoveWorktree and PruneWorktrees
```

#### Phase 4: CLI Flags

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

### Implementation Plan

#### Phase 1: Config and Engine (3 hours)

1. Extend Config struct with Hooks
2. Implement ExecuteHooks function
3. Add timeout handling
4. Test hook execution in isolation

#### Phase 2: Integration (3 hours)

1. Integrate pre/post-create hooks into EnsureWorktree
2. Integrate pre/post-remove hooks into RemoveWorktree
3. Integrate pre/post-prune hooks into PruneWorktrees
4. Maintain backwards compatibility with postCreateCmd

#### Phase 3: CLI and Safety (2 hours)

1. Add `--skip-hooks` flag to all commands
2. Add environment variable support (WT_SKIP_HOOKS)
3. Test hook abortion scenarios
4. Test timeout scenarios

#### Phase 4: Documentation (2 hours)

1. Create `docs/user/guides/hooks.md` with examples
2. Update configuration reference
3. Add security warnings
4. Create example hook scripts

#### Phase 5: Examples and Testing (2 hours)

1. Create example hooks in `examples/hooks/`
2. Add integration tests for hook execution
3. Test pre-hook abortion
4. Test post-hook error handling

### Files to Modify/Create

- `internal/config/config.go` - Add Hooks struct
- `internal/core/hooks.go` - **New file** - Hook execution engine
- `internal/core/core.go` - Integrate hooks into operations
- `cmd/wt/main.go` - Add --skip-hooks flag
- `docs/user/guides/hooks.md` - **New file** - Hook documentation
- `docs/user/api-references/configuration.md` - Update with hooks
- `examples/hooks/` - **New directory** - Example hook scripts

### Example Hook Scripts

#### Pre-create: Validate Branch Name

```bash
#!/bin/bash
# .wt/hooks/validate-branch-name.sh

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

#### Post-create: Notify Slack

```bash
#!/bin/bash
# .wt/hooks/notify-slack.sh

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

#### Pre-remove: Check Open PRs

```bash
#!/bin/bash
# .wt/hooks/check-open-prs.sh

BRANCH="$WT_BRANCH"

# Check if PR exists using gh CLI
if gh pr view "$BRANCH" --json state --jq '.state' | grep -q "OPEN"; then
    echo "Error: Branch $BRANCH has an open PR"
    echo "Please close or merge the PR before removing the worktree"
    exit 1
fi

exit 0
```

#### Post-prune: Update Dashboard

```bash
#!/bin/bash
# .wt/hooks/update-dashboard.sh

API_URL="https://dev-dashboard.company.com/api/worktrees"
PRUNED_BRANCHES="$WT_PRUNE_CANDIDATES"

curl -X POST "$API_URL/prune" \
    -H "Content-Type: application/json" \
    -d "{\"branches\": \"$PRUNED_BRANCHES\", \"count\": $WT_PRUNE_COUNT}"
```

### Usage Examples

#### Basic Configuration

```json
{
  "hooks": {
    "pre-create": ["./scripts/validate.sh"],
    "post-create": ["npm install"],
    "post-remove": ["./scripts/cleanup.sh $WT_BRANCH"]
  }
}
```

#### Bypassing Hooks

```bash
# Skip all hooks in emergency
$ wt remove feature/broken --skip-hooks

# Or via environment variable
$ WT_SKIP_HOOKS=1 wt remove feature/broken
```

#### Debugging Hooks

```bash
# Hooks print their output to stderr
$ wt feature/test
[pre-create] ✓ Branch name is valid
[pre-create] ✓ Disk space sufficient
/Users/dev/repo.wt/feature-test
[post-create] Installing dependencies...
[post-create] ✓ Slack notification sent
```

### Why This is Priority #4

**Impact: 7/10**

- Unlocks powerful customization
- Enables enterprise workflows
- Future-proofs the tool
- But: Only needed by power users and teams

**Pragmatism: 6/10**

- Moderate complexity (~12 hours implementation)
- Requires careful error handling
- Need to handle timeouts and security
- Good pattern to follow (Git hooks, npm scripts)

**Risk: 4/10**

- Medium risk: executing arbitrary user code
- Timeout handling adds complexity
- Backwards compatibility with postCreateCmd
- Security: hooks run with user's permissions

**User Adoption: 6/10**

- Power users will love it
- Most users won't use it initially
- Requires scripting knowledge
- Discoverability: needs good documentation

**Confidence: 85%**

- Command execution is well-understood
- Main uncertainty: edge cases in timeout handling
- Risk: users writing complex hooks that break
- Mitigation: Good documentation and examples

### User Perception

Power users and teams will perceive this as:

- "Now it's a real platform, not just a tool"
- "We can integrate it into our workflows"
- "This is production-ready and enterprise-friendly"
- "The developers thought about extensibility"

Casual users will perceive it as:

- "Nice to have the option if I need it"
- "Shows the tool is mature and well-designed"

This feature positions `wt` as a **platform** rather than just a utility, opening doors for community contributions of hook libraries.

---

## Proposal 5: Bash and Fish Shell Completion Support

### Priority: SHOULD HAVE

### The Problem

Currently, `wt` only supports zsh completions:

```go
// cmd/wt/main.go:332
ValidArgs: []string{"zsh"},
```

This creates adoption barriers:

**Bash Users:**

- Most Linux distributions default to bash
- Many CI/CD environments use bash
- Large corporate environments standardize on bash
- Missing completions feel like incomplete tool

**Fish Users:**

- Growing popularity among developers (modern shell)
- Superior completion and suggestion system
- Active community especially on macOS
- Expect all modern tools to support fish

**Impact:**

- New users hit immediate friction: "completions don't work"
- Tool feels unpolished or abandoned
- Users on non-zsh systems are second-class citizens
- Reduces adoption in bash-heavy environments (most servers)

### The Solution

Leverage cobra's built-in completion generation to add bash and fish support. Cobra already has excellent completion support - we just need to expose it.

#### Update Completion Command

Modify `cmd/wt/main.go`:

```go
var completionCmd = &cobra.Command{
    Use:   "completion [shell]",
    Short: "generate completion script for the specified shell",
    Long: `Generate completion script for wt.

To load completions:

Bash:
  # Load in current session
  source <(wt completion bash)

  # Load permanently (add to ~/.bashrc)
  echo 'source <(wt completion bash)' >> ~/.bashrc

  # Alternative: Save to file
  wt completion bash > ~/.wt-completion.bash
  echo 'source ~/.wt-completion.bash' >> ~/.bashrc

Zsh:
  # Load in current session
  source <(wt completion zsh)

  # Load permanently (add to ~/.zshrc)
  echo 'source <(wt completion zsh)' >> ~/.zshrc

  # Alternative: Install to completions directory
  mkdir -p ~/.zsh/completions
  wt completion zsh > ~/.zsh/completions/_wt
  # Add to ~/.zshrc:
  fpath=(~/.zsh/completions $fpath)
  autoload -Uz compinit && compinit

Fish:
  # Fish completions are installed to a specific directory
  wt completion fish > ~/.config/fish/completions/wt.fish

  # Completions are loaded automatically on next shell start

Note: You may need to restart your shell after installation.
`,
    Args:      cobra.ExactArgs(1),
    ValidArgs: []string{"bash", "zsh", "fish"},
    RunE: func(cmd *cobra.Command, args []string) error {
        switch args[0] {
        case "bash":
            return rootCmd.GenBashCompletion(os.Stdout)
        case "zsh":
            // Keep custom zsh script for now (can migrate to cobra later)
            fmt.Print(zshCompletionScript)
            return nil
        case "fish":
            return rootCmd.GenFishCompletion(os.Stdout, true)
        default:
            return fmt.Errorf("unsupported shell: %s", args[0])
        }
        return nil
    },
}
```

#### Add Dynamic Branch Completion

Enhance completions with branch name suggestions:

```go
func init() {
    // Add dynamic completion for branch names
    rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        if len(args) != 0 {
            return nil, cobra.ShellCompDirectiveNoFileComp
        }

        // Get existing worktree branches
        worktrees, err := git.ListWorktrees()
        if err != nil {
            return nil, cobra.ShellCompDirectiveNoFileComp
        }

        var branches []string
        for _, wt := range worktrees {
            // Skip detached and filter by prefix
            if wt.Branch == "(detached)" {
                continue
            }
            if strings.HasPrefix(wt.Branch, toComplete) {
                branches = append(branches, wt.Branch)
            }
        }

        return branches, cobra.ShellCompDirectiveNoFileComp
    }

    // Add completion for exec command
    execCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        if len(args) != 0 {
            return nil, cobra.ShellCompDirectiveDefault
        }

        // Same as root: complete with worktree branches
        worktrees, err := git.ListWorktrees()
        if err != nil {
            return nil, cobra.ShellCompDirectiveNoFileComp
        }

        var branches []string
        for _, wt := range worktrees {
            if wt.Branch != "(detached)" && strings.HasPrefix(wt.Branch, toComplete) {
                branches = append(branches, wt.Branch)
            }
        }

        return branches, cobra.ShellCompDirectiveNoFileComp
    }

    // Similar for remove, run, etc.
    removeCmd.ValidArgsFunction = rootCmd.ValidArgsFunction
    // runCmd.ValidArgsFunction = rootCmd.ValidArgsFunction  // If implemented
}
```

### Implementation Plan

#### Phase 1: Core Implementation (30 minutes)

1. Update `completionCmd` ValidArgs to include bash and fish
2. Add GenBashCompletion and GenFishCompletion calls
3. Update Long description with instructions for all shells

#### Phase 2: Dynamic Completions (1 hour)

1. Add ValidArgsFunction to root command
2. Add ValidArgsFunction to exec command
3. Add ValidArgsFunction to remove command
4. Test completions with various branch names

#### Phase 3: Testing (30 minutes)

1. Test bash completions:

   ```bash
   source <(wt completion bash)
   wt feat<TAB>  # Should complete with feature/ branches
   ```

2. Test fish completions:

   ```fish
   wt completion fish > ~/.config/fish/completions/wt.fish
   wt feat<TAB>  # Should show suggestions
   ```

3. Test zsh completions (ensure still working):

   ```zsh
   source <(wt completion zsh)
   wt feat<TAB>  # Should complete
   ```

#### Phase 4: Distribution (30 minutes)

1. Update Homebrew formula to install all completion files:

   ```ruby
   def install
     bin.install "wt"

     # Install completions for all shells
     bash_completion.install "completion/bash/wt.bash"
     zsh_completion.install "completion/zsh/_wt"
     fish_completion.install "completion/fish/wt.fish"
   end
   ```

2. Update quickstart documentation

#### Phase 5: Documentation (30 minutes)

1. Update `docs/user/guides/quickstart.md` with all three shells
2. Update README installation section
3. Add completion troubleshooting section
4. Update platform compatibility notes

### Files to Modify

- `cmd/wt/main.go` - Update completionCmd, add ValidArgsFunction
- `docs/user/guides/quickstart.md` - Add bash/fish instructions
- `README.md` - Update completion documentation
- Homebrew formula (external repository) - Install all completions

### Example User Experience

#### Bash Completion

```bash
$ wt completion bash > ~/.wt-completion.bash
$ echo 'source ~/.wt-completion.bash' >> ~/.bashrc
$ source ~/.bashrc

$ wt <TAB><TAB>
completion  exec        health      init        prune       remove      --help      --version

$ wt feat<TAB>
feature/auth  feature/api  feature/payments

$ wt exec fe<TAB>
feature/auth  feature/api

$ wt exec feature/auth -- npm<TAB>
npm  npmrc  npm-doctor  npm-install  npm-test  ...
```

#### Fish Completion

```fish
$ wt completion fish > ~/.config/fish/completions/wt.fish
# Restart shell

$ wt <TAB>
completion  (generate completion script)
exec       (run command in worktree)
health     (check project health)
init       (create .wt.config.json)
prune      (remove merged worktrees)
remove     (remove a worktree)

$ wt feat<TAB>
feature/auth      Created 2 days ago, currently working on authentication
feature/api       Created 1 week ago, API endpoints for v2
feature/payments  Created 3 days ago, payment integration
```

Fish shows additional context in suggestions!

#### Zsh Completion (Existing)

```zsh
$ wt <TAB>
completion  -- generate completion script
exec       -- run command in worktree
health     -- check project health
init       -- create .wt.config.json
prune      -- remove merged worktrees
remove     -- remove a worktree

$ wt feat<TAB>
feature/auth  feature/api  feature/payments
```

### Why This is Priority #5

**Impact: 6/10**

- Removes adoption barrier for bash/fish users
- Improves first-time user experience
- Makes tool feel complete and professional
- But: Most users can live without completions

**Pragmatism: 10/10**

- Extremely easy: cobra does all the work
- ~2 hours total implementation time
- Near-zero risk of breaking anything
- Can be implemented in one sitting

**Risk: 1/10**

- Very low risk: completions are isolated feature
- Cobra's completion generation is battle-tested
- Worst case: completions don't work, user types full command
- No impact on core functionality

**User Adoption: 8/10**

- Bash users are majority of Linux users
- Fish users are enthusiastic and vocal
- Completions are expected in modern CLIs
- Missing completions is noticed immediately

**Confidence: 95%**

- Cobra handles complexity automatically
- Completion generation is well-documented
- Tested by thousands of projects using cobra
- Main uncertainty: dynamic branch completion edge cases

### User Perception

Users will perceive this as:

- "They actually support my shell! This tool is for real"
- "Attention to detail - they care about polish"
- "Professional quality, ready for daily use"
- "The developers respect all platforms equally"

**bash users specifically:**

- "Finally! Now I can use this on my servers"
- "Completions work just like other tools"
- "This is production-ready"

**fish users specifically:**

- "Yes! Fish support!"
- "The suggestions are beautiful"
- "This tool respects modern shells"

This is a **high-impact, low-effort** improvement that removes a common objection: "doesn't support my shell."

---

## Implementation Priority Matrix

### Recommended Implementation Order

#### Phase 1: Quick Wins (Week 1)

1. **Shell cd integration** (1 hour) - Highest impact, zero risk
2. **Bash/Fish completions** (2 hours) - Low effort, removes adoption barrier

#### Phase 2: Core Enhancements (Week 2)

3. **`wt run` command** (3 hours) - High value, straightforward implementation
2. **Rich status display** (6 hours) - Visible improvement, moderate complexity

#### Phase 3: Advanced Features (Week 3-4)

5. **Lifecycle hooks** (12 hours) - Power user feature, needs careful design

### Quick Win Justification

Starting with items #1 and #5 provides:

- **Immediate user gratification** - Noticeable improvements in first day
- **Momentum builder** - Quick successes motivate further development
- **Risk mitigation** - Low-risk changes build confidence
- **User feedback** - Get early feedback before investing in complex features

### Total Effort Estimate

- Phase 1 (Quick Wins): 3 hours
- Phase 2 (Core Enhancements): 9 hours
- Phase 3 (Advanced Features): 12 hours
- **Total: 24 hours (~3 workdays)**

---

## Success Metrics

### How to Measure Impact

#### Quantitative Metrics

1. **GitHub Stars/Forks** - Growth rate after implementing features
2. **Installation count** - Homebrew analytics (if available)
3. **Issue reduction** - Fewer "how do I..." questions
4. **Time to first worktree** - Onboarding friction reduction

#### Qualitative Metrics

1. **User feedback** - GitHub discussions, issues, Twitter mentions
2. **Bug reports** - Quality and nature of reported issues
3. **Community contributions** - External PRs and hook libraries
4. **Documentation searches** - What users look for most

#### Feature-Specific Metrics

**Shell Integration (#1):**

- Reduction in "how to cd?" questions
- User testimonials about workflow improvement

**wt run (#2):**

- Usage analytics (if implemented)
- Reduction in multi-step workflow issues

**Rich Status (#3):**

- User screenshots shared on social media
- Reduction in "how do I check?" questions

**Hooks (#4):**

- Community hook libraries created
- Enterprise adoption mentions

**Completions (#5):**

- Reduction in "doesn't work in bash" issues
- Increased Linux server adoption

---

## Risk Assessment and Mitigation

### Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Shell integration breaks scripts | Low | Low | Document as optional, provide clear instructions |
| Rich status performance issues | Medium | Medium | Add --plain flag, implement caching if needed |
| Hook execution security concerns | Medium | High | Document security model, add timeouts, provide --skip-hooks |
| Completion compatibility issues | Low | Low | Test across shell versions, cobra is battle-tested |
| wt run flag conflicts | Low | Medium | Careful flag design, comprehensive testing |

### User Experience Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Feature overload | Low | Medium | Good documentation, progressive disclosure |
| Breaking changes | Very Low | High | All features are additive, no breaking changes |
| Confusion about hooks | Medium | Low | Clear documentation, examples, security warnings |
| Color output issues | Medium | Low | Respect NO_COLOR, provide --plain flag |

---

## Alternatives Considered

### For Shell Integration (#1)

**Alternative 1: Add special flag `wt --cd <branch>`**

- ❌ Rejected: Still can't change parent shell directory
- ❌ Requires wrapper function anyway
- ✅ Current solution: Document shell function pattern

**Alternative 2: Generate separate `wtcd` binary**

- ❌ Rejected: Extra binary to maintain and install
- ❌ Confusing to have two commands
- ✅ Current solution: Single `wt` with optional alias

### For wt run (#2)

**Alternative 1: Make `exec` auto-ensure**

- ❌ Rejected: Breaking change, violates principle of least surprise
- ❌ Users expect `exec` to fail if worktree missing
- ✅ Current solution: New command, backwards compatible

**Alternative 2: Add `wt <branch> <cmd>` shorthand**

- ❌ Rejected: Parsing ambiguity (branch vs command)
- ❌ PRD explicitly excludes this (section 1.2)
- ✅ Current solution: Explicit `run` subcommand with `--`

### For Rich Status (#3)

**Alternative 1: Separate `wt status` command only**

- ❌ Rejected: Discoverability issue, users won't find it
- ✅ Modified: Rich status by default, `--plain` for scripts

**Alternative 2: Use fzf for interactive selection**

- ❌ Rejected: PRD explicitly excludes fzf dependency (section 1.2)
- ❌ Adds external dependency
- ✅ Current solution: Static rich output, scriptable

### For Hooks (#4)

**Alternative 1: Plugin system with Go plugins**

- ❌ Rejected: Go plugin system is fragile and platform-specific
- ❌ Much more complex implementation
- ✅ Current solution: Shell scripts are universal

**Alternative 2: Webhooks to external services**

- ❌ Rejected: Requires network, adds latency, complex auth
- ❌ Doesn't handle pre-hooks (validation)
- ✅ Current solution: Local scripts, synchronous execution

### For Completions (#5)

**Alternative 1: Maintain custom completions for each shell**

- ❌ Rejected: High maintenance burden
- ❌ Reinventing what cobra already does
- ✅ Current solution: Leverage cobra's built-in generators

**Alternative 2: Only support zsh (status quo)**

- ❌ Rejected: Creates adoption barrier
- ❌ Bash users are majority on Linux
- ✅ Current solution: Support all major shells

---

## Future Considerations (Beyond These 5)

Features that didn't make the top 5 but are worth noting:

### 6. Parallel Execution Across Worktrees

```bash
wt exec --all -- npm test  # Run tests in all worktrees in parallel
```

### 7. Worktree Templates

```bash
wt create feature/auth --template node-api  # Apply template
```

### 8. Orphaned Worktree Detection/Repair

```bash
wt health --repair  # Auto-fix orphaned worktrees
```

### 9. Tmux/Terminal Integration

```bash
wt tmux feature/auth  # Create worktree and tmux session
```

### 10. Interactive TUI Mode

```bash
wt ui  # Launch terminal UI for managing worktrees
```

These can be considered for future iterations based on user feedback and demand.

---

## Conclusion

These 5 proposals represent the highest-value improvements to `wt` that:

1. **Solve real user pain points** - Based on workflow analysis
2. **Are obviously accretive** - No breaking changes, pure additions
3. **Are pragmatically implementable** - Total ~24 hours of work
4. **Have high confidence** - Low technical risk
5. **Enhance user perception** - Transform good tool into great tool

### Recommended Action Plan

1. **Start with Quick Wins** (#1, #5) - Build momentum, 3 hours total
2. **Gather feedback** - See what users think
3. **Implement Core Enhancements** (#2, #3) - 9 hours total
4. **Evaluate demand** - Do users need hooks?
5. **Consider Hooks** (#4) - If power users request it

### Expected Outcome

Implementing all 5 proposals would:

- Dramatically improve daily workflow ergonomics
- Remove adoption barriers (shell support)
- Provide visibility into worktree state
- Enable power users and enterprise workflows
- Position `wt` as production-ready, polished tool

**From "useful worktree tool" to "indispensable development workflow manager."**

---

**Document End**
