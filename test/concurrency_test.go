package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConcurrency(t *testing.T) {
	// Setup temp environment
	tempDir, err := os.MkdirTemp("", "wt-concurrency-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	tempDir, _ = filepath.EvalSymlinks(tempDir)
	repoPath := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Init git repo
	exec.Command("git", "init", "-b", "main").Dir = repoPath
	runCmd(repoPath, "git", "init", "-b", "main")
	runCmd(repoPath, "git", "config", "user.email", "test@example.com")
	runCmd(repoPath, "git", "config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd(repoPath, "git", "add", ".")
	runCmd(repoPath, "git", "commit", "-m", "initial commit")

	// Build the 'wt' binary
	binPath := filepath.Join(tempDir, "wt")
	projectRoot, _ := os.Getwd()
	projectRoot = filepath.Dir(projectRoot)
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/wt")
	buildCmd.Dir = projectRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %s: %v", string(out), err)
	}

	// 1. Setup a slow postCreateCmd
	configContent := `{
		"defaultBranch": "main",
		"postCreateCmd": ["sleep 10"]
	}`
	if err := os.WriteFile(filepath.Join(repoPath, ".wt.config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Start first wt command in background
	cmd1 := exec.Command(binPath, "feature/slow")
	cmd1.Dir = repoPath
	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd1.Run()
	}()

	// Give it time to start and acquire the lock (increased from 500ms for reliability)
	time.Sleep(2 * time.Second)

	// 3. Attempt to run second wt command
	cmd2 := exec.Command(binPath, "feature/fast")
	cmd2.Dir = repoPath
	out2, err := cmd2.CombinedOutput()

	// 4. Verify second command failed with lock error
	if err == nil {
		t.Errorf("expected second wt command to fail due to lock, but it succeeded")
	}
	if !strings.Contains(string(out2), "another wt operation is in progress") {
		t.Errorf("expected lock error message, got: %s", string(out2))
	}

	// Wait for first command to finish
	if err := <-errChan; err != nil {
		t.Errorf("first wt command failed: %v", err)
	}
}

func runCmd(dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
