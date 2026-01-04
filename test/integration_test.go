package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	// 1. Setup temp environment
	tempDir, err := os.MkdirTemp("", "wt-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve symlinks (especially on macOS where /var -> /private/var)
	tempDir, err = filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	repoPath := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}

	// 2. Init git repo
	runGit(t, repoPath, "init", "-b", "main")
	runGit(t, repoPath, "config", "user.email", "test@example.com")
	runGit(t, repoPath, "config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoPath, "add", ".")
	runGit(t, repoPath, "commit", "-m", "initial commit")

	// 3. Build the 'wt' binary
	binPath := filepath.Join(tempDir, "wt")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/wt")
	buildCmd.Dir = "/Users/trungung/Code/trungung/wt" // Path to the project root
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %s: %v", string(out), err)
	}

	// Helper to run wt
	runWt := func(args ...string) string {
		cmd := exec.Command(binPath, args...)
		cmd.Dir = repoPath
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("wt %v failed: %s: %v", args, string(out), err)
		}
		return strings.TrimSpace(string(out))
	}

	// Test 1: Default branch behavior
	t.Run("Default branch", func(t *testing.T) {
		got := runWt("main")
		want, _ := filepath.Abs(repoPath)
		if got != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	})

	// Test 2: Ensure worktree (mapping feature/x -> feature-x)
	t.Run("Mapping and creation", func(t *testing.T) {
		// First create the branch to make it easier for v1
		runGit(t, repoPath, "branch", "feature/x")
		got := runWt("feature/x")

		wantSuffix := "repo.wt/feature-x"
		if !strings.HasSuffix(got, wantSuffix) {
			t.Errorf("expected path to end with %s, got %s", wantSuffix, got)
		}

		if _, err := os.Stat(got); os.IsNotExist(err) {
			t.Errorf("directory %s was not created", got)
		}
	})

	// Test 3: Idempotency
	t.Run("Idempotency", func(t *testing.T) {
		got1 := runWt("feature/x")
		got2 := runWt("feature/x")
		if got1 != got2 {
			t.Errorf("expected same path on second run, got %s and %s", got1, got2)
		}
	})

	// Test 4: List worktrees
	t.Run("List worktrees", func(t *testing.T) {
		got := runWt()
		if !strings.Contains(got, "main") {
			t.Errorf("expected list to contain 'main', got: %s", got)
		}
		if !strings.Contains(got, "feature/x") {
			t.Errorf("expected list to contain 'feature/x', got: %s", got)
		}
	})
}

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %s: %v", args, string(out), err)
	}
}
