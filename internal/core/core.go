package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/trungung/wt/internal/git"
)

// MapBranchToDir converts a branch name like "feature/x" to "feature-x"
func MapBranchToDir(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

// EnsureWorktree ensures a worktree exists for the given branch and returns its path
func EnsureWorktree(branch, base string) (string, error) {
	root, err := git.GetRepoRoot()
	if err != nil {
		return "", err
	}

	defaultBranch, err := git.GetDefaultBranch()
	if err != nil {
		return "", err
	}

	// Case: Default branch prints repo root
	if branch == defaultBranch {
		return root, nil
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		return "", err
	}

	// Case: Worktree already exists
	for _, wt := range worktrees {
		if wt.Branch == branch {
			return filepath.Clean(wt.Path), nil
		}
	}

	// Case: Create new worktree
	dirName := MapBranchToDir(branch)
	// Template: $REPO_PATH.wt/$DIR_NAME
	wtRoot := root + ".wt"
	targetPath := filepath.Clean(filepath.Join(wtRoot, dirName))

	// Collision Policy: Fail if directory exists but isn't registered as a worktree
	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("collision: directory %s already exists", targetPath)
	}

	if err := os.MkdirAll(wtRoot, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree root: %w", err)
	}

	if err := git.CreateWorktree(targetPath, branch, base); err != nil {
		return "", err
	}

	return targetPath, nil
}
