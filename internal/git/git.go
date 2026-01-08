package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
}

func run(dir string, args ...string) ([]byte, error) {
	start := time.Now()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	debugLog(args, time.Since(start))
	return out, err
}

func runWithStderr(dir string, args ...string) ([]byte, []byte, error) {
	start := time.Now()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	debugLog(args, time.Since(start))
	return stdout.Bytes(), stderr.Bytes(), err
}

// parseLines splits output by newlines and filters empty lines
func parseLines(output []byte) []string {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, line := range lines {
		if b := strings.TrimSpace(line); b != "" {
			result = append(result, b)
		}
	}
	return result
}

// ListWorktrees returns a list of existing worktrees
func ListWorktrees() ([]Worktree, error) {
	out, err := run("", "worktree", "list", "--porcelain")
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
	_, stderr, err := runWithStderr("", "fetch", "--prune")
	if err != nil {
		return fmt.Errorf("git fetch --prune failed: %s: %w", string(stderr), err)
	}
	return nil
}

// GetMergedBranches returns a list of branches merged into the given base
func GetMergedBranches(base string) ([]string, error) {
	out, err := run("", "branch", "--merged", base, "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}
	return parseLines(out), nil
}

// GetRepoRoot returns the absolute path to the git repository root
func GetRepoRoot() (string, error) {
	out, err := run("", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get repo root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetDefaultBranch returns the default branch name (e.g., main or master)
func GetDefaultBranch() (string, error) {
	// Only check remote default branch via origin/HEAD
	out, err := run("", "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "/")
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("could not determine default branch via origin/HEAD")
}

// BranchExists checks if a branch exists locally or on origin
func BranchExists(branch string) (bool, bool) {
	_, errLocal := run("", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	_, errRemote := run("", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
	return errLocal == nil, errRemote == nil
}

// CreateWorktree creates a new worktree at the specified path for the given branch
func CreateWorktree(path, branch, base string) error {
	args := []string{"worktree", "add"}
	if base != "" {
		args = append(args, "-b", branch, path, base)
	} else {
		args = append(args, path, branch)
	}

	_, stderr, err := runWithStderr("", args...)
	if err != nil {
		return fmt.Errorf("git worktree add failed: %s: %w", string(stderr), err)
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

	_, stderr, err := runWithStderr("", args...)
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %s: %w", string(stderr), err)
	}
	return nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(branch string) error {
	_, stderr, err := runWithStderr("", "branch", "-D", branch)
	if err != nil {
		return fmt.Errorf("git branch -D failed: %s: %w", string(stderr), err)
	}
	return nil
}

// IsDirty returns true if the worktree at the given path has uncommitted changes
func IsDirty(path string) (bool, error) {
	out, err := run(path, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("git status failed in %s: %w", path, err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// GetCurrentBranchInMainWorktree returns the branch currently checked out in the main repo
func GetCurrentBranchInMainWorktree(root string) (string, error) {
	out, err := run(root, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ListLocalBranches returns a list of all local branch names
func ListLocalBranches() ([]string, error) {
	out, err := run("", "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	return parseLines(out), nil
}

// IsTracked returns true if the given path (relative to repo root) is tracked by git
func IsTracked(repoRoot, path string) bool {
	_, err := run(repoRoot, "ls-files", "--error-unmatch", path)
	return err == nil
}
