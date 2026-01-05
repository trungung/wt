package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
}

// ListWorktrees returns a list of existing worktrees
func ListWorktrees() ([]Worktree, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []Worktree
	var current Worktree

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				// Flush previous worktree if it didn't have a branch (detached)
				if current.Branch == "" {
					current.Branch = "(detached)"
				}
				worktrees = append(worktrees, current)
			}
			current = Worktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		}
	}
	// Final flush
	if current.Path != "" {
		if current.Branch == "" {
			current.Branch = "(detached)"
		}
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// FetchPrune runs git fetch --prune
func FetchPrune() error {
	var stderr bytes.Buffer
	cmd := exec.Command("git", "fetch", "--prune")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch --prune failed: %s: %w", stderr.String(), err)
	}
	return nil
}

// GetMergedBranches returns a list of branches merged into the given base
func GetMergedBranches(base string) ([]string, error) {
	out, err := exec.Command("git", "branch", "--merged", base, "--format=%(refname:short)").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var merged []string
	for _, line := range lines {
		if b := strings.TrimSpace(line); b != "" {
			merged = append(merged, b)
		}
	}
	return merged, nil
}

// GetRepoRoot returns the absolute path to the git repository root
func GetRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repo root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetDefaultBranch returns the default branch name (e.g., main or master)
func GetDefaultBranch() (string, error) {
	// Only check remote default branch via origin/HEAD
	out, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "/")
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("could not determine default branch via origin/HEAD")
}

// BranchExists checks if a branch exists locally or on origin
func BranchExists(branch string) (bool, bool) {
	local := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil
	remote := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch).Run() == nil
	return local, remote
}

// CreateWorktree creates a new worktree at the specified path for the given branch
func CreateWorktree(path, branch, base string) error {
	args := []string{"worktree", "add"}
	if base != "" {
		args = append(args, "-b", branch, path, base)
	} else {
		args = append(args, path, branch)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git worktree add failed: %s: %w", stderr.String(), err)
	}
	return nil
}

// RemoveWorktree removes a worktree at the specified path
func RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	var stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git worktree remove failed: %s: %w", stderr.String(), err)
	}
	return nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(branch string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git branch -D failed: %s: %w", stderr.String(), err)
	}
	return nil
}

// IsDirty returns true if the worktree at the given path has uncommitted changes
func IsDirty(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed in %s: %w", path, err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// GetCurrentBranchInMainWorktree returns the branch currently checked out in the main repo
func GetCurrentBranchInMainWorktree(root string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ListLocalBranches returns a list of all local branch names
func ListLocalBranches() ([]string, error) {
	out, err := exec.Command("git", "branch", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var branches []string
	for _, line := range lines {
		if b := strings.TrimSpace(line); b != "" {
			branches = append(branches, b)
		}
	}
	return branches, nil
}
