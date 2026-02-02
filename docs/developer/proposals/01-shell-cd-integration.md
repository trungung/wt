⚠️ **NOT IMPLEMENTED**

This document describes a proposed feature that is **not yet implemented** in `wt`. 
For currently available features, see the [API Reference](../../user/api-references/index.md).

---

# Proposal 1: Shell Integration Function for Directory Navigation

**Status:** Proposed
**Priority:** MUST HAVE
**Effort:** 1 hour (documentation only)
**Confidence:** 99%

## Quick Reference

| Aspect | Rating | Notes |
|--------|--------|-------|
| Impact | 10/10 | Eliminates #1 daily friction point |
| Pragmatism | 10/10 | Zero Go code changes needed |
| Risk | 0/10 | Completely opt-in, no breaking changes |
| User Adoption | 10/10 | Immediate obvious value |

---

## The Problem

Currently, `wt <branch>` prints the worktree path to stdout, enabling scripting and piping. However, the most common use case is creating a worktree and immediately navigating into it. The current workflow requires:

```bash
$ wt feature/auth
/Users/dev/repo.wt/feature-auth
$ cd /Users/dev/repo.wt/feature-auth  # Manual copy-paste or retyping
```

This is the **#1 friction point** in daily usage. Unix processes cannot change their parent shell's directory, so this requires manual intervention after every worktree creation.

### Pain Points

- Manual copy-paste of paths is error-prone
- Breaks flow state during development
- Requires memorizing or typing long paths
- Particularly frustrating on macOS with long temp directory paths
- New users expect "just go there" behavior

---

## The Solution

Provide official shell function wrappers that intercept `wt` output and automatically `cd` when the command succeeds and outputs a valid directory path.

### For Zsh

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

### For Bash

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

### For Fish

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

---

## Advanced: CLI Command for Shell Integration

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

---

## Implementation Plan

### Phase 1: Documentation (Immediate - 1 hour)

1. Add "Shell Integration" section to `docs/user/guides/quickstart.md`
2. Include copy-paste snippets for zsh, bash, and fish
3. Add to README.md "Quick Start" section
4. Update multi-agent workflow guide with shell integration example

### Phase 2: CLI Command (Optional - 2 hours)

1. Add `shell-init` command to `cmd/wt/main.go`
2. Embed shell scripts as string constants
3. Add tests for output format
4. Update completion command documentation

### Phase 3: Distribution (Homebrew - 1 hour)

1. Update Homebrew formula to suggest shell integration in post-install message
2. Consider automatic shell config detection and prompting

---

## Files to Modify

- `docs/user/guides/quickstart.md` - Add shell integration section
- `README.md` - Add to Quick Start
- `docs/user/guides/multi-agent-workflow.md` - Update examples
- `cmd/wt/main.go` - (Optional) Add shell-init command

---

## Example Usage

### Current Workflow

```bash
$ wt feature/auth
/Users/dev/repo.wt/feature-auth
$ cd /Users/dev/repo.wt/feature-auth
$ ls
```

### With Shell Integration

```bash
$ wt feature/auth
→ /Users/dev/repo.wt/feature-auth
$ ls  # Already in the directory!
```

### Still Script-Friendly

```bash
# Direct command usage preserves scriptability
$ WORKTREE_PATH=$(command wt feature/auth)
$ docker run -v "$WORKTREE_PATH:/app" myimage
```

---

## Why This is Priority #1

### Impact: 10/10

- Eliminates the #1 daily friction point
- Used every single time someone creates or accesses a worktree
- Transforms "useful tool" into "indispensable tool"

### Pragmatism: 10/10

- Zero Go code changes needed for basic version
- Pure shell functions, no binary modifications
- Can be implemented in documentation alone
- Optional CLI command adds convenience but not required

### Risk: 0/10

- No breaking changes whatsoever
- Completely opt-in via shell configuration
- Falls back gracefully if function not installed
- Shell functions are well-understood technology

### User Adoption: 10/10

- Every user will immediately understand the value
- Copy-paste installation takes 10 seconds
- Works with existing muscle memory (still type `wt <branch>`)
- Can be aliased or used as separate command (`wtcd`)

### Confidence: 99%

- Shell functions are battle-tested patterns
- Used by tools like `z`, `autojump`, `direnv`
- No compatibility concerns across Unix shells
- Zero implementation risk

---

## User Perception

Users will perceive this as:

- "Finally! This is how it should have worked from the start"
- "This tool gets me - it removes the annoying parts"
- "Professional quality - they thought about the actual workflow"

This single feature can turn satisfied users into enthusiastic advocates who recommend the tool to others.

---

## Alternatives Considered

### Alternative 1: Add special flag `wt --cd <branch>`

- ❌ Rejected: Still can't change parent shell directory
- ❌ Requires wrapper function anyway
- ✅ Current solution: Document shell function pattern

### Alternative 2: Generate separate `wtcd` binary

- ❌ Rejected: Extra binary to maintain and install
- ❌ Confusing to have two commands
- ✅ Current solution: Single `wt` with optional alias

---

## Success Metrics

### Quantitative

- Reduction in GitHub issues asking "how do I cd?"
- User testimonials mentioning workflow improvement
- Social media mentions of "game changer" or similar

### Qualitative

- User feedback: "This is exactly what I needed"
- Reduced friction in onboarding new users
- Increased daily usage frequency

---

## Related Proposals

- [Proposal 2: wt run command](02-wt-run-command.md) - Another workflow enhancement
- [Proposal 5: Bash/Fish completions](05-bash-fish-completions.md) - Shell support

---

**Status:** Ready for implementation
**Next Steps:** Add documentation to quickstart guide
