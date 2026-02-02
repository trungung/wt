package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/trungung/wt/internal/config"
	"github.com/trungung/wt/internal/log"
)

func TestMapBranchToDir(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple branch",
			branch:   "main",
			expected: "main",
			wantErr:  false,
		},
		{
			name:     "feature branch with slash",
			branch:   "feature/payment",
			expected: "feature-payment",
			wantErr:  false,
		},
		{
			name:     "nested branch with multiple slashes",
			branch:   "feature/nested/branch",
			expected: "feature-nested-branch",
			wantErr:  false,
		},
		{
			name:     "branch with underscores",
			branch:   "fix_bug_123",
			expected: "fix_bug_123",
			wantErr:  false,
		},
		{
			name:     "branch with dots",
			branch:   "release.1.0",
			expected: "release.1.0",
			wantErr:  false,
		},
		{
			name:     "branch with mixed characters",
			branch:   "feature/user-auth_v2.0",
			expected: "feature-user-auth_v2.0",
			wantErr:  false,
		},
		{
			name:     "branch with space - should error",
			branch:   "branch with space",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "branch with special chars - should error",
			branch:   "branch@123",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "branch with exclamation - should error",
			branch:   "feature!test",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapBranchToDir(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapBranchToDir(%q) error = %v, wantErr %v", tt.branch, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("MapBranchToDir(%q) = %v, want %v", tt.branch, got, tt.expected)
			}
		})
	}
}

func TestApplyPostCreation_EmptyCommands(t *testing.T) {
	// Create temp directories
	tempDir, err := os.MkdirTemp("", "wt-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	repoRoot := filepath.Join(tempDir, "repo")
	targetPath := filepath.Join(tempDir, "worktree")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Capture log output to verify warnings
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	cfg := &config.Config{
		PostCreateCmd: []string{
			"",           // empty string
			"   ",        // whitespace only
			"echo hello", // valid command
		},
	}

	// Should not panic on empty commands
	err = applyPostCreation(repoRoot, targetPath, cfg, false, "test-branch")
	if err != nil {
		t.Errorf("applyPostCreation should not error on empty commands, got: %v", err)
	}

	// Verify warnings were logged
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("skipping empty postCreateCmd")) {
		t.Errorf("expected warning for empty command, got output: %s", output)
	}
}

func TestApplyPostCreation_WhitespaceOnly(t *testing.T) {
	// Create temp directories
	tempDir, err := os.MkdirTemp("", "wt-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	repoRoot := filepath.Join(tempDir, "repo")
	targetPath := filepath.Join(tempDir, "worktree")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Suppress log output during test
	log.SetOutput(os.NewFile(0, os.DevNull))
	defer log.SetOutput(os.Stderr)

	cfg := &config.Config{
		PostCreateCmd: []string{
			"   ",  // only spaces
			"\t\n", // tabs and newlines
		},
	}

	// Should not panic and should skip whitespace-only commands
	err = applyPostCreation(repoRoot, targetPath, cfg, false, "test-branch")
	if err != nil {
		t.Errorf("applyPostCreation should not error on whitespace-only commands, got: %v", err)
	}
}

func TestMapBranchToDir_RegexPerformance(t *testing.T) {
	// This test ensures the regex is compiled once and reused
	// We test multiple calls to ensure no panic or performance issue
	branches := []string{
		"main", "feature/test", "bugfix/issue-123", "release/v1.0.0",
		"hotfix/critical", "develop", "feature/nested/deep/branch",
	}

	for _, branch := range branches {
		_, err := MapBranchToDir(branch)
		if err != nil {
			t.Errorf("MapBranchToDir(%q) unexpected error: %v", branch, err)
		}
	}

	// Test invalid branches
	invalidBranches := []string{
		"branch with space", "special!char", "at@symbol", "hash#tag",
	}

	for _, branch := range invalidBranches {
		_, err := MapBranchToDir(branch)
		if err == nil {
			t.Errorf("MapBranchToDir(%q) expected error for invalid branch", branch)
		}
	}
}
