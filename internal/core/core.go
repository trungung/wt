package core

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/trungung/wt/internal/config"
	"github.com/trungung/wt/internal/git"
)

// MapBranchToDir converts a branch name to a sanitized directory name.
// Sanitization rules:
// - Replace '/' with '-'
// - Fail if branch name contains illegal characters (anything other than alphanumeric, '-', '_', '.')
func MapBranchToDir(branch string) (string, error) {
	sanitized := strings.ReplaceAll(branch, "/", "-")
	valid := regexp.MustCompile(`^[a-zA-Z0-9\-_.]+$`)
	if !valid.MatchString(sanitized) {
		return "", fmt.Errorf("branch name %q contains illegal characters for worktree mapping", branch)
	}

	return sanitized, nil
}

// FindWorktree returns the path to an existing worktree for the given branch.
// It handles the default branch special case (returns repo root).
// It returns an error if no worktree exists for the branch.
func FindWorktree(branch string) (string, error) {
	root, err := git.GetRepoRoot()
	if err != nil {
		return "", err
	}

	cfg, err := config.LoadConfig(root)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	defaultBranch := cfg.DefaultBranch
	if defaultBranch == "" {
		defaultBranch, err = git.GetDefaultBranch()
		if err != nil {
			return "", err
		}
	}

	// Case: Default branch is the repo root
	if branch == defaultBranch {
		return root, nil
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		return "", err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return filepath.Clean(wt.Path), nil
		}
	}

	return "", fmt.Errorf("no worktree exists for branch %q", branch)
}

// RollbackError represents a failure that triggered a rollback attempt.
type RollbackError struct {
	OriginalErr    error
	RollbackErr    error
	RollbackStatus string
}

func (e *RollbackError) Error() string {
	return fmt.Sprintf("%v (rollback: %s)", e.OriginalErr, e.RollbackStatus)
}

// EnsureWorktree ensures a worktree exists for the given branch and returns its path
func EnsureWorktree(branch, base string) (string, error) {
	// 1. Try to find existing worktree first
	if path, err := FindWorktree(branch); err == nil {
		return path, nil
	}

	// 2. Not found, proceed with creation
	root, err := git.GetRepoRoot()
	if err != nil {
		return "", err
	}

	cfg, err := config.LoadConfig(root)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	defaultBranch := cfg.DefaultBranch
	if defaultBranch == "" {
		defaultBranch, _ = git.GetDefaultBranch()
	}

	// Case: Create new worktree
	dirName, err := MapBranchToDir(branch)
	if err != nil {
		return "", err
	}

	wtRoot := cfg.GetWorktreeBase(root)
	targetPath := filepath.Clean(filepath.Join(wtRoot, dirName))

	worktrees, err := git.ListWorktrees()
	if err != nil {
		return "", err
	}

	// Collision Policy: Fail if another branch maps to the same directory
	if err := checkCollisions(branch, dirName, worktrees); err != nil {
		return "", err
	}

	// Fail if directory exists but isn't registered as a worktree
	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("collision: directory %s already exists", targetPath)
	}

	if err := os.MkdirAll(wtRoot, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree root: %w", err)
	}

	// Concurrency Safety: Acquire lock before modification
	unlock, err := git.AcquireLock(root, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer unlock()

	// Resolve branch source if it doesn't exist locally
	local, remote := git.BranchExists(branch)
	isNewBranch := !local && !remote

	if isNewBranch {
		if base == "" {
			if defaultBranch == "" {
				return "", fmt.Errorf("branch %s not found and no default branch detected. Use --from", branch)
			}
			base = defaultBranch
		}
	} else {
		// Branch exists (locally or on origin).
		// We ignore 'base' because we are checking out an existing reference.
		base = ""
	}

	if err := git.CreateWorktree(targetPath, branch, base); err != nil {
		return "", err
	}

	// Success from here: attempt post-creation steps
	if err := applyPostCreation(root, targetPath, cfg, isNewBranch, branch); err != nil {
		// Rollback on failure
		status := "started"
		rbErr := git.RemoveWorktree(targetPath, true)
		if rbErr != nil {
			status = fmt.Sprintf("failed to remove worktree: %v", rbErr)
		} else {
			status = "worktree removed"
			if isNewBranch {
				rbErr = git.DeleteBranch(branch)
				if rbErr != nil {
					status += fmt.Sprintf(", failed to delete branch: %v", rbErr)
				} else {
					status += ", branch deleted"
				}
			}
		}

		if status == "worktree removed" || status == "worktree removed, branch deleted" {
			status = "succeeded (" + status + ")"
		}

		return "", &RollbackError{
			OriginalErr:    err,
			RollbackErr:    rbErr,
			RollbackStatus: status,
		}
	}

	return targetPath, nil
}

func checkCollisions(branch, dirName string, existing []git.Worktree) error {
	// Check existing worktrees
	for _, wt := range existing {
		// We already checked for exact branch match.
		// Now check if a different branch maps to the same dir name.
		existingDir := filepath.Base(wt.Path)
		if existingDir == dirName && wt.Branch != branch {
			return fmt.Errorf("collision: branch %q maps to same directory %q as existing worktree for branch %q",
				branch, dirName, wt.Branch)
		}
	}

	// Check all local branches to be thorough (as per PRD 2.3 and 4.7)
	branches, err := git.ListLocalBranches()
	if err != nil {
		return nil // skip if we can't list branches, not fatal here
	}
	for _, b := range branches {
		if b == branch {
			continue
		}
		d, err := MapBranchToDir(b)
		if err != nil {
			continue
		}
		if d == dirName {
			return fmt.Errorf("collision: branch %q maps to same directory %q as another local branch %q",
				branch, dirName, b)
		}
	}

	return nil
}

// RemoveWorktree handles the safety logic for removing a worktree
func RemoveWorktree(branch string, force bool, confirmFn func(string) bool) error {
	root, err := git.GetRepoRoot()
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig(root)
	if err != nil {
		return err
	}

	defaultBranch := cfg.DefaultBranch
	if defaultBranch == "" {
		defaultBranch, err = git.GetDefaultBranch()
		if err != nil {
			return err
		}
	}

	if branch == defaultBranch {
		return fmt.Errorf("refusing to remove default branch/main worktree")
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		return err
	}

	var targetWt *git.Worktree
	for _, wt := range worktrees {
		if wt.Branch == branch {
			targetWt = &wt
			break
		}
	}

	if targetWt == nil {
		return fmt.Errorf("no worktree found for branch %s", branch)
	}

	// Dirty check
	dirty, err := git.IsDirty(targetWt.Path)
	if err != nil {
		return err
	}

	if dirty && !force {
		if confirmFn == nil || !confirmFn(fmt.Sprintf("Worktree %s is dirty. Remove anyway?", branch)) {
			return fmt.Errorf("worktree is dirty; use --force or confirm")
		}
		force = true // If confirmed, we can use --force for the git command
	}

	// Concurrency Safety: Acquire lock before modification
	unlock, err := git.AcquireLock(root, 5*time.Second)
	if err != nil {
		return err
	}
	defer unlock()

	if err := git.RemoveWorktree(targetWt.Path, force); err != nil {
		return err
	}

	if cfg.DeleteBranchWithWorktree {
		mainBranch, _ := git.GetCurrentBranchInMainWorktree(root)
		if branch != mainBranch && branch != defaultBranch {
			_ = git.DeleteBranch(branch)
		}
	}

	return nil
}

type PruneOptions struct {
	DryRun bool
	Force  bool
	Fetch  bool
}

// PruneWorktrees removes worktrees whose branches are merged into the default branch
func PruneWorktrees(opts PruneOptions) (int, []string, error) {
	root, err := git.GetRepoRoot()
	if err != nil {
		return 0, nil, err
	}

	cfg, err := config.LoadConfig(root)
	if err != nil {
		return 0, nil, err
	}

	if opts.Fetch {
		_ = git.FetchPrune()
	}

	defaultBranch := cfg.DefaultBranch
	if defaultBranch == "" {
		defaultBranch, err = git.GetDefaultBranch()
		if err != nil {
			return 0, nil, err
		}
	}

	merged, err := git.GetMergedBranches(defaultBranch)
	if err != nil {
		return 0, nil, err
	}

	mergedSet := make(map[string]bool)
	for _, b := range merged {
		mergedSet[b] = true
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		return 0, nil, err
	}

	mainBranch, _ := git.GetCurrentBranchInMainWorktree(root)

	var candidates []string
	prunedCount := 0

	// Concurrency Safety: Acquire lock before modification (only if not dry run)
	if !opts.DryRun {
		unlock, err := git.AcquireLock(root, 5*time.Second)
		if err != nil {
			return 0, nil, err
		}
		defer unlock()
	}

	for i, wt := range worktrees {
		if i == 0 {
			continue // Skip main worktree
		}
		if wt.Branch == "(detached)" || wt.Branch == defaultBranch {
			continue
		}

		if mergedSet[wt.Branch] {
			dirty, err := git.IsDirty(wt.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to check dirty status for %s: %v\n", wt.Branch, err)
				continue
			}

			if dirty && !opts.Force {
				fmt.Fprintf(os.Stderr, "Skipping %s: worktree is dirty (use --force to prune)\n", wt.Branch)
				continue
			}

			if opts.DryRun {
				candidates = append(candidates, wt.Branch)
			} else {
				if err := git.RemoveWorktree(wt.Path, opts.Force); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to remove worktree for %s: %v\n", wt.Branch, err)
					continue
				}
				if cfg.DeleteBranchWithWorktree && wt.Branch != mainBranch {
					_ = git.DeleteBranch(wt.Branch)
				}
				prunedCount++
			}
		}
	}

	return prunedCount, candidates, nil
}

func applyPostCreation(repoRoot, targetPath string, cfg *config.Config, isNewBranch bool, branch string) error {
	// 1. Copy patterns
	for _, pattern := range cfg.WorktreeCopyPatterns {
		matches, err := filepath.Glob(filepath.Join(repoRoot, pattern))
		if err != nil {
			continue
		}
		for _, src := range matches {
			rel, err := filepath.Rel(repoRoot, src)
			if err != nil {
				continue
			}
			dst := filepath.Join(targetPath, rel)
			if err := copyIfMissing(src, dst); err != nil {
				return fmt.Errorf("failed to copy %s to %s: %w", src, dst, err)
			}
		}
	}

	// 2. PostCreateCmd
	for _, cmdStr := range cfg.PostCreateCmd {
		if cmdStr == "" {
			continue
		}
		parts := strings.Fields(cmdStr)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = targetPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("postCreateCmd '%s' failed: %w", cmdStr, err)
		}
	}

	return nil
}

func copyIfMissing(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil // already exists
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
