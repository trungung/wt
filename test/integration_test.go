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

	// 2.1 Mock origin/HEAD for tests that rely on GetDefaultBranch auto-detection
	runGit(t, repoPath, "update-ref", "refs/remotes/origin/main", "HEAD")
	runGit(t, repoPath, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	// 2.2 Pre-create config
	configContent := `{
		"defaultBranch": "main",
		"worktreePathTemplate": "$REPO_PATH.wt",
		"worktreeCopyPatterns": [],
		"postCreateCmd": [],
		"deleteBranchWithWorktree": false
	}`
	if err := os.WriteFile(filepath.Join(repoPath, ".wt.config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Build the 'wt' binary
	binPath := filepath.Join(tempDir, "wt")
	testDir, _ := os.Getwd()
	projectRoot := filepath.Dir(testDir)
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/wt")
	buildCmd.Dir = projectRoot
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

	// Test 5: Config and PostCreateCmd
	t.Run("Config and PostCreateCmd", func(t *testing.T) {
		configContent := `{
			"defaultBranch": "main",
			"postCreateCmd": ["touch created.txt"],
			"worktreeCopyPatterns": ["README.md"]
		}`
		if err := os.WriteFile(filepath.Join(repoPath, ".wt.config.json"), []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		got := runWt("feature/y")
		if _, err := os.Stat(filepath.Join(got, "created.txt")); os.IsNotExist(err) {
			t.Errorf("postCreateCmd did not run: created.txt missing in %s", got)
		}
		if _, err := os.Stat(filepath.Join(got, "README.md")); os.IsNotExist(err) {
			t.Errorf("copy pattern did not run: README.md missing in %s", got)
		}
	})

	// Test 6: Exec command
	t.Run("Exec command", func(t *testing.T) {
		out := runWt("exec", "feature/y", "--", "ls")
		if !strings.Contains(out, "created.txt") {
			t.Errorf("exec did not run correctly: expected output to contain created.txt, got %s", out)
		}

		// Test: exec should fail if worktree doesn't exist (no on-the-fly creation)
		cmd := exec.Command(binPath, "exec", "non-existent-branch", "--", "ls")
		cmd.Dir = repoPath
		outBytes, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("exec should have failed for non-existent branch, but succeeded")
		}
		if !strings.Contains(string(outBytes), "no worktree exists") {
			t.Errorf("expected error message to contain 'no worktree exists', got: %s", string(outBytes))
		}
	})

	// Test 8: Remove worktree with branch deletion
	t.Run("Remove worktree with branch deletion", func(t *testing.T) {
		configContent := `{
			"defaultBranch": "main",
			"deleteBranchWithWorktree": true
		}`
		if err := os.WriteFile(filepath.Join(repoPath, ".wt.config.json"), []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		runWt("feature/z")
		runWt("remove", "feature/z")

		// Check if branch is deleted
		cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/feature/z")
		cmd.Dir = repoPath
		if err := cmd.Run(); err == nil {
			t.Errorf("branch feature/z should have been deleted")
		}
	})

	// Test 9: Collision detection and strict validation
	t.Run("Strict validation and collisions", func(t *testing.T) {
		// Test illegal characters (whitespace)
		cmd := exec.Command(binPath, "branch with space")
		cmd.Dir = repoPath
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("expected failure for branch with space, but succeeded")
		}
		if !strings.Contains(string(out), "illegal characters") {
			t.Errorf("expected error message to contain 'illegal characters', got: %s", string(out))
		}

		// Test collision: feature/a and feature-a both map to feature-a
		runGit(t, repoPath, "branch", "feature-a")
		runWt("feature-a")

		cmd = exec.Command(binPath, "feature/a")
		cmd.Dir = repoPath
		out, err = cmd.CombinedOutput()
		if err == nil {
			t.Errorf("expected failure for colliding branch feature/a, but succeeded")
		}
		if !strings.Contains(string(out), "collision") {
			t.Errorf("expected error message to contain 'collision', got: %s", string(out))
		}
	})

	// Test 10: Prune worktrees
	t.Run("Prune worktrees", func(t *testing.T) {
		// 1. Create a merged branch
		runGit(t, repoPath, "checkout", "-b", "merged-branch")
		if err := os.WriteFile(filepath.Join(repoPath, "merged.txt"), []byte("merged"), 0644); err != nil {
			t.Fatal(err)
		}
		runGit(t, repoPath, "add", "merged.txt")
		runGit(t, repoPath, "commit", "-m", "merge me")
		runGit(t, repoPath, "checkout", "main")
		runGit(t, repoPath, "merge", "merged-branch")

		// 2. Create an unmerged branch
		runGit(t, repoPath, "checkout", "-b", "unmerged-branch")
		if err := os.WriteFile(filepath.Join(repoPath, "unmerged.txt"), []byte("not merged"), 0644); err != nil {
			t.Fatal(err)
		}
		runGit(t, repoPath, "add", "unmerged.txt")
		runGit(t, repoPath, "commit", "-m", "don't merge me")
		runGit(t, repoPath, "checkout", "main")

		// 3. Ensure worktrees exist for both
		runWt("merged-branch")
		runWt("unmerged-branch")

		// 4. Test dry-run
		out := runWt("prune", "--dry-run")
		if !strings.Contains(out, "merged-branch") {
			t.Errorf("expected dry-run to contain 'merged-branch', got: %s", out)
		}
		if strings.Contains(out, "unmerged-branch") {
			t.Errorf("expected dry-run NOT to contain 'unmerged-branch', got: %s", out)
		}

		// 5. Test actual prune
		runWt("prune", "--force")

		// 6. Verify worktree removed
		cmd := exec.Command("git", "worktree", "list")
		cmd.Dir = repoPath
		wtListOut, _ := cmd.CombinedOutput()
		wtList := string(wtListOut)
		if strings.Contains(wtList, "[merged-branch]") {
			t.Errorf("expected merged-branch worktree to be removed, but it's still in list")
		}
		if !strings.Contains(wtList, "[unmerged-branch]") {
			t.Errorf("expected unmerged-branch worktree to remain, but it's gone")
		}

		// 7. Verify branch deletion (since deleteBranchWithWorktree is true from previous test)
		cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/merged-branch")
		cmd.Dir = repoPath
		if err := cmd.Run(); err == nil {
			t.Errorf("merged-branch should have been deleted")
		}
	})

	// Test 11: Init config
	t.Run("Init config", func(t *testing.T) {
		// Remove existing config if any
		os.Remove(filepath.Join(repoPath, ".wt.config.json"))

		runWt("init", "--yes")
		configPath := filepath.Join(repoPath, ".wt.config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("init --yes did not create config file")
		}

		// Test: init should not overwrite
		out := runWt("init")
		if !strings.Contains(out, ".wt.config.json") {
			t.Errorf("init should print config path if it exists, got: %s", out)
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
