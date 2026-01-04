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
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimPrefix(line, "branch ")
			// In porcelain mode, branch is full ref, e.g., refs/heads/main or refs/heads/feature/x
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
			worktrees = append(worktrees, current)
			current = Worktree{}
		}
	}

	return worktrees, nil
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
	out, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "/")
		return parts[len(parts)-1], nil
	}

	// Fallback to local check if remote isn't set
	out, err = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "main", nil
	}
	return strings.TrimSpace(string(out)), nil
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
