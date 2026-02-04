✅ **IMPLEMENTED**

This proposal has been implemented. Bash and fish completions are now available via `wt completion bash` and `wt completion fish`. The recommended setup is `eval "$(wt shell-setup)"` which configures both completions and the `wt cd` wrapper.

For usage, see the [Completion Reference](../../user/api-references/completion.md).

---

# Proposal 5: Bash and Fish Shell Completion Support

**Status:** Implemented
**Priority:** SHOULD HAVE
**Effort:** 1-2 hours
**Confidence:** 95%

## Quick Reference

| Aspect | Rating | Notes |
|--------|--------|-------|
| Impact | 6/10 | Removes adoption barrier |
| Pragmatism | 10/10 | Extremely easy: cobra does the work |
| Risk | 1/10 | Isolated feature, minimal risk |
| User Adoption | 8/10 | Completions are expected in modern CLIs |

---

## The Problem

Currently, `wt` only supports zsh completions:

```go
// cmd/wt/main.go:332
ValidArgs: []string{"zsh"},
```

This creates adoption barriers for users of other shells.

### Impact on Bash Users

- **Most Linux distributions default to bash**
- Many CI/CD environments use bash
- Large corporate environments standardize on bash
- Missing completions feel like incomplete tool
- Bash users are the majority on Linux servers

### Impact on Fish Users

- Growing popularity among developers (modern shell)
- Superior completion and suggestion system
- Active community especially on macOS
- Expect all modern tools to support fish
- Fish shows richer suggestions with descriptions

### Overall Impact

- New users hit immediate friction: "completions don't work"
- Tool feels unpolished or abandoned
- Users on non-zsh systems are second-class citizens
- Reduces adoption in bash-heavy environments (most servers)

---

## The Solution

Leverage cobra's built-in completion generation to add bash and fish support. Cobra already has excellent completion support - we just need to expose it.

---

## Implementation

### Update Completion Command

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

### Add Dynamic Branch Completion

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

    // Similar for remove command
    removeCmd.ValidArgsFunction = rootCmd.ValidArgsFunction

    // If run command is implemented
    // runCmd.ValidArgsFunction = rootCmd.ValidArgsFunction
}
```

---

## Implementation Plan

### Phase 1: Core Implementation (30 minutes)

1. Update `completionCmd` ValidArgs to include bash and fish
2. Add GenBashCompletion and GenFishCompletion calls
3. Update Long description with instructions for all shells

### Phase 2: Dynamic Completions (1 hour)

1. Add ValidArgsFunction to root command
2. Add ValidArgsFunction to exec command
3. Add ValidArgsFunction to remove command
4. Test completions with various branch names

### Phase 3: Testing (30 minutes)

**Test bash completions:**

```bash
source <(wt completion bash)
wt feat<TAB>  # Should complete with feature/ branches
```

**Test fish completions:**

```fish
wt completion fish > ~/.config/fish/completions/wt.fish
wt feat<TAB>  # Should show suggestions
```

**Test zsh completions (ensure still working):**

```zsh
source <(wt completion zsh)
wt feat<TAB>  # Should complete
```

### Phase 4: Distribution (30 minutes)

Update Homebrew formula to install all completion files:

```ruby
def install
  bin.install "wt"

  # Install completions for all shells
  bash_completion.install "completion/bash/wt.bash"
  zsh_completion.install "completion/zsh/_wt"
  fish_completion.install "completion/fish/wt.fish"
end
```

### Phase 5: Documentation (30 minutes)

1. Update `docs/user/guides/quickstart.md` with all three shells
2. Update README installation section
3. Add completion troubleshooting section
4. Update platform compatibility notes

---

## Files to Modify

- `cmd/wt/main.go` - Update completionCmd, add ValidArgsFunction
- `docs/user/guides/quickstart.md` - Add bash/fish instructions
- `README.md` - Update completion documentation
- Homebrew formula (external repository) - Install all completions

---

## Example User Experience

### Bash Completion

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

### Fish Completion

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

**Note:** Fish shows additional context in suggestions - this is a fish feature, not something we need to implement.

### Zsh Completion (Existing)

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

---

## Why This is Priority #5

### Impact: 6/10

- Removes adoption barrier for bash/fish users
- Improves first-time user experience
- Makes tool feel complete and professional
- But: Most users can live without completions initially

### Pragmatism: 10/10

- Extremely easy: cobra does all the work
- ~2 hours total implementation time
- Near-zero risk of breaking anything
- Can be implemented in one sitting

### Risk: 1/10

- Very low risk: completions are isolated feature
- Cobra's completion generation is battle-tested
- Worst case: completions don't work, user types full command
- No impact on core functionality

### User Adoption: 8/10

- Bash users are majority of Linux users
- Fish users are enthusiastic and vocal
- Completions are expected in modern CLIs
- Missing completions is noticed immediately

### Confidence: 95%

- Cobra handles complexity automatically
- Completion generation is well-documented
- Tested by thousands of projects using cobra
- Main uncertainty: dynamic branch completion edge cases

---

## User Perception

### General users will perceive this as

- "They actually support my shell! This tool is for real"
- "Attention to detail - they care about polish"
- "Professional quality, ready for daily use"
- "The developers respect all platforms equally"

### Bash users specifically

- "Finally! Now I can use this on my servers"
- "Completions work just like other tools"
- "This is production-ready"

### Fish users specifically

- "Yes! Fish support!"
- "The suggestions are beautiful"
- "This tool respects modern shells"

This is a **high-impact, low-effort** improvement that removes a common objection: "doesn't support my shell."

---

## Alternatives Considered

### Alternative 1: Maintain custom completions for each shell

- ❌ Rejected: High maintenance burden
- ❌ Reinventing what cobra already does
- ✅ Current solution: Leverage cobra's built-in generators

### Alternative 2: Only support zsh (status quo)

- ❌ Rejected: Creates adoption barrier
- ❌ Bash users are majority on Linux
- ✅ Current solution: Support all major shells

---

## Success Metrics

### Quantitative

- Reduction in "doesn't work in bash" issues
- Increased Linux server adoption
- Download statistics from different platforms

### Qualitative

- User feedback about shell support
- Reduced onboarding friction
- Platform diversity in user base

---

## Platform-Specific Notes

### Bash

- Bash 4.0+ required for some completion features
- Most Linux distributions have bash 4.2+
- macOS ships with bash 3.2 (but users can upgrade)
- Completions work on older bash, just less rich

### Fish

- Fish 3.0+ recommended
- Fish has superior suggestion engine
- Completions are automatically loaded from `~/.config/fish/completions/`
- Fish users will see rich descriptions automatically

### Zsh

- Already supported (custom completion script)
- Zsh 5.0+ recommended
- Most macOS and modern Linux have zsh 5.8+
- Consider migrating to cobra-generated completions for consistency

---

## Future Enhancements

### Potential improvements (not in scope for initial implementation)

1. **Rich completions with descriptions:**

   ```bash
   $ wt <branch><TAB>
   feature/auth     -- Authentication system (created 2d ago)
   feature/api      -- API endpoints (created 1w ago)
   ```

2. **Context-aware completions:**
   - Complete with remote branches when using `--from`
   - Complete with file patterns for copy patterns

3. **Completion for flag values:**
   - Complete shell names for `completion` command
   - Complete branch names for `--from` flag

---

## Testing Checklist

### Bash Testing

- [ ] Basic command completion (`wt <TAB>`)
- [ ] Subcommand completion (`wt e<TAB>` → exec)
- [ ] Branch name completion (`wt feature/<TAB>`)
- [ ] Flag completion (`wt --<TAB>`)
- [ ] Works with bash 4.0+
- [ ] Works with bash 5.0+

### Fish Testing

- [ ] Command completion with descriptions
- [ ] Branch name completion
- [ ] Flag completion
- [ ] File path completion after `--`
- [ ] Completions auto-load from config directory

### Zsh Testing (Regression)

- [ ] Existing completions still work
- [ ] Branch name completion works
- [ ] No regressions from changes

### Cross-shell Testing

- [ ] All shells show same completion options
- [ ] Dynamic branch completion works in all shells
- [ ] Error handling when git commands fail

---

## Documentation Updates Needed

### Quickstart Guide

- Add bash and fish installation sections
- Update screenshots to show all three shells
- Add troubleshooting for each shell

### README

- Update "Installation" section with all shells
- Add platform compatibility matrix
- Link to detailed completion docs

### New Documentation

- Create completion troubleshooting guide
- Document shell-specific features
- Add FAQ for completion issues

---

## Related Proposals

- [Proposal 1: Shell cd integration](01-shell-cd-integration.md) - Also involves shell functions
- All proposals benefit from better shell integration

---

**Status:** Ready for implementation
**Next Steps:** Update completionCmd and test across shells
**Recommendation:** Implement early (Quick Win) - low effort, high polish
