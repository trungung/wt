# Improvement Proposals

This directory contains detailed proposals for improving the `wt` git worktree manager. Each proposal has been carefully analyzed for impact, pragmatism, risk, and user adoption.

## Overview

These proposals emerged from a comprehensive analysis of the codebase, documentation, and user workflows on 2026-01-10. All proposals are independently implementable with no breaking changes to existing functionality.

### Total Effort: ~24 hours (3 workdays)

---

## Proposals by Priority

### Must Have

1. **[Shell cd Integration](01-shell-cd-integration.md)** ✅ IMPLEMENTED
   - **Effort:** 1 hour | **Confidence:** 99% | **Impact:** 10/10
   - Enable automatic directory navigation after worktree creation
   - Eliminates #1 daily friction point
   - Zero Go code changes - pure documentation/shell functions

### Should Have

1. **[wt run Command](02-wt-run-command.md)**
   - **Effort:** 2-3 hours | **Confidence:** 95% | **Impact:** 9/10
   - Combine ensure + exec in one idempotent command
   - Perfect for automation and CI/CD
   - ~40 lines of code leveraging existing functions

2. **[Rich Status Display](03-rich-status-display.md)**
   - **Effort:** 4-6 hours | **Confidence:** 90% | **Impact:** 8/10
   - Show git awareness (dirty, ahead/behind, merged, stale)
   - Dramatically improves visibility
   - Uses existing lipgloss dependency

3. **[Bash and Fish Completions](05-bash-fish-completions.md)** ✅ IMPLEMENTED
   - **Effort:** 1-2 hours | **Confidence:** 95% | **Impact:** 6/10
   - Remove adoption barrier for non-zsh users
   - Cobra handles all the complexity
   - Professional polish

### Could Have

1. **[Lifecycle Hooks System](04-lifecycle-hooks.md)**
   - **Effort:** 6-8 hours | **Confidence:** 85% | **Impact:** 7/10
   - Enable custom pre/post operation logic
   - Powerful for teams and enterprise workflows
   - Positions tool as extensible platform

---

## Quick Reference Table

| # | Feature | Priority | Effort | Risk | Confidence | Impact |
|---|---------|----------|--------|------|------------|--------|
| 1 | Shell cd integration | ✅ Implemented | 1h | None | 99% | 10/10 |
| 2 | wt run command | Should Have | 2-3h | Very Low | 95% | 9/10 |
| 3 | Rich status display | Should Have | 4-6h | Low | 90% | 8/10 |
| 4 | Lifecycle hooks | Could Have | 6-8h | Medium | 85% | 7/10 |
| 5 | Bash/Fish completions | ✅ Implemented | 1-2h | Very Low | 95% | 6/10 |

---

## Recommended Implementation Order

### Phase 1: Quick Wins (Week 1)

1. **Shell cd integration** (1 hour) - Highest impact, zero risk
2. **Bash/Fish completions** (2 hours) - Low effort, removes adoption barrier

**Total: 3 hours**

### Phase 2: Core Enhancements (Week 2)

3. **wt run command** (3 hours) - High value, straightforward implementation
2. **Rich status display** (6 hours) - Visible improvement, moderate complexity

**Total: 9 hours**

### Phase 3: Advanced Features (Week 3)

5. **Lifecycle hooks** (12 hours) - Power user feature, needs careful design

**Total: 12 hours**

---

## Reading the Proposals

Each proposal document contains:

- **Quick Reference** - Key metrics at a glance
- **The Problem** - Current pain points and user impact
- **The Solution** - Detailed implementation approach
- **Implementation Plan** - Phased breakdown with time estimates
- **Files to Modify** - Specific code locations
- **Example Usage** - Real-world scenarios
- **Why This Priority** - Justification with confidence rating
- **User Perception** - How users will experience the change
- **Alternatives Considered** - Why other approaches were rejected
- **Success Metrics** - How to measure impact

---

## Implementation Guidelines

### Before You Start

1. Read the proposal document completely
2. Review related proposals for dependencies
3. Check current codebase state (files may have changed)
4. Set up local development environment per AGENTS.md

### During Implementation

1. Follow the phased implementation plan
2. Write tests as you go (don't defer testing)
3. Update documentation alongside code changes
4. Run `make ci-check` before committing

### After Implementation

1. Test thoroughly across platforms (macOS, Linux)
2. Update proposal status to "Implemented"
3. Gather user feedback via GitHub issues
4. Consider follow-up enhancements

---

## Cross-References

### Dependencies Between Proposals

- **Proposal 1** (Shell cd) is independent - can implement first
- **Proposal 2** (wt run) is independent - can implement anytime
- **Proposal 3** (Rich status) is independent - can implement anytime
- **Proposal 4** (Hooks) could integrate with Proposal 2 (run command)
- **Proposal 5** (Completions) is independent - can implement first

### Complementary Features

- Proposals 1 & 5: Both improve shell integration experience
- Proposals 2 & 4: Run command could trigger hooks
- Proposals 3 & 4: Status could show hook execution status

---

## Success Criteria

### Overall Goals

Implementing all 5 proposals should result in:

1. **Workflow Transformation**
   - From "useful tool" to "indispensable workflow manager"
   - Eliminate daily friction points
   - Enable automation and CI/CD workflows

2. **User Experience**
   - Professional, polished feel
   - Modern, colorized output
   - Cross-platform support (shells, OSes)

3. **Adoption**
   - Remove barriers (shell support)
   - Enable enterprise use (hooks)
   - Reduce onboarding friction

4. **Community**
   - Position as extensible platform
   - Enable community contributions (hooks)
   - Generate enthusiasm and advocacy

---

## Feedback and Iteration

### How to Provide Feedback

1. **On a specific proposal:** Open GitHub issue with proposal number
2. **New ideas:** Create feature request using proposal template
3. **Implementation questions:** Comment on proposal PR

### Proposal Lifecycle

1. **Proposed** - Initial state, ready for review
2. **Approved** - Accepted for implementation
3. **In Progress** - Active development
4. **Implemented** - Merged to main branch
5. **Deployed** - Released in version X.Y.Z

---

## Additional Resources

- [Comprehensive Improvement Proposals Document](../improvement-proposals.md) - Full analysis with all 30 initial ideas
- [AGENTS.md](../../AGENTS.md) - Development guidelines for contributors
- [PRD](../internal/prd.md) - Product requirements document
- [Contributing Guide](../../CONTRIBUTING.md) - How to contribute

---

## Version History

- **v1.0** (2026-01-10) - Initial proposals based on codebase analysis
- Proposals are living documents - updates tracked via git history

---

**Last Updated:** 2026-01-10
**Status:** All proposals ready for implementation
**Next Steps:** Begin with Phase 1 Quick Wins
